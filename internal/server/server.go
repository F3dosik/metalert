// Package server реализует HTTP-сервер сбора метрик.
//
// Сервер выбирает хранилище в зависимости от конфигурации:
//   - PostgreSQL ([repository.DBMetricsStorage]) — если задан DatabaseDSN
//   - файловое ([repository.FileMetricsStorage]) — если задан FileStoragePath
//   - in-memory ([repository.MemMetricsStorage]) — если ничего не задано
//
// При старте регистрирует наблюдателей аудита ([audit.AuditDispatcher]),
// настраивает маршруты и при необходимости запускает автосохранение метрик.
// Поддерживает graceful shutdown по сигналам SIGINT и SIGTERM.
package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"crypto/rsa"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/F3dosik/metalert/internal/audit"
	cfg "github.com/F3dosik/metalert/internal/config/server"
	"github.com/F3dosik/metalert/internal/crypto"
	grpcserver "github.com/F3dosik/metalert/internal/grpc"
	"github.com/F3dosik/metalert/internal/handler"
	"github.com/F3dosik/metalert/internal/middleware"
	"github.com/F3dosik/metalert/internal/middleware/gzip"
	pb "github.com/F3dosik/metalert/internal/proto"
	"github.com/F3dosik/metalert/internal/repository"
	"github.com/F3dosik/metalert/internal/service"
)

// app holds the domain layer: repository, service, and audit dispatcher.
// It is responsible for building and tearing down domain-level dependencies.
type app struct {
	storage    repository.MetricsStorage
	svc        service.MetricsService
	dispatcher *audit.AuditDispatcher
	logger     *zap.SugaredLogger
}

func newApp(c *cfg.ServerConfig, logger *zap.SugaredLogger) *app {
	storage := buildStorage(c, logger)
	dispatcher := buildDispatcher(c, logger)

	_, isSavable := storage.(repository.Savable)
	asyncSave := isSavable && c.StoreInterval == 0

	svc := service.NewMetricsService(storage, dispatcher, asyncSave, logger)

	return &app{
		storage:    storage,
		svc:        svc,
		dispatcher: dispatcher,
		logger:     logger,
	}
}

// close flushes and closes all domain resources.
func (a *app) close() {
	if closer, ok := a.storage.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			a.logger.Warnw("Ошибка при закрытии storage", "error", err)
		}
	}
	a.dispatcher.Close()
}

// autoSave periodically flushes metrics to disk while ctx is alive.
func (a *app) autoSave(ctx context.Context, interval int) {
	savable, ok := a.storage.(repository.Savable)
	if !ok {
		a.logger.Warn("Хранилище не поддерживает автосохранение")
		return
	}
	a.logger.Infow("Включено автосохранение метрик", "interval", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := savable.Save(); err != nil {
				a.logger.Warnw("Ошибка при автосохранении", "error", err)
			} else {
				a.logger.Debug("Метрики успешно сохранены")
			}
		case <-ctx.Done():
			return
		}
	}
}

// Server is the transport layer: HTTP + gRPC routing wired to the domain via app.
//
// Создаётся через [NewServer], запускается через [Server.Run].
type Server struct {
	config        *cfg.ServerConfig
	app           *app
	router        chi.Router
	listener      net.Listener
	logger        *zap.SugaredLogger
	trustedSubnet *net.IPNet
	privateKey    *rsa.PrivateKey
}

// NewServer создаёт и конфигурирует сервер на основе cfg.
func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) *Server {
	a := newApp(cfg, logger)

	privateKey, err := crypto.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		logger.Fatalw("failed to load private key", "error", err)
	}

	var ipNet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, ipNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			logger.Fatalw("failed to parse CIDR", "error", err)
		}
	}

	lis, err := net.Listen("tcp", cfg.AddrGRPC)
	if err != nil {
		logger.Fatalw("failed to announces on the local network address", "error", err)
	}

	s := &Server{
		config:        cfg,
		app:           a,
		router:        chi.NewRouter(),
		listener:      lis,
		logger:        logger,
		trustedSubnet: ipNet,
		privateKey:    privateKey,
	}
	s.routes()

	return s
}

// routes регистрирует все HTTP-маршруты сервера.
func (s *Server) routes() {
	s.router.Use(middleware.RequireTrustedSubnet(s.logger, s.trustedSubnet))
	s.router.Use(middleware.DecryptMiddleware(s.privateKey, s.logger))
	s.router.Use(gzip.WithCompression(s.logger))
	s.router.Use(middleware.WithLogging(s.logger))

	svc := s.app.svc

	s.router.Get("/", handler.MainHandler(svc))

	s.router.Route("/update/", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.UpdateJSONHandler(svc, s.logger))
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(svc, s.logger))
	})
	s.router.With(middleware.RequireJSON(s.logger)).Post("/updates/", handler.UpdatesJSONHandler(svc, s.logger))
	s.router.Route("/value", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.ValueJSONHandler(svc, s.logger))
		r.Get("/{metType}/{metName}", handler.ValueHandler(svc))
	})
	s.router.Get("/ping", handler.PingDB(svc, s.logger))
	s.router.Mount("/debug", chiMiddleware.Profiler())
}

// Run запускает HTTP-сервер и блокирует выполнение до получения сигнала завершения.
//
// Graceful shutdown по SIGINT / SIGTERM:
//  1. Закрывает domain-ресурсы (storage, dispatcher)
//  2. Даёт серверу 5 секунд на завершение активных запросов
func (s *Server) Run() {
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	s.logger.Infow("Запуск сервера, config:",
		"addr", s.config.Addr,
		"log_mode", s.config.LogMode,
		"store_interval", s.config.StoreInterval,
		"file_path", s.config.FileStoragePath,
		"restore", s.config.Restore,
		"DatabaseDSN", s.config.DatabaseDSN,
		"AuditFile", s.config.AuditFile,
		"AuditURL", s.config.AuditURL,
		"TrustedSubnet", s.config.TrustedSubnet,
	)

	if s.config.StoreInterval > 0 {
		go s.app.autoSave(ctx, s.config.StoreInterval)
	}

	httpSrv := &http.Server{
		Addr:           s.config.Addr,
		Handler:        s.router,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.SubnetInterceptor(s.trustedSubnet)),
	)
	pb.RegisterMetricsServer(grpcSrv, grpcserver.NewMetricsServer(s.app.storage))

	go grpcSrv.Serve(s.listener)

	go func() {
		<-ctx.Done()

		s.logger.Infow("Получен сигнал завершения, сохраняем метрики...")
		s.app.close()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			s.logger.Fatalw(err.Error(), "event", "Принудительное завершение сервера")
		}

		grpcSrv.GracefulStop()

		s.logger.Infow("Сервер завершен")
	}()

	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}

// AutoSave периодически сохраняет метрики в хранилище через интервал StoreInterval.
//
// Запускается из [Server.Run] в отдельной горутине, если StoreInterval > 0
// и хранилище реализует [repository.Savable].
func (s *Server) AutoSave(ctx context.Context) {
	s.app.autoSave(ctx, s.config.StoreInterval)
}

func buildStorage(c *cfg.ServerConfig, logger *zap.SugaredLogger) repository.MetricsStorage {
	if c.DatabaseDSN != "" {
		storage, err := repository.NewDBMetricStorage(c.DatabaseDSN)
		if err != nil {
			logger.Warnw("failed to create DBMetricStorage", "error", err)
		}
		logger.Infow("Storage selected", "type", "DB", "database", c.DatabaseDSN)
		return storage
	}
	if c.FileStoragePath != "" {
		storage, err := repository.NewFileMetricsStorage(c.FileStoragePath, c.Restore)
		if err != nil {
			logger.Warnw("failed to create FileMetricStorage", "error", err)
		}
		logger.Infow("Storage selected", "type", "File", "file_path", c.FileStoragePath)
		return storage
	}
	logger.Infow("Storage selected", "type", "Memory")
	return repository.NewMemMetricsStorage()
}

func buildDispatcher(c *cfg.ServerConfig, logger *zap.SugaredLogger) *audit.AuditDispatcher {
	dispatcher := audit.NewAuditDispatcher(logger)
	if c.AuditFile != "" {
		observer, err := audit.NewFileAuditObserver(c.AuditFile)
		if err != nil {
			logger.Fatalw("failed to create file observer", "error", err)
		}
		dispatcher.Register(observer)
	}
	if c.AuditURL != "" {
		dispatcher.Register(audit.NewURLAuditObserver(c.AuditURL))
	}
	return dispatcher
}

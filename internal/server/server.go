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
	"crypto/rsa"
	"io"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/F3dosik/metalert/internal/audit"
	cfg "github.com/F3dosik/metalert/internal/config/server"
	"github.com/F3dosik/metalert/internal/crypto"
	"github.com/F3dosik/metalert/internal/handler"
	"github.com/F3dosik/metalert/internal/middleware"
	"github.com/F3dosik/metalert/internal/middleware/gzip"
	"github.com/F3dosik/metalert/internal/repository"
)

// Server — HTTP-сервер сбора и хранения метрик.
//
// Агрегирует конфигурацию, хранилище, роутер, логгер и диспетчер аудита.
// Создаётся через [NewServer], запускается через [Server.Run].
type Server struct {
	config        *cfg.ServerConfig
	storage       repository.MetricsStorage
	router        chi.Router
	logger        *zap.SugaredLogger
	dispatcher    *audit.AuditDispatcher
	privateKey    *rsa.PrivateKey
	trustedSubnet *net.IPNet
}

// NewServer создаёт и конфигурирует сервер на основе cfg.
//
// Порядок выбора хранилища:
//  1. DatabaseDSN задан → [repository.NewDBMetricStorage]
//  2. FileStoragePath задан → [repository.NewFileMetricsStorage]
//  3. Иначе → [repository.NewMemMetricsStorage]
//
// Аудит-наблюдатели регистрируются по наличию полей AuditFile и AuditURL в конфиге.
// Маршруты настраиваются вызовом [Server.routes].
func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) *Server {
	var storage repository.MetricsStorage
	var err error

	if cfg.DatabaseDSN != "" {
		storage, err = repository.NewDBMetricStorage(cfg.DatabaseDSN)
		if err != nil {
			logger.Warnw("failed to create DBMetricStorage", "error", err)
		}
		logger.Infow("Storage selected", "type", "DB", "database", cfg.DatabaseDSN)
	} else if cfg.FileStoragePath != "" {
		logger.Infow("Database DSN not set — using file storage only")
		storage, err = repository.NewFileMetricsStorage(cfg.FileStoragePath, cfg.Restore)
		if err != nil {
			logger.Warnw("failed to create New FileMetricStorage", "error", err)
		}
		logger.Infow("Storage selected", "type", "File", "file_path", cfg.FileStoragePath)
	} else {
		logger.Infow("Database DSN and filesStoragePath not set — using memory storage only")
		storage = repository.NewMemMetricsStorage()
		logger.Infow("Storage selected", "type", "Memory")
	}

	dispatcher := audit.NewAuditDispatcher(logger)
	if cfg.AuditFile != "" {
		observer, err := audit.NewFileAuditObserver(cfg.AuditFile)
		if err != nil {
			logger.Fatalw("failed to create file observer", "error", err)
		}
		dispatcher.Register(observer)
	}
	if cfg.AuditURL != "" {
		dispatcher.Register(audit.NewURLAuditObserver(cfg.AuditURL))
	}

	privateKey, err := crypto.LoadPrivateKey(cfg.CryptoKey)
	if err != nil {
		logger.Fatalw("failed to load private key", "error", err)
	}

	var ipNet *net.IPNet
	if cfg.TrustedSubnet != "" {
		var err error
		_, ipNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			logger.Fatalw("failed to parse CIDR", "error", err)
		}
	}

	r := chi.NewRouter()

	server := &Server{
		config:        cfg,
		storage:       storage,
		router:        r,
		logger:        logger,
		dispatcher:    dispatcher,
		privateKey:    privateKey,
		trustedSubnet: ipNet,
	}
	server.routes()

	return server
}

// routes регистрирует все HTTP-маршруты сервера.
//
// Глобальные middleware: gzip-сжатие, структурированное логирование.
//
// Маршруты:
//   - GET  /                              — список всех метрик
//   - POST /update/{metType}/{metName}/{metValue} — обновление через URL
//   - POST /update/                       — обновление одной метрики (JSON)
//   - POST /updates/                      — пакетное обновление (JSON)
//   - POST /value/                        — получение метрики (JSON)
//   - GET  /value/{metType}/{metName}     — получение метрики через URL
//   - GET  /ping                          — проверка соединения с БД
//   - GET  /debug/*                       — pprof-профилировщик
//
// asyncSave включается, если хранилище реализует [repository.Savable]
// и StoreInterval == 0 (сохранение после каждого обновления).
func (s *Server) routes() {
	s.router.Use(middleware.RequireTrustedSubnet(s.logger, s.trustedSubnet))
	s.router.Use(middleware.DecryptMiddleware(s.privateKey, s.logger))
	s.router.Use(gzip.WithCompression(s.logger))
	s.router.Use(middleware.WithLogging(s.logger))

	s.router.Get("/", handler.MainHandler(s.storage))

	_, isSavable := s.storage.(repository.Savable)
	asyncSave := isSavable && s.config.StoreInterval == 0

	s.router.Route("/update/", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.UpdateJSONHandler(s.storage, s.dispatcher, s.logger, asyncSave))
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(s.storage, s.dispatcher, s.logger))
	})
	s.router.With(middleware.RequireJSON(s.logger)).Post("/updates/", handler.UpdatesJSONHandler(s.storage, s.dispatcher, s.logger, asyncSave))
	s.router.Route("/value", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.ValueJSONHandler(s.storage, s.logger))
		r.Get("/{metType}/{metName}", handler.ValueHandler(s.storage))
	})
	s.router.Get("/ping", handler.PingDB(s.storage, s.logger))
	s.router.Mount("/debug", chiMiddleware.Profiler())
}

// Run запускает HTTP-сервер и блокирует выполнение до получения сигнала завершения.
//
// Если StoreInterval > 0 и хранилище реализует [repository.Savable],
// запускает [Server.AutoSave] в отдельной горутине.
//
// Graceful shutdown по SIGINT / SIGTERM:
//  1. Сохраняет метрики (для [repository.FileMetricsStorage] — вызывает Close)
//  2. Даёт серверу 5 секунд на завершение активных запросов
//  3. Ожидает завершения всех горутин аудита через [audit.AuditDispatcher.Wait]
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
	)

	if _, ok := s.storage.(repository.Savable); s.config.StoreInterval > 0 && ok {
		go s.AutoSave(ctx)
	}

	srv := &http.Server{
		Addr:    s.config.Addr,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()

		s.logger.Infow("Получен сигнал завершения, сохраняем метрики...")

		if closer, ok := s.storage.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				s.logger.Warnw("Ошибка при закрытии storage", "error", err)
			}
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			s.logger.Fatalw(err.Error(), "event", "Принудительное завершение сервера")
		}

		s.logger.Infow("Сервер завершен")
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}

	s.dispatcher.Close()

}

// AutoSave периодически сохраняет метрики в хранилище через интервал StoreInterval.
//
// Запускается из [Server.Run] в отдельной горутине, если StoreInterval > 0
// и хранилище реализует [repository.Savable].
// При ошибке сохранения логирует предупреждение и продолжает работу.
func (s *Server) AutoSave(ctx context.Context) {
	savable, ok := s.storage.(repository.Savable)
	if !ok {
		s.logger.Warn("Хранилище не поддерживает автосохранение")
		return
	}
	s.logger.Infow("Включено автосохранение метрик", "interval", s.config.StoreInterval)

	ticker := time.NewTicker(time.Duration(s.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := savable.Save(); err != nil {
				s.logger.Warnw("Ошибка при автосохранении", "error", err)
			} else {
				s.logger.Debug("Метрики успешно сохранены")
			}
		case <-ctx.Done():
			return
		}
	}
}

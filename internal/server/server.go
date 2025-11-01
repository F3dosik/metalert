package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfg "github.com/F3dosik/metalert.git/internal/config/server"
	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/middleware"
	"github.com/F3dosik/metalert.git/internal/middleware/gzip"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	config  *cfg.ServerConfig
	storage repository.MetricsStorage
	router  chi.Router
	logger  *zap.SugaredLogger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) *Server {
	ctx, cancel := context.WithCancel(context.Background())

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

	r := chi.NewRouter()

	server := &Server{
		config:  cfg,
		storage: storage,
		router:  r,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
	server.routes()

	return server
}

func (s *Server) routes() {
	s.router.Use(gzip.WithCompression(s.logger))
	s.router.Use(middleware.WithLogging(s.logger))

	s.router.Get("/", handler.MainHandler(s.storage))
	_, isSavable := s.storage.(repository.Savable)
	asyncSave := isSavable && s.config.StoreInterval == 0
	s.router.Route("/update/", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.UpdateJSONHandler(s.storage, s.logger, asyncSave))
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(s.storage, s.logger))
	})
	s.router.With(middleware.RequireJSON(s.logger)).Post("/updates/", handler.UpdatesJSONHandler(s.storage, s.logger, asyncSave))
	s.router.Route("/value", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.ValueJSONHandler(s.storage, s.logger))
		r.Get("/{metType}/{metName}", handler.ValueHandler(s.storage))
	})
	s.router.Get("/ping", handler.PingDB(s.storage, s.logger))
}

func (s *Server) Run() {
	s.logger.Infow("Запуск сервера, config:",
		"addr", s.config.Addr,
		"log_mode", s.config.LogMode,
		"store_interval", s.config.StoreInterval,
		"file_path", s.config.FileStoragePath,
		"restore", s.config.Restore,
		"DatabaseDSN", s.config.DatabaseDSN,
	)

	if _, ok := s.storage.(repository.Savable); s.config.StoreInterval > 0 && ok {
		go s.AutoSave()
	}

	srv := &http.Server{
		Addr:    s.config.Addr,
		Handler: s.router,
	}

	// graceful shutdown
	go func() {

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		s.logger.Infow("Получен сигнал завершения, сохраняем метрики...")

		if fileStorage, ok := s.storage.(*repository.FileMetricsStorage); ok {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			if err := fileStorage.Close(ctx); err != nil {
				s.logger.Warnw("Ошибка при закрытии storage", "error", err)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			s.logger.Fatalw(err.Error(), "event", "Принудительное завершение сервера")
		}

		s.logger.Infow("Сервер завершен")
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}

func (s *Server) AutoSave() {
	savable, ok := s.storage.(repository.Savable)
	if !ok {
		s.logger.Warn("Хранилище не поддерживает автосохранение")
		return
	}
	s.logger.Infow("Включено автосохранение метрик", "interval", s.config.StoreInterval)

	ticker := time.NewTicker(time.Duration(s.config.StoreInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := savable.Save(); err != nil {
			s.logger.Warnw("Ошибка при автосохранении", "error", err)
		} else {
			s.logger.Debug("Метрики успешно сохранены")
		}
	}

}

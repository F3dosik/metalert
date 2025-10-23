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
	config    *cfg.ServerConfig
	storage   repository.MetricsStorage
	dbStorage *repository.DBMetricStorage
	router    chi.Router
	logger    *zap.SugaredLogger
}

func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) *Server {
	// var storage repository.MetricsStorage
	// var err error

	var DBstorage *repository.DBMetricStorage
	var err error

	if cfg.DatabaseDSN != "" {
		DBstorage, err = repository.NewDBMetricStorage(cfg.DatabaseDSN)
		if err != nil {
			logger.Warnw("failed to create DBMetricStorage", "error", err)
		}
	} else {
		logger.Infow("Database DSN not set — using memory storage only")
	}

	storage, err := repository.NewMemMetricsStorage(cfg.FileStoragePath, cfg.Restore)
	if err != nil {
		logger.Warnw("failed to create New MemMetricStorage", "error", err)
	}
	// if memStorage, ok := storage.(*repository.MemMetricsStorage); ok {
	// 	go func() {
	// 		for err := range memStorage.ErrCh {
	// 			logger.Warnw("failed to save metrics", "error", err)
	// 		}
	// 	}()
	// }

	go func() {
		for err := range storage.ErrCh {
			logger.Warnw("failed to save metrics", "error", err)
		}
	}()

	r := chi.NewRouter()

	server := &Server{
		config:    cfg,
		storage:   storage,
		dbStorage: DBstorage,
		router:    r,
		logger:    logger,
	}
	server.routes()

	return server
}

func (s *Server) routes() {
	s.router.Use(gzip.WithCompression(s.logger))
	s.router.Use(middleware.WithLogging(s.logger))

	s.router.Get("/", handler.MainHandler(s.storage))
	s.router.Route("/update", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.UpdateJSONHandler(s.storage, s.logger, s.config.StoreInterval == 0))
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(s.storage, s.logger))
	})
	s.router.Route("/value", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.ValueJSONHandler(s.storage, s.logger))
		r.Get("/{metType}/{metName}", handler.ValueHandler(s.storage))
	})
	s.router.Get("/ping", handler.PingDB(s.dbStorage, s.config.UseDB, s.logger))
}

func (s *Server) Run() {
	s.logger.Infow("Запуск сервера, config:",
		"addr", s.config.Addr,
		"log_mode", s.config.LogMode,
		"store_interval", s.config.StoreInterval,
		"file_path", s.config.FileStoragePath,
		"restore", s.config.Restore,
		"DatabaseDSN", s.config.DatabaseDSN,
		"UseDB", s.config.UseDB,
	)

	if s.config.StoreInterval > 0 {
		go s.AutoSave()
	}

	srv := &http.Server{Addr: s.config.Addr, Handler: s.router}

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop

		s.logger.Infow("Получен сигнал завершения, сохраняем метрики...")

		if memStorage, ok := s.storage.(*repository.MemMetricsStorage); ok {
			if err := memStorage.Close(); err != nil {
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

	s.logger.Infow("Starting server", "address", s.config.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}

func (s *Server) AutoSave() {
	s.logger.Infow("Включено автосохранение метрик", "interval", s.config.StoreInterval)

	savable, ok := s.storage.(repository.Savable)
	if !ok {
		s.logger.Warn("Хранилище не поддерживает автосохранение")
		return
	}

	ticker := time.NewTicker(s.config.StoreInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := savable.Save(); err != nil {
			s.logger.Warnw("Ошибка при автосохранении", "error", err)
		} else {
			s.logger.Debug("Метрики успешно сохранены")
		}
	}

}

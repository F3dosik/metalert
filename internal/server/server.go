package server

import (
	"net/http"
	"path/filepath"
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
	storage *repository.MemMetricsStorage
	router  chi.Router
	logger  *zap.SugaredLogger
}

func NewServer(cfg *cfg.ServerConfig, logger *zap.SugaredLogger) *Server {
	absPath, err := filepath.Abs(cfg.FileStoragePath)
	if err != nil {
		absPath = cfg.FileStoragePath
	}
	storage, err := repository.NewMemMetricsStorage(absPath, cfg.Restore)
	if err != nil {
		logger.Warnw("failed to restore metrics", "error", err)
	}
	go func() {
		for err := range storage.ErrCh {
			logger.Warnw("failed to save metrics", "error", err)
		}
	}()

	r := chi.NewRouter()

	server := &Server{
		config:  cfg,
		storage: storage,
		router:  r,
		logger:  logger,
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
}

func (s *Server) Run() {
	s.logger.Infow("Запуск сервера, config:",
		"addr", s.config.Addr,
		"log_mode", s.config.LogMode,
		"store_interval", s.config.StoreInterval,
		"file_path", s.config.FileStoragePath,
		"restore", s.config.Restore,
	)

	if s.config.StoreInterval > 0 {
		go s.AutoSave()
	}

	if err := http.ListenAndServe(s.config.Addr, s.router); err != nil {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}

func (s *Server) AutoSave() {
	s.logger.Infow("Включено автосохранение метрик", "interval", s.config.StoreInterval)

	ticker := time.NewTicker(s.config.StoreInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.storage.Save(); err != nil {
			s.logger.Warnw("Ошибка при автосохранении", "error", err)
		} else {
			s.logger.Debug("Метрики успешно сохранены")
		}
	}

}

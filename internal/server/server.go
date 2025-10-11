package server

import (
	"net/http"

	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/middleware"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	storage *repository.MemMetricsStorage
	router  chi.Router
	logger  *zap.SugaredLogger
}

func NewServer(logger *zap.SugaredLogger) *Server {
	storage := repository.NewMemMetricsStorage()
	r := chi.NewRouter()

	server := &Server{
		storage: storage,
		router:  r,
		logger:  logger,
	}
	server.routes()
	return server
}

func (s *Server) routes() {
	s.router.Use(middleware.WithLogging(s.logger))

	s.router.Get("/", handler.MainHandler(s.storage))
	s.router.Route("/update", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.UpdateJSONHandler(s.storage, s.logger))
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(s.storage))
	})
	s.router.Route("/value", func(r chi.Router) {
		r.With(middleware.RequireJSON(s.logger)).Post("/", handler.ValueJSONHandler(s.storage, s.logger))
		r.Get("/{metType}/{metName}", handler.ValueHandler(s.storage))
	})
}

func (s *Server) Run(addr string) {
	s.logger.Infow("Запуск сервера", "addr", addr)

	if err := http.ListenAndServe(addr, s.router); err != nil {
		s.logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}


package server

import (
	"net/http"

	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/middleware"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/go-chi/chi/v5"
)

func Run(addr string) {
	storage := repository.NewMemStorage()
	r := chi.NewRouter()
	baseLogger, logger := middleware.NewLogger()
	defer baseLogger.Sync()

	r.Use(middleware.WithLogging(logger))

	r.Get("/", handler.MainHandler(storage))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(storage))
		//r.Post{"/{metType}/{metName}/{metValue}/*", handler.BadRequestHandler}//Ессли не будет проходить с 404
	})
	r.Route("/value", func(r chi.Router) {
		r.Get("/{metType}/{metName}", handler.ValueHandler(storage))
	})

	logger.Infow(
		"Запуск сервера",
		"addr", addr,
	)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatalw(err.Error(), "event", "Запуск сервера")
	}
}

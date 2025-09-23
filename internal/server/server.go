package server

import (
	"log"
	"net/http"

	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/repository"
	"github.com/go-chi/chi/v5"
)

func Run(addr string) error {
	storage := repository.NewMemStorage()
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.MainHandler(storage))
		r.Route("/update", func(r chi.Router) {
			r.Post("/{metType}/{metName}/{metValue}", handler.UpdateHandler(storage))
			//r.Post{"/{metType}/{metName}/{metValue}/*", handler.BadRequestHandler}//Ессли не будет проходить с 404
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{metType}/{metName}", handler.ValueHandler(storage))
		})
	})

	log.Printf("Сервер запущен на %s", addr)
	return http.ListenAndServe(addr, r)
}

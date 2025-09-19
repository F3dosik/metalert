package server

import (
	"log"
	"net/http"

	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/repository"
)

func Run(addr string) error {
	storage := repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(storage))

	log.Printf("Сервер запущен на %s", addr)
	return http.ListenAndServe(addr, mux)
}

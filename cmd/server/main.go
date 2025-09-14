package main

import (
	"log"
	"net/http"

	"github.com/F3dosik/metalert.git/internal/handler"
	"github.com/F3dosik/metalert.git/internal/repository"
)

func main() {
	storage := repository.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler(storage))

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		log.Fatal(err)
	}
}

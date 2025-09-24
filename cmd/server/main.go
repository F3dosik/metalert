package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/config"
	"github.com/F3dosik/metalert.git/internal/server"
)

func main() {
	cfg := config.LoadServerConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid server configuration: %v", err)
	}

	log.Printf("Starting server on %s", cfg.Addr)
	if err := server.Run(cfg.Addr); err != nil {
		log.Fatal(err)
	}
}

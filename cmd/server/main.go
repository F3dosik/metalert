package main

import (
	"log"

	cfg "github.com/F3dosik/metalert.git/internal/config/server"
	"github.com/F3dosik/metalert.git/internal/server"
)

func main() {
	cfg, err := cfg.LoadServerConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}

	server.Run(cfg.Addr)
}

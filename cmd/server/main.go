package main

import (
	"log"

	cfg "github.com/F3dosik/metalert.git/internal/config/server"
	"github.com/F3dosik/metalert.git/internal/server"
	"github.com/F3dosik/metalert.git/pkg/logger"
)

func main() {
	cfg, err := cfg.LoadServerConfig()
	if err != nil {
		 log.Fatalf("Configuration loading error: %v", err)
	}

	mode := logger.Mode(cfg.LogMode)
	baseLogger, sugarLogger := logger.NewLogger(mode)
	defer func() { _ = baseLogger.Sync() }()

	server := server.NewServer(sugarLogger)
	server.Run(cfg.Addr)
}

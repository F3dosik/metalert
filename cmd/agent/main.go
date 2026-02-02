package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/F3dosik/metalert.git/internal/agent"
	cfg "github.com/F3dosik/metalert.git/internal/config/agent"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := cfg.LoadAgentConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}

	agent.Run(ctx, cfg.Endpoint, cfg.ReportInterval, cfg.PollInterval, cfg.RateLimit)
}

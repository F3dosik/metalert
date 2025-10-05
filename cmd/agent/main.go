package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/agent"
	"github.com/F3dosik/metalert.git/internal/config"
)

func main() {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}
	agent.Run(cfg.Endpoint, cfg.ReportInterval, cfg.PollInterval)
}

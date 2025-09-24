package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/agent"
	"github.com/F3dosik/metalert.git/internal/config"
)

func main() {
	cfg := config.LoadAgentConfig()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	agent.Run(cfg.Endpoint, cfg.ReportInterval, cfg.PollInterval)
}

package main

import (
	"log"

	"github.com/F3dosik/metalert.git/internal/agent"
	cfg "github.com/F3dosik/metalert.git/internal/config/agent"
)

func main() {
	cfg, err := cfg.LoadAgentConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}
	agent.Run(cfg.Endpoint, cfg.ReportInterval, cfg.PollInterval)
}

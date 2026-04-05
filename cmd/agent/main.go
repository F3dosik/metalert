package main

import (
	"cmp"
	"fmt"
	"log"

	"github.com/F3dosik/metalert/internal/agent"
	cfg "github.com/F3dosik/metalert/internal/config/agent"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		cmp.Or(buildVersion, "N/A"),
		cmp.Or(buildDate, "N/A"),
		cmp.Or(buildCommit, "N/A"))

	cfg, err := cfg.LoadAgentConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}
	agent.Run(cfg.Endpoint, cfg.ReportInterval, cfg.PollInterval, cfg.CryptoKey, cfg.GRPCEndpoint)
}

package main

import (
	"cmp"
	"fmt"
	"log"

	cfg "github.com/F3dosik/metalert/internal/config/server"
	"github.com/F3dosik/metalert/internal/server"
	"github.com/F3dosik/metalert/pkg/logger"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		cmp.Or(buildVersion, "N/A"),
		cmp.Or(buildDate, "N/A"),
		cmp.Or(buildCommit, "N/A"))

	cfg, err := cfg.LoadServerConfig()
	if err != nil {
		log.Fatalf("Configuration loading error: %v", err)
	}
	mode := logger.Mode(cfg.LogMode)
	baseLogger, sugarLogger := logger.NewLogger(mode)
	defer func() { _ = baseLogger.Sync() }()
	server := server.NewServer(cfg, sugarLogger)
	server.Run()
}

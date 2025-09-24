package config

import (
	"errors"
	"flag"
	"time"
)

type AgentConfig struct {
	Endpoint       string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func (c *AgentConfig) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint can't be empty")
	}

	if c.ReportInterval <= 0 {
		return errors.New("report interval must be positive")
	}

	if c.PollInterval <= 0 {
		return errors.New("poll interval must be postive")
	}

	return nil
}

func LoadAgentConfig() *AgentConfig {
	var endpoint string
	var reportInterval, pollInterval int

	flag.StringVar(&endpoint, "a", "http://localhost:8080", "HTTP server endpoint address")
	flag.IntVar(&reportInterval, "r", 10, "frequency of sending metrics to the server")
	flag.IntVar(&pollInterval, "p", 2, "frequency of polling metrics from runtime")
	flag.Parse()

	return &AgentConfig{
		Endpoint:       endpoint,
		ReportInterval: time.Duration(reportInterval) * time.Second,
		PollInterval:   time.Duration(pollInterval) * time.Second,
	}
}

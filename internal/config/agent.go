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

var (
	ErrEmptyEndpoint    = errors.New("endpoint can't be empty")
	ErrInvalidReportInt = errors.New("report interval must be positive")
	ErrInvalidPollInt   = errors.New("poll interval must be positive")
)

func (c *AgentConfig) Validate() error {
	if c.Endpoint == "" {
		return ErrEmptyEndpoint
	}

	if c.ReportInterval <= 0 {
		return ErrInvalidReportInt
	}

	if c.PollInterval <= 0 {
		return ErrInvalidPollInt
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

package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
}

var (
	ErrEmptyEndpoint    = errors.New("endpoint can't be empty")
	ErrInvalidReportInt = errors.New("report interval must be positive")
	ErrInvalidPollInt   = errors.New("poll interval must be positive")
)

const (
	defaultEndpoint       = "http://localhost:8080"
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
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

func (c *AgentConfig) String() string {
	return fmt.Sprintf(
		"Endpoint: %s, ReportInterval: %v, PollInterval: %v",
		c.Endpoint, c.ReportInterval, c.PollInterval,
	)
}

func LoadAgentConfig() (*AgentConfig, error) {
	envConfig := parseEnvConfig()
	// log.Println("envConfig:", envConfig)
	flagConfig := parseFlagConfig()
	// log.Println("flagConfig:", flagConfig)

	config := mergeConfigs(envConfig, flagConfig)
	// log.Println("FinalConfig", config)
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func parseEnvConfig() *AgentConfig {
	var config AgentConfig
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Warning: failed to parse env config: %v\n", err)
	}

	return &config
}

type flagConfig struct {
	Endpoint       string
	ReportInterval int
	PollInterval   int
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Endpoint, "a", defaultEndpoint, "HTTP server endpoint address")
	flag.IntVar(&config.ReportInterval, "r", int(defaultReportInterval.Seconds()), "frequency of sending metrics to the server (seconds)")
	flag.IntVar(&config.PollInterval, "p", int(defaultPollInterval.Seconds()), "frequency of polling metrics from runtime (seconds)")
	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *AgentConfig, flagConfig *flagConfig) *AgentConfig {
	config := &AgentConfig{}

	config.Endpoint = resolveEndpoint(envConfig.Endpoint, flagConfig.Endpoint)
	config.ReportInterval = resolveInterval(envConfig.ReportInterval, flagConfig.ReportInterval, defaultReportInterval)
	config.PollInterval = resolveInterval(envConfig.PollInterval, flagConfig.PollInterval, defaultPollInterval)

	return config
}

func resolveEndpoint(envEndpoint, flagEndpoint string) string {
	if envEndpoint != "" {
		return envEndpoint
	}

	if flagEndpoint != defaultEndpoint {
		return flagEndpoint
	}

	return defaultEndpoint
}

func resolveInterval(envInterval time.Duration, flagInterval int, defaultInterval time.Duration) time.Duration {
	if envInterval > 0 {
		return envInterval
	}

	if flagInterval != int(defaultInterval.Seconds()) {
		return time.Duration(flagInterval) * time.Second
	}

	return defaultInterval
}

package agent

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
}

var (
	errEmptyEndpoint    = errors.New("endpoint can't be empty")
	errInvalidReportInt = errors.New("report interval must be positive")
	errInvalidPollInt   = errors.New("poll interval must be positive")
)

const (
	defaultEndpoint       = "http://localhost:8080"
	defaultReportInterval = 10 * time.Second
	defaultPollInterval   = 2 * time.Second
)

func (c *AgentConfig) Validate() error {
	if c.Endpoint == "" {
		return errEmptyEndpoint
	}

	if c.ReportInterval <= 0 {
		return errInvalidReportInt
	}

	if c.PollInterval <= 0 {
		return errInvalidPollInt
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
	flagConfig := parseFlagConfig()

	config := mergeConfigs(envConfig, flagConfig)
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
	CryptoKey      string
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Endpoint, "a", defaultEndpoint, "HTTP server endpoint address")
	flag.IntVar(&config.ReportInterval, "r", int(defaultReportInterval.Seconds()), "frequency of sending metrics to the server (seconds)")
	flag.IntVar(&config.PollInterval, "p", int(defaultPollInterval.Seconds()), "frequency of polling metrics from runtime (seconds)")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "the full path to the file with the public key")
	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *AgentConfig, flagConfig *flagConfig) *AgentConfig {
	config := &AgentConfig{}

	config.Endpoint = resolveEndpoint(envConfig.Endpoint, flagConfig.Endpoint)
	config.ReportInterval = resolveInterval(envConfig.ReportInterval, flagConfig.ReportInterval, defaultReportInterval)
	config.PollInterval = resolveInterval(envConfig.PollInterval, flagConfig.PollInterval, defaultPollInterval)
	config.CryptoKey = resolveString(envConfig.CryptoKey, flagConfig.CryptoKey)

	return config
}

func resolveEndpoint(envEndpoint, flagEndpoint string) string {
	endpoint := ""

	if envEndpoint != "" {
		endpoint = envEndpoint
	} else if flagEndpoint != defaultEndpoint {
		endpoint = flagEndpoint
	} else {
		endpoint = defaultEndpoint
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint

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

func resolveString(envStr, flagStr string) string {
	if envStr != "" {
		return envStr
	}
	return flagStr
}

package agent

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Endpoint       string        `env:"ADDRESS"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	CryptoKey      string        `env:"CRYPTO_KEY"`
	JSONConfigPath string        `env:"CONFIG"`
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

	jsonConfigPath := envConfig.JSONConfigPath
	if jsonConfigPath == "" {
		jsonConfigPath = flagConfig.JSONConfigPath
	}

	var jsonCfg *jsonConfig
	if jsonConfigPath != "" {
		var err error
		jsonCfg, err = parseJSONConfig(jsonConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	}

	config := mergeConfigs(envConfig, flagConfig, jsonCfg)
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
	JSONConfigPath string
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Endpoint, "a", defaultEndpoint, "HTTP server endpoint address")
	flag.IntVar(&config.ReportInterval, "r", int(defaultReportInterval.Seconds()), "frequency of sending metrics to the server (seconds)")
	flag.IntVar(&config.PollInterval, "p", int(defaultPollInterval.Seconds()), "frequency of polling metrics from runtime (seconds)")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "the full path to the file with the public key")
	flag.StringVar(&config.JSONConfigPath, "config", "", "name of the configuration JSON file")
	flag.StringVar(&config.JSONConfigPath, "c", "", "name of the configuration JSON file (shorthand)")
	flag.Parse()

	return &config
}

type jsonConfig struct {
	Endpoint       string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

func parseJSONConfig(path string) (*jsonConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config jsonConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func mergeConfigs(envConfig *AgentConfig, flagConfig *flagConfig, jsonCfg *jsonConfig) *AgentConfig {
	config := &AgentConfig{}

	var jsonEndpoint, jsonCryptoKey string
	var jsonReportInterval, jsonPollInterval time.Duration

	if jsonCfg != nil {
		jsonEndpoint = jsonCfg.Endpoint
		jsonCryptoKey = jsonCfg.CryptoKey
		if d, err := time.ParseDuration(jsonCfg.ReportInterval); err == nil {
			jsonReportInterval = d
		}
		if d, err := time.ParseDuration(jsonCfg.PollInterval); err == nil {
			jsonPollInterval = d
		}
	}

	config.Endpoint = resolveEndpoint(envConfig.Endpoint, flagConfig.Endpoint, jsonEndpoint)
	config.ReportInterval = resolveInterval(envConfig.ReportInterval, flagConfig.ReportInterval, jsonReportInterval, defaultReportInterval)
	config.PollInterval = resolveInterval(envConfig.PollInterval, flagConfig.PollInterval, jsonPollInterval, defaultPollInterval)
	config.CryptoKey = resolveString(envConfig.CryptoKey, flagConfig.CryptoKey, jsonCryptoKey)

	return config
}

func resolveEndpoint(envEndpoint, flagEndpoint, jsonEndpoint string) string {
	endpoint := ""

	if envEndpoint != "" {
		endpoint = envEndpoint
	} else if flagEndpoint != defaultEndpoint {
		endpoint = flagEndpoint
	} else if jsonEndpoint != "" {
		endpoint = jsonEndpoint
	} else {
		endpoint = defaultEndpoint
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint
}

func resolveInterval(envInterval time.Duration, flagInterval int, jsonInterval time.Duration, defaultInterval time.Duration) time.Duration {
	if envInterval > 0 {
		return envInterval
	}

	if flagInterval != int(defaultInterval.Seconds()) {
		return time.Duration(flagInterval) * time.Second
	}

	if jsonInterval > 0 {
		return jsonInterval
	}

	return defaultInterval
}

func resolveString(envStr, flagStr, jsonStr string) string {
	if envStr != "" {
		return envStr
	}
	if flagStr != "" {
		return flagStr
	}
	return jsonStr
}

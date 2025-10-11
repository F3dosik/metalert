package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/F3dosik/metalert.git/pkg/logger"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr    string `env:"ADDRESS"`
	LogMode string `env:"LOG_MODE"`
}

var (
	ErrEmptyAddr = errors.New("address can't be empty")
	ErrEmptyPort = errors.New("server address must contain port")
)

const (
	defaultAddr    = "localhost:8080"
	defaultLogMode = string(logger.ModeDevelopment)
)

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return ErrEmptyAddr
	}

	if !strings.Contains(c.Addr, ":") {
		return ErrEmptyPort
	}

	switch c.LogMode {
	case string(logger.ModeDevelopment), string(logger.ModeProduction):
	default:
		return fmt.Errorf("invalid log mode: %s, allowed: development, production", c.LogMode)
	}

	return nil
}

func LoadServerConfig() (*ServerConfig, error) {
	envConfig := parseEnvConfig()
	flagConfig := parseFlagConfig()

	config := mergeConfigs(envConfig, flagConfig)
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil

}

func parseEnvConfig() *ServerConfig {
	var config ServerConfig
	err := env.Parse(&config)
	if err != nil {
		log.Printf("Warning: failed to parse env config: %v\n", err)
	}
	return &config
}

type flagConfig struct {
	Address string
	LogMode string
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Address, "a", defaultAddr, "address and port to run server")
	flag.StringVar(&config.LogMode, "log-mode", defaultLogMode, "logger mode: development (Debug + Colors), production (Info)")
	flag.Parse()
	return &config
}

func mergeConfigs(envConfig *ServerConfig, flagConfig *flagConfig) *ServerConfig {
	config := &ServerConfig{}

	config.Addr = resolveAddress(envConfig.Addr, flagConfig.Address)
	config.LogMode = resolveLogMode(envConfig.LogMode, flagConfig.LogMode)

	return config
}

func resolveAddress(envAddr, flagAddr string) string {
	if envAddr != "" {
		return envAddr
	}

	if flagAddr != defaultAddr {
		return flagAddr
	}

	return defaultAddr
}

func resolveLogMode(envMode, flagMode string) string {
	if envMode != "" {

		return envMode
	}
	if flagMode != "" {
		return flagMode
	}

	return defaultLogMode
}

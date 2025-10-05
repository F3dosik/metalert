package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr string
}

var (
	ErrEmptyAddr = errors.New("address can't be empty")
	ErrEmptyPort = errors.New("server address must contain port")
)

const (
	defaultAddr = "localhost:8080"
)

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return ErrEmptyAddr
	}

	if !strings.Contains(c.Addr, ":") {
		return ErrEmptyPort
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
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Address, "a", defaultAddr, "address and port to run server")
	flag.Parse()
	return &config
}

func mergeConfigs(envConfig *ServerConfig, flagConfig *flagConfig) *ServerConfig {
	config := &ServerConfig{}

	config.Addr = resolveAddress(envConfig.Addr, flagConfig.Address)

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

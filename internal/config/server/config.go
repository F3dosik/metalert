package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/F3dosik/metalert.git/pkg/logger"
	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Addr            string        `env:"ADDRESS"`
	LogMode         string        `env:"LOG_MODE"`
	StoreInterval   time.Duration `env:"STORE_INTERVAL"`
	FileStoragePath string        `env:"FILE_STORAGE_PATH"`
	Restore         bool          `env:"RESTORE"`
}

var (
	errEmptyAddr = errors.New("address can't be empty")
	errEmptyPort = errors.New("server address must contain port")
)

const (
	defaultAddr            = "localhost:8080"
	defaultLogMode         = string(logger.ModeDevelopment)
	defaultStoreInterval   = 300 * time.Second
	defaultFileStoragePath = "var/metrics.json"
	defaultRestore         = false
)

func (c *ServerConfig) Validate() error {
	if c.Addr == "" {
		return errEmptyAddr
	}

	if !strings.Contains(c.Addr, ":") {
		return errEmptyPort
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
	Address         string
	LogMode         string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Address, "a", "", "address and port to run server")
	flag.StringVar(&config.LogMode, "log-mode", "", "logger mode: development (Debug + Colors), production (Info)")
	flag.IntVar(&config.StoreInterval, "i", 0, "store interval (seconds)")
	flag.StringVar(&config.FileStoragePath, "", "", "file storage path")
	flag.BoolVar(&config.Restore, "r", false, "restore metrics from file")
	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *ServerConfig, flagConfig *flagConfig) *ServerConfig {
	config := &ServerConfig{}

	config.Addr = resolveString(envConfig.Addr, flagConfig.Address, defaultAddr)
	config.LogMode = resolveString(envConfig.LogMode, flagConfig.LogMode, defaultLogMode)
	config.StoreInterval = resolveDuration(envConfig.StoreInterval, flagConfig.StoreInterval, defaultStoreInterval)
	config.FileStoragePath = resolveString(envConfig.FileStoragePath, flagConfig.FileStoragePath, defaultFileStoragePath)
	config.Restore = resolveBool(envConfig.Restore, flagConfig.Restore, defaultRestore)

	return config
}

func resolveString(envVal, flagVal, def string) string {
	if envVal != "" {
		return envVal
	}
	if flagVal != "" {
		return flagVal
	}
	return def
}

func resolveDuration(envVal time.Duration, flagVal int, def time.Duration) time.Duration {
	if envVal >= 0 {
		return envVal
	}
	if flagVal >= 0 {
		return time.Duration(flagVal) * time.Second
	}
	return def
}

func resolveBool(envVal, flagVal, def bool) bool {
	if _, ok := os.LookupEnv("RESTORE"); ok {
		return envVal
	}
	return flagVal || def
}

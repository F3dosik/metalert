package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"

	"github.com/F3dosik/metalert/pkg/logger"
)

// generate:reset
type ServerConfig struct {
	Addr          string `env:"ADDRESS"`
	LogMode       string `env:"LOG_MODE"`
	StoreInterval int    `env:"STORE_INTERVAL"`

	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`

	DatabaseDSN string `env:"DATABASE_DSN"`

	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`
}

var (
	errEmptyAddr = errors.New("address can't be empty")
	errEmptyPort = errors.New("server address must contain port")
)

const (
	defaultAddr            = "localhost:8080"
	defaultLogMode         = string(logger.ModeDevelopment)
	defaultStoreInterval   = 300
	defaultFileStoragePath = ""
	defaultRestore         = false
	defaultDSN             = ""
	defaultAuditFile       = ""
	defaultAuditURL        = ""
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
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
}

func parseFlagConfig() *flagConfig {
	var config flagConfig

	flag.StringVar(&config.Address, "a", "", "address and port to run server")
	flag.StringVar(&config.LogMode, "log-mode", "", "logger mode: development (Debug + Colors), production (Info)")
	flag.IntVar(&config.StoreInterval, "i", -1, "store interval (seconds)")
	flag.StringVar(&config.FileStoragePath, "f", "", "file storage path")
	flag.BoolVar(&config.Restore, "r", false, "restore metrics from file")
	flag.StringVar(&config.DatabaseDSN, "d", "", "PostgreSQL DSN")
	flag.StringVar(&config.AuditFile, "audit-file", "", "the path to the file where the audit logs are saved. If the parameter is not passed, the audit should be disabled")
	flag.StringVar(&config.AuditURL, "audit-url", "", "the full URL where the audit logs are sent. If the parameter is not passed, the audit should be disabled")
	flag.Parse()

	return &config
}

func mergeConfigs(envConfig *ServerConfig, flagConfig *flagConfig) *ServerConfig {
	config := &ServerConfig{}

	config.Addr = resolveString(envConfig.Addr, flagConfig.Address, defaultAddr)
	config.LogMode = resolveString(envConfig.LogMode, flagConfig.LogMode, defaultLogMode)
	config.StoreInterval = resolveDuration("STORE_INTERVAL", envConfig.StoreInterval, flagConfig.StoreInterval, defaultStoreInterval)
	config.FileStoragePath = resolveString(envConfig.FileStoragePath, flagConfig.FileStoragePath, defaultFileStoragePath)
	config.Restore = resolveBool("RESTORE", envConfig.Restore, flagConfig.Restore, defaultRestore)
	config.DatabaseDSN = resolveString(envConfig.DatabaseDSN, flagConfig.DatabaseDSN, defaultDSN)
	config.AuditFile = resolveString(envConfig.AuditFile, flagConfig.AuditFile, defaultAuditFile)
	config.AuditURL = resolveString(envConfig.AuditURL, flagConfig.AuditURL, defaultAuditURL)

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

func resolveDuration(envName string, envVal, flagVal, def int) int {
	if _, ok := os.LookupEnv(envName); ok {
		return envVal
	}
	if flagVal >= 0 {
		return flagVal
	}
	return def
}

func resolveBool(envName string, envVal, flagVal, def bool) bool {
	if _, ok := os.LookupEnv(envName); ok {
		return envVal
	}
	if flagVal {
		return true
	}
	return def
}

package server

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

	"github.com/F3dosik/metalert/pkg/logger"
)

type ServerConfig struct {
	Addr          string `env:"ADDRESS"`
	LogMode       string `env:"LOG_MODE"`
	StoreInterval int    `env:"STORE_INTERVAL"`

	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`

	DatabaseDSN string `env:"DATABASE_DSN"`

	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`

	CryptoKey      string `env:"CRYPTO_KEY"`
	JSONConfigPath string `env:"CONFIG"`

	TrustedSubnet string `env:"TRUSTED_SUBNET"`
	AddrGRPC      string `env:"GRPC_ADDRESS"`
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
	defaultAddrGRPC        = "localhost:3200"
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
	CryptoKey       string
	JSONConfigPath  string
	TrustedSubnet   string
	AddrGRPC        string
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
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "the full path to the file with the public key")
	flag.StringVar(&config.JSONConfigPath, "config", "", "name of the configuration JSON file")
	flag.StringVar(&config.TrustedSubnet, "t", "", "string representation of classless addressing (CIDR)")
	flag.StringVar(&config.AddrGRPC, "grpc-addr", "", "gRPC server listen address")

	flag.Parse()

	return &config
}

type jsonConfig struct {
	Address         string `json:"address"`
	Restore         *bool  `json:"restore"`
	StoreInterval   string `json:"store_interval"`
	FileStoragePath string `json:"store_file"`
	DatabaseDSN     string `json:"database_dsn"`
	CryptoKey       string `json:"crypto_key"`
	TrustedSubnet   string `json:"trusted_subnet"`
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

func mergeConfigs(envConfig *ServerConfig, flagConfig *flagConfig, jsonCfg *jsonConfig) *ServerConfig {
	config := &ServerConfig{}

	var jsonAddr, jsonFile, jsonDSN, jsonCrypto, jsonTrustedSubnet string
	var jsonRestore *bool
	jsonInterval := -1

	if jsonCfg != nil {
		jsonAddr = jsonCfg.Address
		jsonRestore = jsonCfg.Restore
		jsonFile = jsonCfg.FileStoragePath
		jsonDSN = jsonCfg.DatabaseDSN
		jsonCrypto = jsonCfg.CryptoKey
		jsonTrustedSubnet = jsonCfg.TrustedSubnet

		if jsonCfg.StoreInterval != "" {
			if d, err := time.ParseDuration(jsonCfg.StoreInterval); err == nil {
				jsonInterval = int(d.Seconds())
			}
		}
	}

	config.Addr = resolveString(envConfig.Addr, flagConfig.Address, jsonAddr, defaultAddr)
	config.LogMode = resolveString(envConfig.LogMode, flagConfig.LogMode, "", defaultLogMode)
	config.StoreInterval = resolveDuration("STORE_INTERVAL", envConfig.StoreInterval, flagConfig.StoreInterval, jsonInterval, defaultStoreInterval)
	config.FileStoragePath = resolveString(envConfig.FileStoragePath, flagConfig.FileStoragePath, jsonFile, defaultFileStoragePath)
	config.Restore = resolveBool("RESTORE", envConfig.Restore, flagConfig.Restore, jsonRestore, defaultRestore)
	config.DatabaseDSN = resolveString(envConfig.DatabaseDSN, flagConfig.DatabaseDSN, jsonDSN, defaultDSN)
	config.AuditFile = resolveString(envConfig.AuditFile, flagConfig.AuditFile, "", defaultAuditFile)
	config.AuditURL = resolveString(envConfig.AuditURL, flagConfig.AuditURL, "", defaultAuditURL)
	config.CryptoKey = resolveString(envConfig.CryptoKey, flagConfig.CryptoKey, jsonCrypto, "")
	config.TrustedSubnet = resolveString(envConfig.TrustedSubnet, flagConfig.TrustedSubnet, jsonTrustedSubnet, "")
	config.AddrGRPC = resolveString(envConfig.AddrGRPC, flagConfig.AddrGRPC, "", defaultAddrGRPC)

	return config
}

func resolveString(envVal, flagVal, jsonVal, def string) string {
	if envVal != "" {
		return envVal
	}
	if flagVal != "" {
		return flagVal
	}
	if jsonVal != "" {
		return jsonVal
	}
	return def
}

func resolveDuration(envName string, envVal, flagVal, jsonVal, def int) int {
	if _, ok := os.LookupEnv(envName); ok {
		return envVal
	}
	if flagVal >= 0 {
		return flagVal
	}
	if jsonVal >= 0 {
		return jsonVal
	}
	return def
}

func resolveBool(envName string, envVal, flagVal bool, jsonVal *bool, def bool) bool {
	if _, ok := os.LookupEnv(envName); ok {
		return envVal
	}
	if flagVal {
		return true
	}
	if jsonVal != nil {
		return *jsonVal
	}
	return def
}

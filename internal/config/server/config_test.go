package server

import (
	"os"
	"testing"
)

// ── resolveString ────────────────────────────────────────────────────────────

func TestResolveString_EnvFirst(t *testing.T) {
	if got := resolveString("env", "flag", "json", "def"); got != "env" {
		t.Errorf("got %q, want env", got)
	}
}

func TestResolveString_FlagSecond(t *testing.T) {
	if got := resolveString("", "flag", "json", "def"); got != "flag" {
		t.Errorf("got %q, want flag", got)
	}
}

func TestResolveString_JSONThird(t *testing.T) {
	if got := resolveString("", "", "json", "def"); got != "json" {
		t.Errorf("got %q, want json", got)
	}
}

func TestResolveString_Default(t *testing.T) {
	if got := resolveString("", "", "", "def"); got != "def" {
		t.Errorf("got %q, want default", got)
	}
}

// ── resolveDuration ──────────────────────────────────────────────────────────

func TestResolveDuration_EnvSet(t *testing.T) {
	os.Setenv("STORE_INTERVAL", "42")
	defer os.Unsetenv("STORE_INTERVAL")

	got := resolveDuration("STORE_INTERVAL", 42, -1, -1, 300)
	if got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestResolveDuration_FlagOverrides(t *testing.T) {
	os.Unsetenv("STORE_INTERVAL")
	got := resolveDuration("STORE_INTERVAL", 0, 60, -1, 300)
	if got != 60 {
		t.Errorf("got %d, want 60", got)
	}
}

func TestResolveDuration_JSONFallback(t *testing.T) {
	os.Unsetenv("STORE_INTERVAL")
	got := resolveDuration("STORE_INTERVAL", 0, -1, 120, 300)
	if got != 120 {
		t.Errorf("got %d, want 120", got)
	}
}

func TestResolveDuration_Default(t *testing.T) {
	os.Unsetenv("STORE_INTERVAL")
	got := resolveDuration("STORE_INTERVAL", 0, -1, -1, 300)
	if got != 300 {
		t.Errorf("got %d, want 300", got)
	}
}

// ── resolveBool ──────────────────────────────────────────────────────────────

func TestResolveBool_EnvSet(t *testing.T) {
	os.Setenv("RESTORE", "true")
	defer os.Unsetenv("RESTORE")

	got := resolveBool("RESTORE", true, false, nil, false)
	if !got {
		t.Error("want true from env")
	}
}

func TestResolveBool_FlagTrue(t *testing.T) {
	os.Unsetenv("RESTORE")
	b := true
	got := resolveBool("RESTORE", false, true, nil, false)
	_ = b
	if !got {
		t.Error("want true from flag")
	}
}

func TestResolveBool_JSONFallback(t *testing.T) {
	os.Unsetenv("RESTORE")
	b := true
	got := resolveBool("RESTORE", false, false, &b, false)
	if !got {
		t.Error("want true from json")
	}
}

func TestResolveBool_Default(t *testing.T) {
	os.Unsetenv("RESTORE")
	got := resolveBool("RESTORE", false, false, nil, true)
	if !got {
		t.Error("want default true")
	}
}

// ── mergeConfigs ─────────────────────────────────────────────────────────────

func TestMergeConfigs_EnvDominates(t *testing.T) {
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("RESTORE")

	env := &ServerConfig{
		Addr:    "env:8080",
		LogMode: "production",
	}
	flag := &flagConfig{
		Address: "flag:9090",
		LogMode: "development",
	}
	cfg := mergeConfigs(env, flag, nil)

	if cfg.Addr != "env:8080" {
		t.Errorf("Addr = %q, want env", cfg.Addr)
	}
	if cfg.LogMode != "production" {
		t.Errorf("LogMode = %q, want production", cfg.LogMode)
	}
}

func TestMergeConfigs_JSONFallback(t *testing.T) {
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("RESTORE")

	env := &ServerConfig{}
	flag := &flagConfig{StoreInterval: -1}
	restore := true
	json := &jsonConfig{
		Address:         "json:7070",
		Restore:         &restore,
		StoreInterval:   "1m",
		FileStoragePath: "/tmp/metrics.json",
		DatabaseDSN:     "postgres://localhost/db",
		CryptoKey:       "key.pem",
		TrustedSubnet:   "10.0.0.0/8",
	}
	cfg := mergeConfigs(env, flag, json)

	if cfg.Addr != "json:7070" {
		t.Errorf("Addr = %q, want json", cfg.Addr)
	}
	if cfg.Restore != true {
		t.Error("Restore should be true from json")
	}
	if cfg.StoreInterval != 60 {
		t.Errorf("StoreInterval = %d, want 60", cfg.StoreInterval)
	}
	if cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Errorf("TrustedSubnet = %q", cfg.TrustedSubnet)
	}
}

// ── parseJSONConfig ──────────────────────────────────────────────────────────

func TestParseJSONConfig_Valid(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	f.WriteString(`{"address":"localhost:8080","store_interval":"5m","restore":true}`)
	f.Close()

	cfg, err := parseJSONConfig(f.Name())
	if err != nil {
		t.Fatalf("parseJSONConfig: %v", err)
	}
	if cfg.Address != "localhost:8080" {
		t.Errorf("address = %q", cfg.Address)
	}
}

func TestParseJSONConfig_NotFound(t *testing.T) {
	_, err := parseJSONConfig("/no/such/file.json")
	if err == nil {
		t.Error("expected error")
	}
}

func TestParseJSONConfig_InvalidJSON(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	f.WriteString("{bad json}")
	f.Close()

	_, err := parseJSONConfig(f.Name())
	if err == nil {
		t.Error("expected error for bad json")
	}
}

// ── parseEnvConfig ────────────────────────────────────────────────────────────

func TestParseEnvConfig_WithVars(t *testing.T) {
	os.Setenv("ADDRESS", "env-srv:8080")
	os.Setenv("LOG_MODE", "production")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("LOG_MODE")
	}()

	cfg := parseEnvConfig()
	if cfg.Addr != "env-srv:8080" {
		t.Errorf("Addr = %q, want env value", cfg.Addr)
	}
	if cfg.LogMode != "production" {
		t.Errorf("LogMode = %q, want production", cfg.LogMode)
	}
}

func TestParseEnvConfig_Empty(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("LOG_MODE")
	cfg := parseEnvConfig()
	_ = cfg
}

// ── Validate ─────────────────────────────────────────────────────────────────

func TestServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ServerConfig
		wantErr bool
	}{
		{"valid", ServerConfig{Addr: "localhost:8080", LogMode: "development"}, false},
		{"empty addr", ServerConfig{LogMode: "development"}, true},
		{"no port", ServerConfig{Addr: "localhost", LogMode: "development"}, true},
		{"invalid logmode", ServerConfig{Addr: "localhost:8080", LogMode: "verbose"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

package agent

import (
	"os"
	"testing"
	"time"
)

// ── resolveEndpoint ──────────────────────────────────────────────────────────

func TestResolveEndpoint_EnvTakesPriority(t *testing.T) {
	got := resolveEndpoint("http://env:8080", "http://flag:8080", "http://json:8080")
	if got != "http://env:8080" {
		t.Errorf("got %q, want env value", got)
	}
}

func TestResolveEndpoint_FlagOverJSON(t *testing.T) {
	got := resolveEndpoint("", "http://flag:9090", "http://json:8080")
	if got != "http://flag:9090" {
		t.Errorf("got %q, want flag value", got)
	}
}

func TestResolveEndpoint_JSONFallback(t *testing.T) {
	got := resolveEndpoint("", defaultEndpoint, "http://json:7070")
	if got != "http://json:7070" {
		t.Errorf("got %q, want json value", got)
	}
}

func TestResolveEndpoint_Default(t *testing.T) {
	got := resolveEndpoint("", defaultEndpoint, "")
	if got != defaultEndpoint {
		t.Errorf("got %q, want default", got)
	}
}

func TestResolveEndpoint_AddsScheme(t *testing.T) {
	got := resolveEndpoint("localhost:8080", "", "")
	if got != "http://localhost:8080" {
		t.Errorf("got %q, want http scheme added", got)
	}
}

// ── resolveInterval ──────────────────────────────────────────────────────────

func TestResolveInterval_EnvTakesPriority(t *testing.T) {
	got := resolveInterval(30*time.Second, 5, 20*time.Second, defaultReportInterval)
	if got != 30*time.Second {
		t.Errorf("got %v, want env value", got)
	}
}

func TestResolveInterval_FlagOverrides(t *testing.T) {
	got := resolveInterval(0, 15, 20*time.Second, defaultReportInterval)
	if got != 15*time.Second {
		t.Errorf("got %v, want flag value", got)
	}
}

func TestResolveInterval_JSONFallback(t *testing.T) {
	got := resolveInterval(0, int(defaultReportInterval.Seconds()), 25*time.Second, defaultReportInterval)
	if got != 25*time.Second {
		t.Errorf("got %v, want json value", got)
	}
}

func TestResolveInterval_Default(t *testing.T) {
	got := resolveInterval(0, int(defaultReportInterval.Seconds()), 0, defaultReportInterval)
	if got != defaultReportInterval {
		t.Errorf("got %v, want default", got)
	}
}

// ── resolveString ────────────────────────────────────────────────────────────

func TestResolveString_EnvFirst(t *testing.T) {
	if got := resolveString("env", "flag", "json"); got != "env" {
		t.Errorf("got %q, want env", got)
	}
}

func TestResolveString_FlagSecond(t *testing.T) {
	if got := resolveString("", "flag", "json"); got != "flag" {
		t.Errorf("got %q, want flag", got)
	}
}

func TestResolveString_JSONThird(t *testing.T) {
	if got := resolveString("", "", "json"); got != "json" {
		t.Errorf("got %q, want json", got)
	}
}

func TestResolveString_Empty(t *testing.T) {
	if got := resolveString("", "", ""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

// ── mergeConfigs ─────────────────────────────────────────────────────────────

func TestMergeConfigs_EnvDominates(t *testing.T) {
	env := &AgentConfig{
		Endpoint:       "http://env:8080",
		ReportInterval: 5 * time.Second,
		PollInterval:   1 * time.Second,
		CryptoKey:      "env-key",
	}
	flag := &flagConfig{
		Endpoint:       "http://flag:9090",
		ReportInterval: 20,
		PollInterval:   3,
	}
	cfg := mergeConfigs(env, flag, nil)

	if cfg.Endpoint != "http://env:8080" {
		t.Errorf("Endpoint = %q, want env", cfg.Endpoint)
	}
	if cfg.ReportInterval != 5*time.Second {
		t.Errorf("ReportInterval = %v, want 5s", cfg.ReportInterval)
	}
	if cfg.CryptoKey != "env-key" {
		t.Errorf("CryptoKey = %q, want env-key", cfg.CryptoKey)
	}
}

func TestMergeConfigs_JSONFallback(t *testing.T) {
	env := &AgentConfig{}
	flag := &flagConfig{
		Endpoint:       defaultEndpoint,
		ReportInterval: int(defaultReportInterval.Seconds()),
		PollInterval:   int(defaultPollInterval.Seconds()),
	}
	json := &jsonConfig{
		Endpoint:       "http://json:1234",
		ReportInterval: "30s",
		PollInterval:   "3s",
		CryptoKey:      "json-key",
	}
	cfg := mergeConfigs(env, flag, json)

	if cfg.Endpoint != "http://json:1234" {
		t.Errorf("Endpoint = %q, want json", cfg.Endpoint)
	}
	if cfg.CryptoKey != "json-key" {
		t.Errorf("CryptoKey = %q, want json-key", cfg.CryptoKey)
	}
}

// ── parseJSONConfig ──────────────────────────────────────────────────────────

func TestParseJSONConfig_Valid(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	f.WriteString(`{"address":"http://localhost:8080","report_interval":"10s","poll_interval":"2s","crypto_key":"key"}`)
	f.Close()

	cfg, err := parseJSONConfig(f.Name())
	if err != nil {
		t.Fatalf("parseJSONConfig: %v", err)
	}
	if cfg.Endpoint != "http://localhost:8080" {
		t.Errorf("Endpoint = %q", cfg.Endpoint)
	}
	if cfg.ReportInterval != "10s" {
		t.Errorf("ReportInterval = %q", cfg.ReportInterval)
	}
}

func TestParseJSONConfig_NotFound(t *testing.T) {
	_, err := parseJSONConfig("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestParseJSONConfig_InvalidJSON(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "*.json")
	f.WriteString("not json")
	f.Close()

	_, err := parseJSONConfig(f.Name())
	if err == nil {
		t.Error("expected error for invalid json")
	}
}

// ── parseEnvConfig ────────────────────────────────────────────────────────────

func TestParseEnvConfig_WithVars(t *testing.T) {
	os.Setenv("ADDRESS", "http://env-agent:9090")
	os.Setenv("CRYPTO_KEY", "env-key.pem")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("CRYPTO_KEY")
	}()

	cfg := parseEnvConfig()
	if cfg.Endpoint != "http://env-agent:9090" {
		t.Errorf("Endpoint = %q, want env value", cfg.Endpoint)
	}
	if cfg.CryptoKey != "env-key.pem" {
		t.Errorf("CryptoKey = %q, want env-key.pem", cfg.CryptoKey)
	}
}

func TestParseEnvConfig_Empty(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("CRYPTO_KEY")
	cfg := parseEnvConfig()
	_ = cfg // просто не должно паниковать
}

// ── String ────────────────────────────────────────────────────────────────────

func TestAgentConfig_String(t *testing.T) {
	cfg := &AgentConfig{
		Endpoint:       "http://localhost:8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
	}
	s := cfg.String()
	if s == "" {
		t.Error("String() returned empty string")
	}
}

// ── Validate ─────────────────────────────────────────────────────────────────

func TestAgentConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AgentConfig
		wantErr bool
	}{
		{"valid", AgentConfig{Endpoint: "http://localhost:8080", ReportInterval: 10 * time.Second, PollInterval: 2 * time.Second}, false},
		{"empty endpoint", AgentConfig{ReportInterval: 10 * time.Second, PollInterval: 2 * time.Second}, true},
		{"zero report interval", AgentConfig{Endpoint: "http://x", PollInterval: 2 * time.Second}, true},
		{"zero poll interval", AgentConfig{Endpoint: "http://x", ReportInterval: 5 * time.Second}, true},
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

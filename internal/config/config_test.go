package config_test

import (
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/config"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func validEnv() map[string]string {
	return map[string]string{
		"DB_USER":               "agrolang",
		"DB_PASSWORD":           "agrolang",
		"DB_NAME":               "agrolang",
		"DB_HOST":               "localhost",
		"DB_PORT":               "5432",
		"OPEN_METEO_BASE_URL":   "https://api.open-meteo.com/v1",
		"EVOLUTION_API_URL":     "https://evolution.example.com",
		"EVOLUTION_API_KEY":     "secret-key",
		"HTTP_PORT":             "8080",
		"SCAN_INTERVAL_SECONDS": "300",
	}
}

func TestLoad_AllVarsPresent_PopulatesConfig(t *testing.T) {
	setEnv(t, validEnv())

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DBDSN() != "postgres://agrolang:agrolang@localhost:5432/agrolang" {
		t.Fatalf("unexpected DSN: %s", cfg.DBDSN())
	}
	if cfg.OpenMeteoBaseURL != "https://api.open-meteo.com/v1" {
		t.Fatalf("unexpected OpenMeteoBaseURL: %s", cfg.OpenMeteoBaseURL)
	}
	if cfg.EvolutionAPIURL != "https://evolution.example.com" {
		t.Fatalf("unexpected EvolutionAPIURL: %s", cfg.EvolutionAPIURL)
	}
	if cfg.EvolutionAPIKey != "secret-key" {
		t.Fatalf("unexpected EvolutionAPIKey: %s", cfg.EvolutionAPIKey)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("unexpected HTTPPort: %s", cfg.HTTPPort)
	}
	if cfg.ScanInterval.Seconds() != 300 {
		t.Fatalf("unexpected ScanInterval: %v", cfg.ScanInterval)
	}
}

func TestLoad_MissingRequiredVar_ReturnsError(t *testing.T) {
	env := validEnv()
	delete(env, "DB_HOST")
	setEnv(t, env)
	t.Setenv("DB_HOST", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DB_HOST, got nil")
	}
}

func TestLoad_MissingHTTPPort_DefaultsTo8080(t *testing.T) {
	env := validEnv()
	delete(env, "HTTP_PORT")
	setEnv(t, env)
	t.Setenv("HTTP_PORT", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("expected default HTTPPort 8080, got %s", cfg.HTTPPort)
	}
}

func TestLoad_InvalidScanInterval_ReturnsError(t *testing.T) {
	env := validEnv()
	env["SCAN_INTERVAL_SECONDS"] = "not-a-number"
	setEnv(t, env)

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid SCAN_INTERVAL_SECONDS, got nil")
	}
}

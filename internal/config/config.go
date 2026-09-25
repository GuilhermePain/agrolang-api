package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string

	OpenMeteoBaseURL string
	EvolutionAPIURL  string
	EvolutionAPIKey  string

	HTTPPort     string
	ScanInterval time.Duration
}

func (c *Config) DBDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func Load() (*Config, error) {
	cfg := &Config{
		DBUser:           os.Getenv("DB_USER"),
		DBPassword:       os.Getenv("DB_PASSWORD"),
		DBName:           os.Getenv("DB_NAME"),
		DBHost:           os.Getenv("DB_HOST"),
		DBPort:           os.Getenv("DB_PORT"),
		OpenMeteoBaseURL: os.Getenv("OPEN_METEO_BASE_URL"),
		EvolutionAPIURL:  os.Getenv("EVOLUTION_API_URL"),
		EvolutionAPIKey:  os.Getenv("EVOLUTION_API_KEY"),
		HTTPPort:         os.Getenv("HTTP_PORT"),
	}

	required := map[string]string{
		"DB_USER":             cfg.DBUser,
		"DB_PASSWORD":         cfg.DBPassword,
		"DB_NAME":             cfg.DBName,
		"DB_HOST":             cfg.DBHost,
		"DB_PORT":             cfg.DBPort,
		"OPEN_METEO_BASE_URL": cfg.OpenMeteoBaseURL,
	}
	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("config: missing required env var %s", name)
		}
	}

	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}

	intervalSeconds := os.Getenv("SCAN_INTERVAL_SECONDS")
	if intervalSeconds == "" {
		cfg.ScanInterval = 5 * time.Minute
	} else {
		seconds, err := strconv.Atoi(intervalSeconds)
		if err != nil {
			return nil, fmt.Errorf("config: invalid SCAN_INTERVAL_SECONDS: %w", err)
		}
		cfg.ScanInterval = time.Duration(seconds) * time.Second
	}

	return cfg, nil
}

package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

const (
	defaultHTTPPort      = "8080"
	defaultDatabaseURL   = "postgres://onboarding:onboarding@localhost:5434/vendor_onboarding?sslmode=disable"
	defaultLogLevel      = "INFO"
	defaultStuckAfterDay = "7"
)

type Config struct {
	HTTPPort       string
	DatabaseURL    string
	LogLevel       slog.Level
	StuckAfterDays int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:    envOrDefault("HTTP_PORT", defaultHTTPPort),
		DatabaseURL: envOrDefault("DATABASE_URL", defaultDatabaseURL),
	}

	port, err := strconv.Atoi(cfg.HTTPPort)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("HTTP_PORT must be a number between 1 and 65535")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must not be empty")
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(envOrDefault("LOG_LEVEL", defaultLogLevel))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL must be DEBUG, INFO, WARN, or ERROR: %w", err)
	}

	cfg.StuckAfterDays, err = strconv.Atoi(envOrDefault("STUCK_AFTER_DAYS", defaultStuckAfterDay))
	if err != nil || cfg.StuckAfterDays < 1 {
		return Config{}, fmt.Errorf("STUCK_AFTER_DAYS must be a positive integer")
	}

	return cfg, nil
}

func envOrDefault(name, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}

// Package config loads and validates application configuration from environment variables.
package config

import (
	"errors"
	"os"
)

// Config holds all runtime configuration for the API server.
type Config struct {
	DatabaseURL   string
	Port          string
	JWTSecret     string
	AppleTeamID   string
	AppleClientID string
	APNSKeyID     string
	APNSKeyPath   string
	Env           string
}

// Load reads configuration from environment variables and validates required fields.
// Returns an error if any required field is missing.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          os.Getenv("PORT"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		AppleTeamID:   os.Getenv("APPLE_TEAM_ID"),
		AppleClientID: os.Getenv("APPLE_CLIENT_ID"),
		APNSKeyID:     os.Getenv("APNS_KEY_ID"),
		APNSKeyPath:   os.Getenv("APNS_KEY_PATH"),
		Env:           os.Getenv("ENV"),
	}

	if cfg.Port == "" {
		cfg.Port = "9090"
	}

	if cfg.Env == "" {
		cfg.Env = "development"
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("config: DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return nil, errors.New("config: JWT_SECRET is required")
	}

	return cfg, nil
}

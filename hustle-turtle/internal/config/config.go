package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string
	Port        string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/hustle_turtle?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

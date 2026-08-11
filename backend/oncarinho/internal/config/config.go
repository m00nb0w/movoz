package config

import "os"

type Config struct {
	DatabaseURL   string
	Port          string
	AdminPassword string
	SessionSecret string
	CookieSecure  bool
}

func Load() *Config {
	return &Config{
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://localhost/oncarinho?sslmode=disable"),
		Port:          getEnv("PORT", "8081"),
		AdminPassword: getEnv("ADMIN_PASSWORD", ""),
		SessionSecret: getEnv("SESSION_SECRET", ""),
		CookieSecure:  getEnv("COOKIE_SECURE", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

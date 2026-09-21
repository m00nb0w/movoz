package config

import (
	"bufio"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL      string
	Port             string
	JiraEmail        string
	JiraAPIToken     string
	JiraBaseURL      string
	JiraSyncInterval time.Duration
}

func Load() *Config {
	loadDotEnv(".env")

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://localhost/el_storko?sslmode=disable"),
		Port:             getEnv("PORT", "8082"),
		JiraEmail:        getEnv("JIRA_EMAIL", ""),
		JiraAPIToken:     getEnv("JIRA_API_TOKEN", ""),
		JiraBaseURL:      getEnv("JIRA_BASE_URL", ""),
		JiraSyncInterval: getDuration("JIRA_SYNC_INTERVAL", 5*time.Minute),
	}
}

// loadDotEnv sets process environment variables from a git-ignored .env file
// without overriding any variable already set in the real environment, so
// launchd EnvironmentVariables and .env stay consistent regardless of which
// one is present.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

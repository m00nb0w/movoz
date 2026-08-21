package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL   string
	Port          string
	AdminPassword string
	SessionSecret string
	CookieSecure  bool

	AnthropicAPIKey string

	GitHubToken string
	GitHubRepos []string

	JiraBaseURL  string
	JiraEmail    string
	JiraAPIToken string
	JiraProjects []string

	SyncInterval time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://localhost/scout?sslmode=disable"),
		Port:            getEnv("PORT", "8082"),
		AdminPassword:   getEnv("ADMIN_PASSWORD", ""),
		SessionSecret:   getEnv("SESSION_SECRET", ""),
		CookieSecure:    getEnv("COOKIE_SECURE", "false") == "true",
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		GitHubToken:     getEnv("SCOUT_GITHUB_TOKEN", ""),
		GitHubRepos:     splitCSV(getEnv("SCOUT_GITHUB_REPOS", "")),
		JiraBaseURL:     getEnv("SCOUT_JIRA_BASE_URL", ""),
		JiraEmail:       getEnv("SCOUT_JIRA_EMAIL", ""),
		JiraAPIToken:    getEnv("SCOUT_JIRA_API_TOKEN", ""),
		JiraProjects:    splitCSV(getEnv("SCOUT_JIRA_PROJECTS", "")),
		SyncInterval:    getEnvHours("SYNC_INTERVAL_HOURS", 12),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvHours(key string, fallbackHours int) time.Duration {
	if value := os.Getenv(key); value != "" {
		if hours, err := strconv.Atoi(value); err == nil {
			return time.Duration(hours) * time.Hour
		}
	}
	return time.Duration(fallbackHours) * time.Hour
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

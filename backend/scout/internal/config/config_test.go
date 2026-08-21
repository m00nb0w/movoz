package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("SYNC_INTERVAL_HOURS")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://localhost/scout?sslmode=disable" {
		t.Fatalf("unexpected default DatabaseURL: %s", cfg.DatabaseURL)
	}
	if cfg.Port != "8082" {
		t.Fatalf("unexpected default Port: %s", cfg.Port)
	}
	if cfg.SyncInterval != 12*time.Hour {
		t.Fatalf("unexpected default SyncInterval: %s", cfg.SyncInterval)
	}
}

func TestLoadGitHubReposCSV(t *testing.T) {
	os.Setenv("SCOUT_GITHUB_REPOS", "org/repo-a, org/repo-b")
	defer os.Unsetenv("SCOUT_GITHUB_REPOS")

	cfg := Load()

	if len(cfg.GitHubRepos) != 2 || cfg.GitHubRepos[0] != "org/repo-a" || cfg.GitHubRepos[1] != "org/repo-b" {
		t.Fatalf("expected trimmed CSV repos, got %+v", cfg.GitHubRepos)
	}
}

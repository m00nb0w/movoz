package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("ADMIN_PASSWORD")
	os.Unsetenv("SESSION_SECRET")
	os.Unsetenv("COOKIE_SECURE")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://localhost/oncarinho?sslmode=disable" {
		t.Errorf("expected default DatabaseURL, got %q", cfg.DatabaseURL)
	}
	if cfg.Port != "8081" {
		t.Errorf("expected default Port 8081, got %q", cfg.Port)
	}
	if cfg.CookieSecure != false {
		t.Errorf("expected default CookieSecure false, got %v", cfg.CookieSecure)
	}
}

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://example/test")
	os.Setenv("PORT", "9090")
	os.Setenv("ADMIN_PASSWORD", "secret")
	os.Setenv("SESSION_SECRET", "shh")
	os.Setenv("COOKIE_SECURE", "true")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("ADMIN_PASSWORD")
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("COOKIE_SECURE")
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://example/test" {
		t.Errorf("expected env DatabaseURL, got %q", cfg.DatabaseURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected env Port, got %q", cfg.Port)
	}
	if cfg.AdminPassword != "secret" {
		t.Errorf("expected env AdminPassword, got %q", cfg.AdminPassword)
	}
	if cfg.SessionSecret != "shh" {
		t.Errorf("expected env SessionSecret, got %q", cfg.SessionSecret)
	}
	if cfg.CookieSecure != true {
		t.Errorf("expected env CookieSecure true, got %v", cfg.CookieSecure)
	}
}

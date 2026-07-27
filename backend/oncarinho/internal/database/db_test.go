package database

import "testing"

func TestConnectInvalidURL(t *testing.T) {
	_, err := Connect("postgres://invalid-host-that-does-not-exist:5432/db?sslmode=disable&connect_timeout=1")
	if err == nil {
		t.Fatal("expected error connecting to invalid host, got nil")
	}
}

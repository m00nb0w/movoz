package handlers

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDBForHandlers(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/scout_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

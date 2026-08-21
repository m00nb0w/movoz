package store

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
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
		t.Skipf("skipping: test database not available at %s: %v", url, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// truncateTables truncates the given tables in order with RESTART IDENTITY
// CASCADE.
//
// main_attributes is special: migration 000002 seeds the initial 6 rows, and
// TestMainAttributeStoreSeedData asserts they exist without reseeding them
// itself. Every test in this package shares one test binary and one database,
// so a test that truncates main_attributes and walks away leaves whichever
// seed-dependent test runs next looking at an empty table — an intra-package
// ordering bug that `go test -p 1` cannot fix (it only serializes *across*
// packages). Restoring the seed rows on cleanup (rather than immediately)
// keeps the truncating test's own view of the table empty, which is what its
// fixtures expect, while guaranteeing the seed rows are back before the next
// test starts.
func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
		if table == "main_attributes" {
			t.Cleanup(func() { seedMainAttributes(t, db) })
		}
	}
}

func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

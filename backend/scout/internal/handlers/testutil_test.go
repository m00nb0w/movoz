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

// seedMainAttributes restores the 6 main attributes migration 000002 seeds.
// Idempotent (ON CONFLICT DO NOTHING) so it is safe to call after a partial
// truncate.
func seedMainAttributes(t *testing.T, db *sql.DB) {
	t.Helper()
	seedData := []struct {
		key  string
		name string
	}{
		{"technical_expertise", "Technical Expertise"},
		{"critical_thinking", "Critical Thinking"},
		{"communication", "Communication"},
		{"management", "Management"},
		{"product_mindset", "Product Mindset"},
		{"force_multiplier", "Force Multiplier"},
	}
	for _, seed := range seedData {
		if _, err := db.Exec("INSERT INTO main_attributes (key, name) VALUES ($1, $2) ON CONFLICT DO NOTHING", seed.key, seed.name); err != nil {
			t.Fatalf("failed to reseed main_attributes: %v", err)
		}
	}
}

// truncateTables truncates the given tables in order with RESTART IDENTITY
// CASCADE, restoring main_attributes' migration seed rows on cleanup whenever
// that table is one of them.
//
// Handler tests here (e.g. TestMainAttributeHandlerList, which asserts an
// exact row count) need an empty table *during* the test, so the reseed runs
// on cleanup rather than immediately. Without it, a truncate in this package
// leaves the shared scout_test database without the seeded 6 attributes,
// breaking store.TestMainAttributeStoreSeedData and any other seed-dependent
// assertion that happens to run afterwards.
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

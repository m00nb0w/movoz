# Scout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build Scout — a private Go/Next.js tool that tracks each engineer's FIFA-style attribute card via biweekly peer-relative ranking, synced GitHub/Jira metrics, an AI ranking-chat assistant, and a manual highlight/lowlight log — per `wiki/specs/scout.md`.

**Architecture:** `backend/scout` is a Go REST API (Gin + `lib/pq` + `golang-migrate`) following the exact `cmd/server` / `internal/{config,database,models,store,handlers,auth}` layout used by `backend/hustle-turtle` and `backend/oncarinho`, backed by its own PostgreSQL database. Every route sits behind session auth (no public group at all, unlike `oncarinho`) via the same shared-password + HMAC-signed-cookie pattern as `oncarinho`'s `internal/auth`. Main-attribute, Overall, and cycle-view scores are computed on read from `sub_attribute_rankings` via SQL aggregation — mirroring `oncarinho`'s leaderboard/profile computed-on-read pattern — never cached. A ticker-based in-process scheduler (started from `cmd/server/main.go`, no separate binary) polls GitHub and Jira per active engineer on a configurable interval and idempotently upserts `metric_snapshots`. `apps/scout` is a new Next.js 14 multi-zone app (own `basePath`/`assetPrefix`, consuming `@movoz/tailwind-config` + `@movoz/theme`) with no public routes — every page requires an authenticated session.

**Tech Stack:** Go 1.23+, Gin, `lib/pq`, `golang-migrate`, PostgreSQL — identical dependency set to `hustle-turtle`/`oncarinho` (no new Go deps beyond one). For the Claude API integration (F9 conversational ranking chat, F14 semantic duplicate check) this plan adds `github.com/anthropics/anthropic-sdk-go`, the official Anthropic Go SDK — chosen over raw HTTP because the Claude API skill's stated default is "the official SDK for the project's language... whenever a supported SDK exists," and the SDK's streaming helpers (`Messages.NewStreaming`, `stream.Next()`/`Current()`/`Err()`) and structured-output support (`OutputConfig.Format` + `JSONOutputFormatParam`) are exactly what both AI flows need without hand-rolling SSE parsing or JSON-schema validation. Model: `claude-opus-5` (the skill's mandatory default absent an explicit user preference for a cheaper model — flagged in the judgment-calls list below since Scout's volume is low and Sonnet would also be adequate). Frontend: Next.js 14 (App Router), React, `@movoz/tailwind-config`, `@movoz/theme`, TypeScript — same stack as `apps/personal-site`/`apps/drunken-dolphin`, no new frontend dependencies.

## Global Constraints

These apply to every task in this plan; re-verify each task satisfies them before marking it done.

- **NF1 — No public access.** Every `/api/*` route (except `POST /api/auth/login`, which cannot itself require a session, and `GET /health`, an infra liveness probe with no application data — both exempted the same way `oncarinho` exempts them) requires a valid session. There is no `oncarinho`-style public route group for application data.
- **NF2 — Scores always derivable, no drift.** Main-attribute, Overall, cycle-view, and trend scores are computed on read via SQL from `sub_attribute_rankings` on every request. No materialized/cached score table, no background recomputation job.
- **NF3 — AI never silently applies.** The AI ranking chat's `proposed_ranking` is persisted into `sub_attribute_rankings` only via the explicit `POST /api/cycles/:id/ai-sessions/:sessionId/accept` call, which accepts the (possibly admin-edited) ranking in the request body — never auto-applied when the chat produces a proposal. The highlight/lowlight duplicate check never blocks a save; a flagged entry can still be saved by the admin, and if the AI call fails or times out, the save proceeds without a flag rather than blocking.
- **NF4 — Sync worker is idempotent.** Every sync upsert into `metric_snapshots` is keyed on `(engineer_id, period_start, period_end)` via a DB unique constraint + `ON CONFLICT ... DO UPDATE`. A failed or repeated run never duplicates or corrupts rows. Per-engineer sync failures are logged and skipped, not fatal to the run.
- **F6/F7 ranking validation.** Every ranking submission (manual or AI-accepted) must be a strict 1..N permutation of exactly the active roster for that cycle+sub-attribute — no ties, no gaps, no duplicates, no missing/extra engineers — or the API rejects with 400. Rank→score is linear interpolation: rank 1 → 100, rank N → 50, evenly spaced (`score = 100` when `N == 1`, else `100 - (rank-1) * 50/(N-1)`).
- **F8 Overall-score cutover rule.** A main attribute only counts toward Overall for cycles from its creation forward; adding a main attribute never retroactively changes a past cycle's Overall. Implemented as: a main attribute counts toward a cycle's Overall iff `main_attributes.created_at <= rating_cycles.created_at` (flagged as a judgment call below — the spec doesn't pin down "existed as of that cycle" to created-cycle-timestamp vs. period_end).
- **F12 is explicitly out of scope.** Standalone automatic performance rating is Deferred per the spec's own Non-Goals/Open-Questions — no task in this plan implements it.
- **Testing convention.** Go: table-driven tests per `hustle-turtle`/`oncarinho` convention, `go test ./...`, using a `oncarinho_test`-style dedicated `scout_test` database (tests skip gracefully — not silently-pass-as-success — when `TEST_DATABASE_URL`/default is unreachable, exactly like `oncarinho`). Frontend: TypeScript type-checking (`tsc --noEmit` / `next build`) + manual verification only — this repo's Next.js apps have no test framework (confirmed: no `jest`/`vitest`/`playwright` config in `apps/personal-site` or `apps/drunken-dolphin`), so frontend tasks use a type-check-and-verify cycle instead of the red/green test cycle used for Go tasks.
- **Test isolation caveat (unique to Scout vs. its siblings):** unlike `oncarinho`'s tables, `main_attributes` is migration-seeded (the initial 6, F3), and several store/handler tests `TRUNCATE` it (plus dependent tables) before seeding their own fixture rows. This is safe within a single sequential test run, but Go runs different packages' test binaries concurrently by default, and a truncate in one package can race a seed-dependent assertion in another (e.g. `TestMainAttributeStoreSeedData`). Run the full suite with `go test -p 1 ./...` (forces packages to run one at a time) to eliminate this race; the backend README says so explicitly.

## File Map

| Path | Responsibility |
|---|---|
| `backend/scout/go.mod` | Module `scout`, deps: gin, golang-migrate, lib/pq, anthropic-sdk-go |
| `backend/scout/cmd/server/main.go` | Entrypoint: config, migrations CLI, DB connect, router, sync scheduler |
| `backend/scout/cmd/server/router.go` | Route wiring: health + login public, everything else behind `RequireAuth` |
| `backend/scout/internal/config/config.go` | Env-var config (DB, port, auth, Anthropic, GitHub, Jira, sync interval) |
| `backend/scout/internal/database/db.go` | `Connect` |
| `backend/scout/internal/database/migrate.go` | `MigrationManager` (Up/Down/Version) |
| `backend/scout/internal/auth/session.go` | HMAC session token sign/verify (copied pattern from `oncarinho`) |
| `backend/scout/internal/handlers/*.go` | Gin handlers per resource |
| `backend/scout/internal/models/*.go` | Row structs |
| `backend/scout/internal/store/*.go` | SQL access, one file per aggregate |
| `backend/scout/internal/scoring/*.go` | Pure functions: rank→score, permutation validation |
| `backend/scout/internal/aiclient/*.go` | Anthropic SDK wrapper: streaming chat, duplicate check, JSON-block extraction |
| `backend/scout/internal/integrations/*.go` | GitHub + Jira API clients |
| `backend/scout/internal/syncer/*.go` | Sync orchestration + ticker scheduler |
| `backend/scout/migrations/*.sql` | golang-migrate SQL files |
| `backend/scout/README.md` | Setup, env vars, API surface, testing instructions |
| `apps/scout/*` | Next.js 14 zone: login, dashboard, roster, attributes, cycles, engineer card, highlights |

---

### Task 1: Backend scaffold — config, database, health check

**Files:**
- Create: `backend/scout/go.mod`
- Create: `backend/scout/internal/config/config.go`
- Create: `backend/scout/internal/config/config_test.go`
- Create: `backend/scout/internal/database/db.go`
- Create: `backend/scout/internal/database/migrate.go`
- Create: `backend/scout/internal/handlers/health.go`
- Create: `backend/scout/internal/handlers/health_test.go`
- Create: `backend/scout/cmd/server/router.go`
- Create: `backend/scout/cmd/server/main.go`

**Interfaces:**
- Produces: `config.Config{DatabaseURL, Port, AdminPassword, SessionSecret, CookieSecure, AnthropicAPIKey, GitHubToken, GitHubRepos []string, JiraBaseURL, JiraEmail, JiraAPIToken, JiraProjects []string, SyncInterval time.Duration}`, `config.Load() *Config`; `database.Connect(databaseURL string) (*sql.DB, error)`; `database.NewMigrationManager(databaseURL string) *MigrationManager` with `.Up()`/`.Down()`/`.Version()`; `handlers.NewHealthHandler() *HealthHandler` with `.HealthCheck(c *gin.Context)`; `main.buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/config/config_test.go
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
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/config/... -run TestLoad -v`
Expected: FAIL — `package config: no Go files` (package doesn't exist yet)
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/go.mod
module scout

go 1.23.0

require (
	github.com/anthropics/anthropic-sdk-go v1.4.0
	github.com/gin-gonic/gin v1.10.1
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/lib/pq v1.10.9
)
```
```go
// backend/scout/internal/config/config.go
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
```
```go
// backend/scout/internal/database/db.go
package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}
	return db, nil
}
```
```go
// backend/scout/internal/database/migrate.go
package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type MigrationManager struct {
	databaseURL string
}

func NewMigrationManager(databaseURL string) *MigrationManager {
	return &MigrationManager{databaseURL: databaseURL}
}

func (mm *MigrationManager) getMigrationInstance() (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", mm.databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %v", err)
	}

	return m, nil
}

func (mm *MigrationManager) Up() error {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %v", err)
	}
	log.Println("Up migrations completed successfully")
	return nil
}

func (mm *MigrationManager) Down() error {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run down migrations: %v", err)
	}
	log.Println("Down migrations completed successfully")
	return nil
}

func (mm *MigrationManager) Version() (uint, bool, error) {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return 0, false, fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("could not get migration version: %v", err)
	}
	return version, dirty, nil
}
```
```go
// backend/scout/internal/handlers/health.go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "scout"})
}
```
```go
// backend/scout/cmd/server/router.go
package main

import (
	"database/sql"

	"scout/internal/config"
	"scout/internal/handlers"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	healthHandler := handlers.NewHealthHandler()

	r := gin.Default()
	r.GET("/health", healthHandler.HealthCheck)

	return r
}
```
```go
// backend/scout/cmd/server/main.go
package main

import (
	"flag"
	"fmt"
	"log"

	"scout/internal/config"
	"scout/internal/database"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	var (
		migrateDir  = flag.String("migrate", "", "Run database migrations: 'up', 'down'")
		version     = flag.Bool("version", false, "Show current migration version")
		autoMigrate = flag.Bool("auto-migrate", false, "Run up migrations on startup")
	)
	flag.Parse()

	migrationManager := database.NewMigrationManager(cfg.DatabaseURL)

	if *version {
		v, dirty, err := migrationManager.Version()
		if err != nil {
			log.Fatalf("could not get migration version: %v", err)
		}
		status := "clean"
		if dirty {
			status = "dirty"
		}
		fmt.Printf("Current migration version: %d (status: %s)\n", v, status)
		return
	}

	if *migrateDir != "" {
		switch *migrateDir {
		case "up":
			if err := migrationManager.Up(); err != nil {
				log.Fatalf("migration up failed: %v", err)
			}
		case "down":
			if err := migrationManager.Down(); err != nil {
				log.Fatalf("migration down failed: %v", err)
			}
		default:
			log.Fatalf("invalid migration direction: %s", *migrateDir)
		}
		return
	}

	if *autoMigrate {
		if err := migrationManager.Up(); err != nil {
			log.Printf("auto migration failed: %v", err)
		}
	}

	if cfg.AdminPassword == "" || cfg.SessionSecret == "" {
		log.Fatal("ADMIN_PASSWORD and SESSION_SECRET must be set")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	r := buildRouter(db, cfg)

	port := ":" + cfg.Port
	log.Printf("starting server on port %s", cfg.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go mod tidy && go test ./internal/config/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/go.mod backend/scout/go.sum backend/scout/internal/config backend/scout/internal/database backend/scout/internal/handlers/health.go backend/scout/internal/handlers/health_test.go backend/scout/cmd/server
git commit -m "scout: scaffold Go backend (config, database, health check)"
```

---

### Task 2: Engineers migration, model, store (F2)

**Files:**
- Create: `backend/scout/migrations/000001_create_engineers_table.up.sql`
- Create: `backend/scout/migrations/000001_create_engineers_table.down.sql`
- Create: `backend/scout/internal/models/engineer.go`
- Create: `backend/scout/internal/store/engineers.go`
- Create: `backend/scout/internal/store/testutil_test.go`
- Create: `backend/scout/internal/store/engineers_test.go`

**Interfaces:**
- Consumes: nothing new
- Produces: `models.Engineer{ID int, Name string, Role *string, GitHubUsername *string, JiraAccountID *string, StartedAt time.Time, IsActive bool, CreatedAt time.Time}`; `store.EngineerStore` with `NewEngineerStore(db *sql.DB) *EngineerStore`, `.List(activeOnly bool) ([]models.Engineer, error)`, `.GetByID(id int) (*models.Engineer, error)`, `.Create(name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error)`, `.Update(id int, name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error)`, `.Deactivate(id int) (bool, error)`, `.Reactivate(id int) (bool, error)`, `.Exists(id int) (bool, error)`, `.ListActiveIDs() ([]int, error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/testutil_test.go
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
```
```go
// backend/scout/internal/store/engineers_test.go
package store

import (
	"testing"
	"time"
)

func TestEngineerStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewEngineerStore(db)

	role := "Backend Engineer"
	gh := "octocat"
	created, err := s.Create("Alex Kim", &role, &gh, nil, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Name != "Alex Kim" || !created.IsActive {
		t.Fatalf("unexpected created engineer: %+v", created)
	}

	list, err := s.List(true)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one active engineer, got %+v", list)
	}
}

func TestEngineerStoreDeactivateExcludesFromActiveList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewEngineerStore(db)

	created, _ := s.Create("Sam Lee", nil, nil, nil, time.Now())
	ok, err := s.Deactivate(created.ID)
	if err != nil || !ok {
		t.Fatalf("deactivate failed: ok=%v err=%v", ok, err)
	}

	activeIDs, err := s.ListActiveIDs()
	if err != nil {
		t.Fatalf("list active ids failed: %v", err)
	}
	for _, id := range activeIDs {
		if id == created.ID {
			t.Fatalf("deactivated engineer %d still in active list", id)
		}
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestEngineerStore -v`
Expected: FAIL — `undefined: NewEngineerStore` (compile error)
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000001_create_engineers_table.up.sql
CREATE TABLE IF NOT EXISTS engineers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(255),
    github_username VARCHAR(255),
    jira_account_id VARCHAR(255),
    started_at DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_engineers_is_active ON engineers(is_active);
```
```sql
-- backend/scout/migrations/000001_create_engineers_table.down.sql
DROP TABLE IF EXISTS engineers;
```
```go
// backend/scout/internal/models/engineer.go
package models

import "time"

type Engineer struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Role           *string   `json:"role"`
	GitHubUsername *string   `json:"github_username"`
	JiraAccountID  *string   `json:"jira_account_id"`
	StartedAt      time.Time `json:"started_at"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}
```
```go
// backend/scout/internal/store/engineers.go
package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type EngineerStore struct {
	db *sql.DB
}

func NewEngineerStore(db *sql.DB) *EngineerStore {
	return &EngineerStore{db: db}
}

func (s *EngineerStore) List(activeOnly bool) ([]models.Engineer, error) {
	query := `SELECT id, name, role, github_username, jira_account_id, started_at, is_active, created_at FROM engineers`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	engineers := []models.Engineer{}
	for rows.Next() {
		var e models.Engineer
		if err := rows.Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt); err != nil {
			return nil, err
		}
		engineers = append(engineers, e)
	}
	return engineers, rows.Err()
}

func (s *EngineerStore) GetByID(id int) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`SELECT id, name, role, github_username, jira_account_id, started_at, is_active, created_at
		 FROM engineers WHERE id = $1`, id,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Create(name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`INSERT INTO engineers (name, role, github_username, jira_account_id, started_at, is_active)
		 VALUES ($1, $2, $3, $4, $5, true)
		 RETURNING id, name, role, github_username, jira_account_id, started_at, is_active, created_at`,
		name, role, githubUsername, jiraAccountID, startedAt,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Update(id int, name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`UPDATE engineers SET name = $1, role = $2, github_username = $3, jira_account_id = $4, started_at = $5
		 WHERE id = $6
		 RETURNING id, name, role, github_username, jira_account_id, started_at, is_active, created_at`,
		name, role, githubUsername, jiraAccountID, startedAt, id,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Deactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE engineers SET is_active = false WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *EngineerStore) Reactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE engineers SET is_active = true WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *EngineerStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM engineers WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (s *EngineerStore) ListActiveIDs() ([]int, error) {
	rows, err := s.db.Query("SELECT id FROM engineers WHERE is_active = true ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `createdb scout_test 2>/dev/null; cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... -v`
Expected: PASS (or graceful SKIP if no local Postgres)
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations backend/scout/internal/models/engineer.go backend/scout/internal/store/engineers.go backend/scout/internal/store/testutil_test.go backend/scout/internal/store/engineers_test.go
git commit -m "scout: add engineers table, model, and store (F2)"
```

---

### Task 3: Engineer handlers + routes (F2)

**Files:**
- Create: `backend/scout/internal/handlers/engineers.go`
- Create: `backend/scout/internal/handlers/engineers_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.NewEngineerStore`, `models.Engineer`
- Produces: `handlers.NewEngineerHandler(s *store.EngineerStore) *EngineerHandler` with `.List`, `.Get`, `.Create`, `.Update`, `.Deactivate`, `.Reactivate` (`gin.HandlerFunc`-compatible methods)

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/engineers_test.go
package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupEngineerTestRouter(t *testing.T) *gin.Engine {
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
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	engineerStore := store.NewEngineerStore(db)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewEngineerHandler(engineerStore)
	r.GET("/api/engineers", h.List)
	r.GET("/api/engineers/:id", h.Get)
	r.POST("/api/engineers", h.Create)
	r.PUT("/api/engineers/:id", h.Update)
	r.DELETE("/api/engineers/:id", h.Deactivate)
	r.POST("/api/engineers/:id/reactivate", h.Reactivate)
	return r
}

func TestEngineerHandlerCreateAndList(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Alex Kim", "started_at": "2024-01-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/engineers", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var engineers []models.Engineer
	json.Unmarshal(w.Body.Bytes(), &engineers)
	if len(engineers) != 1 || engineers[0].Name != "Alex Kim" {
		t.Fatalf("expected one engineer Alex Kim, got %+v", engineers)
	}
}

func TestEngineerHandlerCreateMissingName(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"started_at": "2024-01-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEngineerHandlerDeactivateNotFound(t *testing.T) {
	r := setupEngineerTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/engineers/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestEngineerHandlerGet(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Alex Kim", "started_at": "2024-01-15"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	var created models.Engineer
	json.Unmarshal(createW.Body.Bytes(), &created)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(created.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var fetched models.Engineer
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.ID != created.ID || fetched.Name != "Alex Kim" {
		t.Fatalf("unexpected fetched engineer: %+v", fetched)
	}
}

func TestEngineerHandlerGetNotFound(t *testing.T) {
	r := setupEngineerTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestEngineerHandler -v`
Expected: FAIL — `undefined: NewEngineerHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/engineers.go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type EngineerHandler struct {
	store *store.EngineerStore
}

func NewEngineerHandler(s *store.EngineerStore) *EngineerHandler {
	return &EngineerHandler{store: s}
}

type engineerRequest struct {
	Name           string  `json:"name" binding:"required"`
	Role           *string `json:"role"`
	GitHubUsername *string `json:"github_username"`
	JiraAccountID  *string `json:"jira_account_id"`
	StartedAt      string  `json:"started_at" binding:"required"`
}

func (h *EngineerHandler) List(c *gin.Context) {
	activeOnly := c.DefaultQuery("active", "true") != "all"
	engineers, err := h.store.List(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list engineers"})
		return
	}
	c.JSON(http.StatusOK, engineers)
}

// Get handles GET /api/engineers/:id — a single engineer, independent of
// any cycle. The frontend engineer card page (a later task) uses this for
// the page header before any cycle-scoped data is fetched.
func (h *EngineerHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	engineer, err := h.store.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if engineer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, engineer)
}

func (h *EngineerHandler) Create(c *gin.Context) {
	var req engineerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and started_at are required"})
		return
	}
	startedAt, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "started_at must be YYYY-MM-DD"})
		return
	}

	engineer, err := h.store.Create(req.Name, req.Role, req.GitHubUsername, req.JiraAccountID, startedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create engineer"})
		return
	}
	c.JSON(http.StatusCreated, engineer)
}

func (h *EngineerHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req engineerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and started_at are required"})
		return
	}
	startedAt, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "started_at must be YYYY-MM-DD"})
		return
	}

	engineer, err := h.store.Update(id, req.Name, req.Role, req.GitHubUsername, req.JiraAccountID, startedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update engineer"})
		return
	}
	if engineer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, engineer)
}

func (h *EngineerHandler) Deactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	ok, err := h.store.Deactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate engineer"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EngineerHandler) Reactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	ok, err := h.store.Reactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reactivate engineer"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
```
```go
// backend/scout/cmd/server/router.go — modify buildRouter
package main

import (
	"database/sql"

	"scout/internal/config"
	"scout/internal/handlers"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	engineerStore := store.NewEngineerStore(db)

	healthHandler := handlers.NewHealthHandler()
	engineerHandler := handlers.NewEngineerHandler(engineerStore)

	r := gin.Default()
	r.GET("/health", healthHandler.HealthCheck)

	api := r.Group("/api")
	{
		api.GET("/engineers", engineerHandler.List)
		api.GET("/engineers/:id", engineerHandler.Get)
		api.POST("/engineers", engineerHandler.Create)
		api.PUT("/engineers/:id", engineerHandler.Update)
		api.DELETE("/engineers/:id", engineerHandler.Deactivate)
		api.POST("/engineers/:id/reactivate", engineerHandler.Reactivate)
	}

	return r
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/engineers.go backend/scout/internal/handlers/engineers_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add engineer CRUD handlers and routes (F2)"
```

---

### Task 4: Main attributes migration, model, store — seeded with initial 6 (F3)

**Files:**
- Create: `backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.up.sql`
- Create: `backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.down.sql`
- Create: `backend/scout/internal/models/attribute.go`
- Create: `backend/scout/internal/store/main_attributes.go`
- Create: `backend/scout/internal/store/main_attributes_test.go`

**Interfaces:**
- Produces: `models.MainAttribute{ID int, Key string, Name string, CreatedAt time.Time}`; `models.SubAttribute{ID int, MainAttributeID int, Name string, Description *string, IsActive bool, CreatedAt time.Time}`; `store.MainAttributeStore` with `NewMainAttributeStore(db *sql.DB) *MainAttributeStore`, `.List() ([]models.MainAttribute, error)`, `.GetByID(id int) (*models.MainAttribute, error)`, `.Create(key, name string) (*models.MainAttribute, error)`, `.Update(id int, name string) (*models.MainAttribute, error)`, `.Exists(id int) (bool, error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/main_attributes_test.go
package store

import "testing"

func TestMainAttributeStoreSeedData(t *testing.T) {
	db := setupTestDB(t)
	s := NewMainAttributeStore(db)

	list, err := s.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) < 6 {
		t.Fatalf("expected at least 6 seeded main attributes, got %d", len(list))
	}
	found := map[string]bool{}
	for _, a := range list {
		found[a.Key] = true
	}
	for _, key := range []string{"technical_expertise", "critical_thinking", "communication", "management", "product_mindset", "force_multiplier"} {
		if !found[key] {
			t.Fatalf("expected seeded main attribute %q, not found in %+v", key, list)
		}
	}
}

func TestMainAttributeStoreCreate(t *testing.T) {
	db := setupTestDB(t)
	s := NewMainAttributeStore(db)

	created, err := s.Create("delivery_speed", "Delivery Speed")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Key != "delivery_speed" || created.Name != "Delivery Speed" {
		t.Fatalf("unexpected created main attribute: %+v", created)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestMainAttributeStore -v`
Expected: FAIL — `undefined: NewMainAttributeStore`
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.up.sql
CREATE TABLE IF NOT EXISTS main_attributes (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sub_attributes (
    id SERIAL PRIMARY KEY,
    main_attribute_id INTEGER NOT NULL REFERENCES main_attributes(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_sub_attributes_main_attribute_id ON sub_attributes(main_attribute_id);
CREATE INDEX idx_sub_attributes_is_active ON sub_attributes(is_active);

INSERT INTO main_attributes (key, name) VALUES
    ('technical_expertise', 'Technical Expertise'),
    ('critical_thinking', 'Critical Thinking'),
    ('communication', 'Communication'),
    ('management', 'Management'),
    ('product_mindset', 'Product Mindset'),
    ('force_multiplier', 'Force Multiplier');
```
```sql
-- backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.down.sql
DROP TABLE IF EXISTS sub_attributes;
DROP TABLE IF EXISTS main_attributes;
```
```go
// backend/scout/internal/models/attribute.go
package models

import "time"

type MainAttribute struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SubAttribute struct {
	ID              int       `json:"id"`
	MainAttributeID int       `json:"main_attribute_id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}
```
```go
// backend/scout/internal/store/main_attributes.go
package store

import (
	"database/sql"

	"scout/internal/models"
)

type MainAttributeStore struct {
	db *sql.DB
}

func NewMainAttributeStore(db *sql.DB) *MainAttributeStore {
	return &MainAttributeStore{db: db}
}

func (s *MainAttributeStore) List() ([]models.MainAttribute, error) {
	rows, err := s.db.Query("SELECT id, key, name, created_at FROM main_attributes ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attrs := []models.MainAttribute{}
	for rows.Next() {
		var a models.MainAttribute
		if err := rows.Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	return attrs, rows.Err()
}

func (s *MainAttributeStore) GetByID(id int) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow("SELECT id, key, name, created_at FROM main_attributes WHERE id = $1", id).
		Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Create(key, name string) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow(
		`INSERT INTO main_attributes (key, name) VALUES ($1, $2)
		 RETURNING id, key, name, created_at`,
		key, name,
	).Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Update(id int, name string) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow(
		"UPDATE main_attributes SET name = $1 WHERE id = $2 RETURNING id, key, name, created_at",
		name, id,
	).Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM main_attributes WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... -run TestMainAttributeStore -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.up.sql backend/scout/migrations/000002_create_main_attributes_and_sub_attributes.down.sql backend/scout/internal/models/attribute.go backend/scout/internal/store/main_attributes.go backend/scout/internal/store/main_attributes_test.go
git commit -m "scout: add main_attributes/sub_attributes tables + main attribute store, seed initial 6 (F3)"
```

---

### Task 5: Main attribute handlers + routes (F3)

**Files:**
- Create: `backend/scout/internal/handlers/main_attributes.go`
- Create: `backend/scout/internal/handlers/main_attributes_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.MainAttributeStore`
- Produces: `handlers.NewMainAttributeHandler(s *store.MainAttributeStore) *MainAttributeHandler` with `.List`, `.Create`, `.Update`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/main_attributes_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestMainAttributeHandlerCreate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	s := store.NewMainAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMainAttributeHandler(s)
	r.GET("/api/main-attributes", h.List)
	r.POST("/api/main-attributes", h.Create)

	body, _ := json.Marshal(map[string]string{"key": "delivery_speed", "name": "Delivery Speed"})
	req := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
```
```go
// backend/scout/internal/handlers/testutil_test.go
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
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestMainAttributeHandler -v`
Expected: FAIL — `undefined: NewMainAttributeHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/main_attributes.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type MainAttributeHandler struct {
	store *store.MainAttributeStore
}

func NewMainAttributeHandler(s *store.MainAttributeStore) *MainAttributeHandler {
	return &MainAttributeHandler{store: s}
}

type mainAttributeRequest struct {
	Key  string `json:"key" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (h *MainAttributeHandler) List(c *gin.Context) {
	attrs, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list main attributes"})
		return
	}
	c.JSON(http.StatusOK, attrs)
}

func (h *MainAttributeHandler) Create(c *gin.Context) {
	var req mainAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key and name are required"})
		return
	}
	attr, err := h.store.Create(req.Key, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create main attribute"})
		return
	}
	c.JSON(http.StatusCreated, attr)
}

func (h *MainAttributeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid main attribute id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	attr, err := h.store.Update(id, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update main attribute"})
		return
	}
	if attr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "main attribute not found"})
		return
	}
	c.JSON(http.StatusOK, attr)
}
```
```go
// backend/scout/cmd/server/router.go — modify buildRouter to add main-attributes routes
	mainAttributeStore := store.NewMainAttributeStore(db)
	mainAttributeHandler := handlers.NewMainAttributeHandler(mainAttributeStore)
	// ... inside the api group, alongside the engineers routes:
		api.GET("/main-attributes", mainAttributeHandler.List)
		api.POST("/main-attributes", mainAttributeHandler.Create)
		api.PUT("/main-attributes/:id", mainAttributeHandler.Update)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestMainAttributeHandler -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/main_attributes.go backend/scout/internal/handlers/main_attributes_test.go backend/scout/internal/handlers/testutil_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add main attribute handlers and routes (F3)"
```

---

### Task 6: Sub-attribute store + handlers, scoped to a main attribute (F3)

**Files:**
- Create: `backend/scout/internal/store/sub_attributes.go`
- Create: `backend/scout/internal/store/sub_attributes_test.go`
- Create: `backend/scout/internal/handlers/sub_attributes.go`
- Create: `backend/scout/internal/handlers/sub_attributes_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `models.SubAttribute` (Task 4), `store.MainAttributeStore.Exists` (Task 4)
- Produces: `store.SubAttributeStore` with `NewSubAttributeStore(db *sql.DB) *SubAttributeStore`, `.ListByMainAttribute(mainAttributeID int, activeOnly bool) ([]models.SubAttribute, error)`, `.ListAllActive() ([]models.SubAttribute, error)`, `.GetByID(id int) (*models.SubAttribute, error)`, `.Create(mainAttributeID int, name string, description *string) (*models.SubAttribute, error)`, `.Update(id int, name string, description *string) (*models.SubAttribute, error)`, `.Deactivate(id int) (bool, error)`, `.Exists(id int) (bool, error)`; `handlers.NewSubAttributeHandler(s *store.SubAttributeStore, mainAttrStore *store.MainAttributeStore) *SubAttributeHandler` with `.List`, `.Create`, `.Update`, `.Deactivate`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/sub_attributes_test.go
package store

import "testing"

func TestSubAttributeStoreCreateAndListByMainAttribute(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, err := mainStore.Create("test_main_"+t.Name(), "Test Main")
	if err != nil {
		t.Fatalf("create main attribute failed: %v", err)
	}

	desc := "Writes clean, well-tested code"
	created, err := subStore.Create(main.ID, "Code Quality", &desc)
	if err != nil {
		t.Fatalf("create sub attribute failed: %v", err)
	}
	if created.MainAttributeID != main.ID || created.Name != "Code Quality" {
		t.Fatalf("unexpected created sub attribute: %+v", created)
	}

	list, err := subStore.ListByMainAttribute(main.ID, true)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one sub attribute under main %d, got %+v", main.ID, list)
	}
}

func TestSubAttributeStoreDeactivateExcludesFromActiveList(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create("test_main2_"+t.Name(), "Test Main 2")
	created, _ := subStore.Create(main.ID, "Ownership", nil)

	ok, err := subStore.Deactivate(created.ID)
	if err != nil || !ok {
		t.Fatalf("deactivate failed: ok=%v err=%v", ok, err)
	}

	active, err := subStore.ListAllActive()
	if err != nil {
		t.Fatalf("list all active failed: %v", err)
	}
	for _, s := range active {
		if s.ID == created.ID {
			t.Fatalf("deactivated sub attribute %d still in active list", s.ID)
		}
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestSubAttributeStore -v`
Expected: FAIL — `undefined: NewSubAttributeStore`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/store/sub_attributes.go
package store

import (
	"database/sql"

	"scout/internal/models"
)

type SubAttributeStore struct {
	db *sql.DB
}

func NewSubAttributeStore(db *sql.DB) *SubAttributeStore {
	return &SubAttributeStore{db: db}
}

func (s *SubAttributeStore) ListByMainAttribute(mainAttributeID int, activeOnly bool) ([]models.SubAttribute, error) {
	query := `SELECT id, main_attribute_id, name, description, is_active, created_at
	          FROM sub_attributes WHERE main_attribute_id = $1`
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query, mainAttributeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []models.SubAttribute{}
	for rows.Next() {
		var sa models.SubAttribute
		if err := rows.Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sa)
	}
	return subs, rows.Err()
}

func (s *SubAttributeStore) ListAllActive() ([]models.SubAttribute, error) {
	rows, err := s.db.Query(
		`SELECT id, main_attribute_id, name, description, is_active, created_at
		 FROM sub_attributes WHERE is_active = true ORDER BY main_attribute_id, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []models.SubAttribute{}
	for rows.Next() {
		var sa models.SubAttribute
		if err := rows.Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sa)
	}
	return subs, rows.Err()
}

func (s *SubAttributeStore) GetByID(id int) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`SELECT id, main_attribute_id, name, description, is_active, created_at
		 FROM sub_attributes WHERE id = $1`, id,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Create(mainAttributeID int, name string, description *string) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`INSERT INTO sub_attributes (main_attribute_id, name, description, is_active)
		 VALUES ($1, $2, $3, true)
		 RETURNING id, main_attribute_id, name, description, is_active, created_at`,
		mainAttributeID, name, description,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Update(id int, name string, description *string) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`UPDATE sub_attributes SET name = $1, description = $2 WHERE id = $3
		 RETURNING id, main_attribute_id, name, description, is_active, created_at`,
		name, description, id,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Deactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE sub_attributes SET is_active = false WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *SubAttributeStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM sub_attributes WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
```
```go
// backend/scout/internal/handlers/sub_attributes.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type SubAttributeHandler struct {
	store         *store.SubAttributeStore
	mainAttrStore *store.MainAttributeStore
}

func NewSubAttributeHandler(s *store.SubAttributeStore, mainAttrStore *store.MainAttributeStore) *SubAttributeHandler {
	return &SubAttributeHandler{store: s, mainAttrStore: mainAttrStore}
}

type subAttributeRequest struct {
	MainAttributeID int     `json:"main_attribute_id" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	Description     *string `json:"description"`
}

func (h *SubAttributeHandler) List(c *gin.Context) {
	mainAttributeID, err := strconv.Atoi(c.Query("main_attribute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "main_attribute_id query param is required"})
		return
	}
	activeOnly := c.DefaultQuery("active", "true") != "all"
	subs, err := h.store.ListByMainAttribute(mainAttributeID, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sub attributes"})
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *SubAttributeHandler) Create(c *gin.Context) {
	var req subAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "main_attribute_id and name are required"})
		return
	}
	exists, err := h.mainAttrStore.Exists(req.MainAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up main attribute"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown main_attribute_id"})
		return
	}

	sub, err := h.store.Create(req.MainAttributeID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sub attribute"})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (h *SubAttributeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	sub, err := h.store.Update(id, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sub attribute"})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *SubAttributeHandler) Deactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}
	ok, err := h.store.Deactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate sub attribute"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return
	}
	c.Status(http.StatusNoContent)
}
```
```go
// backend/scout/internal/handlers/sub_attributes_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestSubAttributeHandlerCreate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Deliberately not truncating main_attributes here — it holds the
	// migration-seeded initial 6 (F3), which other packages' tests (e.g.
	// TestMainAttributeStoreSeedData) assert on against the same shared
	// test database. Read the existing seed instead of wiping it.
	mains, err := mainStore.List()
	if err != nil || len(mains) == 0 {
		t.Fatalf("expected at least one seeded main attribute (run migrations first): mains=%v err=%v", mains, err)
	}
	main := mains[0]

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"main_attribute_id": main.ID, "name": "Testability"})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSubAttributeHandlerCreateUnknownMainAttribute(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"main_attribute_id": 999999, "name": "Testability"})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'SubAttribute' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/store/sub_attributes.go backend/scout/internal/store/sub_attributes_test.go backend/scout/internal/handlers/sub_attributes.go backend/scout/internal/handlers/sub_attributes_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add sub-attribute store and handlers scoped to main attributes (F3)"
```

---

### Task 7: Session auth package (F1)

**Files:**
- Create: `backend/scout/internal/auth/session.go`
- Create: `backend/scout/internal/auth/session_test.go`

**Interfaces:**
- Produces: `auth.SessionDuration` (const, `24*time.Hour`), `auth.NewSessionToken(secret string, now time.Time) string`, `auth.ValidateSessionToken(secret, token string, now time.Time) bool`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/auth/session_test.go
package auth

import (
	"testing"
	"time"
)

func TestSessionTokenRoundTrip(t *testing.T) {
	now := time.Now()
	token := NewSessionToken("s3cr3t", now)

	if !ValidateSessionToken("s3cr3t", token, now.Add(1*time.Hour)) {
		t.Fatal("expected token to validate within session duration")
	}
	if ValidateSessionToken("s3cr3t", token, now.Add(SessionDuration+time.Minute)) {
		t.Fatal("expected token to be expired past session duration")
	}
	if ValidateSessionToken("wrong-secret", token, now) {
		t.Fatal("expected token to fail validation with wrong secret")
	}
	if ValidateSessionToken("s3cr3t", "garbage", now) {
		t.Fatal("expected garbage token to fail validation")
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/auth/... -v`
Expected: FAIL — `package auth: no Go files`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/auth/session.go
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const SessionDuration = 24 * time.Hour

func NewSessionToken(secret string, now time.Time) string {
	expiry := now.Add(SessionDuration).Unix()
	payload := strconv.FormatInt(expiry, 10)
	sig := sign(secret, payload)
	return base64.URLEncoding.EncodeToString([]byte(payload + "." + sig))
}

func ValidateSessionToken(secret, token string, now time.Time) bool {
	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(raw), ".", 2)
	if len(parts) != 2 {
		return false
	}
	payload, sig := parts[0], parts[1]

	if !hmac.Equal([]byte(sig), []byte(sign(secret, payload))) {
		return false
	}

	expiry, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return false
	}
	return now.Unix() < expiry
}

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/auth/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/auth
git commit -m "scout: add HMAC-signed session token package (F1)"
```

---

### Task 8: Login handler + RequireAuth middleware guarding every route (F1, NF1)

**Files:**
- Create: `backend/scout/internal/handlers/auth.go`
- Create: `backend/scout/internal/handlers/auth_test.go`
- Modify: `backend/scout/internal/config/config.go` (no change needed — `AdminPassword`/`SessionSecret`/`CookieSecure` already present from Task 1)
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `auth.NewSessionToken`, `auth.ValidateSessionToken`, `auth.SessionDuration`
- Produces: `handlers.NewAuthHandler(adminPassword, sessionSecret string, cookieSecure bool) *AuthHandler` with `.Login(c *gin.Context)`; `handlers.RequireAuth(sessionSecret string) gin.HandlerFunc`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/auth_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupAuthTestRouter() (*gin.Engine, *AuthHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler("correct-password", "test-secret", false)
	r.POST("/api/auth/login", h.Login)
	protected := r.Group("/api")
	protected.Use(RequireAuth("test-secret"))
	protected.GET("/whoami", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r, h
}

func TestLoginSuccessSetsCookie(t *testing.T) {
	r, _ := setupAuthTestRouter()

	body, _ := json.Marshal(map[string]string{"password": "correct-password"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(w.Result().Cookies()) == 0 {
		t.Fatal("expected a session cookie to be set")
	}
}

func TestLoginWrongPasswordRejected(t *testing.T) {
	r, _ := setupAuthTestRouter()

	body, _ := json.Marshal(map[string]string{"password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProtectedRouteRequiresSession(t *testing.T) {
	r, _ := setupAuthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session cookie, got %d", w.Code)
	}
}

func TestProtectedRouteAllowsValidSession(t *testing.T) {
	r, _ := setupAuthTestRouter()

	loginBody, _ := json.Marshal(map[string]string{"password": "correct-password"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	cookies := loginW.Result().Cookies()

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid session cookie, got %d: %s", w.Code, w.Body.String())
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run 'Login|ProtectedRoute' -v`
Expected: FAIL — `undefined: NewAuthHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/auth.go
package handlers

import (
	"net/http"
	"time"

	"scout/internal/auth"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "scout_session"

type AuthHandler struct {
	adminPassword string
	sessionSecret string
	cookieSecure  bool
}

func NewAuthHandler(adminPassword, sessionSecret string, cookieSecure bool) *AuthHandler {
	return &AuthHandler{adminPassword: adminPassword, sessionSecret: sessionSecret, cookieSecure: cookieSecure}
}

type loginRequest struct {
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	if req.Password != h.adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token := auth.NewSessionToken(h.sessionSecret, time.Now())
	c.SetCookie(sessionCookieName, token, int(auth.SessionDuration.Seconds()), "/", "", h.cookieSecure, true)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RequireAuth guards every route it is applied to — Scout has no public
// application-data route group (NF1), unlike oncarinho.
func RequireAuth(sessionSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(sessionCookieName)
		if err != nil || !auth.ValidateSessionToken(sessionSecret, cookie, time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
```
```go
// backend/scout/cmd/server/router.go — full file after this task
package main

import (
	"database/sql"

	"scout/internal/config"
	"scout/internal/handlers"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	engineerStore := store.NewEngineerStore(db)
	mainAttributeStore := store.NewMainAttributeStore(db)
	subAttributeStore := store.NewSubAttributeStore(db)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg.AdminPassword, cfg.SessionSecret, cfg.CookieSecure)
	engineerHandler := handlers.NewEngineerHandler(engineerStore)
	mainAttributeHandler := handlers.NewMainAttributeHandler(mainAttributeStore)
	subAttributeHandler := handlers.NewSubAttributeHandler(subAttributeStore, mainAttributeStore)

	r := gin.Default()

	// Exempt from auth: infra liveness probe and the login endpoint itself
	// (which by definition cannot require a session). Every other route
	// below sits behind RequireAuth — there is no public application-data
	// group, per NF1.
	r.GET("/health", healthHandler.HealthCheck)
	r.POST("/api/auth/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(handlers.RequireAuth(cfg.SessionSecret))
	{
		api.GET("/engineers", engineerHandler.List)
		api.GET("/engineers/:id", engineerHandler.Get)
		api.POST("/engineers", engineerHandler.Create)
		api.PUT("/engineers/:id", engineerHandler.Update)
		api.DELETE("/engineers/:id", engineerHandler.Deactivate)
		api.POST("/engineers/:id/reactivate", engineerHandler.Reactivate)

		api.GET("/main-attributes", mainAttributeHandler.List)
		api.POST("/main-attributes", mainAttributeHandler.Create)
		api.PUT("/main-attributes/:id", mainAttributeHandler.Update)

		api.GET("/sub-attributes", subAttributeHandler.List)
		api.POST("/sub-attributes", subAttributeHandler.Create)
		api.PUT("/sub-attributes/:id", subAttributeHandler.Update)
		api.DELETE("/sub-attributes/:id", subAttributeHandler.Deactivate)
	}

	return r
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/auth.go backend/scout/internal/handlers/auth_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add login + RequireAuth middleware guarding every route (F1, NF1)"
```

---

### Task 9: Pure rank→score linear interpolation function, isolated (F7)

**Files:**
- Create: `backend/scout/internal/scoring/scoring.go`
- Create: `backend/scout/internal/scoring/scoring_test.go`

**Interfaces:**
- Produces: `scoring.RankToScore(rank, n int) float64`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/scoring/scoring_test.go
package scoring

import "testing"

func TestRankToScore(t *testing.T) {
	cases := []struct {
		name string
		rank int
		n    int
		want float64
	}{
		{"rank 1 of 5 is 100", 1, 5, 100},
		{"last rank of 5 is 50", 5, 5, 50},
		{"middle rank of 5", 3, 5, 75},
		{"only engineer scores 100", 1, 1, 100},
		{"rank 1 of 11 is 100", 1, 11, 100},
		{"rank 11 of 11 is 50", 11, 11, 50},
		{"rank 6 of 11 is midpoint", 6, 11, 75},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RankToScore(tc.rank, tc.n)
			if got != tc.want {
				t.Fatalf("RankToScore(%d, %d) = %v, want %v", tc.rank, tc.n, got, tc.want)
			}
		})
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/scoring/... -run TestRankToScore -v`
Expected: FAIL — `package scoring: no Go files`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/scoring/scoring.go
package scoring

// RankToScore converts a 1..N rank into a 50-100 score via linear
// interpolation: rank 1 -> 100, rank N -> 50, evenly spaced (F7).
// N == 1 is a special case (no interpolation possible) and scores 100.
func RankToScore(rank, n int) float64 {
	if n <= 1 {
		return 100
	}
	return 100 - float64(rank-1)*50/float64(n-1)
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/scoring/... -run TestRankToScore -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/scoring/scoring.go backend/scout/internal/scoring/scoring_test.go
git commit -m "scout: add pure rank-to-score linear interpolation function (F7)"
```

---

### Task 10: Pure strict 1..N permutation validation, isolated (F6)

**Files:**
- Create: `backend/scout/internal/scoring/permutation.go`
- Create: `backend/scout/internal/scoring/permutation_test.go`

**Interfaces:**
- Consumes: nothing new
- Produces: `scoring.RankEntry{EngineerID int, Rank int}`, `scoring.ValidatePermutation(entries []RankEntry, activeEngineerIDs []int) error`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/scoring/permutation_test.go
package scoring

import "testing"

func TestValidatePermutationValid(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 2}, {EngineerID: 2, Rank: 1}, {EngineerID: 3, Rank: 3}}
	active := []int{1, 2, 3}

	if err := ValidatePermutation(entries, active); err != nil {
		t.Fatalf("expected valid permutation, got error: %v", err)
	}
}

func TestValidatePermutationRejectsTie(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 2, Rank: 1}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for duplicate rank (tie)")
	}
}

func TestValidatePermutationRejectsGap(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 2, Rank: 3}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for gap in ranks (1, 3 with only 2 engineers)")
	}
}

func TestValidatePermutationRejectsMissingEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error when submission omits an active engineer")
	}
}

func TestValidatePermutationRejectsUnknownEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 99, Rank: 2}}
	active := []int{1, 2}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for an engineer not in the active roster")
	}
}

func TestValidatePermutationRejectsDuplicateEngineer(t *testing.T) {
	entries := []RankEntry{{EngineerID: 1, Rank: 1}, {EngineerID: 1, Rank: 2}}
	active := []int{1}

	if err := ValidatePermutation(entries, active); err == nil {
		t.Fatal("expected error for duplicate engineer entry")
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/scoring/... -run TestValidatePermutation -v`
Expected: FAIL — `undefined: ValidatePermutation`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/scoring/permutation.go
package scoring

import "fmt"

type RankEntry struct {
	EngineerID int
	Rank       int
}

// ValidatePermutation enforces F6: a ranking submission must be a strict
// 1..N permutation of exactly the active roster — no ties, no gaps, no
// duplicates, no missing or extra engineers.
func ValidatePermutation(entries []RankEntry, activeEngineerIDs []int) error {
	if len(entries) != len(activeEngineerIDs) {
		return fmt.Errorf("expected exactly %d ranked engineers (the active roster), got %d", len(activeEngineerIDs), len(entries))
	}

	activeSet := make(map[int]bool, len(activeEngineerIDs))
	for _, id := range activeEngineerIDs {
		activeSet[id] = true
	}

	seenEngineers := make(map[int]bool, len(entries))
	seenRanks := make(map[int]bool, len(entries))
	for _, e := range entries {
		if !activeSet[e.EngineerID] {
			return fmt.Errorf("engineer %d is not in the active roster for this cycle", e.EngineerID)
		}
		if seenEngineers[e.EngineerID] {
			return fmt.Errorf("engineer %d appears more than once in the submission", e.EngineerID)
		}
		seenEngineers[e.EngineerID] = true

		if e.Rank < 1 || e.Rank > len(entries) {
			return fmt.Errorf("rank %d is out of range 1..%d", e.Rank, len(entries))
		}
		if seenRanks[e.Rank] {
			return fmt.Errorf("rank %d is used more than once (ties are not allowed)", e.Rank)
		}
		seenRanks[e.Rank] = true
	}

	for rank := 1; rank <= len(entries); rank++ {
		if !seenRanks[rank] {
			return fmt.Errorf("rank %d is missing (ranks must be a contiguous 1..%d sequence)", rank, len(entries))
		}
	}

	return nil
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/scoring/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/scoring/permutation.go backend/scout/internal/scoring/permutation_test.go
git commit -m "scout: add pure strict 1..N ranking permutation validation (F6)"
```

---

### Task 11: Rating cycles migration, model, store, handlers (F6 — create/list cycle)

**Files:**
- Create: `backend/scout/migrations/000003_create_rating_cycles_and_rankings.up.sql`
- Create: `backend/scout/migrations/000003_create_rating_cycles_and_rankings.down.sql`
- Create: `backend/scout/internal/models/cycle.go`
- Create: `backend/scout/internal/store/cycles.go`
- Create: `backend/scout/internal/store/cycles_test.go`
- Create: `backend/scout/internal/handlers/cycles.go`
- Create: `backend/scout/internal/handlers/cycles_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Produces: `models.RatingCycle{ID int, PeriodStart time.Time, PeriodEnd time.Time, CreatedAt time.Time}`; `models.SubAttributeRanking{ID int, CycleID int, SubAttributeID int, EngineerID int, Rank int, Score float64}`; `store.CycleStore` with `NewCycleStore(db *sql.DB) *CycleStore`, `.List() ([]models.RatingCycle, error)`, `.Create(periodStart, periodEnd time.Time) (*models.RatingCycle, error)`, `.GetByID(id int) (*models.RatingCycle, error)`, `.Exists(id int) (bool, error)`; `handlers.NewCycleHandler(s *store.CycleStore) *CycleHandler` with `.List`, `.Create`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/cycles_test.go
package store

import (
	"testing"
	"time"
)

func TestCycleStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewCycleStore(db)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	created, err := s.Create(start, end)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !created.PeriodStart.Equal(start) || !created.PeriodEnd.Equal(end) {
		t.Fatalf("unexpected created cycle: %+v", created)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one cycle, got %+v", list)
	}
}
```
```go
// backend/scout/internal/handlers/cycles_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCycleHandlerCreate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)
	r.GET("/api/cycles", h.List)

	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-01", "period_end": "2026-01-14"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'Cycle' -v`
Expected: FAIL — `undefined: NewCycleStore`
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000003_create_rating_cycles_and_rankings.up.sql
CREATE TABLE IF NOT EXISTS rating_cycles (
    id SERIAL PRIMARY KEY,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sub_attribute_rankings (
    id SERIAL PRIMARY KEY,
    cycle_id INTEGER NOT NULL REFERENCES rating_cycles(id) ON DELETE CASCADE,
    sub_attribute_id INTEGER NOT NULL REFERENCES sub_attributes(id) ON DELETE CASCADE,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL,
    score NUMERIC(5,2) NOT NULL,
    UNIQUE (cycle_id, sub_attribute_id, engineer_id),
    UNIQUE (cycle_id, sub_attribute_id, rank)
);

CREATE INDEX idx_sub_attribute_rankings_cycle_id ON sub_attribute_rankings(cycle_id);
CREATE INDEX idx_sub_attribute_rankings_engineer_id ON sub_attribute_rankings(engineer_id);
```
```sql
-- backend/scout/migrations/000003_create_rating_cycles_and_rankings.down.sql
DROP TABLE IF EXISTS sub_attribute_rankings;
DROP TABLE IF EXISTS rating_cycles;
```
```go
// backend/scout/internal/models/cycle.go
package models

import "time"

type RatingCycle struct {
	ID          int       `json:"id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
}

type SubAttributeRanking struct {
	ID             int     `json:"id"`
	CycleID        int     `json:"cycle_id"`
	SubAttributeID int     `json:"sub_attribute_id"`
	EngineerID     int     `json:"engineer_id"`
	Rank           int     `json:"rank"`
	Score          float64 `json:"score"`
}
```
```go
// backend/scout/internal/store/cycles.go
package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type CycleStore struct {
	db *sql.DB
}

func NewCycleStore(db *sql.DB) *CycleStore {
	return &CycleStore{db: db}
}

func (s *CycleStore) List() ([]models.RatingCycle, error) {
	rows, err := s.db.Query("SELECT id, period_start, period_end, created_at FROM rating_cycles ORDER BY period_start DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cycles := []models.RatingCycle{}
	for rows.Next() {
		var c models.RatingCycle
		if err := rows.Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt); err != nil {
			return nil, err
		}
		cycles = append(cycles, c)
	}
	return cycles, rows.Err()
}

func (s *CycleStore) Create(periodStart, periodEnd time.Time) (*models.RatingCycle, error) {
	var c models.RatingCycle
	err := s.db.QueryRow(
		`INSERT INTO rating_cycles (period_start, period_end) VALUES ($1, $2)
		 RETURNING id, period_start, period_end, created_at`,
		periodStart, periodEnd,
	).Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CycleStore) GetByID(id int) (*models.RatingCycle, error) {
	var c models.RatingCycle
	err := s.db.QueryRow("SELECT id, period_start, period_end, created_at FROM rating_cycles WHERE id = $1", id).
		Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CycleStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM rating_cycles WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
```
```go
// backend/scout/internal/handlers/cycles.go
package handlers

import (
	"net/http"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type CycleHandler struct {
	store *store.CycleStore
}

func NewCycleHandler(s *store.CycleStore) *CycleHandler {
	return &CycleHandler{store: s}
}

type cycleRequest struct {
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func (h *CycleHandler) List(c *gin.Context) {
	cycles, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cycles"})
		return
	}
	c.JSON(http.StatusOK, cycles)
}

func (h *CycleHandler) Create(c *gin.Context) {
	var req cycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_start and period_end are required"})
		return
	}
	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_start must be YYYY-MM-DD"})
		return
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be YYYY-MM-DD"})
		return
	}
	if !end.After(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be after period_start"})
		return
	}

	cycle, err := h.store.Create(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cycle"})
		return
	}
	c.JSON(http.StatusCreated, cycle)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		cycleStore := store.NewCycleStore(db)
		cycleHandler := handlers.NewCycleHandler(cycleStore)
		api.GET("/cycles", cycleHandler.List)
		api.POST("/cycles", cycleHandler.Create)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... ./internal/handlers/... -run 'Cycle' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations/000003_create_rating_cycles_and_rankings.up.sql backend/scout/migrations/000003_create_rating_cycles_and_rankings.down.sql backend/scout/internal/models/cycle.go backend/scout/internal/store/cycles.go backend/scout/internal/store/cycles_test.go backend/scout/internal/handlers/cycles.go backend/scout/internal/handlers/cycles_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add rating cycles table, store, and create/list endpoints (F6)"
```

---

### Task 12: Ranking submission store + handler (F6, F7)

**Files:**
- Create: `backend/scout/internal/store/rankings.go`
- Create: `backend/scout/internal/store/rankings_test.go`
- Create: `backend/scout/internal/handlers/rankings.go`
- Create: `backend/scout/internal/handlers/rankings_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `scoring.RankEntry`, `scoring.ValidatePermutation`, `scoring.RankToScore` (Tasks 9, 10); `models.SubAttributeRanking` (Task 11); `store.EngineerStore.ListActiveIDs` (Task 2); `store.CycleStore.Exists`, `store.SubAttributeStore.Exists`
- Produces: `store.RankingStore` with `NewRankingStore(db *sql.DB, engineerStore *store.EngineerStore) *RankingStore`, `.SubmitRanking(cycleID, subAttributeID int, entries []scoring.RankEntry) ([]models.SubAttributeRanking, error)`, `.GetByCycleAndSubAttribute(cycleID, subAttributeID int) ([]models.SubAttributeRanking, error)`; `handlers.NewRankingHandler(s *store.RankingStore, cycleStore *store.CycleStore, subAttrStore *store.SubAttributeStore) *RankingHandler` with `.Submit(c *gin.Context)` registered as `PUT /api/cycles/:id/sub-attributes/:subId/ranking` and `.Get(c *gin.Context)` registered as `GET /api/cycles/:id/sub-attributes/:subId/ranking`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/rankings_test.go
package store

import (
	"testing"
	"time"

	"scout/internal/scoring"
)

func TestRankingStoreSubmitRankingValidPermutation(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main", "Test Main")
	sub, _ := subStore.Create(main.ID, "Code Quality", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	rankings, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	})
	if err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}
	if len(rankings) != 2 {
		t.Fatalf("expected 2 persisted rankings, got %d", len(rankings))
	}
	for _, r := range rankings {
		if r.EngineerID == e1.ID && r.Score != 100 {
			t.Fatalf("expected rank-1 engineer to score 100, got %v", r.Score)
		}
		if r.EngineerID == e2.ID && r.Score != 50 {
			t.Fatalf("expected rank-2-of-2 engineer to score 50, got %v", r.Score)
		}
	}
}

func TestRankingStoreSubmitRankingRejectsInvalidPermutation(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main2", "Test Main 2")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	_, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 1},
	})
	if err == nil {
		t.Fatal("expected error for tied ranks")
	}
}
```
```go
// backend/scout/internal/handlers/rankings_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestRankingHandlerSubmit(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main3", "Test Main 3")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/" + itoa(cycle.ID) + "/sub-attributes/" + itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func itoa(n int) string {
	return fmtInt(n)
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'Ranking' -v`
Expected: FAIL — `undefined: NewRankingStore`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/store/rankings.go
package store

import (
	"database/sql"

	"scout/internal/models"
	"scout/internal/scoring"
)

type RankingStore struct {
	db            *sql.DB
	engineerStore *EngineerStore
}

func NewRankingStore(db *sql.DB, engineerStore *EngineerStore) *RankingStore {
	return &RankingStore{db: db, engineerStore: engineerStore}
}

// SubmitRanking validates entries as a strict 1..N permutation of the active
// roster (F6), converts each rank to a score via linear interpolation (F7),
// and upserts all rows for this cycle+sub-attribute in a single transaction.
func (s *RankingStore) SubmitRanking(cycleID, subAttributeID int, entries []scoring.RankEntry) ([]models.SubAttributeRanking, error) {
	activeIDs, err := s.engineerStore.ListActiveIDs()
	if err != nil {
		return nil, err
	}
	if err := scoring.ValidatePermutation(entries, activeIDs); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"DELETE FROM sub_attribute_rankings WHERE cycle_id = $1 AND sub_attribute_id = $2",
		cycleID, subAttributeID,
	); err != nil {
		return nil, err
	}

	n := len(entries)
	stmt, err := tx.Prepare(
		`INSERT INTO sub_attribute_rankings (cycle_id, sub_attribute_id, engineer_id, rank, score)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, cycle_id, sub_attribute_id, engineer_id, rank, score`,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	results := make([]models.SubAttributeRanking, 0, n)
	for _, e := range entries {
		var r models.SubAttributeRanking
		score := scoring.RankToScore(e.Rank, n)
		if err := stmt.QueryRow(cycleID, subAttributeID, e.EngineerID, e.Rank, score).
			Scan(&r.ID, &r.CycleID, &r.SubAttributeID, &r.EngineerID, &r.Rank, &r.Score); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *RankingStore) GetByCycleAndSubAttribute(cycleID, subAttributeID int) ([]models.SubAttributeRanking, error) {
	rows, err := s.db.Query(
		`SELECT id, cycle_id, sub_attribute_id, engineer_id, rank, score
		 FROM sub_attribute_rankings WHERE cycle_id = $1 AND sub_attribute_id = $2 ORDER BY rank`,
		cycleID, subAttributeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []models.SubAttributeRanking{}
	for rows.Next() {
		var r models.SubAttributeRanking
		if err := rows.Scan(&r.ID, &r.CycleID, &r.SubAttributeID, &r.EngineerID, &r.Rank, &r.Score); err != nil {
			return nil, err
		}
		rankings = append(rankings, r)
	}
	return rankings, rows.Err()
}
```
```go
// backend/scout/internal/handlers/rankings.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type RankingHandler struct {
	store        *store.RankingStore
	cycleStore   *store.CycleStore
	subAttrStore *store.SubAttributeStore
}

func NewRankingHandler(s *store.RankingStore, cycleStore *store.CycleStore, subAttrStore *store.SubAttributeStore) *RankingHandler {
	return &RankingHandler{store: s, cycleStore: cycleStore, subAttrStore: subAttrStore}
}

type rankingEntryRequest struct {
	EngineerID int `json:"engineer_id" binding:"required"`
	Rank       int `json:"rank" binding:"required"`
}

type submitRankingRequest struct {
	Rankings []rankingEntryRequest `json:"rankings" binding:"required,dive"`
}

// Submit handles PUT /api/cycles/:id/sub-attributes/:subId/ranking (F6).
// Past cycles remain editable by re-submission — Scout does not lock a
// cycle's rankings once saved (see the plan's judgment-calls section).
func (h *RankingHandler) Submit(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	subAttributeID, err := strconv.Atoi(c.Param("subId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}

	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}
	if exists, err := h.subAttrStore.Exists(subAttributeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up sub attribute"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return
	}

	var req submitRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings are required"})
		return
	}

	entries := make([]scoring.RankEntry, 0, len(req.Rankings))
	for _, r := range req.Rankings {
		entries = append(entries, scoring.RankEntry{EngineerID: r.EngineerID, Rank: r.Rank})
	}

	rankings, err := h.store.SubmitRanking(cycleID, subAttributeID, entries)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rankings)
}

// Get handles GET /api/cycles/:id/sub-attributes/:subId/ranking — reads
// back whatever has been saved so far for this cycle+sub-attribute (an
// empty array if nothing has been submitted yet). The ranking UI (Task 35)
// uses this to pre-populate rank inputs when reopening a cycle.
func (h *RankingHandler) Get(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	subAttributeID, err := strconv.Atoi(c.Param("subId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}

	rankings, err := h.store.GetByCycleAndSubAttribute(cycleID, subAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ranking"})
		return
	}
	c.JSON(http.StatusOK, rankings)
}

func fmtInt(n int) string {
	return fmt.Sprintf("%d", n)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		rankingStore := store.NewRankingStore(db, engineerStore)
		rankingHandler := handlers.NewRankingHandler(rankingStore, cycleStore, subAttributeStore)
		api.PUT("/cycles/:id/sub-attributes/:subId/ranking", rankingHandler.Submit)
		api.GET("/cycles/:id/sub-attributes/:subId/ranking", rankingHandler.Get)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'Ranking' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/store/rankings.go backend/scout/internal/store/rankings_test.go backend/scout/internal/handlers/rankings.go backend/scout/internal/handlers/rankings_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add ranking submission store and endpoint with strict permutation validation (F6, F7)"
```

---

### Task 13: Computed-on-read score aggregation — main-attribute and Overall scores (F8, NF2)

**Files:**
- Create: `backend/scout/internal/models/score.go`
- Create: `backend/scout/internal/store/scores.go`
- Create: `backend/scout/internal/store/scores_test.go`

**Interfaces:**
- Consumes: `sub_attribute_rankings`/`main_attributes`/`sub_attributes`/`rating_cycles` tables (Tasks 4, 11, 12)
- Produces: `models.MainAttributeScore{MainAttributeID int, Key string, Name string, Score float64}`; `store.ScoreStore` with `NewScoreStore(db *sql.DB) *ScoreStore`, `.MainAttributeScores(engineerID, cycleID int) ([]models.MainAttributeScore, error)`, `.OverallScore(engineerID, cycleID int) (*float64, error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/scores_test.go
package store

import (
	"testing"
	"time"

	"scout/internal/scoring"
)

func TestScoreStoreMainAttributeScoresAndOverall(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_score", "Test Main Score")
	sub1, _ := subStore.Create(main.ID, "Code Quality", nil)
	sub2, _ := subStore.Create(main.ID, "Testing", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Alex ranks 1st on both sub-attributes -> both score 100 -> main attribute avg 100.
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub1.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub2.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub2 ranking failed: %v", err)
	}

	scores, err := scoreStore.MainAttributeScores(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("main attribute scores failed: %v", err)
	}
	if len(scores) != 1 || scores[0].Score != 100 {
		t.Fatalf("expected one main attribute scoring 100 for e1, got %+v", scores)
	}

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	if overall == nil || *overall != 100 {
		t.Fatalf("expected overall score 100 for e1, got %v", overall)
	}

	e2Overall, err := scoreStore.OverallScore(e2.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score for e2 failed: %v", err)
	}
	if e2Overall == nil || *e2Overall != 50 {
		t.Fatalf("expected overall score 50 for e2 (rank 2 of 2 on both sub-attributes), got %v", e2Overall)
	}
}

func TestScoreStoreOverallScoreNilWhenNoData(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	cycleStore := NewCycleStore(db)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("NoData", nil, nil, nil, time.Now())
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	if overall != nil {
		t.Fatalf("expected nil overall score when engineer has no rankings for the cycle, got %v", *overall)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestScoreStore -v`
Expected: FAIL — `undefined: NewScoreStore`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/models/score.go
package models

import "time"

type MainAttributeScore struct {
	MainAttributeID int     `json:"main_attribute_id"`
	Key             string  `json:"key"`
	Name            string  `json:"name"`
	Score           float64 `json:"score"`
}

type EngineerCard struct {
	Engineer       Engineer             `json:"engineer"`
	CycleID        int                  `json:"cycle_id"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type TrendPoint struct {
	CycleID        int                  `json:"cycle_id"`
	PeriodStart    time.Time            `json:"period_start"`
	PeriodEnd      time.Time            `json:"period_end"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type EngineerCycleScore struct {
	Engineer       Engineer             `json:"engineer"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type RosterEntry struct {
	Engineer      Engineer   `json:"engineer"`
	LatestOverall *float64   `json:"latest_overall"`
	LastCycleDate *time.Time `json:"last_cycle_date"`
}
```
```go
// backend/scout/internal/store/scores.go
package store

import (
	"database/sql"

	"scout/internal/models"
)

type ScoreStore struct {
	db *sql.DB
}

func NewScoreStore(db *sql.DB) *ScoreStore {
	return &ScoreStore{db: db}
}

// MainAttributeScores computes each main attribute's score for an engineer
// in a cycle as the average of that main attribute's sub-attribute scores
// (F8). Computed on read from sub_attribute_rankings — never cached (NF2).
func (s *ScoreStore) MainAttributeScores(engineerID, cycleID int) ([]models.MainAttributeScore, error) {
	rows, err := s.db.Query(
		`SELECT ma.id, ma.key, ma.name, AVG(sar.score)
		 FROM sub_attribute_rankings sar
		 JOIN sub_attributes sa ON sa.id = sar.sub_attribute_id
		 JOIN main_attributes ma ON ma.id = sa.main_attribute_id
		 WHERE sar.cycle_id = $1 AND sar.engineer_id = $2
		 GROUP BY ma.id, ma.key, ma.name
		 ORDER BY ma.id`,
		cycleID, engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := []models.MainAttributeScore{}
	for rows.Next() {
		var m models.MainAttributeScore
		if err := rows.Scan(&m.MainAttributeID, &m.Key, &m.Name, &m.Score); err != nil {
			return nil, err
		}
		scores = append(scores, m)
	}
	return scores, rows.Err()
}

// OverallScore computes the average of the main-attribute scores that
// existed as of this cycle (F8): a main attribute counts toward Overall
// only if it was created at or before the cycle was opened, so adding a
// main attribute later never retroactively changes a past cycle's Overall.
// Returns nil if the engineer has no rankings for the cycle at all.
func (s *ScoreStore) OverallScore(engineerID, cycleID int) (*float64, error) {
	var overall sql.NullFloat64
	err := s.db.QueryRow(
		`SELECT AVG(ma_scores.score) FROM (
			SELECT ma.id, AVG(sar.score) AS score
			FROM sub_attribute_rankings sar
			JOIN sub_attributes sa ON sa.id = sar.sub_attribute_id
			JOIN main_attributes ma ON ma.id = sa.main_attribute_id
			JOIN rating_cycles rc ON rc.id = sar.cycle_id
			WHERE sar.cycle_id = $1 AND sar.engineer_id = $2 AND ma.created_at <= rc.created_at
			GROUP BY ma.id
		 ) ma_scores`,
		cycleID, engineerID,
	).Scan(&overall)
	if err != nil {
		return nil, err
	}
	if !overall.Valid {
		return nil, nil
	}
	return &overall.Float64, nil
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... -run TestScoreStore -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/models/score.go backend/scout/internal/store/scores.go backend/scout/internal/store/scores_test.go
git commit -m "scout: add computed-on-read main-attribute and Overall score aggregation (F8, NF2)"
```

---

### Task 14: Engineer card + trend endpoints (F10)

**Files:**
- Modify: `backend/scout/internal/store/scores.go`
- Modify: `backend/scout/internal/store/scores_test.go`
- Create: `backend/scout/internal/handlers/engineer_card.go`
- Create: `backend/scout/internal/handlers/engineer_card_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.ScoreStore.MainAttributeScores`, `store.ScoreStore.OverallScore` (Task 13); `store.EngineerStore.GetByID` (Task 2); `models.EngineerCard`, `models.TrendPoint` (Task 13)
- Produces: `store.ScoreStore.EngineerCard(engineerStore *store.EngineerStore, engineerID, cycleID int) (*models.EngineerCard, error)`, `store.ScoreStore.EngineerTrend(engineerStore *store.EngineerStore, engineerID int) ([]models.TrendPoint, error)`; `handlers.NewEngineerCardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *EngineerCardHandler` with `.Card(c *gin.Context)` (`GET /api/engineers/:id/card?cycleId=`), `.Trend(c *gin.Context)` (`GET /api/engineers/:id/trend`)

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/scores_test.go — append
func TestScoreStoreEngineerCardAndTrend(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_card", "Test Main Card")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	card, err := scoreStore.EngineerCard(engineerStore, e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("engineer card failed: %v", err)
	}
	if card.Engineer.ID != e1.ID || card.Overall == nil || *card.Overall != 100 {
		t.Fatalf("unexpected card: %+v", card)
	}

	trend, err := scoreStore.EngineerTrend(engineerStore, e1.ID)
	if err != nil {
		t.Fatalf("engineer trend failed: %v", err)
	}
	if len(trend) != 1 || trend[0].CycleID != cycle.ID {
		t.Fatalf("expected one trend point for the one cycle, got %+v", trend)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestScoreStoreEngineerCard -v`
Expected: FAIL — `undefined: (*ScoreStore).EngineerCard`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/store/scores.go — top of file, add to imports
import (
	"database/sql"
	"time"

	"scout/internal/models"
)
```
```go
// backend/scout/internal/store/scores.go — append
// EngineerCard returns the engineer's Overall + main-attribute scores for
// one cycle (F10).
func (s *ScoreStore) EngineerCard(engineerStore *EngineerStore, engineerID, cycleID int) (*models.EngineerCard, error) {
	engineer, err := engineerStore.GetByID(engineerID)
	if err != nil {
		return nil, err
	}
	if engineer == nil {
		return nil, nil
	}

	mainScores, err := s.MainAttributeScores(engineerID, cycleID)
	if err != nil {
		return nil, err
	}
	overall, err := s.OverallScore(engineerID, cycleID)
	if err != nil {
		return nil, err
	}

	return &models.EngineerCard{
		Engineer:       *engineer,
		CycleID:        cycleID,
		Overall:        overall,
		MainAttributes: mainScores,
	}, nil
}

// EngineerTrend returns the engineer's scores across every past cycle they
// have at least one ranking in, oldest first (F10).
func (s *ScoreStore) EngineerTrend(engineerStore *EngineerStore, engineerID int) ([]models.TrendPoint, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT rc.id, rc.period_start, rc.period_end
		 FROM rating_cycles rc
		 JOIN sub_attribute_rankings sar ON sar.cycle_id = rc.id
		 WHERE sar.engineer_id = $1
		 ORDER BY rc.period_start`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type cycleRow struct {
		id          int
		periodStart time.Time
		periodEnd   time.Time
	}
	cycles := []cycleRow{}
	for rows.Next() {
		var cr cycleRow
		if err := rows.Scan(&cr.id, &cr.periodStart, &cr.periodEnd); err != nil {
			return nil, err
		}
		cycles = append(cycles, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	trend := make([]models.TrendPoint, 0, len(cycles))
	for _, cr := range cycles {
		mainScores, err := s.MainAttributeScores(engineerID, cr.id)
		if err != nil {
			return nil, err
		}
		overall, err := s.OverallScore(engineerID, cr.id)
		if err != nil {
			return nil, err
		}
		trend = append(trend, models.TrendPoint{
			CycleID:        cr.id,
			PeriodStart:    cr.periodStart,
			PeriodEnd:      cr.periodEnd,
			Overall:        overall,
			MainAttributes: mainScores,
		})
	}
	return trend, nil
}
```
```go
// backend/scout/internal/handlers/engineer_card.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type EngineerCardHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
}

func NewEngineerCardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *EngineerCardHandler {
	return &EngineerCardHandler{scoreStore: s, engineerStore: engineerStore}
}

func (h *EngineerCardHandler) Card(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	cycleID, err := strconv.Atoi(c.Query("cycleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cycleId query param is required"})
		return
	}

	card, err := h.scoreStore.EngineerCard(h.engineerStore, engineerID, cycleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute engineer card"})
		return
	}
	if card == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *EngineerCardHandler) Trend(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}

	trend, err := h.scoreStore.EngineerTrend(h.engineerStore, engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute engineer trend"})
		return
	}
	c.JSON(http.StatusOK, trend)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		scoreStore := store.NewScoreStore(db)
		engineerCardHandler := handlers.NewEngineerCardHandler(scoreStore, engineerStore)
		api.GET("/engineers/:id/card", engineerCardHandler.Card)
		api.GET("/engineers/:id/trend", engineerCardHandler.Trend)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'EngineerCard|EngineerTrend' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/store/scores.go backend/scout/internal/store/scores_test.go backend/scout/internal/handlers/engineer_card.go backend/scout/internal/handlers/engineer_card_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add engineer card and trend endpoints (F10)"
```

---

### Task 15: Cycle view endpoint — all engineers' scores for one cycle (F15)

**Files:**
- Modify: `backend/scout/internal/store/scores.go`
- Modify: `backend/scout/internal/store/scores_test.go`
- Create: `backend/scout/internal/handlers/cycle_view.go`
- Create: `backend/scout/internal/handlers/cycle_view_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.ScoreStore.MainAttributeScores`, `.OverallScore` (Task 13); `store.EngineerStore.GetByID` (Task 2); `models.EngineerCycleScore` (Task 13)
- Produces: `store.ScoreStore.CycleScores(engineerStore *store.EngineerStore, cycleID int) ([]models.EngineerCycleScore, error)`; `handlers.NewCycleViewHandler(s *store.ScoreStore, engineerStore *store.EngineerStore, cycleStore *store.CycleStore) *CycleViewHandler` with `.Get(c *gin.Context)` registered as `GET /api/cycles/:id/scores`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/scores_test.go — append
func TestScoreStoreCycleScores(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_cv", "Test Main CV")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}})

	scores, err := scoreStore.CycleScores(engineerStore, cycle.ID)
	if err != nil {
		t.Fatalf("cycle scores failed: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected 2 engineers in cycle view, got %+v", scores)
	}
}
```
```go
// backend/scout/internal/handlers/cycle_view_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCycleViewHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_cv_handler", "Test Main CV Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	r.GET("/api/cycles/:id/scores", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/cycles/"+itoa(cycle.ID)+"/scores", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCycleViewHandlerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	cycleStore := store.NewCycleStore(db)
	scoreStore := store.NewScoreStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	r.GET("/api/cycles/:id/scores", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/cycles/999999/scores", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'ScoreStoreCycleScores|CycleViewHandler' -v`
Expected: FAIL — `undefined: (*ScoreStore).CycleScores`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/store/scores.go — append
// CycleScores returns every engineer who has at least one ranking in this
// cycle, each with their Overall + main-attribute scores as of that cycle
// (F15) — so the team can be compared at a single point in time.
func (s *ScoreStore) CycleScores(engineerStore *EngineerStore, cycleID int) ([]models.EngineerCycleScore, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT e.id, e.name
		 FROM engineers e
		 JOIN sub_attribute_rankings sar ON sar.engineer_id = e.id
		 WHERE sar.cycle_id = $1
		 ORDER BY e.name`,
		cycleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	engineerIDs := []int{}
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		engineerIDs = append(engineerIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	results := make([]models.EngineerCycleScore, 0, len(engineerIDs))
	for _, id := range engineerIDs {
		engineer, err := engineerStore.GetByID(id)
		if err != nil {
			return nil, err
		}
		mainScores, err := s.MainAttributeScores(id, cycleID)
		if err != nil {
			return nil, err
		}
		overall, err := s.OverallScore(id, cycleID)
		if err != nil {
			return nil, err
		}
		results = append(results, models.EngineerCycleScore{
			Engineer:       *engineer,
			Overall:        overall,
			MainAttributes: mainScores,
		})
	}
	return results, nil
}
```
```go
// backend/scout/internal/handlers/cycle_view.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type CycleViewHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
	cycleStore    *store.CycleStore
}

func NewCycleViewHandler(s *store.ScoreStore, engineerStore *store.EngineerStore, cycleStore *store.CycleStore) *CycleViewHandler {
	return &CycleViewHandler{scoreStore: s, engineerStore: engineerStore, cycleStore: cycleStore}
}

func (h *CycleViewHandler) Get(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	scores, err := h.scoreStore.CycleScores(h.engineerStore, cycleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute cycle scores"})
		return
	}
	c.JSON(http.StatusOK, scores)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		cycleViewHandler := handlers.NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
		api.GET("/cycles/:id/scores", cycleViewHandler.Get)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'CycleScores|CycleView' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/store/scores.go backend/scout/internal/store/scores_test.go backend/scout/internal/handlers/cycle_view.go backend/scout/internal/handlers/cycle_view_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add cycle view endpoint listing all engineers' scores for a cycle (F15)"
```

---

### Task 16: Roster dashboard endpoint (F11)

**Files:**
- Modify: `backend/scout/internal/store/scores.go`
- Modify: `backend/scout/internal/store/scores_test.go`
- Create: `backend/scout/internal/handlers/dashboard.go`
- Create: `backend/scout/internal/handlers/dashboard_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.EngineerStore.List` (Task 2); `store.ScoreStore.OverallScore` (Task 13); `models.RosterEntry` (Task 13)
- Produces: `store.ScoreStore.RosterDashboard(engineerStore *store.EngineerStore) ([]models.RosterEntry, error)`; `handlers.NewDashboardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *DashboardHandler` with `.Get(c *gin.Context)` registered as `GET /api/dashboard`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/scores_test.go — append
func TestScoreStoreRosterDashboard(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_dash", "Test Main Dash")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	dashboard, err := scoreStore.RosterDashboard(engineerStore)
	if err != nil {
		t.Fatalf("roster dashboard failed: %v", err)
	}
	if len(dashboard) != 1 {
		t.Fatalf("expected one active engineer on the dashboard, got %+v", dashboard)
	}
	if dashboard[0].LatestOverall == nil || *dashboard[0].LatestOverall != 100 {
		t.Fatalf("expected latest overall 100, got %+v", dashboard[0])
	}
	if dashboard[0].LastCycleDate == nil {
		t.Fatal("expected a last cycle date")
	}
}
```
```go
// backend/scout/internal/handlers/dashboard_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestDashboardHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	scoreStore := store.NewScoreStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(scoreStore, engineerStore)
	r.GET("/api/dashboard", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'RosterDashboard|DashboardHandler' -v`
Expected: FAIL — `undefined: (*ScoreStore).RosterDashboard`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/store/scores.go — append
// RosterDashboard returns every active engineer with their latest Overall
// score and the most recent cycle they were ranked in (F11).
func (s *ScoreStore) RosterDashboard(engineerStore *EngineerStore) ([]models.RosterEntry, error) {
	engineers, err := engineerStore.List(true)
	if err != nil {
		return nil, err
	}

	entries := make([]models.RosterEntry, 0, len(engineers))
	for _, engineer := range engineers {
		var latestCycleID int
		var periodEnd time.Time
		err := s.db.QueryRow(
			`SELECT rc.id, rc.period_end
			 FROM rating_cycles rc
			 JOIN sub_attribute_rankings sar ON sar.cycle_id = rc.id
			 WHERE sar.engineer_id = $1
			 ORDER BY rc.period_start DESC
			 LIMIT 1`,
			engineer.ID,
		).Scan(&latestCycleID, &periodEnd)

		if err == sql.ErrNoRows {
			entries = append(entries, models.RosterEntry{Engineer: engineer})
			continue
		}
		if err != nil {
			return nil, err
		}

		overall, err := s.OverallScore(engineer.ID, latestCycleID)
		if err != nil {
			return nil, err
		}
		entries = append(entries, models.RosterEntry{
			Engineer:      engineer,
			LatestOverall: overall,
			LastCycleDate: &periodEnd,
		})
	}
	return entries, nil
}
```
```go
// backend/scout/internal/handlers/dashboard.go
package handlers

import (
	"net/http"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
}

func NewDashboardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *DashboardHandler {
	return &DashboardHandler{scoreStore: s, engineerStore: engineerStore}
}

func (h *DashboardHandler) Get(c *gin.Context) {
	dashboard, err := h.scoreStore.RosterDashboard(h.engineerStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute roster dashboard"})
		return
	}
	c.JSON(http.StatusOK, dashboard)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		dashboardHandler := handlers.NewDashboardHandler(scoreStore, engineerStore)
		api.GET("/dashboard", dashboardHandler.Get)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'RosterDashboard|Dashboard' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/store/scores.go backend/scout/internal/store/scores_test.go backend/scout/internal/handlers/dashboard.go backend/scout/internal/handlers/dashboard_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add roster dashboard endpoint (F11)"
```

---

### Task 17: `metric_snapshots` migration, model, idempotent-upsert store (F4, NF4)

**Files:**
- Create: `backend/scout/migrations/000004_create_metric_snapshots_table.up.sql`
- Create: `backend/scout/migrations/000004_create_metric_snapshots_table.down.sql`
- Create: `backend/scout/internal/models/metric_snapshot.go`
- Create: `backend/scout/internal/store/metrics.go`
- Create: `backend/scout/internal/store/metrics_test.go`

**Interfaces:**
- Produces: `models.MetricSnapshot{ID int, EngineerID int, PeriodStart time.Time, PeriodEnd time.Time, PRsRaised int, PRsReviewed int, TicketsClosed int, ComplexityScore float64, SyncedAt time.Time}`; `store.MetricStore` with `NewMetricStore(db *sql.DB) *MetricStore`, `.UpsertSnapshot(engineerID int, periodStart, periodEnd time.Time, prsRaised, prsReviewed, ticketsClosed int, complexityScore float64) error`, `.ListByEngineer(engineerID int) ([]models.MetricSnapshot, error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/metrics_test.go
package store

import (
	"testing"
	"time"
)

func TestMetricStoreUpsertSnapshotIsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)

	if err := metricStore.UpsertSnapshot(e1.ID, start, end, 3, 5, 2, 7.5); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}
	// Re-running the sync for the same period must update in place, not
	// duplicate the row (NF4).
	if err := metricStore.UpsertSnapshot(e1.ID, start, end, 4, 6, 3, 9.0); err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected exactly one snapshot after repeated upsert, got %d", len(snapshots))
	}
	if snapshots[0].PRsRaised != 4 || snapshots[0].TicketsClosed != 3 {
		t.Fatalf("expected the second upsert's values to win, got %+v", snapshots[0])
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestMetricStore -v`
Expected: FAIL — `undefined: NewMetricStore`
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000004_create_metric_snapshots_table.up.sql
CREATE TABLE IF NOT EXISTS metric_snapshots (
    id SERIAL PRIMARY KEY,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    prs_raised INTEGER NOT NULL DEFAULT 0,
    prs_reviewed INTEGER NOT NULL DEFAULT 0,
    tickets_closed INTEGER NOT NULL DEFAULT 0,
    complexity_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (engineer_id, period_start, period_end)
);

CREATE INDEX idx_metric_snapshots_engineer_id ON metric_snapshots(engineer_id);
```
```sql
-- backend/scout/migrations/000004_create_metric_snapshots_table.down.sql
DROP TABLE IF EXISTS metric_snapshots;
```
```go
// backend/scout/internal/models/metric_snapshot.go
package models

import "time"

type MetricSnapshot struct {
	ID              int       `json:"id"`
	EngineerID      int       `json:"engineer_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	PRsRaised       int       `json:"prs_raised"`
	PRsReviewed     int       `json:"prs_reviewed"`
	TicketsClosed   int       `json:"tickets_closed"`
	ComplexityScore float64   `json:"complexity_score"`
	SyncedAt        time.Time `json:"synced_at"`
}
```
```go
// backend/scout/internal/store/metrics.go
package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type MetricStore struct {
	db *sql.DB
}

func NewMetricStore(db *sql.DB) *MetricStore {
	return &MetricStore{db: db}
}

// UpsertSnapshot idempotently writes a metric snapshot keyed on
// (engineer_id, period_start, period_end) — a repeated or retried sync run
// updates the existing row in place rather than duplicating it (NF4).
func (s *MetricStore) UpsertSnapshot(engineerID int, periodStart, periodEnd time.Time, prsRaised, prsReviewed, ticketsClosed int, complexityScore float64) error {
	_, err := s.db.Exec(
		`INSERT INTO metric_snapshots (engineer_id, period_start, period_end, prs_raised, prs_reviewed, tickets_closed, complexity_score, synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (engineer_id, period_start, period_end)
		 DO UPDATE SET prs_raised = $4, prs_reviewed = $5, tickets_closed = $6, complexity_score = $7, synced_at = NOW()`,
		engineerID, periodStart, periodEnd, prsRaised, prsReviewed, ticketsClosed, complexityScore,
	)
	return err
}

func (s *MetricStore) ListByEngineer(engineerID int) ([]models.MetricSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT id, engineer_id, period_start, period_end, prs_raised, prs_reviewed, tickets_closed, complexity_score, synced_at
		 FROM metric_snapshots WHERE engineer_id = $1 ORDER BY period_start DESC`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := []models.MetricSnapshot{}
	for rows.Next() {
		var m models.MetricSnapshot
		if err := rows.Scan(&m.ID, &m.EngineerID, &m.PeriodStart, &m.PeriodEnd, &m.PRsRaised, &m.PRsReviewed, &m.TicketsClosed, &m.ComplexityScore, &m.SyncedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, m)
	}
	return snapshots, rows.Err()
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... -run TestMetricStore -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations/000004_create_metric_snapshots_table.up.sql backend/scout/migrations/000004_create_metric_snapshots_table.down.sql backend/scout/internal/models/metric_snapshot.go backend/scout/internal/store/metrics.go backend/scout/internal/store/metrics_test.go
git commit -m "scout: add metric_snapshots table and idempotent-upsert store (F4, NF4)"
```

---

### Task 18: GitHub integration client (F4)

**Files:**
- Create: `backend/scout/internal/integrations/github.go`
- Create: `backend/scout/internal/integrations/github_test.go`

**Interfaces:**
- Produces: `integrations.GitHubClient` with `NewGitHubClient(token string, httpClient *http.Client) *GitHubClient`, `.FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (prsRaised, prsReviewed int, err error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/integrations/github_test.go
package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGitHubClientFetchPRStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/search/issues" {
			// Raised-PR search includes "author:"; reviewed-PR search
			// includes "reviewed-by:" — respond with a distinct count for each.
			if containsSubstring(q, "author:") {
				json.NewEncoder(w).Encode(map[string]int{"total_count": 3})
				return
			}
			json.NewEncoder(w).Encode(map[string]int{"total_count": 5})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	prsRaised, prsReviewed, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}
	if prsRaised != 3 || prsReviewed != 5 {
		t.Fatalf("expected raised=3 reviewed=5, got raised=%d reviewed=%d", prsRaised, prsReviewed)
	}
}

func containsSubstring(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/integrations/... -run TestGitHubClient -v`
Expected: FAIL — `package integrations: no Go files`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/integrations/github.go
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GitHubClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewGitHubClient(token string, httpClient *http.Client) *GitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GitHubClient{token: token, httpClient: httpClient, baseURL: "https://api.github.com"}
}

type searchIssuesResponse struct {
	TotalCount int `json:"total_count"`
}

// FetchPRStats returns the count of PRs the user raised and reviewed across
// the given repos within [since, until), via the GitHub search API (F4).
func (c *GitHubClient) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (prsRaised, prsReviewed int, err error) {
	repoFilter := ""
	for _, repo := range repos {
		repoFilter += fmt.Sprintf(" repo:%s", repo)
	}
	dateRange := fmt.Sprintf("%s..%s", since.Format("2006-01-02"), until.Format("2006-01-02"))

	raisedQuery := fmt.Sprintf("is:pr author:%s created:%s%s", username, dateRange, repoFilter)
	prsRaised, err = c.searchCount(ctx, raisedQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching raised PRs: %w", err)
	}

	reviewedQuery := fmt.Sprintf("is:pr reviewed-by:%s created:%s%s", username, dateRange, repoFilter)
	prsReviewed, err = c.searchCount(ctx, reviewedQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching reviewed PRs: %w", err)
	}

	return prsRaised, prsReviewed, nil
}

func (c *GitHubClient) searchCount(ctx context.Context, query string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/search/issues", nil)
	if err != nil {
		return 0, err
	}
	q := req.URL.Query()
	q.Set("q", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("github search returned status %d", resp.StatusCode)
	}

	var result searchIssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.TotalCount, nil
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/integrations/... -run TestGitHubClient -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/integrations/github.go backend/scout/internal/integrations/github_test.go
git commit -m "scout: add GitHub PR stats integration client (F4)"
```

---

### Task 19: Jira integration client (F4)

**Files:**
- Create: `backend/scout/internal/integrations/jira.go`
- Create: `backend/scout/internal/integrations/jira_test.go`

**Interfaces:**
- Produces: `integrations.JiraClient` with `NewJiraClient(baseURL, email, apiToken string, httpClient *http.Client) *JiraClient`, `.FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (ticketsClosed int, complexityScore float64, err error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/integrations/jira_test.go
package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJiraClientFetchTicketStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 4,
			"issues": []map[string]interface{}{
				{"fields": map[string]interface{}{"customfield_10016": 3.0}},
				{"fields": map[string]interface{}{"customfield_10016": 5.0}},
				{"fields": map[string]interface{}{"customfield_10016": nil}},
				{"fields": map[string]interface{}{"customfield_10016": 2.0}},
			},
		})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}
	if ticketsClosed != 4 {
		t.Fatalf("expected 4 tickets closed, got %d", ticketsClosed)
	}
	if complexityScore != 10 {
		t.Fatalf("expected complexity score sum 10 (3+5+2, nil skipped), got %v", complexityScore)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/integrations/... -run TestJiraClient -v`
Expected: FAIL — `undefined: NewJiraClient`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/integrations/jira.go
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type JiraClient struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

func NewJiraClient(baseURL, email, apiToken string, httpClient *http.Client) *JiraClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &JiraClient{baseURL: strings.TrimRight(baseURL, "/"), email: email, apiToken: apiToken, httpClient: httpClient}
}

// complexityFieldID is the Jira "Story point estimate" custom field used as
// the complexity figure (F4). Configurable per Jira instance would require
// an env var; hardcoded here as the common default (see plan judgment calls).
const complexityFieldID = "customfield_10016"

type jiraSearchResponse struct {
	Total  int `json:"total"`
	Issues []struct {
		Fields map[string]interface{} `json:"fields"`
	} `json:"issues"`
}

// FetchTicketStats returns the count of tickets the user closed and the sum
// of their complexity (story point) figures within [since, until) across
// the given Jira projects (F4).
func (c *JiraClient) FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (ticketsClosed int, complexityScore float64, err error) {
	projectFilter := ""
	if len(projects) > 0 {
		projectFilter = fmt.Sprintf(" AND project in (%s)", strings.Join(projects, ","))
	}
	jql := fmt.Sprintf(
		`assignee = "%s" AND status = Done AND resolved >= "%s" AND resolved < "%s"%s`,
		accountID, since.Format("2006-01-02"), until.Format("2006-01-02"), projectFilter,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/rest/api/3/search", nil)
	if err != nil {
		return 0, 0, err
	}
	q := req.URL.Query()
	q.Set("jql", jql)
	q.Set("fields", complexityFieldID)
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("jira search returned status %d", resp.StatusCode)
	}

	var result jiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}

	for _, issue := range result.Issues {
		if v, ok := issue.Fields[complexityFieldID].(float64); ok {
			complexityScore += v
		}
	}

	return result.Total, complexityScore, nil
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/integrations/... -run TestJiraClient -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/integrations/jira.go backend/scout/internal/integrations/jira_test.go
git commit -m "scout: add Jira ticket stats integration client (F4)"
```

---

### Task 20: Sync worker orchestration + in-process ticker scheduler (F4, NF4)

**Files:**
- Create: `backend/scout/internal/syncer/syncer.go`
- Create: `backend/scout/internal/syncer/syncer_test.go`
- Create: `backend/scout/internal/syncer/scheduler.go`
- Create: `backend/scout/internal/syncer/scheduler_test.go`
- Create: `backend/scout/internal/syncer/testutil_test.go`
- Modify: `backend/scout/cmd/server/main.go`

**Interfaces:**
- Consumes: `store.EngineerStore.ListActiveIDs`, `.GetByID` (Task 2); `store.MetricStore.UpsertSnapshot` (Task 17); `integrations.GitHubClient.FetchPRStats` (Task 18); `integrations.JiraClient.FetchTicketStats` (Task 19)
- Produces: `syncer.GitHubStatsFetcher` / `syncer.JiraStatsFetcher` interfaces; `syncer.Syncer` with `NewSyncer(engineerStore *store.EngineerStore, metricStore *store.MetricStore, github GitHubStatsFetcher, jira JiraStatsFetcher, repos, projects []string) *Syncer`, `.RunOnce(ctx context.Context, periodStart, periodEnd time.Time) error`; `syncer.RunSyncCycle(ctx context.Context, s *Syncer, now time.Time) error`; `syncer.StartScheduler(ctx context.Context, s *Syncer, interval time.Duration)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/syncer/testutil_test.go
package syncer

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

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
		t.Skipf("skipping: test database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

type fakeGitHub struct {
	raised, reviewed int
	err              error
}

func (f *fakeGitHub) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (int, int, error) {
	return f.raised, f.reviewed, f.err
}

type fakeJira struct {
	closed     int
	complexity float64
	err        error
}

func (f *fakeJira) FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (int, float64, error) {
	return f.closed, f.complexity, f.err
}
```
```go
// backend/scout/internal/syncer/syncer_test.go
package syncer

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/store"
)

func TestSyncerRunOnceUpsertsSnapshotsForActiveEngineers(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	s := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 3, reviewed: 2}, &fakeJira{closed: 4, complexity: 9}, []string{"org/repo"}, []string{"ENG"})

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	if err := s.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("run once failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 || snapshots[0].PRsRaised != 3 || snapshots[0].TicketsClosed != 4 {
		t.Fatalf("unexpected snapshots: %+v", snapshots)
	}
}

func TestSyncerRunOnceSkipsFailingEngineerButContinues(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gh1, gh2 := "broken-user", "working-user"
	jira := "abc123"
	engineerStore.Create("Broken", nil, &gh1, &jira, time.Now())
	e2, _ := engineerStore.Create("Working", nil, &gh2, &jira, time.Now())

	failingGitHub := &failOnceThenSucceedGitHub{failFor: gh1}
	s := NewSyncer(engineerStore, metricStore, failingGitHub, &fakeJira{closed: 1, complexity: 1}, nil, nil)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	if err := s.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("expected RunOnce to log-and-continue past a per-engineer failure, got error: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e2.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected the working engineer to still get a snapshot despite the other's failure, got %+v", snapshots)
	}
}

type failOnceThenSucceedGitHub struct {
	failFor string
}

func (f *failOnceThenSucceedGitHub) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (int, int, error) {
	if username == f.failFor {
		return 0, 0, errors.New("simulated github outage")
	}
	return 1, 1, nil
}
```
```go
// backend/scout/internal/syncer/scheduler_test.go
package syncer

import (
	"context"
	"testing"
	"time"

	"scout/internal/store"
)

func TestRunSyncCycleUsesTrailingBiweeklyPeriod(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	s := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 1, reviewed: 1}, &fakeJira{closed: 1, complexity: 1}, nil, nil)

	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if err := RunSyncCycle(context.Background(), s, now); err != nil {
		t.Fatalf("run sync cycle failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	wantStart := now.AddDate(0, 0, -14)
	if len(snapshots) != 1 || !snapshots[0].PeriodStart.Equal(wantStart) || !snapshots[0].PeriodEnd.Equal(now) {
		t.Fatalf("expected period [%s, %s], got %+v", wantStart, now, snapshots)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/syncer/... -v`
Expected: FAIL — `package syncer: no Go files`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/syncer/syncer.go
package syncer

import (
	"context"
	"log"
	"time"

	"scout/internal/store"
)

type GitHubStatsFetcher interface {
	FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (prsRaised, prsReviewed int, err error)
}

type JiraStatsFetcher interface {
	FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (ticketsClosed int, complexityScore float64, err error)
}

type Syncer struct {
	engineerStore *store.EngineerStore
	metricStore   *store.MetricStore
	github        GitHubStatsFetcher
	jira          JiraStatsFetcher
	repos         []string
	projects      []string
}

func NewSyncer(engineerStore *store.EngineerStore, metricStore *store.MetricStore, github GitHubStatsFetcher, jira JiraStatsFetcher, repos, projects []string) *Syncer {
	return &Syncer{
		engineerStore: engineerStore,
		metricStore:   metricStore,
		github:        github,
		jira:          jira,
		repos:         repos,
		projects:      projects,
	}
}

// RunOnce polls GitHub/Jira for every active engineer and idempotently
// upserts a metric_snapshots row per engineer (F4). A single engineer's
// fetch or upsert failure is logged and skipped rather than aborting the
// whole run — the next scheduled run will retry it, and the idempotent
// upsert means retries never duplicate or corrupt data (NF4).
func (s *Syncer) RunOnce(ctx context.Context, periodStart, periodEnd time.Time) error {
	activeIDs, err := s.engineerStore.ListActiveIDs()
	if err != nil {
		return err
	}

	for _, id := range activeIDs {
		engineer, err := s.engineerStore.GetByID(id)
		if err != nil {
			log.Printf("scout sync: failed to load engineer %d: %v", id, err)
			continue
		}
		if engineer == nil {
			continue
		}

		var prsRaised, prsReviewed int
		if engineer.GitHubUsername != nil {
			prsRaised, prsReviewed, err = s.github.FetchPRStats(ctx, *engineer.GitHubUsername, s.repos, periodStart, periodEnd)
			if err != nil {
				log.Printf("scout sync: github fetch failed for engineer %d (%s): %v", id, *engineer.GitHubUsername, err)
				continue
			}
		}

		var ticketsClosed int
		var complexityScore float64
		if engineer.JiraAccountID != nil {
			ticketsClosed, complexityScore, err = s.jira.FetchTicketStats(ctx, *engineer.JiraAccountID, s.projects, periodStart, periodEnd)
			if err != nil {
				log.Printf("scout sync: jira fetch failed for engineer %d (%s): %v", id, *engineer.JiraAccountID, err)
				continue
			}
		}

		if err := s.metricStore.UpsertSnapshot(id, periodStart, periodEnd, prsRaised, prsReviewed, ticketsClosed, complexityScore); err != nil {
			log.Printf("scout sync: upsert failed for engineer %d: %v", id, err)
			continue
		}
	}

	return nil
}
```
```go
// backend/scout/internal/syncer/scheduler.go
package syncer

import (
	"context"
	"log"
	"time"
)

// RunSyncCycle computes the trailing biweekly period ending at `now` and
// runs one sync pass. Split from StartScheduler so the period math is
// unit-testable without waiting on a real ticker.
func RunSyncCycle(ctx context.Context, s *Syncer, now time.Time) error {
	periodEnd := now
	periodStart := now.AddDate(0, 0, -14)
	return s.RunOnce(ctx, periodStart, periodEnd)
}

// StartScheduler runs an immediate sync pass, then one every `interval`,
// until ctx is cancelled. Runs as an in-process goroutine started from
// cmd/server/main.go — there is no separate cmd/syncer binary (see the
// plan's judgment-calls section).
func StartScheduler(ctx context.Context, s *Syncer, interval time.Duration) {
	go func() {
		if err := RunSyncCycle(ctx, s, time.Now()); err != nil {
			log.Printf("scout sync: initial run failed: %v", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := RunSyncCycle(ctx, s, time.Now()); err != nil {
					log.Printf("scout sync: scheduled run failed: %v", err)
				}
			}
		}
	}()
}
```
```go
// backend/scout/cmd/server/main.go — modify: add imports "context",
// "scout/internal/integrations", "scout/internal/store", "scout/internal/syncer",
// and start the scheduler after the DB connects and before r.Run
	githubClient := integrations.NewGitHubClient(cfg.GitHubToken, nil)
	jiraClient := integrations.NewJiraClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraAPIToken, nil)
	engineerStoreForSync := store.NewEngineerStore(db)
	metricStoreForSync := store.NewMetricStore(db)
	syncWorker := syncer.NewSyncer(engineerStoreForSync, metricStoreForSync, githubClient, jiraClient, cfg.GitHubRepos, cfg.JiraProjects)

	syncCtx, cancelSync := context.WithCancel(context.Background())
	defer cancelSync()
	syncer.StartScheduler(syncCtx, syncWorker, cfg.SyncInterval)

	r := buildRouter(db, cfg)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/syncer/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/syncer backend/scout/cmd/server/main.go
git commit -m "scout: add sync worker orchestration and in-process ticker scheduler (F4, NF4)"
```

---

### Task 21: Metrics-over-time read endpoint (F5)

**Files:**
- Create: `backend/scout/internal/handlers/metrics.go`
- Create: `backend/scout/internal/handlers/metrics_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.MetricStore.ListByEngineer` (Task 17); `store.EngineerStore.Exists` (Task 2)
- Produces: `handlers.NewMetricsHandler(s *store.MetricStore, engineerStore *store.EngineerStore) *MetricsHandler` with `.Get(c *gin.Context)` registered as `GET /api/engineers/:id/metrics`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/metrics_test.go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestMetricsHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	metricStore.UpsertSnapshot(e1.ID, time.Now().AddDate(0, 0, -14), time.Now(), 3, 5, 2, 7.5)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricsHandler(metricStore, engineerStore)
	r.GET("/api/engineers/:id/metrics", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+itoa(e1.ID)+"/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMetricsHandlerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricsHandler(metricStore, engineerStore)
	r.GET("/api/engineers/:id/metrics", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/999999/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestMetricsHandler -v`
Expected: FAIL — `undefined: NewMetricsHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/metrics.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	store         *store.MetricStore
	engineerStore *store.EngineerStore
}

func NewMetricsHandler(s *store.MetricStore, engineerStore *store.EngineerStore) *MetricsHandler {
	return &MetricsHandler{store: s, engineerStore: engineerStore}
}

func (h *MetricsHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	exists, err := h.engineerStore.Exists(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}

	snapshots, err := h.store.ListByEngineer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list metrics"})
		return
	}
	c.JSON(http.StatusOK, snapshots)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		metricStore := store.NewMetricStore(db)
		metricsHandler := handlers.NewMetricsHandler(metricStore, engineerStore)
		api.GET("/engineers/:id/metrics", metricsHandler.Get)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestMetricsHandler -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/metrics.go backend/scout/internal/handlers/metrics_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add engineer metrics-over-time endpoint (F5)"
```

---

### Task 22: `ai_ranking_sessions` migration, model, store (F9)

**Files:**
- Create: `backend/scout/migrations/000005_create_ai_ranking_sessions_table.up.sql`
- Create: `backend/scout/migrations/000005_create_ai_ranking_sessions_table.down.sql`
- Create: `backend/scout/internal/models/ai_session.go`
- Create: `backend/scout/internal/store/ai_sessions.go`
- Create: `backend/scout/internal/store/ai_sessions_test.go`

**Interfaces:**
- Produces: `models.AIRankingSession{ID int, CycleID int, SubAttributeID int, Transcript json.RawMessage, ProposedRanking json.RawMessage, CreatedAt time.Time}`; `store.AISessionStore` with `NewAISessionStore(db *sql.DB) *AISessionStore`, `.Create(cycleID, subAttributeID int) (*models.AIRankingSession, error)`, `.GetByID(id int) (*models.AIRankingSession, error)`, `.UpdateTranscript(id int, transcript json.RawMessage, proposedRanking json.RawMessage) error`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/ai_sessions_test.go
package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAISessionStoreCreateAndUpdateTranscript(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attributes", "main_attributes", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	sessionStore := NewAISessionStore(db)

	main, _ := mainStore.Create("test_main_ai", "Test Main AI")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	session, err := sessionStore.Create(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if string(session.Transcript) != "[]" {
		t.Fatalf("expected new session to start with empty transcript array, got %s", session.Transcript)
	}

	transcript := json.RawMessage(`[{"role":"user","content":"who stood out this cycle?"}]`)
	proposed := json.RawMessage(`{"ranking":[{"engineer_id":1,"rank":1}]}`)
	if err := sessionStore.UpdateTranscript(session.ID, transcript, proposed); err != nil {
		t.Fatalf("update transcript failed: %v", err)
	}

	updated, err := sessionStore.GetByID(session.ID)
	if err != nil {
		t.Fatalf("get by id failed: %v", err)
	}
	if string(updated.Transcript) != string(transcript) {
		t.Fatalf("expected transcript to be updated, got %s", updated.Transcript)
	}
	if string(updated.ProposedRanking) != string(proposed) {
		t.Fatalf("expected proposed ranking to be set, got %s", updated.ProposedRanking)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... -run TestAISessionStore -v`
Expected: FAIL — `undefined: NewAISessionStore`
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000005_create_ai_ranking_sessions_table.up.sql
CREATE TABLE IF NOT EXISTS ai_ranking_sessions (
    id SERIAL PRIMARY KEY,
    cycle_id INTEGER NOT NULL REFERENCES rating_cycles(id) ON DELETE CASCADE,
    sub_attribute_id INTEGER NOT NULL REFERENCES sub_attributes(id) ON DELETE CASCADE,
    transcript JSONB NOT NULL DEFAULT '[]',
    proposed_ranking JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_ranking_sessions_cycle_id ON ai_ranking_sessions(cycle_id);
```
```sql
-- backend/scout/migrations/000005_create_ai_ranking_sessions_table.down.sql
DROP TABLE IF EXISTS ai_ranking_sessions;
```
```go
// backend/scout/internal/models/ai_session.go
package models

import (
	"encoding/json"
	"time"
)

type AIRankingSession struct {
	ID              int             `json:"id"`
	CycleID         int             `json:"cycle_id"`
	SubAttributeID  int             `json:"sub_attribute_id"`
	Transcript      json.RawMessage `json:"transcript"`
	ProposedRanking json.RawMessage `json:"proposed_ranking"`
	CreatedAt       time.Time       `json:"created_at"`
}
```
```go
// backend/scout/internal/store/ai_sessions.go
package store

import (
	"database/sql"
	"encoding/json"

	"scout/internal/models"
)

type AISessionStore struct {
	db *sql.DB
}

func NewAISessionStore(db *sql.DB) *AISessionStore {
	return &AISessionStore{db: db}
}

func (s *AISessionStore) Create(cycleID, subAttributeID int) (*models.AIRankingSession, error) {
	var session models.AIRankingSession
	err := s.db.QueryRow(
		`INSERT INTO ai_ranking_sessions (cycle_id, sub_attribute_id, transcript)
		 VALUES ($1, $2, '[]')
		 RETURNING id, cycle_id, sub_attribute_id, transcript, proposed_ranking, created_at`,
		cycleID, subAttributeID,
	).Scan(&session.ID, &session.CycleID, &session.SubAttributeID, &session.Transcript, &session.ProposedRanking, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *AISessionStore) GetByID(id int) (*models.AIRankingSession, error) {
	var session models.AIRankingSession
	err := s.db.QueryRow(
		`SELECT id, cycle_id, sub_attribute_id, transcript, proposed_ranking, created_at
		 FROM ai_ranking_sessions WHERE id = $1`, id,
	).Scan(&session.ID, &session.CycleID, &session.SubAttributeID, &session.Transcript, &session.ProposedRanking, &session.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateTranscript replaces the session's full transcript (the chat history
// accumulated so far) and, when the assistant's latest reply included one,
// the proposed_ranking payload (F9). Persisting the ranking into
// sub_attribute_rankings only happens later, via the explicit accept
// endpoint (NF3) — this method never writes to sub_attribute_rankings.
func (s *AISessionStore) UpdateTranscript(id int, transcript json.RawMessage, proposedRanking json.RawMessage) error {
	_, err := s.db.Exec(
		"UPDATE ai_ranking_sessions SET transcript = $1, proposed_ranking = $2 WHERE id = $3",
		transcript, proposedRanking, id,
	)
	return err
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... -run TestAISessionStore -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations/000005_create_ai_ranking_sessions_table.up.sql backend/scout/migrations/000005_create_ai_ranking_sessions_table.down.sql backend/scout/internal/models/ai_session.go backend/scout/internal/store/ai_sessions.go backend/scout/internal/store/ai_sessions_test.go
git commit -m "scout: add ai_ranking_sessions table and store (F9)"
```

---

### Task 23: Anthropic client base + pure JSON-block extraction (F9)

**Files:**
- Create: `backend/scout/internal/aiclient/client.go`
- Create: `backend/scout/internal/aiclient/jsonblock.go`
- Create: `backend/scout/internal/aiclient/jsonblock_test.go`

**Interfaces:**
- Produces: `aiclient.Client` with `NewClient(apiKey string, opts ...option.RequestOption) *Client`; `aiclient.ChatMessage{Role string, Content string}`; `aiclient.ExtractJSONBlock(text string) (raw json.RawMessage, ok bool)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/aiclient/jsonblock_test.go
package aiclient

import "testing"

func TestExtractJSONBlockFindsFencedBlock(t *testing.T) {
	text := "Alex clearly stood out this cycle for shipping the auth migration solo.\n\n```json\n{\"ranking\":[{\"engineer_id\":1,\"rank\":1},{\"engineer_id\":2,\"rank\":2}],\"rationale\":\"Alex led the migration\"}\n```"

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected a JSON block to be found")
	}
	if string(raw) != `{"ranking":[{"engineer_id":1,"rank":1},{"engineer_id":2,"rank":2}],"rationale":"Alex led the migration"}` {
		t.Fatalf("unexpected extracted JSON: %s", raw)
	}
}

func TestExtractJSONBlockNoBlockPresent(t *testing.T) {
	_, ok := ExtractJSONBlock("Tell me more about what Sam shipped this cycle before I propose a ranking.")
	if ok {
		t.Fatal("expected no JSON block to be found in a plain clarifying question")
	}
}

func TestExtractJSONBlockRejectsInvalidJSON(t *testing.T) {
	_, ok := ExtractJSONBlock("```json\nnot valid json\n```")
	if ok {
		t.Fatal("expected invalid JSON inside the fence to be rejected")
	}
}

func TestExtractJSONBlockReturnsLastBlockWhenMultiple(t *testing.T) {
	text := "```json\n{\"draft\":1}\n```\n\nActually, here's the revised proposal:\n\n```json\n{\"draft\":2}\n```"
	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected a JSON block to be found")
	}
	if string(raw) != `{"draft":2}` {
		t.Fatalf("expected the last block to win, got %s", raw)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/aiclient/... -run TestExtractJSONBlock -v`
Expected: FAIL — `package aiclient: no Go files`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/aiclient/client.go
package aiclient

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the official Anthropic Go SDK for Scout's two AI flows: the
// conversational ranking chat (F9) and the highlight/lowlight semantic
// duplicate check (F14).
type Client struct {
	anthropic anthropic.Client
}

// NewClient builds a Client from an API key, plus any extra SDK request
// options (tests use this to point the client at an httptest server via
// option.WithBaseURL instead of the real Anthropic API).
func NewClient(apiKey string, opts ...option.RequestOption) *Client {
	allOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	return &Client{anthropic: anthropic.NewClient(allOpts...)}
}

// ChatMessage is one turn in the ranking chat transcript.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
```
```go
// backend/scout/internal/aiclient/jsonblock.go
package aiclient

import (
	"encoding/json"
	"regexp"
)

var jsonBlockPattern = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")

// ExtractJSONBlock finds the last ```json fenced code block in text — the
// convention this plan uses for the ranking chat to carry a proposed_ranking
// payload alongside its conversational rationale (F9) — and returns it as
// raw JSON. ok is false if no block is present or its contents aren't valid
// JSON (e.g. the assistant is still asking a clarifying question).
func ExtractJSONBlock(text string) (raw json.RawMessage, ok bool) {
	matches := jsonBlockPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, false
	}
	candidate := matches[len(matches)-1][1]
	if !json.Valid([]byte(candidate)) {
		return nil, false
	}
	return json.RawMessage(candidate), true
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go mod tidy && go test ./internal/aiclient/... -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/aiclient/client.go backend/scout/internal/aiclient/jsonblock.go backend/scout/internal/aiclient/jsonblock_test.go backend/scout/go.mod backend/scout/go.sum
git commit -m "scout: add Anthropic client base and pure JSON-block extraction for AI ranking chat (F9)"
```

---

### Task 24: Streaming ranking chat via Anthropic Messages API (F9)

**Files:**
- Create: `backend/scout/internal/aiclient/chat.go`
- Create: `backend/scout/internal/aiclient/chat_test.go`

**Interfaces:**
- Consumes: `aiclient.Client`, `aiclient.ChatMessage`, `aiclient.ExtractJSONBlock` (Task 23)
- Produces: `Client.StreamRankingChat(ctx context.Context, w io.Writer, systemPrompt string, history []ChatMessage, userMessage string) (replyText string, proposedRanking json.RawMessage, err error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/aiclient/chat_test.go
package aiclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestStreamRankingChatAccumulatesTextAndExtractsProposal(t *testing.T) {
	sseBody := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Alex should rank 1st. "}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"` + `\`\`\`json\n{\"ranking\":[{\"engineer_id\":1,\"rank\":1}]}\n\`\`\`` + `"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":0}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":25}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseBody)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		context.Background(), &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err != nil {
		t.Fatalf("stream ranking chat failed: %v", err)
	}
	if !strings.Contains(replyText, "Alex should rank 1st.") {
		t.Fatalf("expected accumulated reply text to include the streamed sentence, got %q", replyText)
	}
	if proposedRanking == nil {
		t.Fatal("expected a proposed ranking to be extracted from the trailing JSON block")
	}
	if out.Len() == 0 {
		t.Fatal("expected streamed chunks to be written to the writer")
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/aiclient/... -run TestStreamRankingChat -v`
Expected: FAIL — `undefined: (*Client).StreamRankingChat`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/aiclient/chat.go
package aiclient

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// StreamRankingChat sends the conversation (prior transcript + new user
// message) to Claude and streams the assistant's reply as SSE "data: "
// chunks to w for the frontend chat UI to consume (F9). It returns the full
// accumulated reply text plus any proposed ranking JSON extracted from a
// trailing ```json code block in that reply — the caller (the ai-sessions
// handler) is responsible for persisting the transcript and for only ever
// writing the ranking into sub_attribute_rankings on explicit admin confirm
// (NF3); this method never touches sub_attribute_rankings.
func (c *Client) StreamRankingChat(ctx context.Context, w io.Writer, systemPrompt string, history []ChatMessage, userMessage string) (replyText string, proposedRanking json.RawMessage, err error) {
	messages := make([]anthropic.MessageParam, 0, len(history)+1)
	for _, m := range history {
		role := anthropic.MessageParamRoleUser
		if m.Role == "assistant" {
			role = anthropic.MessageParamRoleAssistant
		}
		messages = append(messages, anthropic.MessageParam{
			Role:    role,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(m.Content)},
		})
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	stream := c.anthropic.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     "claude-opus-5",
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages:  messages,
	})

	var builder strings.Builder
	for stream.Next() {
		event := stream.Current()
		delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent)
		if !ok {
			continue
		}
		textDelta, ok := delta.Delta.AsAny().(anthropic.TextDelta)
		if !ok {
			continue
		}
		builder.WriteString(textDelta.Text)
		if _, werr := w.Write([]byte("data: " + escapeSSE(textDelta.Text) + "\n\n")); werr != nil {
			return "", nil, werr
		}
		if flusher, ok := w.(interface{ Flush() }); ok {
			flusher.Flush()
		}
	}
	if err := stream.Err(); err != nil {
		return "", nil, err
	}

	replyText = builder.String()
	if raw, found := ExtractJSONBlock(replyText); found {
		proposedRanking = raw
	}
	return replyText, proposedRanking, nil
}

func escapeSSE(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/aiclient/... -run TestStreamRankingChat -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/aiclient/chat.go backend/scout/internal/aiclient/chat_test.go
git commit -m "scout: add streaming ranking chat via Anthropic Messages API (F9)"
```

---

### Task 25: `highlight_entries` migration, model, store, list/add handlers (F13)

**Files:**
- Create: `backend/scout/migrations/000006_create_highlight_entries_table.up.sql`
- Create: `backend/scout/migrations/000006_create_highlight_entries_table.down.sql`
- Create: `backend/scout/internal/models/highlight.go`
- Create: `backend/scout/internal/store/highlights.go`
- Create: `backend/scout/internal/store/highlights_test.go`
- Create: `backend/scout/internal/handlers/highlights.go`
- Create: `backend/scout/internal/handlers/highlights_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Produces: `models.HighlightEntry{ID int, EngineerID int, Kind string, Body string, CreatedAt time.Time}`; `store.HighlightStore` with `NewHighlightStore(db *sql.DB) *HighlightStore`, `.List(engineerID int) ([]models.HighlightEntry, error)`, `.Create(engineerID int, kind, body string) (*models.HighlightEntry, error)`; `handlers.NewHighlightHandler(s *store.HighlightStore, engineerStore *store.EngineerStore) *HighlightHandler` with `.List(c *gin.Context)` (`GET /api/engineers/:id/highlights`), `.Create(c *gin.Context)` (`POST /api/engineers/:id/highlights`)

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/store/highlights_test.go
package store

import "testing"

func TestHighlightStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	highlightStore := NewHighlightStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, timeNow())

	created, err := highlightStore.Create(e1.ID, "highlight", "Shipped the auth migration solo, ahead of schedule.")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Kind != "highlight" {
		t.Fatalf("unexpected kind: %s", created.Kind)
	}

	list, err := highlightStore.List(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one highlight entry, got %+v", list)
	}
}
```
```go
// backend/scout/internal/handlers/highlights_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestHighlightHandlerCreateAndList(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.GET("/api/engineers/:id/highlights", h.List)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "lowlight", "body": "Missed the sprint deadline without flagging it early."})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHighlightHandlerCreateRejectsInvalidKind(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "sidenote", "body": "not a valid kind"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/store/... ./internal/handlers/... -run 'Highlight' -v`
Expected: FAIL — `undefined: NewHighlightStore`
- [ ] **Step 3: Write minimal implementation**
```sql
-- backend/scout/migrations/000006_create_highlight_entries_table.up.sql
CREATE TABLE IF NOT EXISTS highlight_entries (
    id SERIAL PRIMARY KEY,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    kind VARCHAR(10) NOT NULL CHECK (kind IN ('highlight', 'lowlight')),
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_highlight_entries_engineer_id ON highlight_entries(engineer_id);
```
```sql
-- backend/scout/migrations/000006_create_highlight_entries_table.down.sql
DROP TABLE IF EXISTS highlight_entries;
```
```go
// backend/scout/internal/models/highlight.go
package models

import "time"

type HighlightEntry struct {
	ID         int       `json:"id"`
	EngineerID int       `json:"engineer_id"`
	Kind       string    `json:"kind"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}
```
```go
// backend/scout/internal/store/highlights.go
package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

// timeNow is a thin indirection so store tests can call a package-local
// helper without importing "time" redundantly in every test file.
func timeNow() time.Time { return time.Now() }

type HighlightStore struct {
	db *sql.DB
}

func NewHighlightStore(db *sql.DB) *HighlightStore {
	return &HighlightStore{db: db}
}

func (s *HighlightStore) List(engineerID int) ([]models.HighlightEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, engineer_id, kind, body, created_at
		 FROM highlight_entries WHERE engineer_id = $1 ORDER BY created_at DESC`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.HighlightEntry{}
	for rows.Next() {
		var e models.HighlightEntry
		if err := rows.Scan(&e.ID, &e.EngineerID, &e.Kind, &e.Body, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *HighlightStore) Create(engineerID int, kind, body string) (*models.HighlightEntry, error) {
	var e models.HighlightEntry
	err := s.db.QueryRow(
		`INSERT INTO highlight_entries (engineer_id, kind, body) VALUES ($1, $2, $3)
		 RETURNING id, engineer_id, kind, body, created_at`,
		engineerID, kind, body,
	).Scan(&e.ID, &e.EngineerID, &e.Kind, &e.Body, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
```
```go
// backend/scout/internal/handlers/highlights.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type HighlightHandler struct {
	store         *store.HighlightStore
	engineerStore *store.EngineerStore
}

func NewHighlightHandler(s *store.HighlightStore, engineerStore *store.EngineerStore) *HighlightHandler {
	return &HighlightHandler{store: s, engineerStore: engineerStore}
}

type highlightRequest struct {
	Kind string `json:"kind" binding:"required"`
	Body string `json:"body" binding:"required"`
}

func (h *HighlightHandler) List(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	entries, err := h.store.List(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list highlights"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *HighlightHandler) Create(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req highlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind and body are required"})
		return
	}
	if req.Kind != "highlight" && req.Kind != "lowlight" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'highlight' or 'lowlight'"})
		return
	}
	exists, err := h.engineerStore.Exists(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}

	entry, err := h.store.Create(engineerID, req.Kind, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create highlight entry"})
		return
	}
	c.JSON(http.StatusCreated, entry)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		highlightStore := store.NewHighlightStore(db)
		highlightHandler := handlers.NewHighlightHandler(highlightStore, engineerStore)
		api.GET("/engineers/:id/highlights", highlightHandler.List)
		api.POST("/engineers/:id/highlights", highlightHandler.Create)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go run ./cmd/server -migrate=up && go test ./internal/store/... ./internal/handlers/... -run 'Highlight' -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/migrations/000006_create_highlight_entries_table.up.sql backend/scout/migrations/000006_create_highlight_entries_table.down.sql backend/scout/internal/models/highlight.go backend/scout/internal/store/highlights.go backend/scout/internal/store/highlights_test.go backend/scout/internal/handlers/highlights.go backend/scout/internal/handlers/highlights_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add highlight/lowlight entries table, store, and list/add endpoints (F13)"
```

---

### Task 26: AI ranking chat handler — streamed, context-aware (F9)

**Files:**
- Create: `backend/scout/internal/handlers/ai_chat.go`
- Create: `backend/scout/internal/handlers/ai_chat_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `aiclient.Client.StreamRankingChat`, `aiclient.ChatMessage` (Tasks 23, 24); `store.AISessionStore` (Task 22); `store.EngineerStore.List` (Task 2); `store.MetricStore.ListByEngineer` (Task 17); `store.HighlightStore.List` (Task 25); `store.SubAttributeStore.GetByID` (Task 6); `store.CycleStore.Exists` (Task 11)
- Produces: `handlers.NewAIChatHandler(aiClient *aiclient.Client, sessionStore *store.AISessionStore, engineerStore *store.EngineerStore, metricStore *store.MetricStore, highlightStore *store.HighlightStore, subAttrStore *store.SubAttributeStore, cycleStore *store.CycleStore) *AIChatHandler` with `.Chat(c *gin.Context)` registered as `POST /api/cycles/:id/ai-sessions`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/ai_chat_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"
)

func TestAIChatHandlerStreamsAndPersistsSession(t *testing.T) {
	sseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte(
			"event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"msg_1\",\"type\":\"message\",\"role\":\"assistant\",\"model\":\"claude-opus-5\",\"content\":[],\"stop_reason\":null,\"stop_sequence\":null,\"usage\":{\"input_tokens\":5,\"output_tokens\":0}}}\n\n" +
				"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n" +
				"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Tell me more about Sam's cycle.\"}}\n\n" +
				"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n" +
				"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\",\"stop_sequence\":null},\"usage\":{\"output_tokens\":8}}\n\n" +
				"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		))
	}))
	defer sseServer.Close()

	db := setupTestDBForHandlers(t)
	for _, table := range []string{"ai_ranking_sessions", "highlight_entries", "metric_snapshots", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	sessionStore := store.NewAISessionStore(db)
	metricStore := store.NewMetricStore(db)
	highlightStore := store.NewHighlightStore(db)

	engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_chat", "Test Main Chat")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(sseServer.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIChatHandler(aiClient, sessionStore, engineerStore, metricStore, highlightStore, subStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions", h.Chat)

	body, _ := json.Marshal(map[string]interface{}{
		"sub_attribute_id": sub.ID,
		"message":          "Who stood out this cycle?",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/"+itoa(cycle.ID)+"/ai-sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "event: session") {
		t.Fatalf("expected the response to open with a session event carrying the session id, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Sam's cycle") {
		t.Fatalf("expected the streamed reply text in the response body, got %s", w.Body.String())
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestAIChatHandler -v`
Expected: FAIL — `undefined: NewAIChatHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/ai_chat.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"scout/internal/aiclient"
	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type AIChatHandler struct {
	aiClient       *aiclient.Client
	sessionStore   *store.AISessionStore
	engineerStore  *store.EngineerStore
	metricStore    *store.MetricStore
	highlightStore *store.HighlightStore
	subAttrStore   *store.SubAttributeStore
	cycleStore     *store.CycleStore
}

func NewAIChatHandler(aiClient *aiclient.Client, sessionStore *store.AISessionStore, engineerStore *store.EngineerStore, metricStore *store.MetricStore, highlightStore *store.HighlightStore, subAttrStore *store.SubAttributeStore, cycleStore *store.CycleStore) *AIChatHandler {
	return &AIChatHandler{
		aiClient:       aiClient,
		sessionStore:   sessionStore,
		engineerStore:  engineerStore,
		metricStore:    metricStore,
		highlightStore: highlightStore,
		subAttrStore:   subAttrStore,
		cycleStore:     cycleStore,
	}
}

type aiChatRequest struct {
	SessionID      *int   `json:"session_id"`
	SubAttributeID int    `json:"sub_attribute_id"`
	Message        string `json:"message" binding:"required"`
}

// Chat handles POST /api/cycles/:id/ai-sessions (F9). It streams the
// assistant's reply as Server-Sent Events: the response opens with a single
// `event: session` frame carrying {"session_id": N} (so the frontend can
// continue the same session on the next message), followed by plain
// `data: ...` text chunks as the reply streams in. The proposed ranking (if
// any) is parsed from the reply and stored on the session — it is never
// written to sub_attribute_rankings here; only the accept endpoint (Task 27)
// does that, and only on explicit admin confirm (NF3).
func (h *AIChatHandler) Chat(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	var req aiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	session, err := h.loadOrCreateSession(cycleID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var history []aiclient.ChatMessage
	if err := json.Unmarshal(session.Transcript, &history); err != nil {
		history = nil
	}

	systemPrompt, err := h.buildSystemPrompt(session.SubAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build chat context"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	fmt.Fprintf(c.Writer, "event: session\ndata: {\"session_id\":%d}\n\n", session.ID)
	if flusher, ok := c.Writer.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	replyText, proposedRanking, err := h.aiClient.StreamRankingChat(c.Request.Context(), c.Writer, systemPrompt, history, req.Message)
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\":%q}\n\n", err.Error())
		return
	}

	newHistory := append(history,
		aiclient.ChatMessage{Role: "user", Content: req.Message},
		aiclient.ChatMessage{Role: "assistant", Content: replyText},
	)
	transcriptJSON, err := json.Marshal(newHistory)
	if err != nil {
		return
	}
	_ = h.sessionStore.UpdateTranscript(session.ID, transcriptJSON, proposedRanking)
}

func (h *AIChatHandler) loadOrCreateSession(cycleID int, req aiChatRequest) (*models.AIRankingSession, error) {
	if req.SessionID != nil {
		session, err := h.sessionStore.GetByID(*req.SessionID)
		if err != nil {
			return nil, err
		}
		if session == nil {
			return nil, fmt.Errorf("ai ranking session not found")
		}
		return session, nil
	}
	if req.SubAttributeID == 0 {
		return nil, fmt.Errorf("sub_attribute_id is required to start a new session")
	}
	return h.sessionStore.Create(cycleID, req.SubAttributeID)
}

// buildSystemPrompt assembles the active roster's synced metrics and
// existing highlights/lowlights as context for the sub-attribute being
// ranked, per the architecture note in the spec.
func (h *AIChatHandler) buildSystemPrompt(subAttributeID int) (string, error) {
	subAttr, err := h.subAttrStore.GetByID(subAttributeID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("You are Scout's ranking assistant. The manager will describe observations about their engineers in natural language for the current biweekly cycle. Your job is to propose a strict 1..N rank ordering (no ties) of the active engineers for one sub-attribute, with a short rationale. ")
	if subAttr != nil {
		fmt.Fprintf(&b, "The sub-attribute being ranked is %q. ", subAttr.Name)
	}
	b.WriteString("When you are ready to propose a ranking, end your reply with a fenced ```json code block containing exactly {\"rationale\": string, \"ranking\": [{\"engineer_id\": int, \"rank\": int}, ...]}. Only include that block once you have enough information — otherwise keep asking clarifying questions.\n\n")

	engineers, err := h.engineerStore.List(true)
	if err != nil {
		return "", err
	}
	b.WriteString("Active roster, with recent synced metrics and logged highlights/lowlights:\n")
	for _, e := range engineers {
		fmt.Fprintf(&b, "- Engineer %d: %s\n", e.ID, e.Name)
		if snapshots, err := h.metricStore.ListByEngineer(e.ID); err == nil && len(snapshots) > 0 {
			latest := snapshots[0]
			fmt.Fprintf(&b, "  Metrics (%s to %s): %d PRs raised, %d PRs reviewed, %d tickets closed, complexity %.1f\n",
				latest.PeriodStart.Format("2006-01-02"), latest.PeriodEnd.Format("2006-01-02"),
				latest.PRsRaised, latest.PRsReviewed, latest.TicketsClosed, latest.ComplexityScore)
		}
		if entries, err := h.highlightStore.List(e.ID); err == nil {
			for _, entry := range entries {
				label := "Highlight"
				if entry.Kind == "lowlight" {
					label = "Lowlight"
				}
				fmt.Fprintf(&b, "  %s (%s): %s\n", label, entry.CreatedAt.Format("2006-01-02"), entry.Body)
			}
		}
	}

	return b.String(), nil
}
```
```go
// backend/scout/cmd/server/router.go — modify: add import "scout/internal/aiclient",
// construct the shared Anthropic client, and wire the chat route inside the
// `api` group. metricStore (Task 21) and highlightStore (Task 25) already
// exist in this function by this point — reuse them, don't redeclare.
	aiClient := aiclient.NewClient(cfg.AnthropicAPIKey)

	aiSessionStore := store.NewAISessionStore(db)
	aiChatHandler := handlers.NewAIChatHandler(aiClient, aiSessionStore, engineerStore, metricStore, highlightStore, subAttributeStore, cycleStore)
	api.POST("/cycles/:id/ai-sessions", aiChatHandler.Chat)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestAIChatHandler -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/ai_chat.go backend/scout/internal/handlers/ai_chat_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add context-aware streaming AI ranking chat endpoint (F9)"
```

---

### Task 27: AI ranking accept endpoint — persist only on explicit confirm (F9, NF3)

**Files:**
- Create: `backend/scout/internal/handlers/ai_accept.go`
- Create: `backend/scout/internal/handlers/ai_accept_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `store.AISessionStore.GetByID` (Task 22); `store.RankingStore.SubmitRanking`, `scoring.RankEntry` (Task 12)
- Produces: `handlers.NewAIAcceptHandler(sessionStore *store.AISessionStore, rankingStore *store.RankingStore, cycleStore *store.CycleStore) *AIAcceptHandler` with `.Accept(c *gin.Context)` registered as `POST /api/cycles/:id/ai-sessions/:sessionId/accept`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/ai_accept_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestAIAcceptHandlerPersistsEditedRanking(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	sessionStore := store.NewAISessionStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_accept", "Test Main Accept")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	session, _ := sessionStore.Create(cycle.ID, sub.ID)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIAcceptHandler(sessionStore, rankingStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions/:sessionId/accept", h.Accept)

	// The admin may have edited the AI's proposed ranking before confirming
	// (NF3) — the request body, not session.ProposedRanking, is the source
	// of truth for what gets persisted.
	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/" + itoa(cycle.ID) + "/ai-sessions/" + itoa(session.ID) + "/accept"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	if len(rankings) != 1 || rankings[0].EngineerID != e1.ID {
		t.Fatalf("expected the accepted ranking to be persisted, got %+v", rankings)
	}
}

func TestAIAcceptHandlerSessionNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	sessionStore := store.NewAISessionStore(db)
	rankingStore := store.NewRankingStore(db, store.NewEngineerStore(db))
	cycleStore := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIAcceptHandler(sessionStore, rankingStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions/:sessionId/accept", h.Accept)

	body, _ := json.Marshal(map[string]interface{}{"rankings": []map[string]int{}})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/1/ai-sessions/999999/accept", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestAIAcceptHandler -v`
Expected: FAIL — `undefined: NewAIAcceptHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/ai_accept.go
package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type AIAcceptHandler struct {
	sessionStore *store.AISessionStore
	rankingStore *store.RankingStore
	cycleStore   *store.CycleStore
}

func NewAIAcceptHandler(sessionStore *store.AISessionStore, rankingStore *store.RankingStore, cycleStore *store.CycleStore) *AIAcceptHandler {
	return &AIAcceptHandler{sessionStore: sessionStore, rankingStore: rankingStore, cycleStore: cycleStore}
}

// Accept handles POST /api/cycles/:id/ai-sessions/:sessionId/accept (F9,
// NF3). The AI's proposed_ranking is never auto-applied — this is the only
// code path that writes an AI session's ranking into sub_attribute_rankings,
// and it only runs on this explicit call. The request body (the admin's
// possibly-edited ranking, not session.ProposedRanking) is the source of
// truth for what gets persisted, and it goes through the exact same strict
// 1..N permutation validation as the manual ranking endpoint (F6).
func (h *AIAcceptHandler) Accept(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	session, err := h.sessionStore.GetByID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up ai ranking session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ai ranking session not found"})
		return
	}

	var req submitRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings are required"})
		return
	}

	entries := make([]scoring.RankEntry, 0, len(req.Rankings))
	for _, r := range req.Rankings {
		entries = append(entries, scoring.RankEntry{EngineerID: r.EngineerID, Rank: r.Rank})
	}

	rankings, err := h.rankingStore.SubmitRanking(cycleID, session.SubAttributeID, entries)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rankings)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
// (aiSessionStore, cycleStore, and rankingStore already exist from Tasks 22,
// 11, and 12 respectively)
		aiAcceptHandler := handlers.NewAIAcceptHandler(aiSessionStore, rankingStore, cycleStore)
		api.POST("/cycles/:id/ai-sessions/:sessionId/accept", aiAcceptHandler.Accept)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestAIAcceptHandler -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/ai_accept.go backend/scout/internal/handlers/ai_accept_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add AI ranking accept endpoint, persisting only on explicit confirm (F9, NF3)"
```

---

### Task 28: AI semantic duplicate-check client method, structured output (F14)

**Files:**
- Create: `backend/scout/internal/aiclient/duplicate.go`
- Create: `backend/scout/internal/aiclient/duplicate_test.go`

**Interfaces:**
- Consumes: `aiclient.Client` (Task 23)
- Produces: `aiclient.ExistingEntry{ID int, Body string, Kind string}`; `aiclient.DuplicateCheckResult{IsDuplicate bool, MatchedEntryID *int, Note string}`; `Client.CheckDuplicate(ctx context.Context, newBody string, existing []ExistingEntry) (*DuplicateCheckResult, error)`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/aiclient/duplicate_test.go
package aiclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestCheckDuplicateParsesStructuredResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_dup",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":true,\"matched_entry_id\":5,\"similarity_note\":\"Both describe missing the Q1 deadline\"}"}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 20, "output_tokens": 15}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Missed the Q1 deadline without flagging it early", []ExistingEntry{
		{ID: 5, Kind: "lowlight", Body: "Slipped the Q1 deadline and didn't raise it in standup"},
	})
	if err != nil {
		t.Fatalf("check duplicate failed: %v", err)
	}
	if !result.IsDuplicate {
		t.Fatal("expected IsDuplicate true")
	}
	if result.MatchedEntryID == nil || *result.MatchedEntryID != 5 {
		t.Fatalf("expected matched entry id 5, got %v", result.MatchedEntryID)
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/aiclient/... -run TestCheckDuplicate -v`
Expected: FAIL — `undefined: (*Client).CheckDuplicate`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/aiclient/duplicate.go
package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type ExistingEntry struct {
	ID   int
	Kind string
	Body string
}

type DuplicateCheckResult struct {
	IsDuplicate    bool   `json:"is_duplicate"`
	MatchedEntryID *int   `json:"matched_entry_id"`
	Note           string `json:"similarity_note"`
}

// CheckDuplicate asks Claude whether newBody is a likely semantic duplicate
// of any of the engineer's existing highlight/lowlight entries (F14), using
// a JSON-schema structured output so the response is always parseable —
// no code-block extraction needed here, unlike the ranking chat.
func (c *Client) CheckDuplicate(ctx context.Context, newBody string, existing []ExistingEntry) (*DuplicateCheckResult, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "New entry to check: %q\n\nExisting entries for this engineer:\n", newBody)
	for _, e := range existing {
		fmt.Fprintf(&b, "- id=%d (%s): %s\n", e.ID, e.Kind, e.Body)
	}
	b.WriteString("\nDetermine whether the new entry is a likely semantic duplicate of any existing entry — same underlying observation, not necessarily the same wording. If so, set matched_entry_id to that entry's id; otherwise null.")

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"is_duplicate": map[string]interface{}{"type": "boolean"},
			"matched_entry_id": map[string]interface{}{
				"anyOf": []map[string]interface{}{
					{"type": "integer"},
					{"type": "null"},
				},
			},
			"similarity_note": map[string]interface{}{"type": "string"},
		},
		"required":             []string{"is_duplicate", "matched_entry_id", "similarity_note"},
		"additionalProperties": false,
	}

	resp, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-opus-5",
		MaxTokens: 512,
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: schema},
		},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(b.String()))},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from duplicate check")
	}
	textBlock, ok := resp.Content[0].AsAny().(anthropic.TextBlock)
	if !ok {
		return nil, fmt.Errorf("unexpected response content type from duplicate check")
	}

	var result DuplicateCheckResult
	if err := json.Unmarshal([]byte(textBlock.Text), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/aiclient/... -run TestCheckDuplicate -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/aiclient/duplicate.go backend/scout/internal/aiclient/duplicate_test.go
git commit -m "scout: add AI semantic duplicate-check client via structured output (F14)"
```

---

### Task 29: Duplicate-check endpoint — never blocks, degrades gracefully (F14, NF3)

**Files:**
- Create: `backend/scout/internal/handlers/duplicate_check.go`
- Create: `backend/scout/internal/handlers/duplicate_check_test.go`
- Modify: `backend/scout/cmd/server/router.go`

**Interfaces:**
- Consumes: `aiclient.Client.CheckDuplicate`, `aiclient.ExistingEntry`, `aiclient.DuplicateCheckResult` (Task 28); `store.HighlightStore.List` (Task 25)
- Produces: `handlers.NewDuplicateCheckHandler(aiClient *aiclient.Client, highlightStore *store.HighlightStore, engineerStore *store.EngineerStore) *DuplicateCheckHandler` with `.Check(c *gin.Context)` registered as `POST /api/engineers/:id/highlights/check-duplicate`

- [ ] **Step 1: Write the failing test**
```go
// backend/scout/internal/handlers/duplicate_check_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"
)

func TestDuplicateCheckHandlerReturnsFlagOnMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_dup2", "type": "message", "role": "assistant", "model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":true,\"matched_entry_id\":1,\"similarity_note\":\"same missed deadline\"}"}],
			"stop_reason": "end_turn", "stop_sequence": null,
			"usage": {"input_tokens": 5, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	highlightStore.Create(e1.ID, "lowlight", "Slipped the Q1 deadline and didn't raise it in standup")

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Missed the Q1 deadline without flagging it early"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if !result.IsDuplicate {
		t.Fatal("expected the duplicate flag to be true")
	}
}

func TestDuplicateCheckHandlerDegradesGracefullyOnAIFailure(t *testing.T) {
	// A server that always errors simulates the AI call failing (F14: the
	// save must proceed without a flag rather than being blocked).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Must still be 200 with no duplicate flag — never a blocking error.
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (degraded, not blocked), got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false when the AI call fails")
	}
}
```
- [ ] **Step 2: Run test to verify it fails**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestDuplicateCheckHandler -v`
Expected: FAIL — `undefined: NewDuplicateCheckHandler`
- [ ] **Step 3: Write minimal implementation**
```go
// backend/scout/internal/handlers/duplicate_check.go
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type DuplicateCheckHandler struct {
	aiClient       *aiclient.Client
	highlightStore *store.HighlightStore
	engineerStore  *store.EngineerStore
}

func NewDuplicateCheckHandler(aiClient *aiclient.Client, highlightStore *store.HighlightStore, engineerStore *store.EngineerStore) *DuplicateCheckHandler {
	return &DuplicateCheckHandler{aiClient: aiClient, highlightStore: highlightStore, engineerStore: engineerStore}
}

type duplicateCheckRequest struct {
	Body string `json:"body" binding:"required"`
}

// Check handles POST /api/engineers/:id/highlights/check-duplicate (F14).
// It runs synchronously before save and returns a similarity flag + matched
// entry, but never blocks: if the AI call fails or times out, this returns
// 200 with is_duplicate=false rather than an error, so the admin's save can
// proceed without a flag (NF3).
func (h *DuplicateCheckHandler) Check(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req duplicateCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}

	existingEntries, err := h.highlightStore.List(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load existing entries"})
		return
	}
	entries := make([]aiclient.ExistingEntry, 0, len(existingEntries))
	for _, e := range existingEntries {
		entries = append(entries, aiclient.ExistingEntry{ID: e.ID, Kind: e.Kind, Body: e.Body})
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	result, err := h.aiClient.CheckDuplicate(ctx, req.Body, entries)
	if err != nil {
		// Graceful degradation: proceed without a flag rather than blocking
		// the admin's save (F14, NF3).
		c.JSON(http.StatusOK, aiclient.DuplicateCheckResult{IsDuplicate: false, Note: "duplicate check unavailable, proceeding without a flag"})
		return
	}
	c.JSON(http.StatusOK, result)
}
```
```go
// backend/scout/cmd/server/router.go — add inside the `api` group
		duplicateCheckHandler := handlers.NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
		api.POST("/engineers/:id/highlights/check-duplicate", duplicateCheckHandler.Check)
```
- [ ] **Step 4: Run test to verify it passes**
Run: `cd backend/scout && go test ./internal/handlers/... -run TestDuplicateCheckHandler -v`
Expected: PASS
- [ ] **Step 5: Commit**
```bash
git add backend/scout/internal/handlers/duplicate_check.go backend/scout/internal/handlers/duplicate_check_test.go backend/scout/cmd/server/router.go
git commit -m "scout: add highlight duplicate-check endpoint with graceful AI degradation (F14, NF3)"
```

---

### Task 30: Backend README (documentation — no test cycle)

**Files:**
- Create: `backend/scout/README.md`

**Interfaces:**
- Consumes: nothing (documentation only)
- Produces: nothing (documentation only)

This task has no code and no red/green cycle — it documents what Tasks 1-29 built. Verification is a manual spot-check that every documented command and env var actually matches the code.

- [ ] **Step 1: Write the README**
```markdown
# Scout

A private Go REST API backing Scout: a FIFA-style engineer attribute-card
tracker with biweekly peer-relative ranking, synced GitHub/Jira metrics, an
AI ranking-chat assistant, and a manual highlight/lowlight log. Built with
Gin, following the same Standard Go Project Layout as `hustle-turtle` and
`oncarinho`.

## Features

- Shared-password admin auth with signed session cookies — **every** route
  requires a valid session except `POST /api/auth/login` and `GET /health`
  (no public application-data group, unlike `oncarinho`)
- Engineer roster CRUD (create, edit, deactivate, reactivate — soft delete only)
- Main/sub-attribute management, seeded with the initial 6 main attributes
- Biweekly rating cycles: open a cycle, submit a strict 1..N ranking per
  sub-attribute (no ties/gaps), converted to a score via linear interpolation
- Computed-on-read scoring: main-attribute score, Overall score, cycle view,
  engineer card + trend — never cached, always derived from stored rankings
- Roster dashboard: latest Overall + last cycle date per active engineer
- In-process ticker scheduler syncing GitHub PR stats and Jira ticket stats
  into `metric_snapshots`, idempotent per (engineer, period)
- AI ranking chat (Claude API, streamed) proposing a rank ordering with
  rationale; persisted only on explicit accept
- Highlight/lowlight log with an AI semantic duplicate check that never
  blocks a save

## Project Structure

```
scout/
├── cmd/
│   └── server/              # Entrypoint + full route wiring
│       ├── main.go
│       └── router.go
├── internal/
│   ├── auth/                # Session token signing/verification
│   ├── config/               # Configuration management
│   ├── database/             # Database connection and migrations
│   ├── handlers/              # HTTP handlers
│   ├── models/                # Data models
│   ├── store/                 # SQL access — one store per aggregate
│   ├── scoring/                # Pure functions: rank->score, permutation validation
│   ├── aiclient/                # Anthropic SDK wrapper (chat + duplicate check)
│   ├── integrations/             # GitHub + Jira API clients
│   └── syncer/                    # Sync orchestration + ticker scheduler
├── migrations/                     # Database migration files
├── go.mod
└── README.md
```

## Setup

```bash
go mod download
createdb scout
```

Environment variables:

```bash
export DATABASE_URL="postgres://username:password@localhost/scout?sslmode=disable"
export PORT="8082"
export ADMIN_PASSWORD="change-me"
export SESSION_SECRET="change-me-too"
export COOKIE_SECURE="false"                # true in production (HTTPS)
export ANTHROPIC_API_KEY="sk-ant-..."
export SCOUT_GITHUB_TOKEN="ghp_..."
export SCOUT_GITHUB_REPOS="org/repo-a,org/repo-b"
export SCOUT_JIRA_BASE_URL="https://your-org.atlassian.net"
export SCOUT_JIRA_EMAIL="manager@example.com"
export SCOUT_JIRA_API_TOKEN="..."
export SCOUT_JIRA_PROJECTS="ENG"
export SYNC_INTERVAL_HOURS="12"
```

**Defaults:**
- `DATABASE_URL`: `postgres://localhost/scout?sslmode=disable`
- `PORT`: `8082`
- `ADMIN_PASSWORD` / `SESSION_SECRET`: none — the server refuses to start without them
- `SYNC_INTERVAL_HOURS`: `12`
- `SCOUT_GITHUB_REPOS` / `SCOUT_JIRA_PROJECTS`: which repos/projects to sync are
  environment-configured rather than hardcoded (see the plan's judgment-calls
  section) — set them to your actual GitHub repos and Jira project keys

## Running

```bash
go build -o bin/scout ./cmd/server
go run ./cmd/server                 # without migrations
go run ./cmd/server -auto-migrate   # with automatic migrations
```

Server starts on port 8082 and immediately begins the sync scheduler
(one run at startup, then every `SYNC_INTERVAL_HOURS`).

## API Endpoints

### Unauthenticated
- `GET /health` — health check
- `POST /api/auth/login` — shared password → signed `scout_session` cookie

### Authenticated (every other route — 401 without a valid session)
- `GET/POST/PUT/DELETE /api/engineers[, /:id]`, `POST /api/engineers/:id/reactivate`
- `GET/POST/PUT /api/main-attributes[, /:id]`
- `GET/POST/PUT/DELETE /api/sub-attributes[, /:id]`
- `GET/POST /api/cycles`
- `PUT /api/cycles/:id/sub-attributes/:subId/ranking`
- `GET /api/cycles/:id/scores`
- `GET /api/engineers/:id/card?cycleId=`
- `GET /api/engineers/:id/trend`
- `GET /api/engineers/:id/metrics`
- `GET /api/dashboard`
- `POST /api/cycles/:id/ai-sessions`
- `POST /api/cycles/:id/ai-sessions/:sessionId/accept`
- `GET/POST /api/engineers/:id/highlights`
- `POST /api/engineers/:id/highlights/check-duplicate`

## Database Migrations

```bash
go run ./cmd/server -migrate=up
go run ./cmd/server -migrate=down
go run ./cmd/server -version
```

### Current Migrations
- `000001_create_engineers_table`
- `000002_create_main_attributes_and_sub_attributes` (seeds the initial 6 main attributes)
- `000003_create_rating_cycles_and_rankings`
- `000004_create_metric_snapshots_table`
- `000005_create_ai_ranking_sessions_table`
- `000006_create_highlight_entries_table`

## Testing

Tests require a running Postgres and a dedicated test database — without one,
`go test ./...` reports `ok` for every package because each test gracefully
skips (intentional, mirroring `oncarinho`'s convention — a green `go test
./...` does NOT by itself prove the tests ran).

```bash
createdb scout_test
go test ./...
```

To point at a non-default test database:
```bash
TEST_DATABASE_URL="postgres://localhost/scout_test?sslmode=disable" go test ./...
```

To confirm tests are actually running (not skipping):
```bash
go test ./... -v 2>&1 | grep -c '^--- SKIP'   # should print 0
```

**Run with `-p 1` for full reliability.** Unlike `oncarinho`'s tables, `main_attributes`
is migration-seeded (the initial 6, F3). Several store/handler tests `TRUNCATE` it
(and dependent tables) to seed their own fixtures, and Go runs different packages'
test binaries concurrently by default — a truncate in one package can race a
seed-dependent assertion in another (e.g. `TestMainAttributeStoreSeedData`). Force
serial package execution to eliminate that race:
```bash
go test -p 1 ./...
```

## Architecture Notes

- **Scoring is computed on read** (NF2): main-attribute, Overall, cycle-view,
  and trend scores are all derived via SQL aggregation over
  `sub_attribute_rankings` on every request — there is no cached/materialized
  score table, avoiding a second source of truth that could drift.
- **Overall score cutover** (F8): a main attribute counts toward a cycle's
  Overall only if `main_attributes.created_at <= rating_cycles.created_at` —
  adding a main attribute later never retroactively changes a past cycle.
- **Sync worker** (F4, NF4): an in-process ticker started from
  `cmd/server/main.go` (no separate `cmd/syncer` binary) polls GitHub/Jira for
  every active engineer and idempotently upserts `metric_snapshots` keyed on
  `(engineer_id, period_start, period_end)`. A single engineer's fetch
  failure is logged and skipped; the next scheduled run retries it safely.
- **AI ranking chat** (F9, NF3): the Anthropic Go SDK streams the assistant's
  reply as Server-Sent Events. The reply's rationale and a trailing fenced
  ` ```json ` block carrying the proposed ranking are both stored on the
  session — the ranking is written to `sub_attribute_rankings` only via the
  separate accept endpoint, on explicit admin confirm.
- **Duplicate check** (F14, NF3): uses the Anthropic Go SDK's structured
  JSON-schema output for a guaranteed-parseable response. If the AI call
  fails or times out, the endpoint returns `200` with `is_duplicate: false`
  rather than an error — the admin's save is never blocked.
```
- [ ] **Step 2: Spot-check the README against the running code**
Run: `cd backend/scout && grep -c "api\." cmd/server/router.go` (confirm the route count roughly matches the endpoints documented above) and `go run ./cmd/server -version` (confirm the migration list matches)
Expected: README's endpoint and migration lists match the actual router and `migrations/` directory
- [ ] **Step 3: Commit**
```bash
git add backend/scout/README.md
git commit -m "scout: add backend README documenting setup, API surface, and architecture"
```

---

## Frontend Tasks (`apps/scout`)

Per the Global Constraints' testing convention, this repo's Next.js apps have no test framework — frontend tasks below replace the red/green TDD cycle with a **write → type-check/build → manual-verify → commit** cycle. "Manual verify" means running `pnpm --filter scout dev` and checking the page in a browser; this is noted per task rather than automated.

### Task 31: `apps/scout` Next.js zone scaffold (basePath, theme, Nginx route)

**Files:**
- Create: `apps/scout/package.json`
- Create: `apps/scout/next.config.mjs`
- Create: `apps/scout/tsconfig.json`
- Create: `apps/scout/tailwind.config.ts`
- Create: `apps/scout/postcss.config.mjs`
- Create: `apps/scout/src/app/globals.css`
- Create: `apps/scout/src/app/layout.tsx`
- Modify: `infra/nginx/nginx.conf`

**Interfaces:**
- Produces: the `apps/scout` Next.js 14 zone, served at basePath `/scout`, consuming `@movoz/tailwind-config` + `@movoz/theme`, with `/api/*` rewritten in dev to the Scout backend (`http://localhost:8082`) — every subsequent frontend task builds pages inside this zone

- [ ] **Step 1: Write the scaffold**
```json
// apps/scout/package.json
{
  "name": "@movoz/scout",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev -p 3002",
    "build": "next build",
    "start": "next start -p 3002",
    "lint": "next lint"
  },
  "dependencies": {
    "@movoz/theme": "workspace:*",
    "next": "catalog:",
    "react": "catalog:",
    "react-dom": "catalog:"
  },
  "devDependencies": {
    "@movoz/tailwind-config": "workspace:*",
    "@movoz/tsconfig": "workspace:*",
    "@types/node": "catalog:",
    "@types/react": "catalog:",
    "@types/react-dom": "catalog:",
    "autoprefixer": "catalog:",
    "eslint": "catalog:",
    "eslint-config-next": "catalog:",
    "postcss": "catalog:",
    "tailwindcss": "catalog:",
    "typescript": "catalog:"
  }
}
```
```js
// apps/scout/next.config.mjs
/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: "/scout",
  assetPrefix: "/scout",
  output: "standalone",
  transpilePackages: ["@movoz/theme", "@movoz/tailwind-config"],
  async rewrites() {
    // Dev-only: proxy API calls to the Scout backend so client code can
    // just call same-origin "/api/*" (cookies stay first-party). In
    // production, Nginx performs the equivalent proxy — see the
    // infra/nginx/nginx.conf change in this task.
    const apiOrigin = process.env.SCOUT_API_URL || "http://localhost:8082";
    return [{ source: "/api/:path*", destination: `${apiOrigin}/api/:path*` }];
  },
};

export default nextConfig;
```
```json
// apps/scout/tsconfig.json
{
  "extends": "@movoz/tsconfig/nextjs.json",
  "compilerOptions": {
    "plugins": [{ "name": "next" }],
    "paths": { "@/*": ["./src/*"] },
    "target": "ES2017"
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```
```ts
// apps/scout/tailwind.config.ts
import type { Config } from "tailwindcss";
import sharedConfig from "@movoz/tailwind-config";

const config: Config = {
  presets: [sharedConfig as Config],
  content: [
    "./src/**/*.{js,ts,jsx,tsx,mdx}",
    "../../packages/theme/src/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  plugins: [],
};

export default config;
```
```js
// apps/scout/postcss.config.mjs
/** @type {import('postcss-load-config').Config} */
const config = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
};

export default config;
```
```css
/* apps/scout/src/app/globals.css */
@import "@movoz/theme/globals.css";

@tailwind base;
@tailwind components;
@tailwind utilities;

* {
  box-sizing: border-box;
  padding: 0;
  margin: 0;
}

body {
  background-color: var(--zen-bg);
  color: var(--zen-text);
  font-family: system-ui, sans-serif;
  transition: background-color 0.3s ease, color 0.3s ease;
}
```
```tsx
// apps/scout/src/app/layout.tsx
import type { Metadata } from "next";
import "./globals.css";
import { ThemeProvider } from "@movoz/theme";

export const metadata: Metadata = {
  title: "Scout",
  description: "Engineer performance tracking — FIFA-style attribute cards, biweekly rankings, and AI-assisted reviews.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="antialiased">
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
```
```nginx
# infra/nginx/nginx.conf — add upstream + location block
upstream scout_frontend {
    server scout:3002;
}

# ... inside the existing `server {}` block, alongside the personal_site location:
    location /scout {
        proxy_pass http://scout_frontend;
    }
    location /scout/_next {
        proxy_pass http://scout_frontend;
    }
    location /scout/api {
        proxy_pass http://scout_backend;
    }

# ... alongside the existing `upstream personal_site` block:
upstream scout_backend {
    server scout-backend:8082;
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm install && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both commands succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` and open `http://localhost:3002/scout` — expect a blank themed page (no crash) at this stage, since no route pages exist yet beyond the root layout
- [ ] **Step 4: Commit**
```bash
git add apps/scout/package.json apps/scout/next.config.mjs apps/scout/tsconfig.json apps/scout/tailwind.config.ts apps/scout/postcss.config.mjs apps/scout/src/app/globals.css apps/scout/src/app/layout.tsx infra/nginx/nginx.conf pnpm-lock.yaml
git commit -m "scout: scaffold apps/scout Next.js zone (basePath, theme, Nginx route)"
```

---

### Task 32: API client wrapper, login page, auth-gated middleware (F1, NF1)

**Files:**
- Create: `apps/scout/src/lib/api.ts`
- Create: `apps/scout/src/app/login/page.tsx`
- Create: `apps/scout/src/middleware.ts`

**Interfaces:**
- Produces: `api.get<T>(path)`, `api.post<T>(path, body?)`, `api.put<T>(path, body?)`, `api.delete<T>(path)`, `APIError{status: number, message: string}` — every later frontend task's data fetching goes through this module

- [ ] **Step 1: Write the API client, login page, and middleware**
```ts
// apps/scout/src/lib/api.ts
export class APIError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`/scout${path}`, {
    ...options,
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options.headers },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new APIError(res.status, body.error || `Request failed with status ${res.status}`);
  }
  if (res.status === 204) {
    return undefined as T;
  }
  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body: body !== undefined ? JSON.stringify(body) : undefined }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PUT", body: body !== undefined ? JSON.stringify(body) : undefined }),
  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
```
```tsx
// apps/scout/src/app/login/page.tsx
"use client";

import { useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import { api, APIError } from "@/lib/api";

export default function LoginPage() {
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const router = useRouter();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await api.post("/api/auth/login", { password });
      router.push("/");
      router.refresh();
    } catch (err) {
      setError(err instanceof APIError ? err.message : "Login failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-zen-bg">
      <form onSubmit={handleSubmit} className="w-full max-w-sm space-y-4 rounded-lg border border-zen-border bg-paper p-8">
        <h1 className="text-xl font-semibold text-zen-text">Scout</h1>
        <p className="text-sm text-zen-muted">Enter the shared password to continue.</p>
        <input
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full rounded border border-zen-border bg-transparent px-3 py-2 text-zen-text"
          placeholder="Password"
          autoFocus
        />
        {error && <p className="text-sm text-red-500">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded bg-accent-600 px-3 py-2 text-white disabled:opacity-50"
        >
          {submitting ? "Signing in..." : "Sign in"}
        </button>
      </form>
    </main>
  );
}
```
```ts
// apps/scout/src/middleware.ts
import { NextRequest, NextResponse } from "next/server";

// Scout's "no public route group" rule (NF1) applies to the frontend too:
// every page redirects to /login unless a session cookie is present. This
// is a UX redirect only — the backend's RequireAuth middleware is the
// authoritative gate (this middleware only checks cookie *presence*, not
// signature validity, since it has no access to SESSION_SECRET).
const SESSION_COOKIE = "scout_session";

export function middleware(request: NextRequest) {
  const isLoginPage = request.nextUrl.pathname === "/login";
  const hasSession = request.cookies.has(SESSION_COOKIE);

  if (!hasSession && !isLoginPage) {
    const loginUrl = request.nextUrl.clone();
    loginUrl.pathname = "/login";
    return NextResponse.redirect(loginUrl);
  }
  if (hasSession && isLoginPage) {
    const homeUrl = request.nextUrl.clone();
    homeUrl.pathname = "/";
    return NextResponse.redirect(homeUrl);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` (with the backend running on port 8082) and open `http://localhost:3002/scout` — expect an automatic redirect to `/scout/login`; entering the correct `ADMIN_PASSWORD` should redirect back to `/scout` (a blank page is fine at this stage — the dashboard page doesn't exist yet)
- [ ] **Step 4: Commit**
```bash
git add apps/scout/src/lib/api.ts apps/scout/src/app/login/page.tsx apps/scout/src/middleware.ts
git commit -m "scout: add API client wrapper, login page, and auth-gated middleware (F1, NF1)"
```

---

### Task 33: Shared API types + roster management UI (F2)

**Files:**
- Create: `apps/scout/src/lib/types.ts`
- Create: `apps/scout/src/app/engineers/page.tsx`

**Interfaces:**
- Consumes: `api.get`/`api.post`/`api.put`/`api.delete` (Task 32)
- Produces: TypeScript interfaces `Engineer`, `MainAttribute`, `SubAttribute`, `RatingCycle`, `SubAttributeRanking`, `MainAttributeScore`, `EngineerCard`, `TrendPoint`, `EngineerCycleScore`, `RosterEntry`, `MetricSnapshot`, `HighlightEntry` — every later frontend task imports its shapes from this one file, matching the backend's JSON field names exactly

- [ ] **Step 1: Write the shared types and the roster page**
```ts
// apps/scout/src/lib/types.ts
export interface Engineer {
  id: number;
  name: string;
  role: string | null;
  github_username: string | null;
  jira_account_id: string | null;
  started_at: string;
  is_active: boolean;
  created_at: string;
}

export interface MainAttribute {
  id: number;
  key: string;
  name: string;
  created_at: string;
}

export interface SubAttribute {
  id: number;
  main_attribute_id: number;
  name: string;
  description: string | null;
  is_active: boolean;
  created_at: string;
}

export interface RatingCycle {
  id: number;
  period_start: string;
  period_end: string;
  created_at: string;
}

export interface SubAttributeRanking {
  id: number;
  cycle_id: number;
  sub_attribute_id: number;
  engineer_id: number;
  rank: number;
  score: number;
}

export interface MainAttributeScore {
  main_attribute_id: number;
  key: string;
  name: string;
  score: number;
}

export interface EngineerCard {
  engineer: Engineer;
  cycle_id: number;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface TrendPoint {
  cycle_id: number;
  period_start: string;
  period_end: string;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface EngineerCycleScore {
  engineer: Engineer;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface RosterEntry {
  engineer: Engineer;
  latest_overall: number | null;
  last_cycle_date: string | null;
}

export interface MetricSnapshot {
  id: number;
  engineer_id: number;
  period_start: string;
  period_end: string;
  prs_raised: number;
  prs_reviewed: number;
  tickets_closed: number;
  complexity_score: number;
  synced_at: string;
}

export interface HighlightEntry {
  id: number;
  engineer_id: number;
  kind: "highlight" | "lowlight";
  body: string;
  created_at: string;
}
```
```tsx
// apps/scout/src/app/engineers/page.tsx
"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { Engineer } from "@/lib/types";

export default function EngineersPage() {
  const [engineers, setEngineers] = useState<Engineer[]>([]);
  const [showAll, setShowAll] = useState(false);
  const [name, setName] = useState("");
  const [role, setRole] = useState("");
  const [githubUsername, setGithubUsername] = useState("");
  const [jiraAccountId, setJiraAccountId] = useState("");
  const [startedAt, setStartedAt] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const list = await api.get<Engineer[]>(`/api/engineers?active=${showAll ? "all" : "true"}`);
    setEngineers(list);
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showAll]);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.post("/api/engineers", {
        name,
        role: role || null,
        github_username: githubUsername || null,
        jira_account_id: jiraAccountId || null,
        started_at: startedAt,
      });
      setName("");
      setRole("");
      setGithubUsername("");
      setJiraAccountId("");
      setStartedAt("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to add engineer");
    }
  }

  async function toggleActive(engineer: Engineer) {
    if (engineer.is_active) {
      await api.delete(`/api/engineers/${engineer.id}`);
    } else {
      await api.post(`/api/engineers/${engineer.id}/reactivate`);
    }
    await load();
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Roster</h1>

      <form onSubmit={handleCreate} className="mb-8 grid grid-cols-2 gap-3 rounded-lg border border-zen-border p-4">
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Role" value={role} onChange={(e) => setRole(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="GitHub username" value={githubUsername} onChange={(e) => setGithubUsername(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" placeholder="Jira account ID" value={jiraAccountId} onChange={(e) => setJiraAccountId(e.target.value)} />
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={startedAt} onChange={(e) => setStartedAt(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Add engineer</button>
        {error && <p className="col-span-2 text-sm text-red-500">{error}</p>}
      </form>

      <label className="mb-3 flex items-center gap-2 text-sm text-zen-muted">
        <input type="checkbox" checked={showAll} onChange={(e) => setShowAll(e.target.checked)} />
        Show deactivated engineers
      </label>

      <ul className="divide-y divide-zen-border">
        {engineers.map((engineer) => (
          <li key={engineer.id} className="flex items-center justify-between py-3">
            <div>
              <Link href={`/engineers/${engineer.id}`} className="font-medium text-zen-text hover:underline">
                {engineer.name}
              </Link>
              <span className="ml-2 text-sm text-zen-muted">{engineer.role}</span>
              {!engineer.is_active && <span className="ml-2 text-xs text-red-500">deactivated</span>}
            </div>
            <button onClick={() => toggleActive(engineer)} className="text-sm text-zen-muted hover:text-zen-text">
              {engineer.is_active ? "Deactivate" : "Reactivate"}
            </button>
          </li>
        ))}
      </ul>
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` and open `http://localhost:3002/scout/engineers` — add an engineer, confirm it appears in the list, deactivate it, confirm it disappears from the default (active-only) view and reappears with "Show deactivated engineers" checked
- [ ] **Step 4: Commit**
```bash
git add apps/scout/src/lib/types.ts apps/scout/src/app/engineers/page.tsx
git commit -m "scout: add shared API types and roster management UI (F2)"
```

---

### Task 34: Main/sub-attribute management UI (F3)

**Files:**
- Create: `apps/scout/src/app/attributes/page.tsx`

**Interfaces:**
- Consumes: `api`, `MainAttribute`, `SubAttribute` (Tasks 32, 33)

- [ ] **Step 1: Write the attribute management page**
```tsx
// apps/scout/src/app/attributes/page.tsx
"use client";

import { useEffect, useState, type FormEvent } from "react";
import { api } from "@/lib/api";
import type { MainAttribute, SubAttribute } from "@/lib/types";

export default function AttributesPage() {
  const [mainAttributes, setMainAttributes] = useState<MainAttribute[]>([]);
  const [subAttributesByMain, setSubAttributesByMain] = useState<Record<number, SubAttribute[]>>({});
  const [newMainKey, setNewMainKey] = useState("");
  const [newMainName, setNewMainName] = useState("");
  const [newSubName, setNewSubName] = useState<Record<number, string>>({});

  async function load() {
    const mains = await api.get<MainAttribute[]>("/api/main-attributes");
    setMainAttributes(mains);
    const entries = await Promise.all(
      mains.map(async (m) => [m.id, await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${m.id}&active=all`)] as const)
    );
    setSubAttributesByMain(Object.fromEntries(entries));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreateMain(e: FormEvent) {
    e.preventDefault();
    await api.post("/api/main-attributes", { key: newMainKey, name: newMainName });
    setNewMainKey("");
    setNewMainName("");
    await load();
  }

  async function handleCreateSub(mainAttributeId: number, e: FormEvent) {
    e.preventDefault();
    const name = newSubName[mainAttributeId];
    if (!name) return;
    await api.post("/api/sub-attributes", { main_attribute_id: mainAttributeId, name });
    setNewSubName((prev) => ({ ...prev, [mainAttributeId]: "" }));
    await load();
  }

  async function toggleSubActive(sub: SubAttribute) {
    if (sub.is_active) {
      await api.delete(`/api/sub-attributes/${sub.id}`);
      await load();
    }
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Attributes</h1>

      <form onSubmit={handleCreateMain} className="mb-8 flex gap-3 rounded-lg border border-zen-border p-4">
        <input className="flex-1 rounded border border-zen-border bg-transparent p-2" placeholder="key (e.g. delivery_speed)" value={newMainKey} onChange={(e) => setNewMainKey(e.target.value)} required />
        <input className="flex-1 rounded border border-zen-border bg-transparent p-2" placeholder="Name (e.g. Delivery Speed)" value={newMainName} onChange={(e) => setNewMainName(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Add main attribute</button>
      </form>

      <div className="space-y-6">
        {mainAttributes.map((main) => (
          <section key={main.id} className="rounded-lg border border-zen-border p-4">
            <h2 className="mb-3 text-lg font-medium text-zen-text">{main.name}</h2>
            <ul className="mb-3 divide-y divide-zen-border">
              {(subAttributesByMain[main.id] ?? []).map((sub) => (
                <li key={sub.id} className="flex items-center justify-between py-2">
                  <span className={sub.is_active ? "text-zen-text" : "text-zen-muted line-through"}>{sub.name}</span>
                  {sub.is_active && (
                    <button onClick={() => toggleSubActive(sub)} className="text-sm text-zen-muted hover:text-zen-text">
                      Deactivate
                    </button>
                  )}
                </li>
              ))}
            </ul>
            <form onSubmit={(e) => handleCreateSub(main.id, e)} className="flex gap-2">
              <input
                className="flex-1 rounded border border-zen-border bg-transparent p-2 text-sm"
                placeholder="New sub-attribute name"
                value={newSubName[main.id] ?? ""}
                onChange={(e) => setNewSubName((prev) => ({ ...prev, [main.id]: e.target.value }))}
              />
              <button type="submit" className="rounded bg-accent-600 px-3 py-1 text-sm text-white">Add</button>
            </form>
          </section>
        ))}
      </div>
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` and open `http://localhost:3002/scout/attributes` — confirm the seeded 6 main attributes render, add a sub-attribute under one, confirm it appears, deactivate it, confirm it shows struck through
- [ ] **Step 4: Commit**
```bash
git add apps/scout/src/app/attributes/page.tsx
git commit -m "scout: add main/sub-attribute management UI (F3)"
```

---

### Task 35: Rating cycle list/create + ranking UI (F6)

**Files:**
- Create: `apps/scout/src/app/cycles/page.tsx`
- Create: `apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx`

**Interfaces:**
- Consumes: `api`, `RatingCycle`, `Engineer`, `MainAttribute`, `SubAttribute`, `SubAttributeRanking` (Tasks 32, 33); `GET/PUT /api/cycles/:id/sub-attributes/:subId/ranking` (Task 12, including the `.Get` endpoint added above)

Ranking is assigned via explicit numeric rank inputs (1..N) per engineer rather than drag-to-reorder — the spec allows either; explicit assignment is simpler to validate client-side and is the judgment call flagged for review.

- [ ] **Step 1: Write the cycle list/create page and the ranking page**
```tsx
// apps/scout/src/app/cycles/page.tsx
"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { RatingCycle } from "@/lib/types";

export default function CyclesPage() {
  const [cycles, setCycles] = useState<RatingCycle[]>([]);
  const [periodStart, setPeriodStart] = useState("");
  const [periodEnd, setPeriodEnd] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function load() {
    setCycles(await api.get<RatingCycle[]>("/api/cycles"));
  }

  useEffect(() => {
    load();
  }, []);

  async function handleCreate(e: FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      await api.post("/api/cycles", { period_start: periodStart, period_end: periodEnd });
      setPeriodStart("");
      setPeriodEnd("");
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create cycle");
    }
  }

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Rating Cycles</h1>

      <form onSubmit={handleCreate} className="mb-8 flex gap-3 rounded-lg border border-zen-border p-4">
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={periodStart} onChange={(e) => setPeriodStart(e.target.value)} required />
        <input className="rounded border border-zen-border bg-transparent p-2" type="date" value={periodEnd} onChange={(e) => setPeriodEnd(e.target.value)} required />
        <button type="submit" className="rounded bg-accent-600 px-3 py-2 text-white">Open cycle</button>
      </form>
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <ul className="divide-y divide-zen-border">
        {cycles.map((cycle) => (
          <li key={cycle.id} className="flex items-center justify-between py-3">
            <span className="text-zen-text">{cycle.period_start} — {cycle.period_end}</span>
            <Link href={`/cycles/${cycle.id}`} className="text-sm text-accent-600 hover:underline">
              View scores
            </Link>
          </li>
        ))}
      </ul>
    </main>
  );
}
```
```tsx
// apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx
"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import type { Engineer, MainAttribute, SubAttribute, SubAttributeRanking } from "@/lib/types";

export default function RankSubAttributePage() {
  const params = useParams<{ id: string; subId: string }>();
  const cycleId = Number(params.id);
  const subAttributeId = Number(params.subId);

  const [engineers, setEngineers] = useState<Engineer[]>([]);
  const [subAttributeName, setSubAttributeName] = useState("");
  const [ranks, setRanks] = useState<Record<number, number>>({});
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    async function load() {
      const activeEngineers = await api.get<Engineer[]>("/api/engineers");
      setEngineers(activeEngineers);

      const mains = await api.get<MainAttribute[]>("/api/main-attributes");
      for (const main of mains) {
        const subs = await api.get<SubAttribute[]>(`/api/sub-attributes?main_attribute_id=${main.id}&active=all`);
        const match = subs.find((s) => s.id === subAttributeId);
        if (match) {
          setSubAttributeName(match.name);
          break;
        }
      }

      const existing = await api.get<SubAttributeRanking[]>(
        `/api/cycles/${cycleId}/sub-attributes/${subAttributeId}/ranking`
      );
      const initialRanks: Record<number, number> = {};
      existing.forEach((r) => {
        initialRanks[r.engineer_id] = r.rank;
      });
      setRanks(initialRanks);
    }
    load();
  }, [cycleId, subAttributeId]);

  const usedRanks = useMemo(() => Object.values(ranks), [ranks]);
  const hasDuplicateRank = new Set(usedRanks).size !== usedRanks.length;
  const allRanked = engineers.length > 0 && engineers.every((e) => ranks[e.id] != null);

  function setRank(engineerId: number, rank: number) {
    setRanks((prev) => ({ ...prev, [engineerId]: rank }));
    setSaved(false);
  }

  async function handleSubmit() {
    setError(null);
    setSaved(false);
    try {
      const rankings = engineers.map((e) => ({ engineer_id: e.id, rank: ranks[e.id] }));
      await api.put(`/api/cycles/${cycleId}/sub-attributes/${subAttributeId}/ranking`, { rankings });
      setSaved(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save ranking");
    }
  }

  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-2 text-2xl font-semibold text-zen-text">
        Rank: {subAttributeName || `Sub-attribute #${subAttributeId}`}
      </h1>
      <p className="mb-6 text-sm text-zen-muted">
        Assign each active engineer a unique rank from 1 (best) to {engineers.length} (last) — no ties. Use the AI
        chat assistant (below, once built) to get a starting proposal, then adjust here before saving.
      </p>

      <ul className="mb-6 space-y-2">
        {engineers.map((engineer) => (
          <li key={engineer.id} className="flex items-center justify-between rounded border border-zen-border p-3">
            <span className="text-zen-text">{engineer.name}</span>
            <input
              type="number"
              min={1}
              max={engineers.length}
              value={ranks[engineer.id] ?? ""}
              onChange={(e) => setRank(engineer.id, Number(e.target.value))}
              className="w-16 rounded border border-zen-border bg-transparent p-1 text-center"
            />
          </li>
        ))}
      </ul>

      {hasDuplicateRank && (
        <p className="mb-4 text-sm text-red-500">
          Two engineers share the same rank — ranks must be unique 1..{engineers.length}.
        </p>
      )}
      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}
      {saved && <p className="mb-4 text-sm text-green-600">Ranking saved.</p>}

      <button
        onClick={handleSubmit}
        disabled={!allRanked || hasDuplicateRank}
        className="rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
      >
        Save ranking
      </button>
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev`, open `http://localhost:3002/scout/cycles`, open a cycle, create one if none exist, then navigate to `/scout/cycles/<id>/sub-attributes/<subId>` for a real sub-attribute id — assign unique ranks to every active engineer, confirm the Save button is disabled until all are ranked with no ties, save, reload the page, and confirm the saved ranks pre-populate
- [ ] **Step 4: Commit**
```bash
git add apps/scout/src/app/cycles/page.tsx "apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx"
git commit -m "scout: add rating cycle list/create and ranking UI (F6)"
```

---

### Task 36: AI ranking chat UI, streamed (F9, NF3)

**Files:**
- Create: `apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/chat/page.tsx`
- Modify: `apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx`

**Interfaces:**
- Consumes: `POST /api/cycles/:id/ai-sessions` (SSE stream, Task 26), `POST /api/cycles/:id/ai-sessions/:sessionId/accept` (Task 27)

The SSE stream is read with the raw Fetch `ReadableStream` API (not the `api` wrapper, which assumes a single JSON response) — frames are split on `"\n\n"`, and a leading `event: session` frame carries `{"session_id": N}` while every other frame is a plain `data: ` text chunk, per the wire contract documented in Task 26's handler comment.

- [ ] **Step 1: Write the chat page and link to it from the ranking page**
```tsx
// apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/chat/page.tsx
"use client";

import { useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { api } from "@/lib/api";

interface ChatTurn {
  role: "user" | "assistant";
  content: string;
}

interface ProposedRankingEntry {
  engineer_id: number;
  rank: number;
}

function extractJSONBlock(text: string): { rationale?: string; ranking: ProposedRankingEntry[] } | null {
  const matches = [...text.matchAll(/```json\s*([\s\S]*?)\s*```/g)];
  if (matches.length === 0) return null;
  try {
    return JSON.parse(matches[matches.length - 1][1]);
  } catch {
    return null;
  }
}

export default function AIRankingChatPage() {
  const params = useParams<{ id: string; subId: string }>();
  const cycleId = Number(params.id);
  const subAttributeId = Number(params.subId);
  const router = useRouter();

  const [sessionId, setSessionId] = useState<number | null>(null);
  const [turns, setTurns] = useState<ChatTurn[]>([]);
  const [input, setInput] = useState("");
  const [streaming, setStreaming] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [proposedRanking, setProposedRanking] = useState<ProposedRankingEntry[] | null>(null);
  const [accepting, setAccepting] = useState(false);
  const assistantBufferRef = useRef("");

  async function sendMessage() {
    if (!input.trim() || streaming) return;
    const userMessage = input;
    setInput("");
    setError(null);
    setTurns((prev) => [...prev, { role: "user", content: userMessage }, { role: "assistant", content: "" }]);
    setStreaming(true);
    assistantBufferRef.current = "";

    try {
      const res = await fetch(`/scout/api/cycles/${cycleId}/ai-sessions`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ session_id: sessionId, sub_attribute_id: subAttributeId, message: userMessage }),
      });
      if (!res.body) throw new Error("no response body");

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });

        const frames = buffer.split("\n\n");
        buffer = frames.pop() ?? "";

        for (const frame of frames) {
          const lines = frame.split("\n");
          const eventLine = lines.find((l) => l.startsWith("event: "));
          const dataLine = lines.find((l) => l.startsWith("data: "));
          if (!dataLine) continue;
          const data = dataLine.slice("data: ".length);

          if (eventLine?.includes("session")) {
            setSessionId(JSON.parse(data).session_id);
          } else if (eventLine?.includes("error")) {
            setError(JSON.parse(data).error);
          } else {
            assistantBufferRef.current += data.replace(/\\n/g, "\n");
            const snapshot = assistantBufferRef.current;
            setTurns((prev) => {
              const next = [...prev];
              next[next.length - 1] = { role: "assistant", content: snapshot };
              return next;
            });
          }
        }
      }

      const block = extractJSONBlock(assistantBufferRef.current);
      if (block) {
        setProposedRanking(block.ranking);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Chat failed");
    } finally {
      setStreaming(false);
    }
  }

  function updateProposedRank(engineerId: number, rank: number) {
    setProposedRanking((prev) => (prev ?? []).map((entry) => (entry.engineer_id === engineerId ? { ...entry, rank } : entry)));
  }

  async function acceptRanking() {
    if (!sessionId || !proposedRanking) return;
    setAccepting(true);
    setError(null);
    try {
      await api.post(`/api/cycles/${cycleId}/ai-sessions/${sessionId}/accept`, { rankings: proposedRanking });
      router.push(`/cycles/${cycleId}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to accept ranking");
    } finally {
      setAccepting(false);
    }
  }

  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">AI Ranking Assistant</h1>

      <div className="mb-4 max-h-96 space-y-3 overflow-y-auto rounded-lg border border-zen-border p-4">
        {turns.length === 0 && (
          <p className="text-sm text-zen-muted">
            Describe what you observed this cycle — who stood out, who struggled — and the assistant will propose a
            ranking with rationale.
          </p>
        )}
        {turns.map((turn, i) => (
          <p key={i} className={turn.role === "user" ? "text-zen-text" : "text-zen-muted"}>
            <strong>{turn.role === "user" ? "You" : "Assistant"}:</strong> {turn.content}
          </p>
        ))}
      </div>

      {proposedRanking && (
        <div className="mb-4 rounded-lg border border-accent-600 p-4">
          <h2 className="mb-2 font-medium text-zen-text">Proposed ranking (edit before accepting)</h2>
          <ul className="space-y-2">
            {proposedRanking.map((entry) => (
              <li key={entry.engineer_id} className="flex items-center justify-between">
                <span className="text-zen-text">Engineer #{entry.engineer_id}</span>
                <input
                  type="number"
                  min={1}
                  max={proposedRanking.length}
                  value={entry.rank}
                  onChange={(e) => updateProposedRank(entry.engineer_id, Number(e.target.value))}
                  className="w-16 rounded border border-zen-border bg-transparent p-1 text-center"
                />
              </li>
            ))}
          </ul>
          {/* NF3: the ranking is only persisted when the admin explicitly clicks
              this button — nothing here is auto-applied from the chat. */}
          <button
            onClick={acceptRanking}
            disabled={accepting}
            className="mt-3 rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
          >
            {accepting ? "Saving..." : "Accept & save ranking"}
          </button>
        </div>
      )}

      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <div className="flex gap-2">
        <input
          className="flex-1 rounded border border-zen-border bg-transparent p-2"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && sendMessage()}
          placeholder="Describe this cycle's observations..."
          disabled={streaming}
        />
        <button
          onClick={sendMessage}
          disabled={streaming}
          className="rounded bg-accent-600 px-4 py-2 text-white disabled:opacity-50"
        >
          {streaming ? "..." : "Send"}
        </button>
      </div>
    </main>
  );
}
```
```tsx
// apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx — add a link,
// e.g. near the top of the returned <main>:
      <Link
        href={`/cycles/${cycleId}/sub-attributes/${subAttributeId}/chat`}
        className="mb-4 inline-block text-sm text-accent-600 hover:underline"
      >
        Open AI ranking assistant →
      </Link>
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` (with `ANTHROPIC_API_KEY` set on the backend), open `/scout/cycles/<id>/sub-attributes/<subId>/chat`, send a message describing observations, confirm the reply streams in token-by-token, confirm a proposed-ranking panel appears once the assistant includes a ```json block, edit a rank, click Accept, and confirm the cycle's ranking endpoint now reflects it
- [ ] **Step 4: Commit**
```bash
git add "apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/chat/page.tsx" "apps/scout/src/app/cycles/[id]/sub-attributes/[subId]/page.tsx"
git commit -m "scout: add streamed AI ranking chat UI with explicit accept-to-save (F9, NF3)"
```

---

### Task 37: Engineer card page + trend visualization (F10)

**Files:**
- Create: `apps/scout/src/app/engineers/[id]/page.tsx`

**Interfaces:**
- Consumes: `api`, `Engineer`, `RatingCycle`, `EngineerCard`, `TrendPoint` (Tasks 32, 33); `GET /api/engineers/:id` (Task 3, added above), `GET /api/cycles`, `GET /api/engineers/:id/card`, `GET /api/engineers/:id/trend` (Tasks 11, 14)

No charting library is added (per the Tech Stack's no-new-frontend-dependencies scope) — the trend is a hand-rolled inline SVG polyline. Tasks 40 and 41 extend this same page with a metrics panel and a highlights/lowlights section.

- [ ] **Step 1: Write the engineer card page**
```tsx
// apps/scout/src/app/engineers/[id]/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import type { Engineer, EngineerCard as EngineerCardData, RatingCycle, TrendPoint } from "@/lib/types";

export default function EngineerCardPage() {
  const params = useParams<{ id: string }>();
  const engineerId = Number(params.id);

  const [engineer, setEngineer] = useState<Engineer | null>(null);
  const [cycles, setCycles] = useState<RatingCycle[]>([]);
  const [selectedCycleId, setSelectedCycleId] = useState<number | null>(null);
  const [card, setCard] = useState<EngineerCardData | null>(null);
  const [trend, setTrend] = useState<TrendPoint[]>([]);

  useEffect(() => {
    async function loadStatic() {
      const [eng, cycleList, trendData] = await Promise.all([
        api.get<Engineer>(`/api/engineers/${engineerId}`),
        api.get<RatingCycle[]>("/api/cycles"),
        api.get<TrendPoint[]>(`/api/engineers/${engineerId}/trend`),
      ]);
      setEngineer(eng);
      setCycles(cycleList);
      setTrend(trendData);
      if (cycleList.length > 0) setSelectedCycleId(cycleList[0].id);
    }
    loadStatic();
  }, [engineerId]);

  useEffect(() => {
    if (selectedCycleId == null) return;
    api.get<EngineerCardData>(`/api/engineers/${engineerId}/card?cycleId=${selectedCycleId}`).then(setCard);
  }, [engineerId, selectedCycleId]);

  if (!engineer) return null;

  const scoredPoints = trend.filter((t) => t.overall != null);
  const points = scoredPoints
    .map((t, i) => {
      const x = scoredPoints.length > 1 ? (i / (scoredPoints.length - 1)) * 300 : 150;
      const y = 100 - (((t.overall as number) - 50) / 50) * 100;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-1 text-2xl font-semibold text-zen-text">{engineer.name}</h1>
      <p className="mb-6 text-sm text-zen-muted">{engineer.role}</p>

      <div className="mb-6 flex items-center gap-3">
        <label className="text-sm text-zen-muted">Cycle:</label>
        <select
          value={selectedCycleId ?? ""}
          onChange={(e) => setSelectedCycleId(Number(e.target.value))}
          className="rounded border border-zen-border bg-transparent p-2"
        >
          {cycles.map((c) => (
            <option key={c.id} value={c.id}>
              {c.period_start} — {c.period_end}
            </option>
          ))}
        </select>
      </div>

      {card && (
        <section className="mb-8 rounded-lg border border-zen-border p-4">
          <p className="mb-3 text-lg text-zen-text">
            Overall: <strong>{card.overall != null ? card.overall.toFixed(1) : "—"}</strong>
          </p>
          <ul className="space-y-1">
            {card.main_attributes.map((m) => (
              <li key={m.main_attribute_id} className="flex justify-between text-sm">
                <span className="text-zen-text">{m.name}</span>
                <span className="text-zen-muted">{m.score.toFixed(1)}</span>
              </li>
            ))}
          </ul>
        </section>
      )}

      <section className="rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Overall trend</h2>
        {points ? (
          <svg viewBox="0 0 300 100" className="h-32 w-full">
            <polyline points={points} fill="none" stroke="currentColor" strokeWidth={2} className="text-accent-600" />
          </svg>
        ) : (
          <p className="text-sm text-zen-muted">No scored cycles yet.</p>
        )}
      </section>
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev`, open `/scout/engineers/<id>` for an engineer with at least one scored cycle — confirm the Overall + main-attribute scores render for the selected cycle, switching the cycle dropdown updates them, and the trend polyline renders across scored cycles
- [ ] **Step 4: Commit**
```bash
git add "apps/scout/src/app/engineers/[id]/page.tsx"
git commit -m "scout: add engineer card page with cycle selector and trend chart (F10)"
```

---

### Task 38: Cycle view page — all engineers for one cycle (F15)

**Files:**
- Create: `apps/scout/src/app/cycles/[id]/page.tsx`

**Interfaces:**
- Consumes: `api`, `RatingCycle`, `EngineerCycleScore` (Tasks 32, 33); `GET /api/cycles/:id/scores` (Task 15)

- [ ] **Step 1: Write the cycle view page**
```tsx
// apps/scout/src/app/cycles/[id]/page.tsx
"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import type { EngineerCycleScore } from "@/lib/types";

export default function CycleViewPage() {
  const params = useParams<{ id: string }>();
  const cycleId = Number(params.id);
  const [scores, setScores] = useState<EngineerCycleScore[]>([]);

  useEffect(() => {
    api.get<EngineerCycleScore[]>(`/api/cycles/${cycleId}/scores`).then(setScores);
  }, [cycleId]);

  return (
    <main className="mx-auto max-w-3xl p-8">
      <h1 className="mb-6 text-2xl font-semibold text-zen-text">Cycle #{cycleId} — Team Scores</h1>

      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-zen-border text-zen-muted">
            <th className="py-2">Engineer</th>
            <th className="py-2">Overall</th>
            {scores[0]?.main_attributes.map((m) => (
              <th key={m.main_attribute_id} className="py-2">
                {m.name}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {scores.map((row) => (
            <tr key={row.engineer.id} className="border-b border-zen-border">
              <td className="py-2">
                <Link href={`/engineers/${row.engineer.id}`} className="text-zen-text hover:underline">
                  {row.engineer.name}
                </Link>
              </td>
              <td className="py-2 text-zen-text">{row.overall != null ? row.overall.toFixed(1) : "—"}</td>
              {row.main_attributes.map((m) => (
                <td key={m.main_attribute_id} className="py-2 text-zen-muted">
                  {m.score.toFixed(1)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {scores.length === 0 && <p className="text-sm text-zen-muted">No rankings submitted for this cycle yet.</p>}
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev`, open `/scout/cycles/<id>` for a cycle with saved rankings — confirm every ranked engineer appears with Overall + per-main-attribute columns
- [ ] **Step 4: Commit**
```bash
git add "apps/scout/src/app/cycles/[id]/page.tsx"
git commit -m "scout: add cycle view page listing all engineers' scores for one cycle (F15)"
```

---

### Task 39: Roster dashboard page (F11)

**Files:**
- Create: `apps/scout/src/app/page.tsx`

**Interfaces:**
- Consumes: `api`, `RosterEntry` (Tasks 32, 33); `GET /api/dashboard` (Task 16)

- [ ] **Step 1: Write the dashboard page**
```tsx
// apps/scout/src/app/page.tsx
"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api } from "@/lib/api";
import type { RosterEntry } from "@/lib/types";

export default function DashboardPage() {
  const [roster, setRoster] = useState<RosterEntry[]>([]);

  useEffect(() => {
    api.get<RosterEntry[]>("/api/dashboard").then(setRoster);
  }, []);

  return (
    <main className="mx-auto max-w-3xl p-8">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-zen-text">Dashboard</h1>
        <nav className="flex gap-4 text-sm text-accent-600">
          <Link href="/engineers" className="hover:underline">
            Roster
          </Link>
          <Link href="/attributes" className="hover:underline">
            Attributes
          </Link>
          <Link href="/cycles" className="hover:underline">
            Cycles
          </Link>
        </nav>
      </div>

      <ul className="divide-y divide-zen-border">
        {roster.map((entry) => (
          <li key={entry.engineer.id} className="flex items-center justify-between py-3">
            <Link href={`/engineers/${entry.engineer.id}`} className="font-medium text-zen-text hover:underline">
              {entry.engineer.name}
            </Link>
            <div className="text-right text-sm">
              <div className="text-zen-text">{entry.latest_overall != null ? entry.latest_overall.toFixed(1) : "—"}</div>
              <div className="text-zen-muted">{entry.last_cycle_date ?? "no cycles yet"}</div>
            </div>
          </li>
        ))}
      </ul>

      {roster.length === 0 && <p className="text-sm text-zen-muted">No active engineers yet.</p>}
    </main>
  );
}
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev` and open `/scout` — confirm every active engineer appears with their latest Overall score and last cycle date, and the nav links to Roster/Attributes/Cycles work
- [ ] **Step 4: Commit**
```bash
git add apps/scout/src/app/page.tsx
git commit -m "scout: add roster dashboard page (F11)"
```

---

### Task 40: Metrics stats panel on the engineer card page (F5)

**Files:**
- Modify: `apps/scout/src/app/engineers/[id]/page.tsx`

**Interfaces:**
- Consumes: `api`, `MetricSnapshot` (Tasks 32, 33); `GET /api/engineers/:id/metrics` (Task 21)

- [ ] **Step 1: Add the metrics panel**
```tsx
// apps/scout/src/app/engineers/[id]/page.tsx — add to imports:
import type { Engineer, EngineerCard as EngineerCardData, MetricSnapshot, RatingCycle, TrendPoint } from "@/lib/types";

// add state, alongside the existing useState calls:
  const [metrics, setMetrics] = useState<MetricSnapshot[]>([]);

// add to the loadStatic() Promise.all — change it to:
    async function loadStatic() {
      const [eng, cycleList, trendData, metricSnapshots] = await Promise.all([
        api.get<Engineer>(`/api/engineers/${engineerId}`),
        api.get<RatingCycle[]>("/api/cycles"),
        api.get<TrendPoint[]>(`/api/engineers/${engineerId}/trend`),
        api.get<MetricSnapshot[]>(`/api/engineers/${engineerId}/metrics`),
      ]);
      setEngineer(eng);
      setCycles(cycleList);
      setTrend(trendData);
      setMetrics(metricSnapshots);
      if (cycleList.length > 0) setSelectedCycleId(cycleList[0].id);
    }

// add a new <section> after the trend chart section, before the closing </main>:
      <section className="mt-8 rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Synced metrics</h2>
        {metrics.length === 0 ? (
          <p className="text-sm text-zen-muted">No synced metrics yet.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-zen-border text-zen-muted">
                <th className="py-2">Period</th>
                <th className="py-2">PRs raised</th>
                <th className="py-2">PRs reviewed</th>
                <th className="py-2">Tickets closed</th>
                <th className="py-2">Complexity</th>
              </tr>
            </thead>
            <tbody>
              {metrics.map((m) => (
                <tr key={m.id} className="border-b border-zen-border">
                  <td className="py-2 text-zen-text">
                    {m.period_start} – {m.period_end}
                  </td>
                  <td className="py-2 text-zen-muted">{m.prs_raised}</td>
                  <td className="py-2 text-zen-muted">{m.prs_reviewed}</td>
                  <td className="py-2 text-zen-muted">{m.tickets_closed}</td>
                  <td className="py-2 text-zen-muted">{m.complexity_score.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev`, open `/scout/engineers/<id>` for an engineer with synced metrics — confirm the metrics table renders one row per synced period
- [ ] **Step 4: Commit**
```bash
git add "apps/scout/src/app/engineers/[id]/page.tsx"
git commit -m "scout: add synced-metrics stats panel to the engineer card page (F5)"
```

---

### Task 41: Highlights/lowlights UI with duplicate-flag warning (F13, F14)

**Files:**
- Modify: `apps/scout/src/app/engineers/[id]/page.tsx`

**Interfaces:**
- Consumes: `api`, `HighlightEntry` (Tasks 32, 33); `GET/POST /api/engineers/:id/highlights`, `POST /api/engineers/:id/highlights/check-duplicate` (Tasks 25, 29)

- [ ] **Step 1: Add the highlights/lowlights section**
```tsx
// apps/scout/src/app/engineers/[id]/page.tsx — add to imports:
import { useState, type FormEvent } from "react";
import type { HighlightEntry } from "@/lib/types";

// add state:
  const [highlights, setHighlights] = useState<HighlightEntry[]>([]);
  const [newKind, setNewKind] = useState<"highlight" | "lowlight">("highlight");
  const [newBody, setNewBody] = useState("");
  const [duplicateWarning, setDuplicateWarning] = useState<string | null>(null);
  const [checkingDuplicate, setCheckingDuplicate] = useState(false);

// add a loader (alongside the other useEffect calls):
  useEffect(() => {
    api.get<HighlightEntry[]>(`/api/engineers/${engineerId}/highlights`).then(setHighlights);
  }, [engineerId]);

// add handlers:
  async function handleAddEntry(e: FormEvent) {
    e.preventDefault();
    if (!newBody.trim()) return;

    setCheckingDuplicate(true);
    setDuplicateWarning(null);
    const check = await api.post<{ is_duplicate: boolean; matched_entry_id: number | null; similarity_note: string }>(
      `/api/engineers/${engineerId}/highlights/check-duplicate`,
      { body: newBody }
    );
    setCheckingDuplicate(false);

    // F14/NF3: the flag never blocks the save — it's shown so the admin can
    // choose to save anyway or cancel and rewrite.
    if (check.is_duplicate) {
      setDuplicateWarning(`Possible duplicate: ${check.similarity_note}`);
      return;
    }
    await saveEntry();
  }

  async function saveEntry() {
    await api.post(`/api/engineers/${engineerId}/highlights`, { kind: newKind, body: newBody });
    setNewBody("");
    setDuplicateWarning(null);
    const updated = await api.get<HighlightEntry[]>(`/api/engineers/${engineerId}/highlights`);
    setHighlights(updated);
  }

// add a new <section> after the metrics panel, before the closing </main>:
      <section className="mt-8 rounded-lg border border-zen-border p-4">
        <h2 className="mb-3 font-medium text-zen-text">Highlights &amp; lowlights</h2>

        <form onSubmit={handleAddEntry} className="mb-4 space-y-2">
          <div className="flex gap-2">
            <select
              value={newKind}
              onChange={(e) => setNewKind(e.target.value as "highlight" | "lowlight")}
              className="rounded border border-zen-border bg-transparent p-2 text-sm"
            >
              <option value="highlight">Highlight</option>
              <option value="lowlight">Lowlight</option>
            </select>
            <input
              className="flex-1 rounded border border-zen-border bg-transparent p-2 text-sm"
              placeholder="What happened?"
              value={newBody}
              onChange={(e) => setNewBody(e.target.value)}
            />
            <button type="submit" disabled={checkingDuplicate} className="rounded bg-accent-600 px-3 py-2 text-sm text-white disabled:opacity-50">
              {checkingDuplicate ? "Checking..." : "Add"}
            </button>
          </div>
          {duplicateWarning && (
            <div className="rounded border border-yellow-500 p-2 text-sm text-yellow-700">
              <p>{duplicateWarning}</p>
              <button type="button" onClick={saveEntry} className="mt-1 underline">
                Save anyway
              </button>
            </div>
          )}
        </form>

        <ul className="space-y-2">
          {highlights.map((h) => (
            <li key={h.id} className="text-sm">
              <span className={h.kind === "highlight" ? "text-green-600" : "text-red-500"}>
                {h.kind === "highlight" ? "★" : "▼"}
              </span>{" "}
              <span className="text-zen-muted">{h.created_at.slice(0, 10)}</span>{" "}
              <span className="text-zen-text">{h.body}</span>
            </li>
          ))}
        </ul>
      </section>
```
- [ ] **Step 2: Type-check and build**
Run: `cd /Users/lto/repos/personal/movoz && pnpm --filter @movoz/scout exec tsc --noEmit && pnpm --filter @movoz/scout build`
Expected: both succeed with no type errors
- [ ] **Step 3: Manual verify**
Run: `pnpm --filter @movoz/scout dev`, open `/scout/engineers/<id>` — add a highlight, confirm it appears in the list; add a near-duplicate of an existing entry and confirm the yellow duplicate warning appears with a "Save anyway" override that still saves the entry; stop the backend's Anthropic connectivity (or use an invalid `ANTHROPIC_API_KEY`) and confirm adding an entry still succeeds without blocking (degraded, no flag)
- [ ] **Step 4: Commit**
```bash
git add "apps/scout/src/app/engineers/[id]/page.tsx"
git commit -m "scout: add highlights/lowlights UI with AI duplicate-flag warning (F13, F14)"
```

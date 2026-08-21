# Scout

A private Go REST API backing Scout: a FIFA-style engineer attribute-card tracker with biweekly peer-relative ranking, synced GitHub/Jira metrics, an AI ranking-chat assistant, and a manual highlight/lowlight log. Built with Gin, following the same Standard Go Project Layout as `hustle-turtle` and `oncarinho`.

## Features

- Shared-password admin auth with signed session cookies — **every** route requires a valid session except `POST /api/auth/login` and `GET /health` (no public application-data group, unlike `oncarinho`)
- Engineer roster CRUD (create, edit, deactivate, reactivate — soft delete only)
- Main/sub-attribute management, seeded with the initial 6 main attributes
- Biweekly rating cycles: open a cycle, submit a strict 1..N ranking per sub-attribute (no ties/gaps), converted to a score via linear interpolation
- Computed-on-read scoring: main-attribute score, Overall score, cycle view, engineer card + trend — never cached, always derived from stored rankings
- Roster dashboard: latest Overall + last cycle date per active engineer
- In-process ticker scheduler syncing GitHub PR stats and Jira ticket stats into `metric_snapshots`, idempotent per (engineer, period)
- AI ranking chat (Claude API, streamed) proposing a rank ordering with rationale; persisted only on explicit accept
- Highlight/lowlight log with an AI semantic duplicate check that never blocks a save

## Project Structure

```
scout/
├── cmd/
│   └── server/              # Entrypoint + full route wiring
│       ├── main.go
│       └── router.go
├── internal/
│   ├── auth/                # Session token signing/verification
│   ├── config/              # Configuration management
│   ├── database/            # Database connection and migrations
│   ├── handlers/            # HTTP handlers
│   ├── models/              # Data models
│   ├── store/               # SQL access — one store per aggregate
│   ├── scoring/             # Pure functions: rank->score, permutation validation
│   ├── aiclient/            # Anthropic SDK wrapper (chat + duplicate check)
│   ├── integrations/        # GitHub + Jira API clients
│   └── syncer/              # Sync orchestration + ticker scheduler
├── migrations/              # Database migration files
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
- `SCOUT_GITHUB_REPOS` / `SCOUT_JIRA_PROJECTS`: which repos/projects to sync are environment-configured rather than hardcoded — set them to your actual GitHub repos and Jira project keys

## Running

### Build the application:
```bash
go build -o bin/scout ./cmd/server
```

### Start the server (without migrations):
```bash
go run ./cmd/server
# OR using the binary
./bin/scout
```

### Start the server with automatic migrations:
```bash
go run ./cmd/server -auto-migrate
# OR using the binary
./bin/scout -auto-migrate
```

The server starts on port 8082 and immediately begins the sync scheduler (one run at startup, then every `SYNC_INTERVAL_HOURS`).

## API Endpoints

### Unauthenticated
- `GET /health` — health check
- `POST /api/auth/login` — shared password → signed `scout_session` cookie

### Authenticated (every other route — 401 without a valid session)
- `GET /api/engineers` — list active engineers
- `GET /api/engineers/:id` — get engineer details
- `POST /api/engineers` — create engineer
- `PUT /api/engineers/:id` — update engineer
- `DELETE /api/engineers/:id` — deactivate engineer (soft delete)
- `POST /api/engineers/:id/reactivate` — reactivate a deactivated engineer
- `GET /api/engineers/:id/highlights` — list highlights/lowlights for an engineer
- `POST /api/engineers/:id/highlights` — add a highlight/lowlight entry
- `POST /api/engineers/:id/highlights/check-duplicate` — check if a highlight is semantically duplicate
- `GET /api/main-attributes` — list all main attributes
- `POST /api/main-attributes` — create a main attribute
- `PUT /api/main-attributes/:id` — update a main attribute
- `GET /api/sub-attributes` — list all sub-attributes
- `POST /api/sub-attributes` — create a sub-attribute
- `PUT /api/sub-attributes/:id` — update a sub-attribute
- `DELETE /api/sub-attributes/:id` — deactivate a sub-attribute
- `GET /api/cycles` — list rating cycles
- `POST /api/cycles` — create a new rating cycle
- `PUT /api/cycles/:id/sub-attributes/:subId/ranking` — submit a ranking for a sub-attribute in a cycle
- `GET /api/cycles/:id/sub-attributes/:subId/ranking` — get submitted ranking for a sub-attribute
- `GET /api/cycles/:id/scores` — get all engineer scores for a cycle
- `GET /api/engineers/:id/card` — engineer card (attributes + scores for a cycle)
- `GET /api/engineers/:id/trend` — engineer trend data (scores over time)
- `GET /api/engineers/:id/metrics` — synced GitHub/Jira metrics for an engineer
- `GET /api/dashboard` — roster dashboard (Overall score + last cycle date per engineer)
- `POST /api/cycles/:id/ai-sessions` — start AI ranking chat session (streamed response)
- `POST /api/cycles/:id/ai-sessions/:sessionId/accept` — accept and persist the AI's proposed ranking

## Database Migrations

Migrations are stored in the `migrations/` directory and can be run separately from the application.

### Migration Commands

#### Run migrations up (apply new migrations):
```bash
go run ./cmd/server -migrate=up
# OR using the binary
./bin/scout -migrate=up
```

#### Run migrations down (rollback migrations):
```bash
go run ./cmd/server -migrate=down
# OR using the binary
./bin/scout -migrate=down
```

#### Check current migration version:
```bash
go run ./cmd/server -version
# OR using the binary
./bin/scout -version
```

### Current Migrations:
- `000001_create_engineers_table` — Creates engineers table
- `000002_create_main_attributes_and_sub_attributes` — Creates main_attributes and sub_attributes tables (seeds the initial 6 main attributes)
- `000003_create_rating_cycles_and_rankings` — Creates rating_cycles and sub_attribute_rankings tables
- `000004_create_metric_snapshots_table` — Creates metric_snapshots table for synced GitHub/Jira metrics
- `000005_create_ai_ranking_sessions_table` — Creates ai_ranking_sessions table for AI chat history
- `000006_create_highlight_entries_table` — Creates highlight_entries table for manual logs

### Migration Best Practices:
- **Development**: Use `-auto-migrate` flag for convenience
- **Production**: Run migrations separately before deploying new code:
  1. `./bin/scout -migrate=up` (run migrations)
  2. `./bin/scout` (start the service)
- **Rollback**: Use `-migrate=down` if you need to rollback migrations

## Testing

Tests require a running Postgres and a dedicated test database — without one, `go test ./...` reports `ok` for every package because each test gracefully skips (intentional, mirroring `oncarinho`'s convention — a green `go test ./...` does NOT by itself prove the tests ran).

The test database needs the schema **and** migration-seeded data (migration `000002` seeds the initial 6 `main_attributes`, which `TestMainAttributeStoreSeedData` asserts on), so create it and run the migrations against it once before the first test run:

```bash
createdb scout_test
DATABASE_URL="postgres://localhost/scout_test?sslmode=disable" go run ./cmd/server -migrate=up
go test -p 1 ./...
```

(`-migrate=up` runs before the `ADMIN_PASSWORD`/`SESSION_SECRET` check in `main.go`, so no other env vars are needed for this step. Re-run it after adding a migration.)

To point at a non-default test database:
```bash
TEST_DATABASE_URL="postgres://localhost/scout_test?sslmode=disable" go test -p 1 ./...
```

To confirm tests are actually running (not skipping):
```bash
go test -p 1 ./... -v 2>&1 | grep -c '^--- SKIP'   # should print 0
```

**`-p 1` is required, not just recommended.** Every package's tests share the *same* `scout_test` database, and most of them `TRUNCATE ... RESTART IDENTITY CASCADE` the tables they own to build fixtures. Go runs different packages' test binaries concurrently by default, so without `-p 1` a truncate in one package wipes rows another package's test is mid-way through asserting on (e.g. `internal/store`'s scores/ai-session tests vs. `internal/handlers`'s cycle-view/AI-chat tests). This is a cross-package data-sharing problem, not something the test code can fix on its own — short of giving every package its own database.

*Intra*-package ordering is handled in code: `truncateTables` (in each package's `testutil_test.go`) restores `main_attributes`' 6 seed rows on test cleanup whenever a test truncates that table, so seed-dependent tests such as `TestMainAttributeStoreSeedData` pass regardless of the order tests run in within a package (verified with `go test -shuffle=on`). Always truncate through that helper rather than issuing raw `TRUNCATE main_attributes` SQL in a test.

## Architecture

- **Scoring is computed on read** (NF2): main-attribute, Overall, cycle-view, and trend scores are all derived via SQL aggregation over `sub_attribute_rankings` on every request — there is no cached/materialized score table, avoiding a second source of truth that could drift.
- **Overall score cutover** (F8): a main attribute counts toward a cycle's Overall only if `main_attributes.created_at <= rating_cycles.created_at` — adding a main attribute later never retroactively changes a past cycle.
- **Sync worker** (F4, NF4): an in-process ticker started from `cmd/server/main.go` (no separate `cmd/syncer` binary) polls GitHub/Jira for every active engineer and idempotently upserts `metric_snapshots` keyed on `(engineer_id, period_start, period_end)`. A single engineer's fetch failure is logged and skipped; the next scheduled run retries it safely.
- **Sync window = the current rating cycle** (F4): each run syncs the newest `rating_cycles` row's `period_start`/`period_end`, not a rolling "now minus 14 days" range. `metric_snapshots`' key uses DATE columns, so a rolling window created a fresh near-duplicate row on every new calendar day; anchoring to the cycle means every run inside one cycle upserts the same row, and a new row appears only when the admin opens the next cycle. It also keeps the GitHub inclusive-boundary offset meaningful (consecutive cycles are genuinely back-to-back). Before the first cycle exists, a run logs `no rating cycle exists yet, skipping metrics sync` and does nothing.
- **AI ranking chat** (F9, NF3): the Anthropic Go SDK streams the assistant's reply as Server-Sent Events. The reply's rationale and a trailing fenced ` ```json ` block carrying the proposed ranking are both stored on the session — the ranking is written to `sub_attribute_rankings` only via the separate accept endpoint, on explicit admin confirm.
- **Duplicate check** (F14, NF3): uses the Anthropic Go SDK's structured JSON-schema output for a guaranteed-parseable response. If the AI call fails or times out, the endpoint returns `200` with `is_duplicate: false` rather than an error — the admin's save is never blocked.

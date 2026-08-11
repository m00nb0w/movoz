# Oncarinho

A Go REST API built with Gin framework following Go project layout best practices. Backs the Oncarinho football team stats-tracking site: an admin logs each matchday's per-player goals, assists, and cards; the team and public browse leaderboards, player profiles, and a team dashboard.

## Features

- Health check endpoint at `/health`
- Shared-password admin auth with signed session cookies
- Player roster management (create, edit, deactivate, reactivate — soft delete only)
- Matchday creation and per-player stat entry (bulk upsert)
- Public leaderboards (goals / assists / cards), filterable by year or all-time
- Public player profiles (all-time totals + year-by-year breakdown)
- Public team summary/dashboard totals
- Database migrations with CLI control
- PostgreSQL support
- Clean architecture with separated concerns
- Environment-based configuration

## Project Structure

```
oncarinho/
├── cmd/
│   └── server/           # Application entrypoint + route wiring
│       ├── main.go
│       └── router.go
├── internal/             # Private application code
│   ├── auth/            # Session token signing/verification
│   ├── config/          # Configuration management
│   ├── database/        # Database operations and migrations
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── store/           # SQL access (players, matchdays, stats, leaderboard, profile, summary)
├── migrations/          # Database migration files
├── bin/                 # Compiled binaries
├── go.mod
├── go.sum
└── README.md
```

## Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb oncarinho
```

3. Set environment variables:
```bash
export DATABASE_URL="postgres://username:password@localhost/oncarinho?sslmode=disable"
export PORT="8081"
export ADMIN_PASSWORD="change-me"
export SESSION_SECRET="change-me-too"
```

**Defaults:**
- `DATABASE_URL`: `postgres://localhost/oncarinho?sslmode=disable`
- `PORT`: `8081`
- `ADMIN_PASSWORD`: none — the server refuses to start without it
- `SESSION_SECRET`: none — the server refuses to start without it

## Running

### Build the application:
```bash
go build -o bin/oncarinho ./cmd/server
```

### Start the server (without migrations):
```bash
go run ./cmd/server
# OR using the binary
./bin/oncarinho
```

### Start the server with automatic migrations:
```bash
go run ./cmd/server -auto-migrate
# OR using the binary
./bin/oncarinho -auto-migrate
```

The server will start on port 8081.

## API Endpoints

### Public (no auth)
- `GET /health` — Health check endpoint
- `GET /api/players` — Active roster
- `GET /api/players/:id` — Player profile (all-time totals + year-by-year breakdown)
- `GET /api/matchdays` — List of matchdays
- `GET /api/matchdays/:id/stats` — Stats entered for a matchday
- `GET /api/leaderboard?year=YYYY&stat=goals|assists|cards` — Ranked table (omit `year` for all-time)
- `GET /api/summary?year=YYYY` — Team dashboard totals (defaults to current year, UTC)
- `POST /api/auth/login` — Shared password → sets a signed `admin_session` cookie

### Admin (requires a valid `admin_session` cookie — returns 401 otherwise)
- `POST /api/players` — Create player
- `PUT /api/players/:id` — Edit player
- `DELETE /api/players/:id` — Soft delete (`is_active = false`)
- `POST /api/players/:id/reactivate` — Reactivate a soft-deleted player
- `POST /api/matchdays` — Create a matchday (date only)
- `PUT /api/matchdays/:id/stats` — Bulk upsert per-player stats for a matchday
- `DELETE /api/matchdays/:id/stats/:playerId` — Remove a player's stat row for a matchday

## Database Migrations

Migrations are stored in the `migrations/` directory and can be run separately from the application.

### Migration Commands

#### Run migrations up (apply new migrations):
```bash
go run ./cmd/server -migrate=up
# OR using the binary
./bin/oncarinho -migrate=up
```

#### Run migrations down (rollback migrations):
```bash
go run ./cmd/server -migrate=down
# OR using the binary
./bin/oncarinho -migrate=down
```

#### Check current migration version:
```bash
go run ./cmd/server -version
# OR using the binary
./bin/oncarinho -version
```

### Current Migrations:
- `000001_create_players_table` — Creates players table (name, position, is_active, created_at)
- `000002_create_matchdays_table` — Creates matchdays table (played_on, created_at)
- `000003_create_match_stats_table` — Creates match_stats table (goals, assists, yellow_cards, red_cards; unique per matchday+player)
- `000004_add_match_stats_non_negative_check` — Adds a CHECK constraint requiring goals/assists/yellow_cards/red_cards to be non-negative

### Migration Best Practices:
- **Development**: Use `-auto-migrate` flag for convenience
- **Production**: Run migrations separately before deploying new code:
  1. `./bin/oncarinho -migrate=up` (run migrations)
  2. `./bin/oncarinho` (start the service)
- **Rollback**: Use `-migrate=down` if you need to rollback migrations

## Testing

Tests require a running Postgres and a dedicated test database — without one, `go test ./...` reports `ok` for every package because each test gracefully skips (this is intentional so the suite doesn't hard-fail in environments without Postgres, but it means a green `go test ./...` does NOT by itself prove the tests ran).

To run the real suite:

```bash
createdb oncarinho_test
go test ./...
```

To point at a non-default test database:
```bash
TEST_DATABASE_URL="postgres://localhost/oncarinho_test?sslmode=disable" go test ./...
```

To confirm tests are actually running (not skipping), use `-v` and check for `--- PASS` lines rather than `--- SKIP`:
```bash
go test ./... -v 2>&1 | grep -c '^--- SKIP'   # should print 0
```

## Architecture

The project follows the **Standard Go Project Layout**:

- **`cmd/`**: Main applications for this project — entrypoint and route wiring (public vs. admin route groups)
- **`internal/`**: Private application and library code
- **`migrations/`**: Database schema migration files

### Internal Package Structure:

- **`auth/`**: Session token signing and verification for admin auth
- **`config/`**: Configuration management and environment variables
- **`database/`**: Database connection, migrations, and database-related utilities
- **`handlers/`**: HTTP handlers (controllers) for different endpoints
- **`models/`**: Data structures and business logic models
- **`store/`**: SQL access layer — one store per aggregate (players, matchdays, stats, leaderboard, profile, summary)

### Aggregation strategy

Leaderboards and player profiles are computed on read via SQL `GROUP BY` queries over `match_stats` joined to `matchdays` (filtered by year, or unfiltered for all-time). There is no materialized/cached totals table, avoiding a second source of truth that could drift from the raw entries.

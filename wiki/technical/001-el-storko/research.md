# Research: el-storko

Phase 0 output for `wiki/technical/001-el-storko/plan.md`. No `[NEEDS CLARIFICATION]` markers
remained after `/speckit-specify`; the items below confirm technical choices against existing
repo conventions rather than resolving open unknowns.

## Backend framework and DB driver

- **Decision**: Gin + `github.com/lib/pq` + `github.com/golang-migrate/migrate/v4`, same versions
  already used by `backend/oncarinho`.
- **Rationale**: `oncarinho` (`backend/oncarinho/go.mod`) is the most recent Go service in this
  repo and already establishes this exact stack; matching it avoids introducing a second Postgres
  driver or migration tool into the monorepo for no reason (Constitution Principle II/IV).
- **Alternatives considered**: `pgx` — more modern but not currently used anywhere in this repo;
  rejected to avoid an unjustified new dependency for a feature that has no need for `pgx`-specific
  features (e.g. no COPY-based bulk load, no advanced type mapping).

## Auth / secrets

- **Decision**: No session/password auth (unlike `oncarinho`'s `ADMIN_PASSWORD`/`SESSION_SECRET`).
  The only secret is the Jira API token, loaded from a git-ignored `.env` via a small loader in
  `internal/config`, and also settable as a launchd `EnvironmentVariables` entry.
- **Rationale**: FR-013 and the spec's "single user, no public exposure" assumption mean there is
  no other actor to authenticate against; adding session auth would be complexity with no
  corresponding requirement (Principle II).
- **Alternatives considered**: Reuse `oncarinho`'s session-auth middleware — rejected, since there
  is no login flow or second user to gate.

## Jira integration

- **Decision**: Poll Jira REST API v3 (`/rest/api/3/search` with JQL `assignee = currentUser()`,
  `/rest/api/3/issue/{key}` for updates) via a hand-rolled HTTP client using the Jira personal API
  token (HTTP Basic auth: email + token), on a goroutine ticking every 5 minutes by default
  (configurable via env var), rather than a third-party Jira SDK.
- **Rationale**: The Jira REST surface needed here is small (search assigned issues, read/update
  an issue's status transition and description); a full SDK is unjustified weight. Polling is
  required, not a preference — the spec (FR-005, and the assumption that the service isn't
  publicly reachable) rules out webhooks since Jira Cloud cannot deliver a webhook to an
  unexposed localhost service.
- **Conflict resolution**: Last-write-wins by comparing the local `updated_at` and the Jira
  issue's `fields.updated` timestamp at each poll (FR-007), matching the spec's accepted
  simplification for a single-user tool.
- **Alternatives considered**: ngrok/tunnel-based webhook delivery — rejected as unneeded
  operational complexity when a 5-minute poll already satisfies the "within one polling interval"
  acceptance criteria (SC-002, SC-003).

## Frontend zone conventions

- **Decision**: Follow `apps/oncarinho`'s `next.config.mjs` shape: `transpilePackages` for the
  shared packages actually used (`@movoz/theme`, `@movoz/tailwind-config`), an `/api/:path*`
  rewrite to an `EL_STORKO_API_URL` env var (default `http://localhost:8082`), `output:
  "standalone"`. Skip `next-intl` (oncarinho's i18n setup) since el-storko has no localization
  requirement in the spec.
- **Rationale**: Reusing the proven pattern avoids re-deriving Multi-Zones wiring; omitting i18n
  keeps the zone minimal per Principle II since nothing in the spec calls for it.
- **Port choice**: Backend on `8082`, frontend `dev`/`start` on `3101` — next free ports after
  `hustle-turtle` (8080), `oncarinho` (8081/3100).

## Always-on service

- **Decision**: Two macOS `launchd` user agents (`~/Library/LaunchAgents/`), one per process
  (backend, frontend), each with `RunAtLoad` and `KeepAlive` (restart on crash), `StandardOutPath`/
  `StandardErrorPath` pointed at log files outside the repo (to avoid ever committing logs that
  could contain the Jira token).
- **Rationale**: `launchd` is the native macOS mechanism for "always running, survives
  reboot/crash" (FR-011, SC-004) without adding a process manager dependency (e.g. pm2, systemd —
  the latter isn't even available on macOS).
- **Alternatives considered**: `pm2` — rejected, adds a Node-based process-manager dependency for
  a two-process macOS-only setup where launchd is already sufficient and native.

## CLI

- **Decision**: A second `main` package (`cmd/cli/`) in the same Go module as the server, calling
  the REST API over HTTP exactly like the web UI would — no direct DB or store access from the
  CLI.
- **Rationale**: FR-010 requires the CLI and web UI to operate on the same data with no
  divergence; routing both through the same REST API is the only way to guarantee that by
  construction, and reusing the module's `internal/models` types keeps request/response shapes in
  sync without duplicating structs.
- **Alternatives considered**: A separate Rust CLI (mirroring `drunken-dolphin`'s style) —
  rejected: it would need to duplicate the Go request/response types in Rust and adds a second
  language toolchain for a thin HTTP client with no engineering benefit here.

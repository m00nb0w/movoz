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

## Jira status mapping under the five-state model — 2026-09-30 revision

- **Decision**: Pulling from Jira, map its status to `backlog`/`in_progress`/`blocked`/`done` only
  (never `picked_for_today` — FR-005); anything unrecognized falls back to `backlog` (previously
  `todo`). Pushing to Jira, a local `picked_for_today` item is pushed as Jira's "In Progress"
  transition — the closest real-world meaning ("I'm actively working this today") — since Jira has
  no equivalent state to push into.
- **Rationale**: `picked_for_today` is an explicitly personal, Jira-less daily-triage concept per
  the spec's Assumptions; the pull direction already had to pick a default bucket for
  unrecognized statuses, and reusing `backlog` for that (instead of inventing a second fallback)
  keeps the mapping table small. The push direction needs *some* answer whenever a Jira-sourced
  item is picked for today locally and the sync loop later pushes its status, so "In Progress" is
  the least-surprising equivalent rather than leaving the push unspecified.
- **Alternatives considered**: Refusing to push status at all for `picked_for_today` items (only
  pushing description changes) — rejected as inconsistent with FR-006's blanket requirement that
  local status changes on Jira-sourced items sync back, and would silently diverge the two systems
  exactly in the case (daily triage) the tracker exists to make effortless.

## Reference keys (EPIC-#/TASK-#) — 2026-09-30 revision

- **Decision**: Two Postgres `SEQUENCE` objects (`epic_reference_seq`, `task_reference_seq`),
  each feeding a single `reference_number INTEGER NOT NULL` column on `work_items`. The store
  layer picks the sequence to draw from based on the row's `type` at insert time
  (`nextval('epic_reference_seq')` for an Epic, `nextval('task_reference_seq')` for a Task) —
  consistent with this codebase's existing preference for validation/business logic living in
  Go rather than DB triggers (see parent/type validation in `data-model.md`). The display string
  (`EPIC-<n>` / `TASK-<n>`) is computed at the API-response layer from `type` + `reference_number`,
  not stored redundantly.
- **Rationale**: FR-018 requires two independent, collision-free numbering sequences. Postgres
  sequences are the simplest mechanism that's atomic under concurrent inserts (irrelevant at this
  project's single-user scale, but free correctness) and need no additional bookkeeping table.
- **Alternatives considered**: Deriving the key from the row's own `id` (the original design) —
  rejected per the spec's later revision, since a single shared `id` column can't produce two
  independently-numbered sequences. A single sequence with a computed offset per type — rejected
  as needlessly clever compared to two plain sequences.

## Drag-and-drop and charts — 2026-09-30 revision

- **Decision**: Implement "drag Backlog item onto the Board" (FR-014) with the native HTML5
  Drag and Drop API (`draggable`, `onDragStart`/`onDragOver`/`onDrop`) rather than a
  drag-and-drop library. Implement the Stats tab's status-breakdown bar chart (FR-017) as a small
  custom SVG component, following the same hand-rolled-chart convention already used by
  `TrendLine` in this app and by `drunken-dolphin`/`scout` elsewhere in the repo.
- **Rationale**: Both are one-off, low-complexity interactions/visuals for a single-user tool;
  adding `@dnd-kit` or a charting library (e.g. `recharts`) would be new dependencies with no
  corresponding requirement beyond what a few dozen lines of native API usage already covers
  (Principle II). The "Pick for today" button remains the primary, keyboard-friendly path;
  drag-and-drop is a convenience on top of it, not the only way to do it (FR-014 requires both).
- **Alternatives considered**: `@dnd-kit/core` — more robust (touch support, accessibility
  affordances) but unjustified weight for a desktop-only, single-user tool; can be revisited if
  drag interactions turn out to need more than the native API offers in practice.

## Item detail drawer — 2026-09-30 revision

- **Decision**: Implement the "item detail drawer" (FR-019) using `@movoz/ui-web`'s existing
  `Modal` component (`size="lg"` or `"xl"`), not a new slide-out-drawer primitive.
- **Rationale**: `@movoz/ui-web` has no dedicated Drawer component, and FR-019's actual
  requirement — every field visible, no internal scrolling — is a sizing/layout constraint that
  `Modal` already satisfies; building and maintaining a new shared primitive for one feature's
  cosmetic preference (slide-from-side vs. center-pop) isn't justified (Principle II). If a real
  slide-out drawer becomes valuable across multiple apps later, it belongs in `packages/ui-web`
  as a new shared primitive then, not a one-off in `apps/el-storko`.
- **Alternatives considered**: A bespoke fixed-position side panel — rejected as one more
  hand-rolled overlay/portal/focus-trap implementation this repo would then own, when `Modal`
  already solves the same problem.

## Mine/Agent scope filter — 2026-09-30 revision

- **Decision**: A `scope=mine|agent` query parameter on `GET /api/work-items` and the stats
  endpoints, translated server-side into `source IN ('personal','jira')` for `mine` or
  `source = 'agent'` for `agent` — see `contracts/rest-api.md`. The existing single-value
  `source=` filter is kept as-is for finer-grained filtering (e.g. Jira-only), independent of the
  scope switch.
- **Rationale**: FR-022/FR-023 describe one global, coarse-grained toggle covering both Board and
  Stats; encoding it as its own param keeps the frontend's toggle state simple (one enum) rather
  than needing to know `personal,jira` is the "Mine" set in multiple places.
- **Alternatives considered**: A repeatable `source=personal&source=jira` query — rejected as more
  ceremony on both client and server for a concept (Mine vs. Agent) the product spec already
  treats as a single switch, not an arbitrary multi-source filter.

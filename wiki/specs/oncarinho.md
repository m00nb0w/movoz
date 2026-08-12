# Product Spec: Oncarinho

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-07-24
> **Project**: oncarinho
> **Location**: `apps/oncarinho/` (frontend), `backend/oncarinho/` (backend)
> **Stack**: Next.js 14 + Go (REST API) + PostgreSQL

## Overview

Oncarinho is a stats-tracking website for a football team. An admin logs each matchday's per-player goals, assists, and cards; the whole team — and the public — can browse leaderboards, player profiles, and a team dashboard.

## Problem Statement

The team currently has no shared, always-available record of who scored, who assisted, and who's racking up cards across matchdays and years. Tracking this informally (spreadsheets, memory) makes it hard to see season leaders or a player's career history at a glance. A small public site gives the team a single source of truth for stats, with a low-friction admin flow for entering them after each matchday.

## Goals

- Give an admin a fast way to record per-player stats for a matchday
- Give the team and public a browsable view of leaderboards and player profiles, by year and all-time
- Keep data entry to "who played and what did they do" — no match metadata overhead (no opponent/score tracking)
- Support both English and Vietnamese, so non-English-speaking teammates get a fully native experience, not just an afterthought

## Non-Goals

- Not a full match-management system (no opponents, scores, fixtures, lineups/formations)
- Not a multi-team platform — single team only
- Not a per-admin audit trail — a single shared admin password, not named admin accounts
- No self-serve accounts for players — players are view-only, roster is admin-managed
- No languages beyond English and Vietnamese, and no localized URLs — language is a display-layer concern only

## User Stories

- As an **admin**, I want to create a matchday and enter each player's goals/assists/cards so that stats stay current after every game.
- As an **admin**, I want to add, edit, and deactivate players on the roster so that it reflects who's actually on the team.
- As a **teammate or visitor**, I want to see the current year's leaderboard (top scorers, assists, cards) so that I know how the season is shaping up.
- As a **teammate or visitor**, I want to view a player's profile so that I can see their career totals and year-by-year breakdown.
- As a **visitor**, I want a team dashboard summarizing the year at a glance (matches played, goals scored, roster) so I get an overview without digging into tables.

## Requirements

### Functional

| # | Requirement | Priority |
|---|---|---|
| F1 | Admin can log in with a shared password to unlock write access | Must have |
| F2 | Admin can create a matchday (date only) | Must have |
| F3 | Admin can enter/edit per-player goals, assists, yellow cards, red cards for a matchday | Must have |
| F4 | Admin can add, edit, and deactivate players (name, position) | Must have |
| F5 | Public leaderboard view: rank players by goals, assists, or cards, filterable by year or all-time | Must have |
| F6 | Public player profile: all-time totals + year-by-year stat breakdown | Must have |
| F7 | Public team dashboard: current-year summary (matches played, goals scored, roster size), roster list, leaderboard preview | Must have |
| F8 | Public matchday history list (dates played, filterable by year) | Should have |
| F9 | Full UI (public + admin) available in English and Vietnamese, defaulting to the visitor's browser language, with a manual toggle that persists the choice | Must have |

### Non-Functional

| # | Requirement | Metric |
|---|---|---|
| NF1 | Public pages require no authentication | All `GET` endpoints and public routes are unauthenticated |
| NF2 | Admin routes are protected | All admin `POST`/`PUT`/`DELETE` endpoints reject requests without a valid session cookie (401) |
| NF3 | Stat aggregates are always correct, no cached drift | Leaderboards/profiles computed via SQL aggregation over raw stat rows at request time, not a precomputed table |
| NF4 | No hardcoded UI strings | All user-facing static text sourced from `en`/`vi` translation dictionaries, no strings baked directly into components |

## Architecture

- **`apps/oncarinho`** — new Next.js 14 zone, following the multi-zone pattern used by `personal-site` (own `basePath`/`assetPrefix`, Nginx route). Uses `@movoz/tailwind-config` and `@movoz/theme`. Calls the Go backend over HTTP for all reads/writes — no direct DB access from the frontend.
- **`backend/oncarinho`** — new Go REST API, following the `hustle-turtle` layout (`cmd/server/`, `internal/{config,database,handlers,models}/`, `migrations/`), backed by its own PostgreSQL database (new DB on the existing Terraform-managed RDS instance, or local Postgres for dev — same convention as `hustle-turtle`).
- **Aggregation strategy**: leaderboards and player profiles are computed on read via SQL `GROUP BY` queries over `match_stats` joined to `matchdays` (filtered by year, or unfiltered for all-time). No materialized/cached totals table — avoids a second source of truth that could drift from the raw entries. At hobby-team scale (dozens of players, at most a few hundred matchdays/year) this requires no caching.
- **i18n**: `apps/oncarinho` uses `next-intl`, no locale-prefixed URLs. Middleware resolves the locale per request — a `locale` cookie (set by the manual toggle) takes precedence, otherwise the `Accept-Language` header is negotiated against the two supported locales (`en`, `vi`), defaulting to `en`. Translation dictionaries (`messages/en.json`, `messages/vi.json`) are namespaced by page/section; player names and the team name are never translated. Dates render locale-aware automatically via `next-intl`'s `Intl.DateTimeFormat` integration.

## Data Model

```
players
  id, name, position (nullable enum: goalkeeper|defender|midfielder|forward), is_active (bool, default true), created_at

matchdays
  id, played_on (date), created_at

match_stats
  id, matchday_id (FK -> matchdays), player_id (FK -> players),
  goals (int, default 0), assists (int, default 0),
  yellow_cards (int, default 0), red_cards (int, default 0)
  unique(matchday_id, player_id)
```

- A row in `match_stats` means that player played that matchday; "matches played" for a player = count of their `match_stats` rows (optionally filtered by year).
- **Year** for grouping is derived from `matchdays.played_on` (`EXTRACT(YEAR FROM played_on)`) — no separate season/competition table.
- **All-time** totals = same aggregate query with no year filter.
- Removing a player from the roster is a soft delete (`is_active = false`), not a hard delete, so historical stats and profile pages stay intact. Inactive players are hidden from the active roster and from new-entry pickers, but their profile/history remains browsable.
- `position` is a fixed enum, not free text, so it can be translated: the DB stores a canonical English key (`goalkeeper`/`defender`/`midfielder`/`forward`), enforced by a DB `CHECK` constraint and API validation; the frontend maps the key to a localized label via the `positions` translation namespace. Player names are never translated.

## API Design

Public (no auth):
```
GET /api/players?active=all                       — roster (active only by default; ?active=all includes deactivated)
GET /api/players/:id                              — profile: all-time totals + year-by-year breakdown (deactivated players included)
GET /api/leaderboard?year=YYYY|all&stat=goals|assists|cards  — ranked table
GET /api/summary?year=YYYY                        — team dashboard totals (defaults to current year, UTC)
GET /api/matchdays?year=YYYY                       — list of matchdays
GET /api/matchdays/:id/stats                      — a matchday's per-player stats
```

Admin (session-cookie protected):
```
POST /api/auth/login                     — shared password -> signed session cookie
POST /api/players                        — create player (position validated against the enum)
PUT  /api/players/:id                    — edit player
DELETE /api/players/:id                  — soft delete (is_active = false)
POST /api/players/:id/reactivate         — undo a soft delete
POST /api/matchdays                      — create a matchday (date)
PUT  /api/matchdays/:id/stats             — bulk upsert per-player stats for that matchday (goals/assists/cards must be >= 0)
DELETE /api/matchdays/:id/stats/:playerId — remove one player's stat row for a matchday
```

`ADMIN_PASSWORD` and `SESSION_SECRET` are required env vars — the server refuses to start without them. Auth middleware guards all admin-prefixed routes and returns 401 on a missing/invalid session cookie. The session cookie's `Secure` flag is configurable via `COOKIE_SECURE` (default `false` for local dev; should be `true` behind HTTPS in production).

## Frontend Pages

| Route | Description |
|---|---|
| `/` | Team dashboard — summary tiles, roster list, leaderboard preview |
| `/leaderboard` | Full ranked tables (goals / assists / cards) with a year selector (each year + all-time) |
| `/players/[id]` | Player profile — all-time totals + year-by-year table |
| `/admin` | Password gate |
| `/admin/matchdays` | List/create matchdays; opens a per-player stat entry grid for a selected matchday |
| `/admin/players` | Add/edit/deactivate/reactivate roster; position picked from a fixed dropdown |

Global nav (all pages, public and admin) includes an EN/VI language toggle — see F9 and the [UI design brief](../technical/oncarinho-ui-design-brief.md).

## Error Handling & Testing

- Validation: player `name` required; `position` (if provided) must be one of the 4 enum values; matchday `played_on` required; stat values (goals/assists/cards) must be >= 0, enforced at both the API and DB layer. Admin write endpoints return 400 on invalid input, 401 on missing/invalid session.
- Testing: Go table-driven tests for handlers and aggregation queries (`go test ./...`, mirroring `hustle-turtle`). Frontend covered by TypeScript type-checking and manual verification — no test framework currently exists in this repo's Next.js apps.

## Success Metrics

| Metric | Target | How Measured |
|---|---|---|
| Admin can log a full matchday's stats | < 2 minutes for a ~15-player squad | Manual timing during first real use |
| Public pages load correctly with no auth | Leaderboard, player profile, dashboard all render for a signed-out visitor | Manual check post-deploy |

## Open Questions

- [ ] Does the Go backend get its own RDS instance/DB, or share the existing Postgres instance with a separate database/schema?
- [ ] Domain/path for the new zone (e.g. `/oncarinho` under the existing multi-zone Nginx setup)?

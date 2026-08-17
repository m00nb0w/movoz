# Oncarinho (frontend)

A Next.js 14 (App Router) frontend for the Oncarinho football team stats tracker. Public visitors browse a team dashboard, leaderboards, and player profiles; an admin logs in with a shared password to record matchdays, per-player stats, and manage the roster. Bilingual (English/Vietnamese) and themeable (light/dark).

Backed by the Go REST API in `backend/oncarinho` — see that package's README for the API contract, data model, and migrations.

## Features

- Team dashboard with season-selectable stat tiles, top scorers, and active roster
- Public leaderboards (goals / assists / cards), filterable by year or all-time, with Active/Inactive status badges
- Public player profiles (all-time totals + year-by-year breakdown)
- Shared-password admin login backed by the API's signed `admin_session` cookie
- Admin matchday creation and a per-player stat-entry grid (bulk save)
  - A player left at all-zero stats for a matchday doesn't count as having played it — no stat row is created, and if one already existed it's deleted on save
- Admin roster management: create, edit, deactivate, and reactivate players (soft delete only — deactivated players disappear from the public roster/leaderboard-by-default but their profile page and historical stats remain reachable)
- i18n: English and Vietnamese, toggle persists in a cookie; player and team names are never translated
- Light/dark theme toggle

## Project Structure

```
apps/oncarinho/
├── src/
│   ├── app/                        # App Router routes
│   │   ├── page.tsx                 # / — dashboard
│   │   ├── leaderboard/page.tsx     # /leaderboard
│   │   ├── players/[id]/            # /players/[id] — profile (+ not-found.tsx)
│   │   └── admin/
│   │       ├── page.tsx             # /admin — login
│   │       ├── matchdays/page.tsx   # /admin/matchdays — list + create
│   │       ├── matchdays/[id]/      # /admin/matchdays/[id] — stat-entry grid
│   │       └── players/page.tsx     # /admin/players — roster management
│   ├── components/                  # Nav, LanguageToggle, StatTile, SeasonSelector, StatTypeTabs
│   ├── lib/
│   │   ├── api/                     # Typed fetch client + response types
│   │   └── years.ts
│   ├── i18n/                        # next-intl config + request handler
│   └── middleware.ts                # Locale negotiation/cookie handling
├── messages/                        # en.json, vi.json — message dictionaries
├── next.config.mjs                  # Proxies /api/* to the Go backend
└── package.json
```

## Setup

1. Install dependencies (from the monorepo root):
```bash
pnpm install
```

2. Have the Go backend running locally (see `backend/oncarinho/README.md`) — by default on `http://localhost:8081`.

3. Set environment variables:
```bash
cp apps/oncarinho/.env.local.example apps/oncarinho/.env.local
```
```
ONCARINHO_API_URL=http://localhost:8081
```

**Defaults:**
- `ONCARINHO_API_URL`: `http://localhost:8081`

Server-side code (Server Components, the Next.js rewrite proxy) reads `ONCARINHO_API_URL` directly. Browser code always calls same-origin `/api/*`, which `next.config.mjs` rewrites to `${ONCARINHO_API_URL}/api/*` — this keeps the `admin_session` cookie first-party from the browser's perspective regardless of what host/port the Go API runs on.

## Running

### Development server:
```bash
pnpm --filter oncarinho dev
```
Starts on `http://localhost:3100`.

### Production build:
```bash
pnpm --filter oncarinho build
pnpm --filter oncarinho start
```

### Lint:
```bash
pnpm --filter oncarinho lint
```

## Routes

### Public
- `/` — Team dashboard: stat tiles, top scorers, and roster for a selected year
- `/leaderboard` — Goals/Assists/Cards leaderboard, filterable by year or all-time
- `/players/[id]` — Player profile: all-time totals + year-by-year breakdown

### Admin (requires the `admin_session` cookie for write actions; set via `/admin` login)
- `/admin` — Shared-password login form
- `/admin/matchdays` — List existing matchdays; create a new one
- `/admin/matchdays/[id]` — Per-player stat-entry grid for a matchday (bulk save)
- `/admin/players` — Roster management: create, edit, deactivate, reactivate

Note: the admin pages themselves render for anyone (the data they list — players, matchdays — comes from public GET endpoints); only mutating actions (create/edit/deactivate/reactivate/save stats) require a valid session and will bounce the browser back to `/admin` on a 401.

## i18n

- Locales: English (`en`, default) and Vietnamese (`vi`)
- Cookie: `locale` — set by `middleware.ts` on first visit (negotiated from `Accept-Language`), and updated client-side by the language toggle in the nav bar
- Message dictionaries: `messages/en.json` and `messages/vi.json`, loaded via `next-intl` (`src/i18n/config.ts`, `src/i18n/request.ts`)
- Player and team names are data, not UI strings, and are never translated — only labels, positions, and status badges switch language

## Testing

There is no unit test framework wired up for this app. Verification is a manual walkthrough against a running backend (dashboard/season switching, leaderboard filters, player profiles, language toggle, theme toggle, and the full admin flow — login, matchday creation, stat entry including the zero-stat exclusion behavior, roster deactivate/reactivate, and session expiry).

## Architecture

- **App Router, Server Components by default**: pages that only read data (dashboard, leaderboard, player profile) are Server Components fetching directly from the Go API using `ONCARINHO_API_URL`; anything interactive (season selector, stat-entry grid, admin forms, nav toggles) is a Client Component.
- **`src/lib/api/client.ts`**: single typed fetch wrapper used by both server and client code. On the server it calls the Go API directly; in the browser it calls same-origin `/api/*` (proxied by Next.js) so the `admin_session` cookie stays same-origin. A 401 on any authenticated call redirects the browser to `/admin`.
- **No client-side data store**: each page/component fetches what it needs; there's no global cache or state management layer, matching the read-mostly, low-traffic nature of the app.
- **Styling**: Tailwind with the shared `@movoz/tailwind-config` and `@movoz/theme` design tokens (`zen-*` CSS variables), so light/dark mode is a single class toggle on `<html>` rather than per-component light/dark variants.

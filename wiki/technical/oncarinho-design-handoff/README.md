# Handoff: Oncarinho Stats Site — Public Prototype

> **Archival note**: this is the raw design-tool handoff as received, kept for reference/provenance. Its "Modernist" visual system (colors, Archivo font, zero border-radius, red accent — see Design Tokens below) was **not adopted** for Oncarinho, which uses the existing shared `@movoz/theme` system instead. Only its layout, content, and interaction decisions were incorporated, into [oncarinho-ui-design-brief.md](../oncarinho-ui-design-brief.md) — treat that file, not this one, as the authoritative frontend design reference. `prototype-reference.html` in this folder is the original interactive prototype (open directly in a browser) — useful for seeing the intended layout/interactions in motion, not for copying its styling. The design-system bundle it depended on (`_ds/`) was not archived, since it represents the un-adopted visual direction.

## Overview
Interactive prototype of Oncarinho, a public stats-tracking site for a football team: a team dashboard, a filterable leaderboard, and player profiles. Built against the PRD `Product Spec: Oncarinho` (Next.js 14 + Go REST API + PostgreSQL target stack). This prototype covers only the public-facing pages (F5, F6, F7 from the spec) — no admin/auth flows were built.

## About the Design Files
The bundled file (`Oncarinho Prototype.dc.html`) is a **design reference built in HTML** — a clickable prototype showing intended layout, content, and interaction, not production code to copy directly. The task is to recreate this UI in the target codebase's actual stack (Next.js 14 / React, per the PRD) using its established patterns, componentizing what's here rather than porting the HTML/inline-style markup as-is.

## Fidelity
**High-fidelity.** Colors, typography, spacing, and component styling are final, drawn from the attached "Modernist" design system (see Design Tokens below). Recreate pixel-close using the codebase's real design-system implementation (or, if none exists yet, implement these exact tokens as the Tailwind/CSS config).

## Screens / Views

### 1. Team Dashboard (`/`)
- **Purpose**: At-a-glance summary of the current season; jump-off point to leaderboard and player profiles.
- **Layout**: Max-width 1100px centered container, 48px top / 32px side / 96px bottom padding. Header row: page title (left) + season segmented-control (right), space-between, wraps on narrow widths. Below a 2px full-width divider (`.hr`), a two-column grid (1.2fr / 1fr, 48px gap): left column is "Top Scorers" (table, top 5 by goals for the selected season, linked player names, "View full leaderboard →" ghost button below); right column is "Roster" — a flush list of active players, name (link) left / position tag right, rows separated by 1px `--color-divider` lines.
- **Components**: Nav bar (see Global Nav below) · Season segmented control (`.seg`/`.seg-opt`, one option per season present in the data) · Table (`.table`) with columns #, Player, Goals · Tag (`.tag tag-neutral`) for position labels · Ghost button (`.btn.btn-ghost`).
- **Content/copy**: Headings "Team Dashboard", "Top Scorers — {season}", "Roster". Empty state: "No goals recorded yet for this season."

### 2. Leaderboard (`/leaderboard`)
- **Purpose**: Full ranked view across all players, filterable by season and stat.
- **Layout**: Page title, then a flex row of two filter fields (Season, Stat), each a `.field` + `.seg` segmented control. Below, a full-width table.
- **Components**: Season filter options = "All-time" + one per season in the data. Stat filter options = Goals / Assists / Cards. Table columns: #, Player, Position, Status, {selected stat} (right-aligned). Status is a tag (`tag-accent` = Active, `tag-neutral` = Inactive). Rows sorted descending by the selected stat value; players with 0 for the selected stat are excluded.
- **Content/copy**: Heading "Leaderboard". Empty state: "No stats recorded for this selection."

### 3. Player Profile (`/players/[id]`)
- **Purpose**: One player's career record.
- **Layout**: Ghost "← Back" button top-left. Header row: player name (h1) + position tag + status tag, inline, wrapping. 2px divider. "All-time Totals": 5-column grid of stat cards (Matches, Goals, Assists, Yellow Cards, Red Cards), each a `.card` with `.card-kicker` label + `.card-title` big number. "Season by Season": table with one row per season the player has stats in, columns Season, Matches, Goals, Assists, Yellow, Red (numeric columns right-aligned).
- **Components**: `.card` grid, `.table`.

### Global Nav (all pages)
`.nav` bar: brand wordmark "ONCARINHO" (left), then text links Dashboard / Leaderboard (current page marked via `aria-current="page"` — style per design system's nav active state), an EN/VI language segmented toggle, and a dark/light mode icon-only ghost button (sun icon in dark mode → switch to light; moon-arc icon in light mode → switch to dark). No admin link — admin flows are out of scope for this prototype.

## Interactions & Behavior
- **Season/Stat filters**: native radio-button segmented controls; selecting one recomputes the table client-side (no navigation).
- **Row/roster player names**: clicking navigates to that player's profile (client-side view swap in the prototype; should be a real route/link in production, e.g. `/players/[id]`).
- **Language toggle (EN/VI)**: swaps all UI strings instantly via a lookup dictionary keyed by locale; no page reload. Player names and the team name are never translated (per PRD). Locale choice should persist (PRD calls for a `locale` cookie, `Accept-Language` negotiation as fallback) — the prototype only holds it in memory.
- **Dark mode toggle**: swaps `--color-bg`, `--color-surface`, `--color-text`, `--color-divider` to dark-ramp steps from the same design-system token set (steps 900/800/100 and a translucent divider). Nothing else changes. Prototype holds this in memory only; production should probably persist to `localStorage` or a cookie.
- No loading/error states are modeled (the prototype uses static in-memory data) — production should add standard loading and error handling around the real API calls (see PRD's `GET /api/leaderboard`, `/api/players/:id`, `/api/summary`, `/api/players`).

## State Management
Prototype state (React component state) — mirror the *shape*, not the mechanism, in production (drive dashboard/leaderboard/profile from API responses instead):
- `players`: id, name, position (enum or null), active (bool)
- `matchdays`: id, played_on (date)
- `stats`: matchday_id, player_id, goals, assists, yellow_cards, red_cards
- `selectedPlayerId`, `lbYear` (season filter or `'all'`), `lbStat` (`goals`/`assists`/`cards`), `dashYear`
- `darkMode` (bool), `locale` (`en`/`vi`)
- Aggregation (season totals, all-time totals, leaderboard ranking) is computed client-side in the prototype by grouping `stats` by `matchday_id`'s year; in production this maps directly to the PRD's SQL `GROUP BY` aggregation endpoints (NF3 — always computed at request time, never cached).

## Design Tokens
Modernist design system — link the system's `styles.css` and use its CSS variables; do not hardcode values.
- **Color**: ground `--color-bg` `#f3f2f2`, text `--color-text` `#201e1d`, single accent `--color-accent` `#ec3013` (mono scheme — accent-2 is a stand-in, treat as the same role). Each role has a 100–900 OKLCH tonal ramp (`--color-neutral-100…900`, `--color-accent-100…900`). Light steps (100–300) for tints/hovers/borders, 500 = base, 700–900 for text-on-tint and pressed states.
- **Type**: `--font-heading` and `--font-body` both set to Archivo.
- **Radius**: `--radius-md` = 0 everywhere — never round a corner.
- **Rules**: 2px solid dividers (`.hr`, `--color-divider`) between major sections — not hairlines.
- **Components used**: `.nav`/`.nav-brand`, `.seg`/`.seg-opt`, `.field`, `.table`, `.card`/`.card-kicker`/`.card-title`, `.tag`/`.tag-accent`/`.tag-neutral`, `.btn`/`.btn-primary`/`.btn-secondary`/`.btn-ghost`/`.btn-icon`. Full component reference and source lives in the design-system bundle (`_ds/` folder, includes `styles.css`, `readme.md`, `theme.json`, and example pages per component).
- Button labels are flush-left, never centered (design-system rule).

## Assets
No custom images — Lucide-style inline SVG icons for sun/moon (dark mode toggle) only. No photography used in this prototype.

## Files
- `Oncarinho Prototype.dc.html` — the full interactive prototype (single file; inline styles + inline React-like state logic).
- `_ds/` — the bound Modernist design-system bundle (stylesheet, tokens, component reference pages) the prototype consumes. Treat this as the source of truth for exact color/spacing/type values.

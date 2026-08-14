# Oncarinho — UI Design Brief

> **Purpose**: A self-contained product design brief for generating the Oncarinho frontend UI. Written to be pasted into an external design tool — includes the visual system, page-by-page content/layout, and component inventory needed to design without further repo access.
> **Source of truth**: [wiki/specs/oncarinho.md](../specs/oncarinho.md) (product spec) and `backend/oncarinho/README.md` (live API contract).
> **Status**: Backend complete and merged. This brief covers the frontend (`apps/oncarinho`), not yet built.
> **Revision note**: layout/content decisions for the 3 public pages incorporate a design-tool handoff (archived at [oncarinho-design-handoff/](./oncarinho-design-handoff/)). Its visual system ("Modernist" — flat, zero-radius, Archivo, red accent) was **not adopted**; Oncarinho keeps the existing shared `@movoz/theme` system below. Only its layout/content/interaction decisions were folded in.

## Product in one paragraph

Oncarinho is a stats-tracking site for a football team. An admin logs each matchday's per-player goals, assists, and cards behind a simple password gate. Everyone else — teammates and the public — browses a team dashboard, year/all-time leaderboards, and individual player profiles. No accounts for players, no match opponents/scores — just who played, and what they did.

## Design system to reuse

This app is a new zone in an existing multi-zone Next.js site (movoz) and must visually match its siblings. Reuse this exact system rather than inventing a new one — the "design work" for Oncarinho is applying these tokens to sports-stats-specific layouts, not choosing new colors/type.

**Color** (light / dark — CSS-variable driven, both must be designed):

| Role | Light | Dark |
|---|---|---|
| Background | `#fdf6e3` (warm cream) | `#1a1a1a` |
| Text | `#3d3520` | `#e8e4dc` |
| Muted text | `#857a5c` | `#a09080` |
| Subtle surface (cards, table stripes) | `#f5edd8` | `#252525` |
| Border | `#e8dfc5` | `#3a3a3a` |
| Paper (elevated surface) | `#fffaed` | `#222222` |
| Accent (primary/CTA) | `#d4775c` (warm rust/orange), light `#e08a70`, dark `#c46448` | same |

Overall feel: warm, paper/parchment-like in light mode; Kindle-inspired dark mode. Not a stark "sports app" palette — keep it calm and editorial, accent color used sparingly for emphasis (top scorer highlight, active nav, primary buttons).

**Typography**: Rubik (sans, body/headings), Libre Baskerville (serif, display headings), Inter (ui, small UI text/labels). Sizes follow a standard type scale (12–72px). Headings can use the serif for a bit of editorial warmth (e.g., team name, page titles); body/data uses sans or ui.

**Spacing**: 4px base scale (Tailwind-standard: 0, 2, 4, 6, 8, 10px... up).

**Existing component primitives already available** (design should compose these, not redesign them): `Button` (primary/secondary/ghost/danger, 3 sizes), `Text` (polymorphic, size/weight/color/font props), `Input` (label/error/helper), `Card`, `Badge`, `Avatar`, `IconButton`, `Stack`, `Container`, `Divider`, `Modal`, `Toast`, `Dropdown`. **No Table/DataGrid component exists yet** — the leaderboard and stat-entry grid need a new tabular component; design it in the same visual language as Card/Badge (subtle borders, `--zen-subtle` row striping, not heavy grid lines).

## Global layout (all pages)

- Top nav bar: team/site name (left, serif wordmark), primary nav links — Dashboard / Leaderboard (center or left-aligned), a language toggle (EN | VI, or a globe-icon `Dropdown` — see Internationalization below), theme toggle (light/dark, sun/moon icon button, syncs across the whole movoz site via shared localStorage), and an unobtrusive "Admin" link (right, low visual weight — this is not a consumer-facing CTA).
- Served under a path prefix (e.g. `/oncarinho`) per the site's multi-zone architecture — footer can be minimal or shared with the rest of the site.
- Fully responsive: single-column stacking on mobile, nav collapses to a simple bar (no complex hamburger menu needed — there are only 2-3 public nav items plus the language/theme toggles, which can collapse into a single overflow menu on narrow screens).

## Internationalization

Full UI (public and admin) is available in English and Vietnamese — see spec F9/NF4. Defaults to the visitor's browser language; a manual EN/VI toggle in the nav overrides it and persists across visits (cookie-based). No localized URLs — `/leaderboard` is the same path in both languages.

Design implications:
- Every string in every mock/screen should be treated as a translation key, not fixed text — expect Vietnamese text to run noticeably longer than English for some labels (e.g. button/badge text), so avoid fixed-width containers around short text.
- Dates render locale-aware (e.g. "March 15, 2026" vs "15 Th3, 2026") — don't hardcode a date format in any mockup.
- Player names and the team name are never translated — they're user data, not UI chrome.
- Position labels (Goalkeeper/Defender/Midfielder/Forward) ARE translated — the admin player form is a dropdown of these 4 fixed options, not free text (this is a backend contract change, not just a UI one — see spec Data Model).

## Public pages

### 1. Team Dashboard — `/`

**Purpose**: landing page, at-a-glance year overview.

**Data** (from `GET /api/summary?year=YYYY`, `GET /api/players`, `GET /api/leaderboard?year=YYYY&stat=goals`):
- Summary stat tiles, current year: Matches Played, Goals Scored, Roster Size (3 tiles in a row, stacking to 1 column on mobile). Big number + small label, e.g. "14 / Matches Played (2026)".
- Top Scorers: top 5 by goals for the selected season, as a table (#, Player, Goals — player names link to profile), with a "View full leaderboard →" ghost-style button below it.
- Roster: a flush list of active players — name (link) left, position tag right, rows separated by a 1px divider (not a card grid).

**Layout**: hero-ish header (team name, tagline), a header row below it with the page title and a **season selector** (segmented control — one option per year present in the data, not a dropdown; this app never has enough years to need one), then the 3 stat tiles, then a strong divider, then a two-column section on desktop — **1.2fr Top Scorers / 1fr Roster, ~48px gap** — collapsing to stacked on mobile. Centered container, generous top/bottom padding (roomier than a typical dashboard — this is closer to an editorial page than a data console).

**Empty state**: brand-new team with zero matchdays — tiles show 0, Top Scorers shows "No goals recorded yet for this season," roster list still shows if players exist.

### 2. Leaderboard — `/leaderboard`

**Purpose**: full ranked stat tables.

**Data** (`GET /api/leaderboard?year=YYYY|all&stat=goals|assists|cards`):
- Controls row: a **season segmented control** (options: "All-time" + one per year in the data) and a **stat segmented control** (Goals / Assists / Cards) — both native-feeling segmented controls, not dropdowns.
- Ranked table columns: **#, Player, Position, Status, {selected stat}** (stat column right-aligned). Player name links to profile. Status is a tag: Active vs. Inactive (deactivated players still appear — their historical stats count, per spec NF3/data model — but the tag makes it clear who's still on the team).
- **Players with a value of 0 for the currently-selected stat are excluded from the ranking** — a "Goals" leaderboard shouldn't list everyone who's ever recorded a single assist. This means the backend leaderboard query needs a value > 0 filter, and the response needs to carry each player's `position` and `is_active` alongside the existing `player_id`/`player_name`/`value` (not in the current API response shape — a small backend addition to make when this page is built).
- Highlight rank 1 (subtle accent-colored row or badge — use `Badge`, keep tone consistent with the calm palette, not a literal trophy emoji).
- "Cards" combines yellow + red into one number per the API — no need to show them split in the table (a future iteration could add a breakdown, not in v1 scope).

**Layout**: controls row pinned at top, table fills remaining width, full-width on mobile with columns kept tight (this is the page most likely to be checked on a phone after a match) — Position/Status can collapse to a single compact column on narrow screens if the full row doesn't fit.

**Empty state**: no data for the selected year/stat → "No stats recorded for this selection." (simple centered message, not a broken-looking empty table).

### 3. Player Profile — `/players/[id]`

**Purpose**: one player's career record.

**Data** (`GET /api/players/:id`):
- A ghost-style "← Back" button, top-left, above everything else (returns to leaderboard or dashboard, whichever referred).
- Header row: name (as a heading), position tag, and an "Inactive" badge if the player has left the team (deactivated players are still browsable per spec — this must be visually clear but not alarming, e.g. a neutral gray `Badge`, not red) — inline, wrapping on narrow screens. A strong divider below the header.
- All-time totals: same stat-tile treatment as the dashboard (Matches Played, Goals, Assists, Yellow Cards, Red Cards) — 5 tiles, wrap to 2-3 columns on smaller screens.
- Year-by-year table: one row per year the player has stats for, columns = Year, Matches, Goals, Assists, Yellow, Red (numeric columns right-aligned). Most recent year first.

**Layout**: back button, header block, divider, all-time tiles, then the year-by-year table below.

## Admin pages (behind password gate, low visual polish priority relative to public pages — utilitarian is fine)

### 4. Admin Login — `/admin`

Single centered card: "Admin" heading, one password `Input`, one `Button` ("Log in"). Error state: inline message below the input on wrong password ("Incorrect password"). No "forgot password" flow (single shared password, out of scope per spec).

### 5. Admin Matchdays — `/admin/matchdays`

**Purpose**: create matchdays, enter/edit/delete a matchday's stats.

**List view**: a simple list/table of matchdays (date, descending), each row clickable, a "+ New Matchday" `Button` at the top opening a small `Modal` or inline form with a single date `Input`.

**Stat-entry view** (after clicking a matchday): a grid/table with one row per active player — columns: Player, Goals, Assists, Yellow Cards, Red Cards, each a small number `Input` (default 0), plus a per-row "remove" `IconButton` for players who shouldn't be in this matchday (maps to `DELETE .../stats/:playerId`). A single "Save" `Button` at the bottom submits the whole grid as one bulk upsert (`PUT /api/matchdays/:id/stats`). Pre-fill existing values if editing a previously-entered matchday (`GET /api/matchdays/:id/stats`). This is the single most-used admin screen — optimize for fast keyboard-driven data entry (tab order flows naturally row by row), not visual flourish.

**Validation feedback**: negative numbers rejected inline (API returns 400) — show a simple inline error rather than a toast for this, since it's a form-validation case not a system error.

### 6. Admin Players — `/admin/players`

**Purpose**: manage the roster.

A table: Name, Position (translated label), Status (Active/Inactive badge), actions (Edit, Deactivate/Reactivate depending on current state). "+ Add Player" button opens a small form/modal (Name required text input; Position optional — a dropdown of the 4 fixed, translated options, not free text). Toggle or filter to show inactive players too (maps to `?active=all`) — default view shows active only, matching the public roster.

## Component inventory to design (beyond existing `@movoz/ui-web` primitives)

1. **StatTile** — big number + label, optional trend/context sub-label (e.g. "(2026)"). Used on Dashboard and Player Profile.
2. **LeaderboardTable** — #/name/position/status/value columns, rank-1 highlight treatment, zero-value rows excluded.
3. **YearStatTable** — year/matches/goals/assists/yellow/red, used on Player Profile.
4. **RosterListItem** — flush list row: name (link) left, position tag right, divider between rows (not a card grid).
5. **MatchdayStatGrid** — the admin stat-entry table described above; this is the most complex custom component in the app.
6. **SeasonSelector** — segmented control (radio-group style, not a dropdown), options = each year with data + "All-time". Used on both Dashboard and Leaderboard.
7. **StatTypeTabs** — Goals / Assists / Cards segmented control for the leaderboard.
8. **LanguageToggle** — EN/VI switch in the global nav, compact enough to sit alongside the theme toggle without crowding.

## Interaction/tone notes

- No player self-service accounts, so there's no "your stats" personalization — every visitor sees the same public view.
- Admin surface should feel clearly distinct from the public surface (e.g., a persistent thin colored bar or "Admin Mode" label) so it's never ambiguous which mode you're in, especially since both are served from the same domain.
- Keep celebratory/game-y flourishes minimal — this is a small hobby team's record book, not a competitive esports dashboard. Warm and understated over flashy.

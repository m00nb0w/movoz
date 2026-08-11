# Oncarinho — UI Design Brief

> **Purpose**: A self-contained product design brief for generating the Oncarinho frontend UI. Written to be pasted into an external design tool — includes the visual system, page-by-page content/layout, and component inventory needed to design without further repo access.
> **Source of truth**: [wiki/specs/oncarinho.md](../specs/oncarinho.md) (product spec) and `backend/oncarinho/README.md` (live API contract).
> **Status**: Backend complete and merged. This brief covers the frontend (`apps/oncarinho`), not yet built.

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

- Top nav bar: team/site name (left, serif wordmark), primary nav links — Dashboard / Leaderboard (center or left-aligned), theme toggle (light/dark, sun/moon icon button, syncs across the whole movoz site via shared localStorage), and an unobtrusive "Admin" link (right, low visual weight — this is not a consumer-facing CTA).
- Served under a path prefix (e.g. `/oncarinho`) per the site's multi-zone architecture — footer can be minimal or shared with the rest of the site.
- Fully responsive: single-column stacking on mobile, nav collapses to a simple bar (no complex hamburger menu needed — there are only 2-3 public nav items).

## Public pages

### 1. Team Dashboard — `/`

**Purpose**: landing page, at-a-glance year overview.

**Data** (from `GET /api/summary?year=YYYY`, `GET /api/players`, `GET /api/leaderboard?year=YYYY&stat=goals`):
- Summary stat tiles, current year: Matches Played, Goals Scored, Roster Size (3 tiles in a row, stacking to 1 column on mobile). Big number + small label, e.g. "14 / Matches Played (2026)".
- Leaderboard preview: top 3-5 scorers for the current year, name + goal count, with a "View full leaderboard →" link.
- Roster list: simple grid/list of active player names (+ position if set), each linking to their profile. No photos in v1 (spec has no player photo field) — use initials-in-circle avatars (the existing `Avatar` component, likely supports initials fallback) for visual texture.

**Layout**: hero-ish header (team name, maybe a one-line tagline), then the 3 stat tiles, then a two-column section on desktop (leaderboard preview | roster list) collapsing to stacked on mobile.

**Empty state**: brand-new team with zero matchdays — tiles show 0, leaderboard preview shows "No matches recorded yet for 2026", roster list still shows if players exist.

### 2. Leaderboard — `/leaderboard`

**Purpose**: full ranked stat tables.

**Data** (`GET /api/leaderboard?year=YYYY|all&stat=goals|assists|cards`):
- Controls row: a year selector (dropdown: current year, prior years, "All-time") and a stat selector (tabs or segmented control: Goals / Assists / Cards).
- Ranked table: rank number, player name (linking to profile), value. Highlight rank 1 (subtle accent-colored row or badge, e.g. a small "🏆"-style badge — use `Badge` component, not an emoji necessarily, keep tone consistent with the calm palette).
- "Cards" combines yellow + red into one number per the API — no need to show them split in the table (a future iteration could add a breakdown, not in v1 scope).

**Layout**: controls row pinned at top, table fills remaining width, full-width on mobile with the rank/name/value columns kept tight (this is the page most likely to be checked on a phone after a match).

**Empty state**: no data for the selected year/stat → simple centered message, not a broken-looking empty table.

### 3. Player Profile — `/players/[id]`

**Purpose**: one player's career record.

**Data** (`GET /api/players/:id`):
- Header: avatar (initials), name, position (if set), an "Inactive" badge if the player has left the team (deactivated players are still browsable per spec — this must be visually clear but not alarming, e.g. a neutral gray `Badge`, not red).
- All-time totals: same stat-tile treatment as the dashboard (Matches Played, Goals, Assists, Yellow Cards, Red Cards) — 5 tiles, wrap to 2-3 columns on smaller screens.
- Year-by-year table: one row per year the player has stats for, columns = Year, Matches, Goals, Assists, Yellow, Red. Most recent year first.

**Layout**: header block, then all-time tiles, then the year-by-year table below.

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

A table: Name, Position, Status (Active/Inactive badge), actions (Edit, Deactivate/Reactivate depending on current state). "+ Add Player" button opens a small form/modal (Name required, Position optional). Toggle or filter to show inactive players too (maps to `?active=all`) — default view shows active only, matching the public roster.

## Component inventory to design (beyond existing `@movoz/ui-web` primitives)

1. **StatTile** — big number + label, optional trend/context sub-label (e.g. "(2026)"). Used on Dashboard and Player Profile.
2. **LeaderboardTable** — rank/name/value, sortable-feeling but not necessarily interactive-sortable in v1, rank-1 highlight treatment.
3. **YearStatTable** — year/matches/goals/assists/yellow/red, used on Player Profile.
4. **RosterListItem / PlayerCard** — avatar + name (+ position), links to profile; needs an "inactive" visual variant.
5. **MatchdayStatGrid** — the admin stat-entry table described above; this is the most complex custom component in the app.
6. **YearSelector** — dropdown/segmented control, options = each year with data + "All-time".
7. **StatTypeTabs** — Goals / Assists / Cards toggle for the leaderboard.

## Interaction/tone notes

- No player self-service accounts, so there's no "your stats" personalization — every visitor sees the same public view.
- Admin surface should feel clearly distinct from the public surface (e.g., a persistent thin colored bar or "Admin Mode" label) so it's never ambiguous which mode you're in, especially since both are served from the same domain.
- Keep celebratory/game-y flourishes minimal — this is a small hobby team's record book, not a competitive esports dashboard. Warm and understated over flashy.

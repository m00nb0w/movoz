# Product Spec: Scout

> **Status**: Draft
> **Author**: To Ngoc Long
> **Date**: 2026-08-14
> **Project**: scout
> **Location**: `apps/scout/` (frontend), `backend/scout/` (backend)
> **Stack**: Next.js 14 + Go (REST API) + PostgreSQL

## Overview

Scout is a private tool for tracking your engineering team's performance. It combines a FIFA-style attribute card per engineer (a growable set of main attributes, each broken into sub-attributes, scored via biweekly peer-relative ranking), data-driven metrics synced from GitHub and Jira (PRs, tickets, complexity), a conversational AI assistant that helps you turn observations into rankings each cycle, and a manually-logged highlight/lowlight history per engineer.

## Problem Statement

Tracking engineer performance informally (memory, scattered notes) makes it hard to see how someone is trending over time, to compare strengths/weaknesses across the team consistently, or to ground performance conversations in both quantitative signals (PR/ticket activity) and your own qualitative judgment. Scout gives you one place to log both, on a recurring cadence, with a low-friction AI-assisted ranking flow instead of a blank form every two weeks.

## Goals

- Give each engineer a FIFA-style attribute card (starting with 6 main attributes, each with a growing list of sub-attributes, and the ability to add further main attributes later) that updates every 2 weeks
- Score sub-attributes via forced ranking against the rest of the active roster, not absolute judgment — this keeps ratings comparable across the team
- Automatically sync PR (raised/reviewed) and Jira ticket (closed, complexity) metrics per engineer as supporting context, shown as their own stats view
- Provide a conversational AI assistant that takes your natural-language observations for the cycle and proposes a ranking (with rationale) for you to review and adjust before saving
- Let you manually log dated highlights/lowlights per engineer, with the system catching likely duplicates before you save
- Keep a full history of past cycles so attribute/overall scores show a trend over time, like a FIFA player card across seasons

## Non-Goals

- Not a public-facing tool — every route requires authentication, there is no visitor-facing view at all
- Not a multi-manager/org platform — single admin (you), single team
- No integration with or reuse of the existing `axon/team-metrics` tool — this syncs its own GitHub/Jira data independently
- No automatic/standalone performance rating in v1 — you want this eventually but haven't defined the scale/cadence yet (see Open Questions)

## User Stories

- As the **manager**, I want to add/edit/deactivate engineers on my roster, linked to their GitHub username and Jira account, so metrics sync correctly.
- As the **manager**, I want to add new main attributes at any time (in addition to sub-attributes), so the card can grow beyond the initial 6 categories.
- As the **manager**, I want to add new sub-attributes under any main attribute at any time, so the card can grow to reflect what I actually want to track.
- As the **manager**, I want the system to automatically pull each engineer's PR and ticket activity so I have data-driven context without manual tracking.
- As the **manager**, I want to start a new biweekly cycle and, for each sub-attribute, rank all active engineers 1..N against each other, so scores stay relative and comparable.
- As the **manager**, I want a chat interface where I describe my observations for the cycle in plain language, and get a proposed ranking (with rationale) back that I can edit before saving, so I don't have to fill out a bare ranking form from scratch.
- As the **manager**, I want to manually log a dated highlight or lowlight for an engineer, with the system flagging likely-duplicate entries (via semantic comparison) before I save, so the history stays clean and doesn't accumulate near-identical notes.
- As the **manager**, I want to view an engineer's card (overall + main-attribute scores) and see how it's trended across past cycles, so I can track growth over time.
- As the **manager**, I want to view all engineers' overall scores and ratings for a given review cycle (not just one engineer at a time), so I can compare the whole team as of that cycle.
- As the **manager**, I want a dashboard of the whole roster with each engineer's latest overall score and last-updated cycle, so I get a quick team-wide view.

## Requirements

### Functional

| # | Requirement | Priority |
|---|---|---|
| F1 | Single shared-password login gates every route; no public/unauthenticated access anywhere | Must have |
| F2 | Admin can add, edit, and deactivate engineers (name, role, GitHub username, Jira account ID, start date) | Must have |
| F3 | Admin can add new main attributes at any time (seeded initially with 6: Technical Expertise, Critical Thinking, Communication, Management, Product Mindset, Force Multiplier), and add/edit/deactivate sub-attributes under any main attribute | Must have |
| F4 | A scheduled sync worker pulls PR-raised, PR-reviewed, tickets-closed, and a complexity figure per engineer from GitHub and Jira into periodic metric snapshots | Must have |
| F5 | Admin can view an engineer's raw metrics over time (PRs, tickets, complexity) as a standalone stats panel | Must have |
| F6 | Admin can open a new biweekly rating cycle and, per sub-attribute, produce a strict 1..N ranking of all active engineers (no ties) | Must have |
| F7 | Each sub-attribute ranking is converted to a score via linear interpolation: rank 1 -> 100, rank N -> 50, evenly spaced across ranks | Must have |
| F8 | Main-attribute score = average of its sub-attributes' scores for that cycle; Overall score = average of the main attributes that existed as of that cycle, computed per engineer per cycle | Must have |
| F9 | Conversational AI assistant: admin describes observations for engineers in natural language; assistant proposes a rank ordering per sub-attribute with a short rationale, which the admin can edit before the ranking is saved | Must have |
| F10 | Engineer card view: current Overall + main-attribute scores, plus a trend view across past cycles | Must have |
| F11 | Roster dashboard: all engineers with latest Overall score and most recent cycle date | Must have |
| F12 | Standalone automatic performance rating (distinct from the Overall attribute score) | Deferred — TBD, see Open Questions |
| F13 | Admin can manually add a dated highlight or lowlight entry for an engineer | Must have |
| F14 | When adding a highlight/lowlight, the AI assistant checks the new entry against that engineer's existing entries for semantic duplicates and flags likely matches before save (admin can override and save anyway) | Must have |
| F15 | Cycle view: for any past review cycle, list all engineers with their Overall and main-attribute scores as of that cycle, so the team can be compared at a single point in time | Must have |

### Non-Functional

| # | Requirement | Metric |
|---|---|---|
| NF1 | No public access | Every route (frontend and API) requires a valid session; no unauthenticated group exists |
| NF2 | Scores are always derivable, no drift | Main-attribute and Overall scores are computed on read from stored per-cycle rankings, not cached/precomputed tables |
| NF3 | AI suggestions/flags never silently block or auto-apply | A ranking proposed by the AI assistant is only persisted after explicit admin save/confirm; a flagged possible-duplicate highlight/lowlight can still be saved if the admin confirms it isn't a duplicate |
| NF4 | Sync worker is safe to re-run | GitHub/Jira sync upserts metric snapshots idempotently; a failed or repeated sync run does not duplicate or corrupt existing data |

## Architecture

- **`apps/scout`** — new Next.js 14 zone, following the multi-zone pattern used by `personal-site`/`oncarinho` (own `basePath`/`assetPrefix`, Nginx route). Uses `@movoz/tailwind-config` and `@movoz/theme`. All pages sit behind auth — there is no public route group at all, unlike `oncarinho`.
- **`backend/scout`** — new Go REST API, following the `hustle-turtle`/`oncarinho` layout (`cmd/server/`, `internal/{config,database,handlers,models,store}/`, `migrations/`), backed by its own PostgreSQL database.
- **Sync worker**: a scheduled job (ticker-based within the server process, or a separate `cmd/syncer` entrypoint — exact shape decided at plan stage) polls GitHub and Jira APIs per active engineer and upserts into `metric_snapshots`. Requires GitHub/Jira API credentials as env vars.
- **AI assistant**: the Go backend calls the Claude API (Anthropic Messages API) server-side for two flows: (1) the conversational ranking chat, streaming responses to a chat UI in `apps/scout`, given the cycle's synced metrics and existing highlights/lowlights as context; and (2) a semantic duplicate check on new highlight/lowlight entries against that engineer's existing entries. Requires an Anthropic API key env var.
- **Scoring**: computed on read (see NF2) — small scale (roughly a dozen engineers, ~24 cycles/year) means no caching is needed, mirroring `oncarinho`'s aggregation approach.

## Data Model

```
engineers
  id, name, role, github_username, jira_account_id, started_at, is_active, created_at

main_attributes
  id, key, name, created_at
  -- admin-managed; seeded with the initial 6 (technical_expertise, critical_thinking,
  -- communication, management, product_mindset, force_multiplier), more can be added later

sub_attributes
  id, main_attribute_id (FK), name, description, is_active, created_at
  -- admin-managed, grows over time

rating_cycles
  id, period_start (date), period_end (date), created_at

sub_attribute_rankings
  id, cycle_id (FK), sub_attribute_id (FK), engineer_id (FK),
  rank (int, 1..N), score (numeric, derived: 100 - (rank-1) * 50/(N-1))
  unique(cycle_id, sub_attribute_id, engineer_id)
  unique(cycle_id, sub_attribute_id, rank)

ai_ranking_sessions
  id, cycle_id (FK), sub_attribute_id (FK), transcript (jsonb), proposed_ranking (jsonb), created_at
  -- stores the chat + AI's proposed ranking; accepted rankings are written to sub_attribute_rankings on save

metric_snapshots
  id, engineer_id (FK), period_start (date), period_end (date),
  prs_raised (int), prs_reviewed (int), tickets_closed (int), complexity_score (numeric),
  synced_at

highlight_entries
  id, engineer_id (FK), kind (highlight|lowlight), body (text), created_at
```

- N in the ranking formula = count of active engineers ranked in that cycle for that sub-attribute.
- Main-attribute score per engineer per cycle = avg of `sub_attribute_rankings.score` for that engineer/cycle across sub-attributes under that main attribute.
- Overall score per engineer per cycle = avg of the main attributes that **existed as of that cycle** — adding a new main attribute later only affects cycles from that point forward; past cycles' Overall is never recalculated (see F8).
- Deactivating an engineer is a soft delete (`is_active = false`) so historical cycles/rankings stay intact, mirroring `oncarinho`'s roster pattern.
- A `highlight_entries` row is a standalone, manually-added note (not tied to a rating cycle) tagged as either a highlight or a lowlight; duplicate-checking (F14) happens at write time via the AI assistant and is not itself persisted as a separate record — only the accepted entry is stored.

## API Design (indicative — refined at plan stage)

```
POST /api/auth/login                                    — shared password -> session cookie

GET/POST/PUT   /api/engineers[, /:id]                    — roster CRUD
GET/POST/PUT   /api/main-attributes[, /:id]               — add/edit main attributes
GET/POST/PUT   /api/sub-attributes[, /:id]                — manage sub-attributes per main attribute

GET  /api/cycles                                          — list all rating cycles
POST /api/cycles                                          — open a new rating cycle
PUT  /api/cycles/:id/sub-attributes/:subId/ranking        — save the 1..N ranking for a sub-attribute
GET  /api/cycles/:id/scores                                — all engineers' overall + main-attribute scores for that cycle
GET  /api/engineers/:id/card?cycleId=                     — overall + main-attribute scores for a cycle
GET  /api/engineers/:id/trend                             — scores across all past cycles

GET  /api/engineers/:id/metrics                           — raw synced GitHub/Jira metrics over time

POST /api/cycles/:id/ai-sessions                          — start/continue an AI ranking chat for a sub-attribute
POST /api/cycles/:id/ai-sessions/:sessionId/accept        — persist the (possibly edited) proposed ranking

GET/POST /api/engineers/:id/highlights                    — list / add a highlight or lowlight entry
POST /api/engineers/:id/highlights/check-duplicate         — AI semantic duplicate check against existing entries (F14), run before save
```

## Error Handling & Testing

- Validation: ranking submissions must cover exactly the active roster with a strict 1..N permutation (no gaps, no ties, no duplicates) or the API rejects with 400.
- Sync worker: idempotent upserts keyed on `(engineer_id, period_start, period_end)`; failures are logged and retried on the next scheduled run rather than partially applied.
- Highlight/lowlight duplicate check: the AI call runs synchronously before save and returns a similarity flag + the matched existing entry (if any); the admin can proceed anyway. If the AI call fails/times out, the save proceeds without a duplicate flag rather than blocking the admin.
- Testing: Go table-driven tests for handlers, scoring math, and sync upsert logic (`go test ./...`), mirroring `hustle-turtle`/`oncarinho`. Frontend covered by TypeScript type-checking and manual verification.

## Success Metrics

| Metric | Target | How Measured |
|---|---|---|
| Full biweekly cycle completion time | Rank all sub-attributes for ~11 engineers via the AI assistant in under 20 minutes | Manual timing during first real use |
| Metrics sync freshness | GitHub/Jira metrics no more than 1 sync interval stale when viewed | Manual check post-deploy |

## Open Questions

- [ ] Standalone automatic performance rating (F12): scale, cadence, and inputs are still TBD — revisit before this is scoped into a plan
- [ ] Which specific GitHub repos and Jira projects/boards to sync — decide at plan/implementation stage
- [ ] Are past rating cycles editable after the fact, or immutable once saved?
- [ ] Sync worker shape: in-process scheduler vs. separate `cmd/syncer` binary

---

description: "Task list for el-storko implementation"
---

# Tasks: el-storko — Personal Work-Item Tracker

**Input**: Design documents from `wiki/technical/001-el-storko/` (plan.md, research.md,
data-model.md, contracts/rest-api.md, quickstart.md) and `wiki/specs/001-el-storko/spec.md`

**Tests**: Included — Constitution Principle VI (Disciplined Implementation Workflow) mandates
TDD (red-green-refactor); tests are written before implementation for each resource, matching the
existing `oncarinho` store/handler test-pair convention.

**Organization**: Tasks are grouped by user story (US1/US2/US3, matching spec.md priorities) so
each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1 (P1, personal tasks/epics), US2 (P2, Jira sync), US3 (P2, CLI + web UI)

## Phase 1: Setup

**Purpose**: Repo scaffolding for all three deliverables, per plan.md's Project Structure.

- [X] T001 Create `backend/el-storko/` Go module (`go.mod`), directories
  `cmd/server/`, `cmd/cli/`, `internal/{config,database,models,store,handlers,jirasync}/`,
  `migrations/`, add `gin-gonic/gin`, `lib/pq`, `golang-migrate/migrate/v4` deps (versions per
  `research.md`, matching `backend/oncarinho/go.mod`)
- [X] T002 [P] Create `backend/el-storko/.env.example` (`DATABASE_URL`, `JIRA_EMAIL`,
  `JIRA_API_TOKEN`, `JIRA_BASE_URL`, `PORT`, `JIRA_SYNC_INTERVAL`) and add
  `backend/el-storko/.env` to the repo root `.gitignore`
- [X] T003 [P] Create `apps/el-storko/` Next.js 14 zone skeleton: `package.json` (deps
  `@movoz/theme`, `@movoz/tailwind-config`, `@movoz/tsconfig` as `workspace:*`, matching
  `apps/oncarinho/package.json` minus `next-intl`), `next.config.mjs` (`transpilePackages`,
  `/api/:path*` rewrite to `EL_STORKO_API_URL`, default `http://localhost:8082`), `tsconfig.json`
  extending `@movoz/tsconfig/nextjs.json`, dev port `3101`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared model, schema, config, and server bootstrap that every user story depends on.

- [X] T004 Define `WorkItem` struct and `Type`/`Status`/`Source` enum types in
  `backend/el-storko/internal/models/work_item.go`, matching `data-model.md`'s column list
- [X] T005 Write migration `backend/el-storko/migrations/0001_create_work_items.up.sql` /
  `.down.sql` creating `work_items` with the columns, `CHECK` constraints, and
  `ON DELETE SET NULL` FK from `data-model.md`
- [X] T006 [P] Implement `backend/el-storko/internal/config/config.go`: load `.env` (git-ignored)
  plus real env vars, expose `DATABASE_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`, `JIRA_BASE_URL`,
  `PORT` (default 8082), `JIRA_SYNC_INTERVAL` (default 5m); never log the token
- [X] T007 Implement `backend/el-storko/internal/database/database.go`: Postgres connection
  setup + migration runner, mirroring `backend/oncarinho/internal/database`
- [X] T008 Implement `backend/el-storko/cmd/server/main.go` entrypoint: `-migrate=up|down`,
  `-version`, `-auto-migrate` flags (matching `oncarinho`/`hustle-turtle`), Gin router bootstrap,
  listen on `PORT`; route registration and sync-goroutine startup are wired in later phases

**Checkpoint**: `go build ./...` succeeds; `go run ./cmd/server -migrate=up` creates the schema
against a local Postgres DB.

---

## Phase 3: User Story 1 - Track personal tasks and epics (P1) 🎯 MVP

**Goal**: CRUD for Epics/Tasks with parent validation, status, and description, durable across
restarts (spec.md User Story 1).

**Independent Test**: Create an Epic, create two Tasks under it, edit one Task, delete the other,
delete the Epic, restart the server, confirm state persisted and the remaining Task is unparented.

- [X] T009 [P] [US1] Write store tests in
  `backend/el-storko/internal/store/work_item_store_test.go` covering: create Epic, create Task
  with Epic parent, reject Task parent pointing at a non-Epic row, list with `source`/`type`/
  `status`/`parent_id` filters, update, delete Epic unparents its Tasks, delete Task
- [X] T010 [P] [US1] Write handler tests in
  `backend/el-storko/internal/handlers/work_item_handler_test.go` covering the endpoints in
  `contracts/rest-api.md`: `POST/GET/PATCH/DELETE /api/work-items[/:id]`, including the 400 cases
  (Epic with parent, Task parent referencing a non-Epic)
- [X] T011 [US1] Implement `backend/el-storko/internal/store/work_item_store.go`
  (Create/Get/List/Update/Delete + parent/type/status/source validation per `data-model.md`) to
  make T009 pass
- [X] T012 [US1] Implement `backend/el-storko/internal/handlers/work_item_handler.go` and
  register `/api/work-items` routes in `cmd/server/main.go` to make T010 pass
- [X] T013 [US1] Wire `GET /api/work-items` list filters (`source`, `parent_id`, `type`, `status`)
  end-to-end through handler → store
- [X] T014 [US1] Verify: `go test ./...` green; `curl` smoke test from `quickstart.md`
  (create Epic, create Task under it, delete Epic, confirm Task's `parent_id` is now null);
  restart the server process and confirm the same rows are still present

**Checkpoint**: User Story 1 is independently complete and demoable — personal task/epic tracking
works end-to-end without Jira, CLI, or web UI.

---

## Phase 4: User Story 2 - Jira bidirectional sync (P2)

**Goal**: Assigned Jira issues appear locally; local edits on Jira-sourced items push back;
conflicts resolve last-write-wins (spec.md User Story 2).

**Independent Test**: Against a real or sandboxed Jira issue assigned to the user, run one poll
cycle, confirm it appears locally with `source=jira`; edit its status locally, confirm the Jira
issue updates within one cycle; edit the Jira issue directly, confirm the next cycle pulls it in.

- [X] T015 [P] [US2] Write pull-sync tests with a mocked Jira HTTP client in
  `backend/el-storko/internal/jirasync/sync_test.go`: new assigned issue creates a local row with
  `source=jira`/`jira_key`/`jira_url`; an issue already local with an older `updated_at` gets
  overwritten by newer Jira data
- [X] T016 [P] [US2] Write push-sync and conflict tests in the same file: a local `source=jira`
  row edited more recently than Jira's `fields.updated` triggers a PATCH to Jira; the reverse
  (Jira newer) does not push
- [X] T017 [US2] Implement `backend/el-storko/internal/jirasync/client.go`: Jira REST API v3
  client using HTTP Basic auth (`JIRA_EMAIL` + `JIRA_API_TOKEN`) — search assigned issues
  (`assignee = currentUser()`), get issue, update issue status/description; never logs the token
- [X] T018 [US2] Implement `backend/el-storko/internal/jirasync/sync.go`: one poll-cycle function
  doing pull-then-push with last-write-wins comparison, to make T015/T016 pass
- [X] T019 [US2] Wire a ticker goroutine (interval = `JIRA_SYNC_INTERVAL`) in `cmd/server/main.go`
  that calls the sync cycle and logs failures (redacted) without crashing the server (FR-013)
- [ ] T020 [US2] Verify: point at a real/sandbox Jira issue per `quickstart.md`; confirm it syncs
  in both directions within one interval; `grep` the log output and confirm the token never
  appears
- [X] T020a [P] [US2] Write a failure-isolation test in `sync_test.go` that simulates a Jira
  client error (network/401) during a poll cycle and asserts `POST`/`GET /api/work-items` on a
  `personal` item still succeeds and the server does not crash (FR-013)

**Checkpoint**: User Story 2 is independently complete — Jira sync works on top of US1's data
model without touching CLI or web UI code.

---

## Phase 5: User Story 3 - CLI capture and kanban/list views (P2)

**Goal**: Quick CLI capture, kanban board (by status), and filterable flat list, all against the
same data as the API (spec.md User Story 3).

**Independent Test**: `el-storko add`/`list` from a terminal, then open the web UI and confirm the
same item appears on both the board and the list, and that source/Epic filters narrow correctly.

- [X] T021 [P] [US3] Write CLI tests for `add`/`list` in `backend/el-storko/cmd/cli/main_test.go`
  (or an extracted `internal/clicmd` package if that's cleaner to test), using a mock/stub of the
  REST API
- [X] T022 [P] [US3] Implement `backend/el-storko/cmd/cli/main.go`: `el-storko add "<title>"
  [--epic <id>] [--description "..."]` and `el-storko list [--source ...] [--status ...]
  [--epic <id>]`, calling the same `/api/work-items` endpoints as the web UI, per
  `contracts/rest-api.md`'s CLI mapping
- [X] T023 [P] [US3] Implement `apps/el-storko/src/lib/api.ts`: typed client for
  `GET/POST/PATCH/DELETE /api/work-items[/:id]`
- [X] T024 [US3] Implement `apps/el-storko/src/components/KanbanBoard.tsx`: columns for
  `todo`/`in_progress`/`blocked`/`done`, each listing its items
- [X] T025 [US3] Implement `apps/el-storko/src/components/WorkItemList.tsx`: flat list with
  source and Epic filter controls
- [X] T026 [US3] Implement `apps/el-storko/src/app/page.tsx` (board) and a `/list` route, wired
  with `@movoz/theme`, consuming `src/lib/api.ts`
- [ ] T027 [US3] Verify: `pnpm --filter el-storko dev`; in the browser, create/move/filter items;
  separately run the CLI `add` command and confirm the new item appears in the browser without a
  divergent data source

**Checkpoint**: User Story 3 is independently complete — day-to-day CLI + web interaction works
against US1's API (and, if implemented, shows US2's Jira-sourced items too).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Always-on deployment and knowledge-base updates (Constitution Principle V).

- [X] T028 [P] Create `infra/launchd/com.movoz.el-storko-backend.plist` and
  `infra/launchd/com.movoz.el-storko-frontend.plist` (`RunAtLoad`, `KeepAlive`, log paths outside
  the repo) per `quickstart.md`
- [X] T029 [P] Update `wiki/technical/architecture.md` and `wiki/specs/index.md` to reference
  el-storko, per Constitution Principle V (docs updated in the same unit of work)
- [ ] T030 Verify launchd survival: `launchctl load` both plists, `kill -9` the backend process
  and confirm it restarts within seconds, then reboot the machine and confirm both services come
  up without manual intervention
- [X] T031 Final verification pass per `plan.md`'s Verification section: `go build ./... && go
  test ./...`, fresh-DB `-auto-migrate`, full CRUD + Jira sync + CLI/UI parity check, confirm
  `.env` is git-ignored and the token never appears in logs/commits/API responses

---

## Phase 7: User Story 4 - Burn-rate stats (P3)

**Goal**: A Stats tab showing completion throughput and open-backlog trends over a trailing
window, built on a `completed_at` timestamp that's accurate under edits and status flip-flops.

**Independent Test**: Complete items on different days (including one edited afterward, and one
completed/reopened/re-completed), open the Stats tab, and confirm both trends match the actual
history per `spec.md` User Story 4's acceptance scenarios.

- [X] T032 Write migration `backend/el-storko/migrations/000002_add_completed_at.up.sql` /
  `.down.sql` adding nullable `completed_at TIMESTAMPTZ` to `work_items`
- [X] T033 [P] [US4] Add `CompletedAt *time.Time` to `models.WorkItem`
- [X] T034 [P] [US4] Write store tests in `work_item_store_test.go`: creating with `status=done`
  sets `completed_at`; updating status to `done` sets it; updating a `done` item's
  title/description without changing status leaves `completed_at` unchanged; moving a `done` item
  to any other status clears `completed_at` to `NULL`; re-completing after that sets a new,
  later `completed_at`
- [X] T035 [US4] Implement the `completed_at` transition logic in `work_item_store.go`'s
  `Create`/`Update`/`UpdateFromSync` to make T034 pass
- [X] T036 [P] [US4] Write a pure-function test for `internal/stats` (no DB): given a slice of
  `WorkItem`-shaped records with known `created_at`/`completed_at`, assert the computed
  per-day `completed`/`open` series matches by hand-calculated expected values, including a day
  with zero completions and a reopened-then-recompleted item
- [X] T037 [US4] Implement `internal/stats/burnrate.go`'s pure computation function to make T036
  pass, per `data-model.md`'s Burn-Rate Stats section
- [X] T038 [P] [US4] Write handler test for `GET /api/stats/burn-rate` (default window, custom
  `days`, empty-tracker all-zero case) in `internal/handlers/stats_handler_test.go`
- [X] T039 [US4] Implement `internal/handlers/stats_handler.go` and register
  `GET /api/stats/burn-rate` in `cmd/server/router.go` to make T038 pass
- [X] T040 [P] [US4] Add `getBurnRate(days)` to `apps/el-storko/src/lib/api.ts`
- [X] T041 [US4] Implement `apps/el-storko/src/app/stats/page.tsx` plus a small
  `StatCard`/trend-line component under `src/components/`, structured so a future metric is
  another card, not a rework
- [X] T042 [US4] Add a `/stats` entry to `NavTabs.tsx`'s route list
- [X] T043 [US4] Verify: `go test ./...` green; `curl localhost:8082/api/stats/burn-rate` sanity
  check against seeded data; `pnpm --filter el-storko build` succeeds; manual check that editing a
  done item doesn't shift its stats-day, and that reopen→recomplete moves it to the new day

---

## Phase 8: Redesign Foundational — five-state workflow, Estimate/Due date, reference sequences

**Purpose**: Blocking prerequisite for the whole 2026-09-30 redesign (spec User Stories 2, 3
revision, 5, 6, 7) — the schema/model changes every later redesign phase depends on.

- [X] T044 Write migration `backend/el-storko/migrations/000003_workflow_states_backlog_and_picked_for_today.up.sql`
  / `.down.sql`: drop and recreate the `status` CHECK constraint for
  `backlog`/`picked_for_today`/`in_progress`/`blocked`/`done`, backfill existing `todo` rows to
  `backlog`, change the column default to `backlog`. `.down.sql` backfills `backlog`/
  `picked_for_today` rows back to `todo` before restoring the old four-value constraint.
- [X] T045 [P] Write migration `000004_add_estimate_and_due_date.up.sql` / `.down.sql`: add nullable
  `estimate_hours NUMERIC(6,2)` and nullable `due_date DATE` to `work_items`.
- [X] T046 Write migration `000005_add_reference_sequences.up.sql` / `.down.sql`: create
  `epic_reference_seq`/`task_reference_seq`, add nullable `reference_number INTEGER`, backfill
  existing rows per-type via a `row_number() OVER (PARTITION BY type ORDER BY id)` window query,
  advance both sequences past the backfilled max, then set the column `NOT NULL`. `.down.sql`
  drops the column and both sequences.
- [X] T047 [P] Update `models.WorkItem`/`models.Status` in `internal/models/work_item.go`: rename
  `StatusTodo` → `StatusBacklog` (`"backlog"`), add `StatusPickedForToday` (`"picked_for_today"`);
  add `EstimateHours *float64`, `DueDate *string`, `ReferenceNumber int` fields
- [X] T048 [P] Write store tests in `work_item_store_test.go`: an Epic and a Task created in
  sequence get independent `reference_number`s from their own counters (interleave epic/task
  creates and assert no shared counter); `estimate_hours`/`due_date` round-trip through
  Create/Update; a negative `estimate_hours` is rejected
- [X] T049 Implement per-type `reference_number` assignment (`nextval` on the matching sequence)
  and `estimate_hours`/`due_date` persistence in `work_item_store.go` to make T048 pass; add a
  `ReferenceKey()` helper (`EPIC-<n>`/`TASK-<n>`) used by the handler's JSON response, not stored
- [X] T050 [P] [US3] Update `jiraStatusToLocal`/`localStatusToJiraTransitionName` in
  `internal/jirasync/client.go` for the five-state model (unrecognized Jira statuses, including
  "To Do", fall back to `backlog`; a local `picked_for_today` item pushes as Jira's "In Progress"
  transition) — update `sync_test.go`/`client.go` tests first to assert the new mapping, confirm
  red, then implement
- [X] T051 Verify: fresh-DB `-auto-migrate` runs `000001`→`000005` cleanly; `go test ./...` green;
  curl-create an Epic and a Task and confirm their `reference_key`s are independent
  (`EPIC-1`/`TASK-1` on a fresh DB, not sharing a counter)

---

## Phase 9: User Story 2 (spec) - Plan today's work

**Goal**: Board shows only the four active-status columns with the full Backlog listed beneath
it; "Pick for today" works via button or drag; the logo returns to Board.

**Independent Test**: Put items in Backlog, use "Pick for today" on some and drag others onto the
Board, confirm the Board's four columns show exactly those items and the Backlog list below still
shows the rest — matching `spec.md` User Story 2's Independent Test.

- [X] T052 [US2] Update `KanbanBoard.tsx`: columns limited to Picked for today/In progress/
  Blocked/Done (remove Backlog as a column)
- [X] T052a [P] [US2] Delete `app/list/page.tsx` and the now-unused `WorkItemList.tsx`; remove
  the List entry from `NavTabs.tsx`'s `ROUTES` array
- [X] T053 [US2] Add a `BacklogList` section below the board rendering `status=backlog` items,
  each with a "Pick for today" button (`PATCH` status → `picked_for_today` via the existing
  `updateWorkItem` client call)
- [X] T054 [US2] Implement drag-from-Backlog-onto-Board using the native HTML5 Drag and Drop API
  (per `research.md`) with the same status-transition effect as the button
- [X] T055 [P] [US2] Add a small clickable "el-storko" wordmark to `layout.tsx`'s nav bar linking
  to `/` (Board) — FR-021
- [X] T056 [US2] Verify: manual pass confirming exactly 4 Board columns, the Backlog list beneath
  it, both the button and drag paths moving an item to Picked for today, and logo-click landing
  on Board

---

## Phase 10: User Story 3 (spec, revised) - Jira card shows its real key

**Goal**: A Jira-sourced card's badge shows the real Jira key (e.g. `AUTH-142`), not a generic
"Jira" label — the tint from the original FR-014 stays as-is.

- [X] T057 [US3] Update `WorkItemCard.tsx`: the source badge on a Jira-sourced item shows
  `item.jira_key` instead of the literal `source` string; keep the existing left-border tint
- [X] T058 [US3] Verify: seed a `source=jira` row with a `jira_key`, confirm its card badge shows
  the key, not "jira"

---

## Phase 11: User Story 4 (spec) - CLI stays correct under the new status set

- [X] T059 [P] [US4] Audit `cmd/cli`/`internal/clicmd` for any hardcoded status strings from the
  old four-value set and update to the five-value set if found
- [X] T060 [US4] Verify: `el-storko-cli list --status picked_for_today` round-trips correctly

---

## Phase 12: User Story 5 (spec) - Item detail drawer, Estimate/Due date, searchable Epic picker

**Goal**: Every field — including the new Estimate/Due date and the reference key — is visible in
one drawer without internal scrolling; assigning an Epic is a type-to-filter search, not a fixed
list.

**Independent Test**: Open an item's drawer, set an Estimate and Due date, close and reopen it,
confirm both persisted and every field was visible without scrolling — matching `spec.md` User
Story 5's Independent Test.

- [X] T061 [US5] Implement `ItemDrawer.tsx` using `@movoz/ui-web`'s `Modal` (per `research.md`):
  title, description, status, Estimate, Due date, reference key, and Jira key when present, sized
  so nothing scrolls internally at realistic field counts
- [X] T062 [P] [US5] Implement `EpicPicker.tsx`: client-side type-to-filter search over the
  already-fetched Epic list (no new endpoint — per `research.md`), with a "Clear" action (FR-024)
- [X] T063 [US5] Wire Estimate (number input, hours) and Due date (date input) fields in the
  drawer to `PATCH /api/work-items/:id`
- [X] T064 [P] [US5] Implement a shared Due-date color-coding helper (red if overdue, yellow if
  due within 3 days, else unstyled — FR-020) used by both `WorkItemCard.tsx` and `ItemDrawer.tsx`
- [X] T065 [US5] Wire drawer open/close from clicking a card on the Board or Backlog list
- [X] T066 [US5] Verify: set Estimate + Due date, close/reopen the drawer, confirm persistence and
  no internal scrolling; confirm overdue/soon-due color coding on both card and drawer

---

## Phase 13: User Story 6 (spec) - Workload & burn-rate stats expansion

**Goal**: The Stats tab adds open-item count, completed-this-week, completion rate, total
tracked, and a status-breakdown bar chart, alongside the existing throughput/backlog trend.

**Independent Test**: Seed items across all five statuses with a mix of completion dates, open
the Stats tab, confirm every number and the bar chart match a hand count — matching `spec.md`
User Story 6's Independent Test.

- [X] T067 [P] [US6] Write a pure-function test for `internal/stats`'s new summary reducer (open
  count, completed-this-week count, completion rate, total tracked, status breakdown across all
  five statuses including zero-count ones) against hand-calculated values, including an
  empty-tracker all-zero case
- [X] T068 [US6] Implement the summary reducer in `internal/stats` to make T067 pass, per
  `data-model.md`'s Workload & Burn-Rate Stats section
- [X] T069 [P] [US6] Write handler tests for `GET /api/stats/summary` (empty tracker, a tracker
  with a known status/completion composition)
- [X] T070 [US6] Implement `StatsHandler.Summary` and register `GET /api/stats/summary` in
  `router.go` to make T069 pass
- [X] T071 [P] [US6] Add `getStatsSummary(scope)` to `apps/el-storko/src/lib/api.ts`
- [X] T072 [US6] Add `StatusBreakdownChart.tsx` (custom SVG bar chart, per `research.md`) and new
  `StatCard`s (open items, completed this week, completion rate, total tracked) to
  `apps/el-storko/src/app/stats/page.tsx`
- [X] T073 [US6] Verify: seed items across all five statuses with known completion dates, confirm
  every summary number and the bar chart against a hand count

---

## Phase 14: User Story 7 (spec) - Mine/Agent scope toggle

**Goal**: One global switch (default Mine) scopes both Board and Stats to `personal`+`jira` or
`agent` sourced items, with both scopes sharing identical UI for now.

**Independent Test**: Seed a personal, a Jira, and an agent-sourced item; confirm Mine scope
shows the first two and Agent scope shows only the third, on both Board and Stats — matching
`spec.md` User Story 7's Independent Test.

- [X] T074 [P] [US7] Write handler tests for `scope=mine`/`scope=agent` on the existing
  `GET /api/work-items` List handler
- [X] T075 [US7] Implement `scope` query-param handling in `work_item_handler.go`'s `List`
  (translate to a multi-source filter in `store.ListFilters`) to make T074 pass
- [X] T076 [P] [US7] Extend `/api/stats/burn-rate` and `/api/stats/summary` (handlers + tests) to
  accept and honor `scope`, defaulting to `mine`
- [X] T077 [US7] Implement `ScopeToggle.tsx` (Mine/Agent switch, default Mine), wired to shared
  scope state consumed by both the Board and Stats pages' data fetches
- [X] T078 [US7] Verify: seed a personal, a Jira, and an agent-sourced row; confirm Mine shows the
  first two and Agent shows only the third, on both Board and Stats

---

## Phase 15: Redesign Polish

- [X] T079 [P] Update `wiki/technical/architecture.md`'s el-storko summary line if it still
  describes the pre-redesign workflow
- [X] T080 Final full verification pass: `go build ./... && go test ./...`; fresh-DB migrate
  `000001` → `000005`; full CRUD across the five-state status set; Jira status-mapping check
  (mocked); CLI round-trip under the new statuses; `pnpm --filter el-storko build`; manual UI
  pass covering every acceptance scenario across all seven user stories in `spec.md`

---

## Dependencies

- **Setup (T001-T003)** blocks **Foundational (T004-T008)**.
- **Foundational** blocks all user stories (T009-T027).
- **US1 (T009-T014)** must complete before **US2** and **US3**, since both need the
  `work_items` store/API to exist.
- **US2 (T015-T020)** and **US3 (T021-T027)** touch disjoint files (`internal/jirasync/` vs.
  `cmd/cli/` + `apps/el-storko/`) and can proceed in parallel once US1 is done.
- **Polish (T028-T031)** depends on US1 at minimum; T030/T031 depend on US2 and US3 also being
  done for a complete verification pass.
- **Redesign Foundational (T044-T051)** blocks every redesign phase (T052-T078) — all of them
  read or write the new status values, fields, or reference sequences.
- **Board redesign (T052-T056)**, **Jira card key (T057-T058)**, **CLI audit (T059-T060)**,
  **Item drawer (T061-T066)**, and **Stats expansion (T067-T073)** touch disjoint files and can
  proceed in parallel once Redesign Foundational is done.
- **Scope toggle (T074-T078)** touches the same Board/Stats page files as T052-T056 and T067-T073,
  so it should land after those two phases rather than run fully in parallel with them.
- **Redesign Polish (T079-T080)** depends on all of T044-T078.

## Parallel Execution Examples

- After Foundational: T009 and T010 in parallel (different test files).
- After US1 checkpoint: the entire US2 phase (T015-T020) and the entire US3 phase (T021-T027) can
  be assigned to two different subagents running in parallel.
- Within US3: T021, T022, T023 can run in parallel (CLI vs. frontend API client are independent
  files); T024/T025/T026 depend on T023 existing.
- After Redesign Foundational (T044-T051): Board redesign (T052-T056), Jira card key (T057-T058),
  CLI audit (T059-T060), Item drawer (T061-T066), and Stats expansion (T067-T073) can be assigned
  to up to five different subagents running in parallel; bring Scope toggle (T074-T078) in only
  after Board redesign and Stats expansion land.

## Implementation Strategy

**MVP first**: Phase 1 → 2 → 3 (US1) alone is a working, demoable personal tracker — this is the
suggested MVP checkpoint before investing in Jira sync or the CLI/web layer.

**Redesign MVP**: Phase 8 (Redesign Foundational) → Phase 9 (Board redesign) alone gets the core
daily-planning workflow working; Phases 10-14 layer on refinements (Jira key display, drawer,
expanded stats, scope toggle) that are each independently valuable and independently deployable.

**Incremental delivery after MVP**: US2 and US3 can each be added independently and in parallel
once US1 lands; Polish (launchd + docs) is the final step before calling the feature done.

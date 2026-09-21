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

## Dependencies

- **Setup (T001-T003)** blocks **Foundational (T004-T008)**.
- **Foundational** blocks all user stories (T009-T027).
- **US1 (T009-T014)** must complete before **US2** and **US3**, since both need the
  `work_items` store/API to exist.
- **US2 (T015-T020)** and **US3 (T021-T027)** touch disjoint files (`internal/jirasync/` vs.
  `cmd/cli/` + `apps/el-storko/`) and can proceed in parallel once US1 is done.
- **Polish (T028-T031)** depends on US1 at minimum; T030/T031 depend on US2 and US3 also being
  done for a complete verification pass.

## Parallel Execution Examples

- After Foundational: T009 and T010 in parallel (different test files).
- After US1 checkpoint: the entire US2 phase (T015-T020) and the entire US3 phase (T021-T027) can
  be assigned to two different subagents running in parallel.
- Within US3: T021, T022, T023 can run in parallel (CLI vs. frontend API client are independent
  files); T024/T025/T026 depend on T023 existing.

## Implementation Strategy

**MVP first**: Phase 1 → 2 → 3 (US1) alone is a working, demoable personal tracker — this is the
suggested MVP checkpoint before investing in Jira sync or the CLI/web layer.

**Incremental delivery after MVP**: US2 and US3 can each be added independently and in parallel
once US1 lands; Polish (launchd + docs) is the final step before calling the feature done.

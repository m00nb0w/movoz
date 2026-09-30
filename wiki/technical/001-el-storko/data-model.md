# Data Model: el-storko

Phase 1 output for `wiki/technical/001-el-storko/plan.md`. Revised 2026-09-30 for the five-state
workflow, Estimate/Due date fields, per-type reference keys, and Mine/Agent scope filtering.

## Entity: WorkItem

Single table `work_items` backs both Epic and Task (`type` discriminates), matching the spec's
Key Entities section (Work Item / Epic / Task).

| Column             | Type                        | Nullable | Notes |
|--------------------|-----------------------------|----------|-------|
| `id`               | `BIGSERIAL PRIMARY KEY`     | no       | Internal PK only — no longer used to derive the display key (see Reference Keys) |
| `type`              | `TEXT`                      | no       | `epic` \| `task`; `CHECK` constraint |
| `parent_id`         | `BIGINT REFERENCES work_items(id) ON DELETE SET NULL` | yes | Only set on `task` rows; must reference a row where `type = 'epic'` (enforced in application layer — see Validation Rules) |
| `title`             | `TEXT`                      | no       | Non-empty |
| `description`       | `TEXT`                      | yes      | Defaults to empty string |
| `status`            | `TEXT`                      | no       | `backlog` \| `picked_for_today` \| `in_progress` \| `blocked` \| `done`; `CHECK` constraint; default `backlog`. Renamed/expanded from the old four-value set (`todo` → `backlog`, `picked_for_today` added) in migration `000003` |
| `source`            | `TEXT`                      | no       | `personal` \| `jira` \| `agent`; `CHECK` constraint; default `personal` |
| `jira_key`          | `TEXT`                      | yes      | e.g. `TCAT-123`; set only when `source = 'jira'`; `UNIQUE` when not null |
| `jira_url`          | `TEXT`                      | yes      | Full link to the issue; set only when `source = 'jira'` |
| `estimate_hours`    | `NUMERIC(6,2)`              | yes      | Optional; FR-019. Added in migration `000004`. No validation beyond "positive if set" — a personal estimate, not billed time |
| `due_date`          | `DATE`                      | yes      | Optional; FR-019/FR-020. Added in migration `000004`. Date only, no time-of-day |
| `reference_number`  | `INTEGER NOT NULL`          | no       | Drawn from `epic_reference_seq` (if `type = 'epic'`) or `task_reference_seq` (if `type = 'task'`) at insert time. Added in migration `000005`. Combined with `type` at the API-response layer into the display key `EPIC-<n>` / `TASK-<n>` (FR-018) — never stored as a string |
| `created_at`        | `TIMESTAMPTZ`               | no       | Default `now()` |
| `updated_at`        | `TIMESTAMPTZ`               | no       | Default `now()`; updated on every write, including by the Jira sync goroutine — this is the field FR-007's last-write-wins comparison reads |
| `completed_at`      | `TIMESTAMPTZ`               | yes      | Set to `now()` exactly when `status` transitions to `done`; cleared (`NULL`) if `status` moves away from `done` (FR-016). Distinct from `updated_at` on purpose — `updated_at` is overwritten by unrelated edits (e.g. a description tweak after completion), which would otherwise corrupt the stats in `contracts/rest-api.md`. |

### Reference Keys (`EPIC-#` / `TASK-#`)

- Two Postgres sequences, `epic_reference_seq` and `task_reference_seq`, each starting at 1 with
  no gaps other than the normal sequence-skip-on-rollback behavior common to all Postgres
  sequences.
- At `Create` time, the store picks the sequence matching the row's `type` and reads
  `nextval(...)` into `reference_number` as part of the same insert.
- The two sequences are fully independent: `EPIC-3` and `TASK-3` can both exist simultaneously —
  this is expected, not a collision (spec SC-011).
- The API never accepts `reference_number` or a reference key string on write; it's
  system-assigned and immutable.

### Validation Rules (application layer, in `internal/store`/`internal/models`)

- `type = 'epic'` rows MUST have `parent_id IS NULL` (FR-002: an Epic has no parent).
- `type = 'task'` rows MAY have `parent_id` set, and if set it MUST reference a row with
  `type = 'epic'` (FR-002: a Task has at most one Epic parent). Enforced in the store layer at
  write time rather than a DB trigger, consistent with the simple, hand-written SQL style already
  used in `oncarinho/internal/store`.
- `status` MUST be one of the five fixed values for every row regardless of `type` or `source`
  (FR-003) — enforced by a `CHECK` constraint plus a Go enum type.
- `source` MUST be one of `personal` / `jira` / `agent` (FR-004) — `CHECK` constraint plus Go enum.
- `jira_key`/`jira_url` are only meaningful when `source = 'jira'`; the API never accepts them on
  a `personal` or `agent` write.
- `estimate_hours`, if provided, MUST be a positive number; the API rejects zero/negative values
  with `400`.
- Deleting an Epic sets `parent_id = NULL` on any Task rows that referenced it (`ON DELETE SET
  NULL`), matching the spec's edge case: Tasks become unparented, not deleted, when their Epic is
  deleted.

### State Transitions

`status` has no enforced transition graph — any of the five states can move to any other directly
(spec defines a fixed set, not a workflow graph, matching the "fixed workflow states" non-goal of
configurable schemes). The one side effect any transition can trigger is `completed_at`: entering
`done` sets it to `now()`, leaving `done` clears it to `NULL`. This applies uniformly whether the
transition comes from the REST API or the Jira sync goroutine, so stats stay correct regardless of
which side made the change.

`Backlog → Picked for today` additionally has a dedicated UX entry point (FR-014: an explicit
"Pick for today" action, or drag-onto-Board) on top of the general-purpose status edit — but at
the data layer it's just another status transition, not a distinct operation.

## Mine/Agent Scope Filtering

No new column. `source` already distinguishes `personal` / `jira` / `agent`; the `Mine`/`Agent`
split (FR-022) is a view-level grouping over that existing field:

- `scope = mine` → `source IN ('personal', 'jira')`
- `scope = agent` → `source = 'agent'`

Applied identically wherever the API accepts a `scope` param (`GET /api/work-items`, the stats
endpoints) — see `contracts/rest-api.md`.

## Workload & Burn-Rate Stats (computed, no new table)

FR-017's Stats tab is served by aggregating existing `work_items` rows on request — no snapshot
table, matching the same "single source of truth" reasoning as Jira Sync Bookkeeping below. All
stats respect the same `scope` filter as the Board.

- **Throughput** (a day's completed count): rows where `completed_at` falls on that calendar day.
- **Backlog trend** (a day's open count): rows where `created_at <= end of that day` AND
  (`completed_at IS NULL` OR `completed_at > end of that day`) — i.e. items that existed and
  weren't yet done as of that day. This reconstructs a historical backlog curve purely from
  `created_at`/`completed_at` without ever having stored a daily snapshot.
- **Open-item count** (a snapshot, not a trend): count of rows where `status != 'done'`, as of now.
- **Completed this week**: count of rows where `completed_at` falls within the trailing 7 days.
- **Completion rate**: `(count where status = 'done') / (total row count)`, all-time, as a
  percentage — see spec Assumptions for why this is distinct from "completed this week."
- **Total tracked**: total row count, all-time, no time window.
- **Status breakdown**: row count grouped by `status`, across all five values (including zero
  counts for statuses with no items, so the bar chart always renders five bars).

All of the above are computed by fetching the scoped rows once (`store.List` with the relevant
scope filter — cheap at this project's scale, see `plan.md`'s Scale/Scope) and reducing them in Go,
in pure functions with no DB dependency of their own — see `internal/stats` in the implementation.
The trend series (throughput/backlog) and the summary numbers (open/completed-this-week/rate/
total/breakdown) are two separate pure functions sharing the same input rows, so adding another
summary metric later means adding a field to the summary struct and its reducer, not touching the
trend logic (FR-017's "structured to grow" requirement).

## Jira Sync Bookkeeping

No separate table. The sync loop treats `work_items` as authoritative:

- **Pull**: for each Jira issue returned by `assignee = currentUser()`, upsert by `jira_key`
  (insert if no row has that key; else compare `fields.updated` from Jira against the local
  `updated_at` and overwrite the local row only if Jira is newer — FR-007). The Jira status maps
  to `backlog` / `in_progress` / `blocked` / `done` only; anything unrecognized (including Jira's
  own "To Do") falls back to `backlog`. Pull never produces `picked_for_today` — see `research.md`.
- **Push**: for each local row with `source = 'jira'` where `updated_at` is newer than the last
  known Jira `fields.updated` recorded at the start of that poll cycle, PATCH the Jira issue's
  status/description. A local `picked_for_today` status pushes as Jira's "In Progress" transition
  (see `research.md` for why).
- No polling-state table is needed because "newer than what Jira last reported" is derived by
  re-reading Jira's `fields.updated` each cycle — keeping the schema exactly the one table the
  spec's Key Entities describe (Principle II).

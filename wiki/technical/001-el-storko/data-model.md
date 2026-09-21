# Data Model: el-storko

Phase 1 output for `wiki/technical/001-el-storko/plan.md`.

## Entity: WorkItem

Single table `work_items` backs both Epic and Task (`type` discriminates), matching the spec's
Key Entities section (Work Item / Epic / Task).

| Column        | Type                        | Nullable | Notes |
|---------------|-----------------------------|----------|-------|
| `id`          | `BIGSERIAL PRIMARY KEY`     | no       | |
| `type`        | `TEXT`                      | no       | `epic` \| `task`; `CHECK` constraint |
| `parent_id`   | `BIGINT REFERENCES work_items(id) ON DELETE SET NULL` | yes | Only set on `task` rows; must reference a row where `type = 'epic'` (enforced in application layer — see Validation Rules) |
| `title`       | `TEXT`                      | no       | Non-empty |
| `description` | `TEXT`                      | yes      | Defaults to empty string |
| `status`      | `TEXT`                      | no       | `todo` \| `in_progress` \| `blocked` \| `done`; `CHECK` constraint; default `todo` |
| `source`      | `TEXT`                      | no       | `personal` \| `jira` \| `agent`; `CHECK` constraint; default `personal` |
| `jira_key`    | `TEXT`                      | yes      | e.g. `TCAT-123`; set only when `source = 'jira'`; `UNIQUE` when not null |
| `jira_url`    | `TEXT`                      | yes      | Full link to the issue; set only when `source = 'jira'` |
| `created_at`  | `TIMESTAMPTZ`               | no       | Default `now()` |
| `updated_at`  | `TIMESTAMPTZ`               | no       | Default `now()`; updated on every write, including by the Jira sync goroutine — this is the field FR-007's last-write-wins comparison reads |

### Validation Rules (application layer, in `internal/store`/`internal/models`)

- `type = 'epic'` rows MUST have `parent_id IS NULL` (FR-002: an Epic has no parent).
- `type = 'task'` rows MAY have `parent_id` set, and if set it MUST reference a row with
  `type = 'epic'` (FR-002: a Task has at most one Epic parent). Enforced in the store layer at
  write time rather than a DB trigger, consistent with the simple, hand-written SQL style already
  used in `oncarinho/internal/store`.
- `status` MUST be one of the four fixed values for every row regardless of `type` or `source`
  (FR-003) — enforced by a `CHECK` constraint plus a Go enum type.
- `source` MUST be one of `personal` / `jira` / `agent` (FR-004) — `CHECK` constraint plus Go enum.
- `jira_key`/`jira_url` are only meaningful when `source = 'jira'`; the API never accepts them on
  a `personal` or `agent` write.
- Deleting an Epic sets `parent_id = NULL` on any Task rows that referenced it (`ON DELETE SET
  NULL`), matching the spec's edge case: Tasks become unparented, not deleted, when their Epic is
  deleted.

### State Transitions

`status` has no enforced transition graph — any of the four states can move to any other
directly (spec defines a fixed set, not a workflow graph, matching the "fixed workflow states"
non-goal of configurable schemes).

## Jira Sync Bookkeeping

No separate table. The sync loop treats `work_items` as authoritative:

- **Pull**: for each Jira issue returned by `assignee = currentUser()`, upsert by `jira_key`
  (insert if no row has that key; else compare `fields.updated` from Jira against the local
  `updated_at` and overwrite the local row only if Jira is newer — FR-007).
- **Push**: for each local row with `source = 'jira'` where `updated_at` is newer than the last
  known Jira `fields.updated` recorded at the start of that poll cycle, PATCH the Jira issue's
  status/description.
- No polling-state table is needed because "newer than what Jira last reported" is derived by
  re-reading Jira's `fields.updated` each cycle — keeping the schema exactly the one table the
  spec's Key Entities describe (Principle II).

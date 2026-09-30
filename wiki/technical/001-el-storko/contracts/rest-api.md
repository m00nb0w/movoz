# Contract: el-storko REST API

Base URL (local): `http://localhost:8082/api`. Consumed identically by the Next.js zone (via its
`/api/:path*` rewrite) and the CLI (FR-010: no divergence between clients).

All endpoints return JSON. The Jira API token is never present in any response (FR-012).

## WorkItem (response shape)

```json
{
  "id": 12,
  "type": "task",
  "reference_key": "TASK-7",
  "parent_id": 3,
  "title": "Draft el-storko launchd plists",
  "description": "",
  "status": "backlog",
  "source": "personal",
  "estimate_hours": 2.5,
  "due_date": "2026-10-03",
  "jira_key": null,
  "jira_url": null,
  "created_at": "2026-09-21T10:00:00Z",
  "updated_at": "2026-09-21T10:00:00Z",
  "completed_at": null
}
```

- `reference_key` is computed from `type` + the item's internal `reference_number` (FR-018) —
  `EPIC-<n>` for an Epic, `TASK-<n>` for a Task. Read-only; never accepted on write.
- `status` is one of `backlog` \| `picked_for_today` \| `in_progress` \| `blocked` \| `done`.
- `estimate_hours` and `due_date` are both optional (`null` when unset).

## Endpoints

### `GET /api/work-items`

List work items. Query params (all optional, combinable):
- `scope=mine|agent` — `mine` = `source IN (personal, jira)`, `agent` = `source = agent` (FR-022).
  Omitting `scope` returns all sources, unfiltered by scope (used internally by Jira sync; the web
  UI always sends one explicitly).
- `source=personal|jira|agent` — finer-grained than `scope`; combinable with `scope` if ever
  needed, though the web UI only ever sends one or the other.
- `parent_id=<id>` (Epic id — list Tasks under that Epic)
- `type=epic|task`
- `status=backlog|picked_for_today|in_progress|blocked|done`

Returns `200 OK` with `{"items": [WorkItem, ...]}`.

### `POST /api/work-items`

Create a Task or Epic. Body:

```json
{
  "type": "task",
  "title": "...",
  "description": "",
  "parent_id": 3,
  "status": "backlog",
  "estimate_hours": 2.5,
  "due_date": "2026-10-03"
}
```

- `type` required. `title` required, non-empty.
- `parent_id` only accepted when `type = "task"`; `400` if set on an `epic`, or if it references a
  non-`epic` row.
- `estimate_hours`, if present, MUST be positive; `due_date`, if present, MUST be a valid date —
  `400` otherwise.
- `source` is not settable via this endpoint — always created as `personal` (Jira/agent rows are
  only ever written by their respective producers).
- `reference_number`/`reference_key` are not settable — assigned server-side from the type's
  sequence (see `data-model.md`).
- Returns `201 Created` with the created `WorkItem`.

### `GET /api/work-items/:id`

Returns `200 OK` with the `WorkItem`, or `404` if not found.

### `PATCH /api/work-items/:id`

Partial update. Body may include any of `title`, `description`, `status`, `parent_id`,
`estimate_hours`, `due_date`.

- If the item's `source = "jira"`, a successful `title`/`description`/`status` update also queues
  a push to Jira on the next sync tick (FR-006). A `status` of `picked_for_today` pushes as Jira's
  "In Progress" transition (see `research.md`).
- `parent_id` validation rules are the same as create. Sending `parent_id: null` explicitly clears
  the Epic assignment (FR-024's "Clear" action).
- Returns `200 OK` with the updated `WorkItem`, or `404`/`400` as appropriate.

### `DELETE /api/work-items/:id`

Deletes the item. If it is an `epic`, any Task rows with that `parent_id` are unparented
(`parent_id` set to `null`) rather than deleted (spec edge case). Returns `204 No Content`.

### `GET /api/stats/burn-rate`

Daily trend series for the Stats tab (FR-017). Query params: `days` (optional, default `30`), and
`scope=mine|agent` (optional, default `mine` — FR-022).

```json
{
  "days": 30,
  "points": [
    { "date": "2026-08-24", "completed": 2, "open": 14 },
    { "date": "2026-08-25", "completed": 0, "open": 14 },
    ...
  ]
}
```

- `points` has exactly `days` entries, one per calendar day, oldest first, ending today — days
  with zero completions are included with `completed: 0` (no gaps).
- `completed` is the throughput metric: count of items whose `completed_at` fell on that day.
- `open` is the backlog metric: count of items that existed and weren't `done` as of the end of
  that day (see `data-model.md`'s Workload & Burn-Rate Stats section for the exact computation).
- Returns `200 OK`. Never fails due to missing data — an empty tracker returns all-zero points.

### `GET /api/stats/summary`

Point-in-time workload numbers for the Stats tab (FR-017), separate from the trend series above so
future summary metrics extend this endpoint without touching the trend one. Query params:
`scope=mine|agent` (optional, default `mine`).

```json
{
  "open_items": 8,
  "completed_this_week": 3,
  "completion_rate": 0.62,
  "total_tracked": 21,
  "status_breakdown": {
    "backlog": 5,
    "picked_for_today": 2,
    "in_progress": 1,
    "blocked": 0,
    "done": 13
  }
}
```

- `open_items`: count where `status != done`, as of now.
- `completed_this_week`: count where `completed_at` falls in the trailing 7 days.
- `completion_rate`: `(done count) / (total count)`, all-time, as a fraction (`0.62`, not `62`) —
  the frontend formats it as a percentage.
- `total_tracked`: total row count in the current scope, all-time.
- `status_breakdown`: count per status, all five keys always present (zero-filled), for the bar
  chart.
- Returns `200 OK`. An empty tracker returns all zeros and `completion_rate: 0`.

## CLI mapping

| CLI command | Backing call |
|---|---|
| `el-storko add "<title>" [--epic <id>] [--description "..."]` | `POST /api/work-items` |
| `el-storko list [--source ...] [--status ...] [--epic <id>]` | `GET /api/work-items` |

## Jira sync (internal, not user-facing HTTP)

Not exposed as an API endpoint — runs as a background goroutine inside `cmd/server`. Documented
here because it reads/writes the same `work_items` rows the API above serves:

- Poll interval: every 5 minutes (env-configurable), calling Jira REST API v3 with the user's
  personal API token (HTTP Basic auth).
- Pull: upsert local rows for issues matching `assignee = currentUser()`. Jira status maps to
  `backlog` / `in_progress` / `blocked` / `done` only; unrecognized statuses (including Jira's own
  "To Do") fall back to `backlog`. Pull never sets `picked_for_today`.
- Push: for local `source = "jira"` rows updated more recently than Jira's last-known `updated`
  timestamp, PATCH the Jira issue. A local `picked_for_today` status pushes as "In Progress."
- Failures (network, 401/403) are logged (token redacted) and do not affect the REST API's
  availability for `personal`/`agent` rows (FR-013).

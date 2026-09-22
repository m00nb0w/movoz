# Contract: el-storko REST API

Base URL (local): `http://localhost:8082/api`. Consumed identically by the Next.js zone (via its
`/api/:path*` rewrite) and the CLI (FR-010: no divergence between clients).

All endpoints return JSON. The Jira API token is never present in any response (FR-012).

## WorkItem (response shape)

```json
{
  "id": 12,
  "type": "task",
  "parent_id": 3,
  "title": "Draft el-storko launchd plists",
  "description": "",
  "status": "todo",
  "source": "personal",
  "jira_key": null,
  "jira_url": null,
  "created_at": "2026-09-21T10:00:00Z",
  "updated_at": "2026-09-21T10:00:00Z",
  "completed_at": null
}
```

## Endpoints

### `GET /api/work-items`

List work items. Query params (all optional, combinable):
- `source=personal|jira|agent`
- `parent_id=<id>` (Epic id — list Tasks under that Epic)
- `type=epic|task`
- `status=todo|in_progress|blocked|done`

Returns `200 OK` with `{"items": [WorkItem, ...]}`.

### `POST /api/work-items`

Create a Task or Epic. Body:

```json
{ "type": "task", "title": "...", "description": "", "parent_id": 3, "status": "todo" }
```

- `type` required. `title` required, non-empty.
- `parent_id` only accepted when `type = "task"`; `400` if set on an `epic`, or if it references a
  non-`epic` row.
- `source` is not settable via this endpoint — always created as `personal` (Jira/agent rows are
  only ever written by their respective producers).
- Returns `201 Created` with the created `WorkItem`.

### `GET /api/work-items/:id`

Returns `200 OK` with the `WorkItem`, or `404` if not found.

### `PATCH /api/work-items/:id`

Partial update. Body may include any of `title`, `description`, `status`, `parent_id`.

- If the item's `source = "jira"`, a successful `title`/`description`/`status` update also queues
  a push to Jira on the next sync tick (FR-006).
- `parent_id` validation rules are the same as create.
- Returns `200 OK` with the updated `WorkItem`, or `404`/`400` as appropriate.

### `DELETE /api/work-items/:id`

Deletes the item. If it is an `epic`, any Task rows with that `parent_id` are unparented
(`parent_id` set to `null`) rather than deleted (spec edge case). Returns `204 No Content`.

### `GET /api/stats/burn-rate`

Burn-rate stats for the Stats tab (FR-016). Query params: `days` (optional, default `30`) — size
of the trailing window in days.

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
  that day (see `data-model.md`'s Burn-Rate Stats section for the exact computation).
- Returns `200 OK`. Never fails due to missing data — an empty tracker returns all-zero points.

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
- Pull: upsert local rows for issues matching `assignee = currentUser()`.
- Push: for local `source = "jira"` rows updated more recently than Jira's last-known `updated`
  timestamp, PATCH the Jira issue.
- Failures (network, 401/403) are logged (token redacted) and do not affect the REST API's
  availability for `personal`/`agent` rows (FR-013).

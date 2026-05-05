# Outlook → Google Calendar one-way sync

**Date:** 2026-04-30
**Owner:** lto@axon.com
**Status:** approved (design)

## Goal

Mirror Axon work calendar (Microsoft Exchange / Outlook) into a personal Google Calendar so the user gets a single unified view across work and personal commitments. Read-only; no editing of work events from the Google side.

## Non-goals

- Two-way sync. Outlook is the source of truth.
- Mirroring full event details (body, attendees, location, attachments). Title + time only.
- Editing or responding to invitations from Google.
- Replacing Outlook for actually attending or managing meetings.

## Constraints driving the design

- **Axon IT does not allow publishing the Outlook calendar as a public ICS link**, so the native "subscribe by URL" path is unavailable.
- **Axon Azure AD app registration likely requires admin approval**, so designs that depend on Microsoft Graph API with a personal app registration are avoided.
- Work-event content (titles, attendees, bodies) is sensitive. Only event titles and times cross the work/personal boundary.
- The user's primary computer is a Mac that already has the Axon Exchange account synced into macOS Calendar via System Settings → Internet Accounts.

## High-level architecture

```
launchd (every 20 min, also on wake from sleep)
    │
    ▼
sync.py
    │
    ├─► EventKit (PyObjC) ──► macOS Calendar Store
    │                            ↳ Outlook events for window [today-7d, today+30d]
    │
    ├─► diff (pure function over plain dicts)
    │
    └─► Google Calendar API ──► dedicated Google account's primary calendar
                                  ↳ each mirrored event tagged with
                                    extendedProperties.private.mirroredFrom = "outlook"
                                    extendedProperties.private.outlookEventId = "<id>|<start>"
```

The user's primary Gmail subscribes to the dedicated account's calendar via Google Calendar's "subscribe to other calendar" share link. The primary account never holds an OAuth token to anything; it only sees a read-only subscribed feed.

## Components and file layout

```
~/repos/personal/outlook-google-sync/
├── pyproject.toml
├── README.md
├── src/sync/
│   ├── __init__.py
│   ├── __main__.py        # CLI entry: `python -m sync {auth|once|install-launchd}`
│   ├── eventkit_reader.py # PyObjC bridge to macOS EventKit
│   ├── google_writer.py   # Google Calendar API client
│   ├── diff.py            # pure function over plain dicts
│   ├── filters.py         # OOO-skip predicate
│   └── config.py          # window, calendar id, paths
├── tests/
│   ├── test_diff.py
│   └── test_filters.py
└── launchd/
    └── com.lto.outlook-google-sync.plist  # template; installed copy lives at ~/Library/LaunchAgents/
```

Tooling: `uv` for venv and dependency management (`uv sync` produces `.venv/`, which the launchd plist references). Python 3.12+. Key dependencies: `pyobjc-framework-EventKit`, `google-api-python-client`, `google-auth-oauthlib`.

`diff.py` and `filters.py` are pure functions over plain dicts. EventKit and Google API calls are isolated in their own modules so the diff is unit-testable without mocking either side.

`config.py` constants: sync window (`PAST_DAYS=7`, `FUTURE_DAYS=30`), Google `calendar_id='primary'` (the dedicated account's primary calendar), config path `~/.config/outlook-google-sync/`, log path `~/Library/Logs/outlook-google-sync.log`.

## Data flow per sync run

1. **Read source events** from EventKit for the window `[today − 7 days, today + 30 days]`. For each event, extract:
   - `outlookEventId` — derived from `EKEvent.calendarItemExternalIdentifier` plus the occurrence start time for recurring events: `"{externalIdentifier}|{occurrenceStartIsoUtc}"`
   - `title`
   - `start`, `end` (UTC)
   - `availability` (Free / Busy / Tentative / Unavailable)
   - `organizerEmail`
2. **Apply OOO filter** (see Filtering below). Events that match are dropped.
3. **List target events** from the dedicated Google calendar over the same window using `events.list` with `privateExtendedProperty=mirroredFrom=outlook` (Google's `privateExtendedProperty` filter requires an exact value, not a wildcard, so we tag every mirrored event with `mirroredFrom=outlook` for filtering and carry the unique `outlookEventId` in a second private extended property for identity matching). Build a map keyed by `outlookEventId`.
4. **Compute diff** in pure code:
   - For each source `id` not in target → `CREATE`
   - For each source `id` in target where `(title, start, end)` differs → `UPDATE`
   - For each target `id` not in source → `DELETE`
5. **Apply changes** via Google Calendar API. Per-event failures log and continue; they do not abort the run.
6. **Log a summary line**: `synced: N created, M updated, K deleted, U unchanged, S skipped`.

## Filtering: skipping OOO events

An event is skipped if any of:

- It is an all-day event (`EKEvent.isAllDay() == true`)
- `availability == Unavailable` (Exchange's "Out of Office" maps to this in EventKit)
- `title` matches the case-insensitive regex: `\b(OOO|OOF|out\s*of\s*office|vacation|PTO|on\s*leave|holiday|maker\s*blocks)\b`

This applies regardless of organizer. The user's own OOO blocks are intentionally also skipped — they don't need to appear in the unified view.

## Recurring events

Each occurrence is mirrored as an independent Google Calendar event rather than as one event with a recurrence rule. The unique key is `"{externalIdentifier}|{occurrenceStartIsoUtc}"`, which is stable per instance.

Rationale:
- Title + time only loses nothing by flattening.
- Per-instance modifications and cancellations naturally fall out of the diff (an instance whose start moved appears as a new id; the old one is no longer in source and gets deleted).
- No recurrence-exception bookkeeping in our code.

Tradeoff: a 5×/week meeting over 30 days produces ~22 mirrored rows instead of 1. Acceptable for unified-view purposes.

## Authentication

- **Source side (EventKit):** no auth. macOS already syncs the Exchange account in the background. First run prompts for Calendar TCC permission via the macOS privacy framework; user grants once.
- **Destination side (Google):** OAuth 2.0 Desktop client for the dedicated Google account.
  - User creates the OAuth client in Google Cloud Console (Calendar API enabled), downloads `credentials.json` to `~/.config/outlook-google-sync/`.
  - First-time setup: `python -m sync auth` runs the local-server flow, exchanges for a refresh token, writes `~/.config/outlook-google-sync/token.json`.
  - The refresh token is long-lived as long as the script runs at least every few months. Loss of the token requires re-running `auth`.
  - Important OAuth consent-screen note: for an `External` user type project in `Testing` status, Google issues refresh tokens that expire after **7 days**. To get long-lived refresh tokens, move the consent screen to `In production` status. The Calendar scope is classified as Sensitive, but Google does not require formal app verification for a self-published app used only by its developer — the user sees an "unverified app" warning on first OAuth and clicks through "Advanced → Continue". This is acceptable for a single-user personal tool. The setup checklist below assumes `In production`.
- **Subscription on primary Gmail:** the user manually shares the dedicated account's calendar to their primary Gmail address with read-only permission, then accepts the share. No OAuth involved on the primary account.

## launchd job

`~/Library/LaunchAgents/com.lto.outlook-google-sync.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.lto.outlook-google-sync</string>

    <key>ProgramArguments</key>
    <array>
        <string>/Users/lto/repos/personal/outlook-google-sync/.venv/bin/python</string>
        <string>-m</string>
        <string>sync</string>
        <string>once</string>
    </array>

    <key>WorkingDirectory</key>
    <string>/Users/lto/repos/personal/outlook-google-sync</string>

    <key>StartInterval</key>
    <integer>1200</integer>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/Users/lto/Library/Logs/outlook-google-sync.log</string>

    <key>StandardErrorPath</key>
    <string>/Users/lto/Library/Logs/outlook-google-sync.log</string>
</dict>
</plist>
```

Key behaviors that come from this:

- **Sleep / wake handling:** `StartInterval` (not `StartCalendarInterval`) fires the job on wake from sleep if any tick was missed during sleep. Multiple missed ticks coalesce into a single fire.
- **After login or reboot:** `RunAtLoad: true` runs once at agent load, then 20-min cadence resumes.
- **No-network on wake:** Google API call fails, error handling exits cleanly, the next 20-min tick catches it.

## Error handling

| Failure | Behavior |
|---|---|
| EventKit access denied (TCC not granted) | Fail loud, log instructions: System Settings → Privacy & Security → Calendars → enable for the script's terminal/python binary |
| Google API 401 (refresh token revoked) | Fail loud, log instructions to re-run `python -m sync auth` |
| Google API 429 / 5xx | Exponential backoff with up to 3 retries inside the run, then exit cleanly. Next launchd tick retries. |
| Single event create/update/delete fails | Log `outlookEventId` and HTTP status, continue with remaining events |
| Outlook account hasn't synced to macOS recently | Out of scope. We read whatever EventKit returns. |

All output goes to `~/Library/Logs/outlook-google-sync.log`. Log rotation: at the start of each run, if the file exceeds 5 MB, rename to `outlook-google-sync.log.1` (overwriting any previous `.1`) and start fresh. No multi-generation rotation needed.

## Testing

- `tests/test_diff.py` — fixtures of synthetic source / target dicts, asserts correct partitioning into create / update / delete sets. Covers:
  - All-create (empty target)
  - All-delete (empty source)
  - Mixed
  - No-op (source == target)
  - Title changed but time unchanged
  - Time changed but title unchanged
  - Recurring event with one moved occurrence (id appears different, old gone)
- `tests/test_filters.py` — fixtures of fake events asserting the skip predicate for various availability + title combinations, including: own OOO, others' OOO, holiday with matching title, normal busy meeting, tentative event.
- Manual smoke test before installing launchd: `python -m sync once --dry-run` lists planned creates/updates/deletes without writing to Google. User verifies output, then runs without `--dry-run` once, then installs launchd.

## One-time setup checklist

1. `cd ~/repos/personal/outlook-google-sync && uv sync` (installs Python deps from `pyproject.toml`)
2. Create the dedicated Google account in a browser (manual)
3. In Google Cloud Console for that account: create a project, enable Calendar API, configure OAuth consent screen (External user type, set status to `In production` to avoid 7-day refresh-token expiry; only the `https://www.googleapis.com/auth/calendar` scope is needed and it is non-sensitive), create OAuth Desktop client, download `credentials.json` to `~/.config/outlook-google-sync/`
4. `python -m sync auth` — runs OAuth flow, writes `token.json`
5. `python -m sync once --dry-run` — verifies events read correctly, prints planned changes
6. `python -m sync once` — first real run; macOS prompts for Calendar TCC, click Allow
7. `python -m sync install-launchd` — copies plist to `~/Library/LaunchAgents/` and `launchctl load`s it
8. In the dedicated account's Google Calendar, share the primary calendar to the user's primary Gmail with read-only permission
9. In the primary Gmail's Google Calendar, accept the share — it now appears as a subscribed calendar in the unified view

## Out of scope (explicit)

- Two-way sync
- Mirroring event bodies, attendees, locations, attachments, or color
- Conflict resolution between Outlook and Google (there is no conflict — Outlook is authoritative, Google is a mirror)
- Cross-device sync state (the launchd job runs on one Mac; that's the only writer)
- Sync windows beyond 30 days into the future
- Past events older than 7 days

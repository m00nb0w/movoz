# outlook-google-sync

One-way sync from Axon work calendar (Outlook / Exchange, via macOS EventKit) to a dedicated Google Calendar account, run every 20 minutes by launchd.

Design: `docs/superpowers/specs/2026-04-30-outlook-google-calendar-sync-design.md`

## What gets mirrored

- Title + start + end only. No body, attendees, location.
- All-day events skipped.
- OOO events skipped (availability=Unavailable, or title matching `OOO|OOF|out of office|vacation|PTO|on leave|holiday`).
- Recurring events flattened: one mirrored event per occurrence.
- Window: past 7 days through future 30 days.

## One-time setup

### 1. Make sure your Outlook account is in macOS Calendar

System Settings → Internet Accounts → Exchange / Office 365. Confirm the work calendar shows up in macOS Calendar.app.

### 2. Install Python deps

```bash
cd ~/repos/personal/outlook-google-sync
uv sync --extra dev
```

### 3. Create a dedicated Google account

In a browser, sign up for a new Gmail/Google account that will hold the mirror calendar. This account never holds any other personal data.

### 4. Configure Google Cloud Console for the dedicated account

1. Go to <https://console.cloud.google.com>, signed in as the dedicated account.
2. Create a new project.
3. APIs & Services → Library → enable "Google Calendar API".
4. APIs & Services → OAuth consent screen:
   - User type: **External**
   - App name: anything (e.g., "outlook-google-sync")
   - Add the dedicated account as a test user (only relevant in Testing status)
   - Add scope: `https://www.googleapis.com/auth/calendar`
   - **Publish app** (Production status). Required to get long-lived refresh tokens — Testing-status tokens expire in 7 days. You'll see an "unverified app" warning on first sign-in; click "Advanced → Continue", which is acceptable for a self-published app used only by its developer.
5. APIs & Services → Credentials → Create Credentials → OAuth client ID → Application type: **Desktop app**.
6. Download the JSON, save to `.config-local/credentials.json` inside this repo. (`.config-local/` is gitignored.)

### 5. Run OAuth flow

```bash
mkdir -p .config-local
# put credentials.json there first
uv run python -m sync auth
```

A browser window opens, you sign in as the dedicated account, click through the unverified-app warning, grant Calendar access. The script saves `.config-local/token.json`.

### 6. Dry-run

```bash
uv run python -m sync once --dry-run
```

Verifies events are read correctly and prints planned creates/updates/deletes without touching Google. Output lands in `~/Library/Logs/outlook-google-sync.log` and on stdout.

### 7. First real run

```bash
uv run python -m sync once
```

macOS will prompt your Terminal app for Calendar access on first run — click **OK**. After this, the dedicated account's primary calendar in Google starts populating.

### 8. Install the launchd job

```bash
uv run python -m sync install-launchd
```

Job runs every 20 minutes thereafter, including immediately on wake from sleep if a tick was missed.

### 9. Subscribe from your primary Gmail

1. In Google Calendar as the dedicated account: Settings → My calendars → primary calendar → Share with specific people → add your primary Gmail address with "See all event details" permission.
2. Open your inbox on the primary Gmail; accept the share.
3. The work calendar now shows as a subscribed calendar in your unified Google Calendar view.

## Day-to-day

| Action | Command |
|---|---|
| Manual sync | `uv run python -m sync once` |
| Dry-run | `uv run python -m sync once --dry-run` |
| View logs | `tail -f ~/Library/Logs/outlook-google-sync.log` |
| Uninstall job | `launchctl unload ~/Library/LaunchAgents/com.lto.outlook-google-sync.plist` |
| Reinstall job | `uv run python -m sync install-launchd` |
| Re-auth | `uv run python -m sync auth` |

## Troubleshooting

**EventKit access denied / no events read**
Run `uv run python -m sync once` from Terminal. macOS pops a Calendar access dialog for Terminal — click OK. launchd inherits the grant from there.

**Google API 401 / invalid_grant**
Refresh token revoked or expired. Re-run `uv run python -m sync auth`.

**Mirror has events I no longer have in Outlook**
Outlook hasn't synced to macOS yet. Open Calendar.app, force a refresh, run `uv run python -m sync once`.

**Want to wipe and re-sync from scratch**
Delete all events on the dedicated account's primary calendar (manually or via the API), then `uv run python -m sync once` will re-create everything.

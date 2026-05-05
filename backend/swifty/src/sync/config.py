# src/sync/config.py
"""Compile-time-ish constants and resolved paths for the sync tool."""
from pathlib import Path

# Sync window
PAST_DAYS = 7
FUTURE_DAYS = 30

# Google Calendar
GOOGLE_CALENDAR_ID = "primary"  # the dedicated account's primary calendar
GOOGLE_SCOPES = ["https://www.googleapis.com/auth/calendar"]

# Extended-property keys on mirrored Google events
EP_MIRRORED_FROM_KEY = "mirroredFrom"
EP_MIRRORED_FROM_VALUE = "outlook"
EP_OUTLOOK_EVENT_ID_KEY = "outlookEventId"

# launchd
LAUNCHD_LABEL = "com.lto.outlook-google-sync"

# Filesystem paths
HOME = Path.home()
PROJECT_ROOT = Path(__file__).resolve().parents[2]  # src/sync/config.py -> repo root
CONFIG_DIR = PROJECT_ROOT / ".config-local"
CREDENTIALS_PATH = CONFIG_DIR / "credentials.json"  # OAuth client (user-supplied)
TOKEN_PATH = CONFIG_DIR / "token.json"              # OAuth refresh token (we write)
LOG_DIR = HOME / "Library" / "Logs"
LOG_PATH = LOG_DIR / "outlook-google-sync.log"
LOG_ROTATE_AT_BYTES = 5 * 1024 * 1024  # 5 MB

LAUNCHD_AGENTS_DIR = HOME / "Library" / "LaunchAgents"
LAUNCHD_PLIST_PATH = LAUNCHD_AGENTS_DIR / f"{LAUNCHD_LABEL}.plist"

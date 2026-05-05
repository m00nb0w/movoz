# src/sync/auth.py
"""OAuth Desktop flow + token persistence."""
from __future__ import annotations

import logging

from google.auth.exceptions import RefreshError
from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build

from sync.config import (
    CONFIG_DIR,
    CREDENTIALS_PATH,
    GOOGLE_SCOPES,
    TOKEN_PATH,
)

log = logging.getLogger(__name__)


def load_credentials() -> Credentials | None:
    if not TOKEN_PATH.exists():
        return None
    creds = Credentials.from_authorized_user_file(str(TOKEN_PATH), GOOGLE_SCOPES)
    if creds.valid:
        return creds
    if creds.expired and creds.refresh_token:
        try:
            creds.refresh(Request())
            _save(creds)
            return creds
        except RefreshError as err:
            log.error("token refresh failed: %s; re-run `python -m sync auth`", err)
            return None
    return None


def _save(creds: Credentials) -> None:
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    TOKEN_PATH.write_text(creds.to_json())
    TOKEN_PATH.chmod(0o600)


def run_auth_flow() -> int:
    """Interactive: opens browser, exchanges for refresh token, saves to TOKEN_PATH."""
    if not CREDENTIALS_PATH.exists():
        print(
            f"Missing {CREDENTIALS_PATH}.\n"
            "Download the OAuth Desktop client JSON from Google Cloud Console for the "
            "dedicated account and save it there. See README for full setup steps.",
        )
        return 1

    flow = InstalledAppFlow.from_client_secrets_file(
        str(CREDENTIALS_PATH), GOOGLE_SCOPES
    )
    creds = flow.run_local_server(port=0)
    _save(creds)
    print(f"Saved refresh token to {TOKEN_PATH}")
    return 0


def build_calendar_service():
    """Returns an authenticated googleapiclient discovery service.

    Raises RuntimeError if no valid credentials exist.
    """
    creds = load_credentials()
    if creds is None:
        raise RuntimeError(
            "No valid Google credentials. Run `python -m sync auth` first."
        )
    return build("calendar", "v3", credentials=creds, cache_discovery=False)

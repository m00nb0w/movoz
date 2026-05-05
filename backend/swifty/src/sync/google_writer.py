# src/sync/google_writer.py
"""Google Calendar API client wrapper.

Pure boundary: takes/returns SourceEvent and TargetEvent values; no other module
imports the googleapiclient library.
"""
from __future__ import annotations

import logging
from typing import Any

from sync.config import (
    EP_MIRRORED_FROM_KEY,
    EP_MIRRORED_FROM_VALUE,
    EP_OUTLOOK_EVENT_ID_KEY,
)
from sync.iso_utc import canonicalize
from sync.models import SourceEvent, TargetEvent

log = logging.getLogger(__name__)


def _source_to_request_body(src: SourceEvent) -> dict[str, Any]:
    return {
        "summary": src.title,
        "start": {"dateTime": src.start, "timeZone": "UTC"},
        "end": {"dateTime": src.end, "timeZone": "UTC"},
        "extendedProperties": {
            "private": {
                EP_MIRRORED_FROM_KEY: EP_MIRRORED_FROM_VALUE,
                EP_OUTLOOK_EVENT_ID_KEY: src.outlook_event_id,
            }
        },
        # Mark visibility as private so the title isn't surfaced more broadly
        # than necessary on the dedicated account.
        "visibility": "private",
        # No reminders for mirrored events; you'll get the real ones from Outlook.
        "reminders": {"useDefault": False, "overrides": []},
    }


def _google_event_to_target(google_event: dict[str, Any]) -> TargetEvent | None:
    private = (
        google_event.get("extendedProperties", {}).get("private", {})
    )
    if private.get(EP_MIRRORED_FROM_KEY) != EP_MIRRORED_FROM_VALUE:
        return None
    outlook_id = private.get(EP_OUTLOOK_EVENT_ID_KEY)
    if not outlook_id:
        return None
    raw_start = google_event.get("start", {}).get("dateTime", "")
    raw_end = google_event.get("end", {}).get("dateTime", "")
    if not raw_start or not raw_end:
        # All-day events use start.date / end.date; we don't currently mirror
        # those (they'd appear as 24h timed blocks otherwise — known limitation).
        # If we encounter one with our extended properties anyway, skip it so
        # we don't produce malformed TargetEvents the diff can't compare.
        return None
    return TargetEvent(
        outlook_event_id=outlook_id,
        google_event_id=google_event["id"],
        title=google_event.get("summary", ""),
        start=canonicalize(raw_start),
        end=canonicalize(raw_end),
    )


class GoogleCalendarWriter:
    def __init__(self, service: Any, calendar_id: str) -> None:
        self._service = service
        self._calendar_id = calendar_id

    def list_mirrored_events(self, *, time_min: str, time_max: str) -> list[TargetEvent]:
        events = self._service.events()
        page_token: str | None = None
        out: list[TargetEvent] = []
        while True:
            kwargs: dict[str, Any] = dict(
                calendarId=self._calendar_id,
                timeMin=time_min,
                timeMax=time_max,
                singleEvents=True,
                privateExtendedProperty=[
                    f"{EP_MIRRORED_FROM_KEY}={EP_MIRRORED_FROM_VALUE}"
                ],
                maxResults=2500,
            )
            if page_token:
                kwargs["pageToken"] = page_token
            resp = events.list(**kwargs).execute()
            for item in resp.get("items", []):
                target = _google_event_to_target(item)
                if target is not None:
                    out.append(target)
            page_token = resp.get("nextPageToken")
            if not page_token:
                break
        return out

    def create_event(self, desired: SourceEvent) -> str:
        body = _source_to_request_body(desired)
        resp = self._service.events().insert(
            calendarId=self._calendar_id, body=body
        ).execute()
        return resp["id"]

    def update_event(self, *, google_event_id: str, desired: SourceEvent) -> None:
        body = _source_to_request_body(desired)
        self._service.events().patch(
            calendarId=self._calendar_id, eventId=google_event_id, body=body
        ).execute()

    def delete_event(self, *, google_event_id: str) -> None:
        self._service.events().delete(
            calendarId=self._calendar_id, eventId=google_event_id
        ).execute()

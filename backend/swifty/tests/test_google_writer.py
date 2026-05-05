# tests/test_google_writer.py
from unittest.mock import MagicMock

import pytest

from sync.google_writer import (
    GoogleCalendarWriter,
    _google_event_to_target,
    _source_to_request_body,
)
from sync.models import SourceEvent


def test_source_to_request_body_includes_extended_properties():
    src = SourceEvent(
        outlook_event_id="abc|2026-05-01T15:00:00+00:00",
        title="Standup",
        start="2026-05-01T15:00:00+00:00",
        end="2026-05-01T15:15:00+00:00",
    )
    body = _source_to_request_body(src)
    assert body["summary"] == "Standup"
    assert body["start"] == {"dateTime": "2026-05-01T15:00:00+00:00", "timeZone": "UTC"}
    assert body["end"] == {"dateTime": "2026-05-01T15:15:00+00:00", "timeZone": "UTC"}
    assert body["extendedProperties"]["private"] == {
        "mirroredFrom": "outlook",
        "outlookEventId": "abc|2026-05-01T15:00:00+00:00",
    }


def test_google_event_to_target_extracts_fields():
    google_event = {
        "id": "g_abc123",
        "summary": "Standup",
        "start": {"dateTime": "2026-05-01T15:00:00Z"},
        "end":   {"dateTime": "2026-05-01T15:15:00Z"},
        "extendedProperties": {
            "private": {
                "mirroredFrom": "outlook",
                "outlookEventId": "abc|2026-05-01T15:00:00+00:00",
            }
        },
    }
    target = _google_event_to_target(google_event)
    assert target is not None
    assert target.google_event_id == "g_abc123"
    assert target.outlook_event_id == "abc|2026-05-01T15:00:00+00:00"
    assert target.title == "Standup"
    # canonicalize() converts the trailing "Z" to "+00:00"
    assert target.start == "2026-05-01T15:00:00+00:00"
    assert target.end == "2026-05-01T15:15:00+00:00"


def test_google_event_to_target_returns_none_when_missing_extended_properties():
    google_event = {
        "id": "g_abc123",
        "summary": "Native Google event",
        "start": {"dateTime": "2026-05-01T15:00:00Z"},
        "end":   {"dateTime": "2026-05-01T15:15:00Z"},
    }
    assert _google_event_to_target(google_event) is None


def test_google_event_to_target_returns_none_for_all_day_events():
    # all-day events use start.date instead of start.dateTime
    google_event = {
        "id": "g_allday",
        "summary": "Holiday",
        "start": {"date": "2026-05-01"},
        "end":   {"date": "2026-05-02"},
        "extendedProperties": {"private": {
            "mirroredFrom": "outlook", "outlookEventId": "anything"}},
    }
    assert _google_event_to_target(google_event) is None


def test_list_mirrored_events_paginates_and_filters_non_mirrored():
    fake_service = MagicMock()
    page1 = {
        "items": [
            {
                "id": "g1",
                "summary": "A",
                "start": {"dateTime": "2026-05-01T15:00:00Z"},
                "end":   {"dateTime": "2026-05-01T15:30:00Z"},
                "extendedProperties": {"private": {
                    "mirroredFrom": "outlook", "outlookEventId": "id1"}},
            }
        ],
        "nextPageToken": "tok2",
    }
    page2 = {
        "items": [
            {
                "id": "g2",
                "summary": "B",
                "start": {"dateTime": "2026-05-02T15:00:00Z"},
                "end":   {"dateTime": "2026-05-02T15:30:00Z"},
                "extendedProperties": {"private": {
                    "mirroredFrom": "outlook", "outlookEventId": "id2"}},
            }
        ],
    }
    list_call = MagicMock()
    list_call.execute.side_effect = [page1, page2]
    fake_service.events.return_value.list.return_value = list_call

    writer = GoogleCalendarWriter(service=fake_service, calendar_id="primary")
    targets = writer.list_mirrored_events(time_min="2026-05-01T00:00:00Z", time_max="2026-06-01T00:00:00Z")

    assert {t.outlook_event_id for t in targets} == {"id1", "id2"}
    # Verify pagination param was used on the second call
    calls = fake_service.events.return_value.list.call_args_list
    assert calls[1].kwargs.get("pageToken") == "tok2"


def test_create_event_calls_insert_with_body():
    fake_service = MagicMock()
    insert_call = MagicMock()
    insert_call.execute.return_value = {"id": "g_new"}
    fake_service.events.return_value.insert.return_value = insert_call

    writer = GoogleCalendarWriter(service=fake_service, calendar_id="primary")
    src = SourceEvent(
        outlook_event_id="x|2026-05-01T15:00:00+00:00",
        title="X",
        start="2026-05-01T15:00:00+00:00",
        end="2026-05-01T15:30:00+00:00",
    )
    google_id = writer.create_event(src)

    assert google_id == "g_new"
    fake_service.events.return_value.insert.assert_called_once()
    kwargs = fake_service.events.return_value.insert.call_args.kwargs
    assert kwargs["calendarId"] == "primary"
    assert kwargs["body"]["summary"] == "X"


def test_update_event_calls_patch():
    fake_service = MagicMock()
    patch_call = MagicMock()
    patch_call.execute.return_value = {"id": "g_existing"}
    fake_service.events.return_value.patch.return_value = patch_call

    writer = GoogleCalendarWriter(service=fake_service, calendar_id="primary")
    src = SourceEvent(
        outlook_event_id="y|2026-05-01T15:00:00+00:00",
        title="Y2",
        start="2026-05-01T15:00:00+00:00",
        end="2026-05-01T15:30:00+00:00",
    )
    writer.update_event(google_event_id="g_existing", desired=src)

    fake_service.events.return_value.patch.assert_called_once()
    kwargs = fake_service.events.return_value.patch.call_args.kwargs
    assert kwargs["calendarId"] == "primary"
    assert kwargs["eventId"] == "g_existing"
    assert kwargs["body"]["summary"] == "Y2"


def test_delete_event_calls_delete():
    fake_service = MagicMock()
    delete_call = MagicMock()
    fake_service.events.return_value.delete.return_value = delete_call

    writer = GoogleCalendarWriter(service=fake_service, calendar_id="primary")
    writer.delete_event(google_event_id="g_existing")

    fake_service.events.return_value.delete.assert_called_once_with(
        calendarId="primary", eventId="g_existing"
    )
    delete_call.execute.assert_called_once()

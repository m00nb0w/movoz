"""Frozen dataclasses passed between pure modules.

These are values, not entities. Equality and hashing are by field, which lets
diff.py use them in sets and dicts without ceremony.
"""
from dataclasses import dataclass


@dataclass(frozen=True)
class SourceEvent:
    """An event read from EventKit, after filtering, in canonical form for diffing."""
    outlook_event_id: str  # "<calendarItemExternalIdentifier>|<startIsoUtc>"
    title: str
    start: str  # ISO 8601 UTC, e.g. "2026-05-01T15:00:00+00:00"
    end: str    # ISO 8601 UTC


@dataclass(frozen=True)
class TargetEvent:
    """An existing mirrored event read from Google Calendar."""
    outlook_event_id: str
    google_event_id: str
    title: str
    start: str
    end: str


@dataclass(frozen=True)
class DiffResult:
    creates: tuple[SourceEvent, ...]
    updates: tuple[tuple[TargetEvent, SourceEvent], ...]  # (existing, desired)
    deletes: tuple[TargetEvent, ...]

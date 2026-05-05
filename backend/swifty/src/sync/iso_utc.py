"""Canonicalize ISO 8601 datetimes to a single string form.

Both the EventKit reader and Google Calendar reader feed timestamps into the
diff. Without this canonicalization, "2026-05-01T15:00:00Z" (Google) would not
equal "2026-05-01T15:00:00+00:00" (EventKit) even though they're the same
instant — and the diff would mark every unchanged event as an UPDATE.

Canonical form: UTC, second-precision, "+00:00" suffix.
"""
from datetime import datetime, timezone


def canonicalize(s: str) -> str:
    # Python's fromisoformat handles "+00:00" but not the "Z" shorthand pre-3.11.
    # 3.11+ does handle "Z" — replacing is harmless either way and keeps support
    # consistent across CPython patch versions.
    s_norm = s.replace("Z", "+00:00")
    dt = datetime.fromisoformat(s_norm)
    if dt.tzinfo is None:
        raise ValueError(f"missing timezone info: {s!r}")
    dt = dt.astimezone(timezone.utc).replace(microsecond=0)
    return dt.isoformat()


def from_unix_timestamp(ts: float) -> str:
    """Build a canonical UTC string from a POSIX timestamp (seconds since epoch)."""
    return datetime.fromtimestamp(ts, tz=timezone.utc).replace(microsecond=0).isoformat()

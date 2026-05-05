# src/sync/eventkit_reader.py
"""Read calendar events from macOS EventKit.

Uses PyObjC to bridge to Apple's EventKit framework. Requires Calendar TCC
permission, granted to the binary that first triggers the prompt — for
launchd-driven runs, this means you must run `python -m sync once` once from
Terminal first so Terminal's TCC entry is set; launchd then inherits access.
"""
from __future__ import annotations

import logging
import threading
from datetime import datetime, timedelta, timezone

from sync.config import FUTURE_DAYS, PAST_DAYS
from sync.filters import is_ooo
from sync.iso_utc import from_unix_timestamp
from sync.models import SourceEvent

log = logging.getLogger(__name__)

# These imports are intentionally inside functions so this module is importable
# on non-macOS systems for `pytest --collect-only` etc. PyObjC is only required
# at runtime when `read_source_events` is actually called.


_AVAILABILITY_NAMES = {
    0: "Busy",
    1: "Free",
    2: "Tentative",
    3: "Unavailable",
}


def _request_access_synchronously(store) -> bool:
    sema = threading.Semaphore(0)
    granted_box: list[bool] = [False]

    def callback(granted, err):
        granted_box[0] = bool(granted)
        sema.release()

    if hasattr(store, "requestFullAccessToEventsWithCompletion_"):
        store.requestFullAccessToEventsWithCompletion_(callback)  # macOS 14+
    else:
        from EventKit import EKEntityTypeEvent
        store.requestAccessToEntityType_completion_(EKEntityTypeEvent, callback)

    if not sema.acquire(timeout=30):
        log.error("timed out waiting for Calendar access decision")
        return False
    return granted_box[0]


def _ns_date_to_iso_utc(ns_date) -> str:
    # NSDate.timeIntervalSince1970() returns seconds since epoch as Cocoa double
    ts = float(ns_date.timeIntervalSince1970())
    return from_unix_timestamp(ts)


def read_source_events() -> list[SourceEvent]:
    """Return SourceEvents in the configured window, with OOO filter applied."""
    from EventKit import EKEventStore  # type: ignore[import-not-found]
    from Foundation import NSDate  # type: ignore[import-not-found]

    store = EKEventStore.alloc().init()
    if not _request_access_synchronously(store):
        raise PermissionError(
            "Calendar access denied. Grant the parent terminal app Calendar access "
            "in System Settings → Privacy & Security → Calendars, then re-run."
        )

    now = datetime.now(timezone.utc)
    start_dt = now - timedelta(days=PAST_DAYS)
    end_dt = now + timedelta(days=FUTURE_DAYS)
    start_ns = NSDate.dateWithTimeIntervalSince1970_(start_dt.timestamp())
    end_ns = NSDate.dateWithTimeIntervalSince1970_(end_dt.timestamp())

    predicate = store.predicateForEventsWithStartDate_endDate_calendars_(
        start_ns, end_ns, None
    )
    raw_events = store.eventsMatchingPredicate_(predicate) or []

    out: list[SourceEvent] = []
    skipped_ooo = 0
    skipped_all_day = 0
    for ev in raw_events:
        ext_id = ev.calendarItemExternalIdentifier()
        if not ext_id:
            continue
        if bool(ev.isAllDay()):
            skipped_all_day += 1
            continue
        title = str(ev.title() or "")
        availability_int = int(ev.availability())
        availability = _AVAILABILITY_NAMES.get(availability_int, "Busy")

        if is_ooo(title, availability):
            skipped_ooo += 1
            continue

        start_iso = _ns_date_to_iso_utc(ev.startDate())
        end_iso = _ns_date_to_iso_utc(ev.endDate())
        outlook_event_id = f"{str(ext_id)}|{start_iso}"
        out.append(
            SourceEvent(
                outlook_event_id=outlook_event_id,
                title=title,
                start=start_iso,
                end=end_iso,
            )
        )

    log.info(
        "read %d source events (skipped %d OOO, %d all-day)",
        len(out), skipped_ooo, skipped_all_day,
    )
    return out

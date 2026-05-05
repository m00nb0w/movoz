# src/sync/orchestrator.py
"""Run a single sync cycle: read source → filter → list target → diff → apply."""
from __future__ import annotations

import logging
from datetime import datetime, timedelta, timezone

from googleapiclient.errors import HttpError

from sync.auth import build_calendar_service
from sync.config import FUTURE_DAYS, GOOGLE_CALENDAR_ID, PAST_DAYS
from sync.diff import compute_diff
from sync.eventkit_reader import read_source_events
from sync.google_writer import GoogleCalendarWriter
from sync.logging_setup import configure_logging
from sync.retry import retry_with_backoff

log = logging.getLogger(__name__)


def _apply(
    writer: GoogleCalendarWriter,
    diff,
    *,
    dry_run: bool,
) -> tuple[int, int, int, int]:
    """Apply diff. Returns (created, updated, deleted, errors)."""
    created = updated = deleted = errors = 0

    for src in diff.creates:
        if dry_run:
            log.info("[dry-run] CREATE %s '%s' %s", src.outlook_event_id, src.title, src.start)
            created += 1
            continue
        try:
            new_id = retry_with_backoff(
                lambda src=src: writer.create_event(src),
                retries=3,
                base_delay=1.0,
                retry_on=(HttpError,),
            )
            log.info("CREATE %s -> %s", src.outlook_event_id, new_id)
            created += 1
        except HttpError as err:
            log.error("CREATE failed for %s: %s", src.outlook_event_id, err)
            errors += 1

    for existing, desired in diff.updates:
        if dry_run:
            log.info(
                "[dry-run] UPDATE %s '%s' -> '%s'",
                existing.google_event_id, existing.title, desired.title,
            )
            updated += 1
            continue
        try:
            retry_with_backoff(
                lambda existing=existing, desired=desired: writer.update_event(
                    google_event_id=existing.google_event_id, desired=desired
                ),
                retries=3,
                base_delay=1.0,
                retry_on=(HttpError,),
            )
            log.info("UPDATE %s", existing.google_event_id)
            updated += 1
        except HttpError as err:
            log.error("UPDATE failed for %s: %s", existing.google_event_id, err)
            errors += 1

    for existing in diff.deletes:
        if dry_run:
            log.info("[dry-run] DELETE %s '%s'", existing.google_event_id, existing.title)
            deleted += 1
            continue
        try:
            retry_with_backoff(
                lambda existing=existing: writer.delete_event(google_event_id=existing.google_event_id),
                retries=3,
                base_delay=1.0,
                retry_on=(HttpError,),
            )
            log.info("DELETE %s", existing.google_event_id)
            deleted += 1
        except HttpError as err:
            log.error("DELETE failed for %s: %s", existing.google_event_id, err)
            errors += 1

    return created, updated, deleted, errors


def run_once(*, dry_run: bool = False) -> int:
    configure_logging()
    log.info("sync run starting (dry_run=%s)", dry_run)

    # Validate Google credentials first so a missing/expired token surfaces
    # before we trigger the macOS Calendar TCC prompt on the user.
    try:
        service = build_calendar_service()
    except RuntimeError as err:
        log.error("%s", err)
        return 2

    try:
        sources = read_source_events()
    except PermissionError as err:
        log.error("source read failed: %s", err)
        return 2
    except Exception as err:
        log.exception("source read failed: %s", err)
        return 2

    writer = GoogleCalendarWriter(service=service, calendar_id=GOOGLE_CALENDAR_ID)

    now = datetime.now(timezone.utc)
    time_min = (now - timedelta(days=PAST_DAYS)).isoformat()
    time_max = (now + timedelta(days=FUTURE_DAYS)).isoformat()

    try:
        targets = retry_with_backoff(
            lambda: writer.list_mirrored_events(time_min=time_min, time_max=time_max),
            retries=3,
            base_delay=1.0,
            retry_on=(HttpError,),
        )
    except HttpError as err:
        log.error("target list failed: %s", err)
        return 2

    diff = compute_diff(sources, targets)
    log.info(
        "diff computed: %d create, %d update, %d delete, %d unchanged",
        len(diff.creates), len(diff.updates), len(diff.deletes),
        len(sources) - len(diff.creates) - len(diff.updates),
    )

    created, updated, deleted, errors = _apply(writer, diff, dry_run=dry_run)
    unchanged = len(sources) - len(diff.creates) - len(diff.updates)

    log.info(
        "synced: %d created, %d updated, %d deleted, %d unchanged, %d errors%s",
        created, updated, deleted, unchanged, errors,
        " (dry-run)" if dry_run else "",
    )
    return 0 if errors == 0 else 1

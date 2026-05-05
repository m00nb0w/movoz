# Outlook → Google Calendar one-way sync — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A small Python tool that reads work calendar events from macOS EventKit (where the Axon Exchange account is already synced) and mirrors them title+time-only into a dedicated Google account's primary calendar, run every 20 minutes by launchd including on wake from sleep.

**Architecture:** Three pure modules (`diff`, `filters`, `config`) over plain dataclasses are unit-tested in isolation. Two I/O modules (`eventkit_reader`, `google_writer`) wrap their respective frameworks. A `__main__` orchestrator wires them together with subcommands `auth`, `once`, `install-launchd`. A launchd plist template plus an installer subcommand handles the schedule.

**Tech Stack:** Python 3.12, `uv` for venv/deps, `pyobjc-framework-EventKit` for macOS Calendar access, `google-api-python-client` + `google-auth-oauthlib` for Google Calendar, `pytest` for tests, launchd for scheduling.

**Reference spec:** [`../specs/swifty.md`](../specs/swifty.md)

---

## File map

| File | Responsibility |
|---|---|
| `pyproject.toml` | Project metadata, dependencies, pytest config |
| `.gitignore` | Exclude `.venv`, `__pycache__`, `*.pyc`, build artifacts |
| `README.md` | One-time setup checklist + day-to-day commands |
| `src/sync/__init__.py` | Package marker (empty) |
| `src/sync/__main__.py` | CLI: `python -m sync {auth\|once\|install-launchd}` |
| `src/sync/config.py` | Constants: window, paths, calendar id, scopes |
| `src/sync/models.py` | Frozen dataclasses: `SourceEvent`, `TargetEvent`, `DiffResult` |
| `src/sync/diff.py` | Pure: `compute_diff(sources, targets) -> DiffResult` |
| `src/sync/filters.py` | Pure: `is_ooo(title, availability) -> bool` |
| `src/sync/eventkit_reader.py` | Reads from macOS EventKit, returns `list[SourceEvent]` |
| `src/sync/google_writer.py` | Google Calendar API: list/create/update/delete |
| `src/sync/iso_utc.py` | Canonicalize ISO 8601 datetimes to a single string form so diffs are stable |
| `src/sync/retry.py` | Exponential backoff helper for Google API calls |
| `src/sync/logging_setup.py` | Configure stdout logging + size-based log rotation |
| `src/sync/orchestrator.py` | The `once` command: read → filter → list → diff → apply → summarize |
| `src/sync/auth.py` | OAuth Desktop flow + token persistence |
| `src/sync/installer.py` | The `install-launchd` command: render and load plist |
| `tests/test_diff.py` | Unit tests for diff |
| `tests/test_filters.py` | Unit tests for filter |
| `tests/test_google_writer.py` | Unit tests with mocked HTTP |
| `tests/test_iso_utc.py` | Unit tests for canonicalization |
| `tests/test_retry.py` | Unit tests for backoff |
| `launchd/com.lto.outlook-google-sync.plist.template` | launchd job template (with `__PYTHON__` and `__WORKDIR__` placeholders the installer fills in) |

---

### Task 1: Scaffold project

**Files:**
- Create: `pyproject.toml`
- Create: `.gitignore`
- Create: `src/sync/__init__.py`
- Create: `src/sync/__main__.py`
- Create: `tests/__init__.py`

- [ ] **Step 1: Write `pyproject.toml`**

```toml
[project]
name = "outlook-google-sync"
version = "0.1.0"
description = "One-way sync from macOS Calendar (Outlook) to Google Calendar"
requires-python = ">=3.12"
dependencies = [
    "pyobjc-framework-EventKit>=10.0",
    "google-api-python-client>=2.140",
    "google-auth-oauthlib>=1.2",
    "google-auth-httplib2>=0.2",
]

[project.optional-dependencies]
dev = [
    "pytest>=8.0",
    "pytest-mock>=3.12",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.hatch.build.targets.wheel]
packages = ["src/sync"]

[tool.pytest.ini_options]
pythonpath = ["src"]
testpaths = ["tests"]
```

- [ ] **Step 2: Write `.gitignore`**

```
.venv/
__pycache__/
*.pyc
*.egg-info/
dist/
build/
.pytest_cache/
.config-local/
```

- [ ] **Step 3: Create empty package files**

```python
# src/sync/__init__.py
```

```python
# src/sync/__main__.py
"""CLI entry point. Subcommands wired in later tasks."""
import sys


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: python -m sync {auth|once|install-launchd}", file=sys.stderr)
        return 2
    cmd = sys.argv[1]
    if cmd == "auth":
        from sync.auth import run_auth_flow
        return run_auth_flow()
    if cmd == "once":
        from sync.orchestrator import run_once
        dry_run = "--dry-run" in sys.argv[2:]
        return run_once(dry_run=dry_run)
    if cmd == "install-launchd":
        from sync.installer import install_launchd
        return install_launchd()
    print(f"unknown command: {cmd}", file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main())
```

```python
# tests/__init__.py
```

- [ ] **Step 4: Verify `uv sync` works**

Run: `cd ~/repos/personal/outlook-google-sync && uv sync --extra dev`
Expected: creates `.venv/`, installs deps without error.

- [ ] **Step 5: Verify CLI scaffold loads**

Run: `uv run python -m sync`
Expected: stderr `usage: python -m sync {auth|once|install-launchd}`, exit code 2.

- [ ] **Step 6: Commit**

```bash
git add pyproject.toml .gitignore src/sync/__init__.py src/sync/__main__.py tests/__init__.py
git commit -m "chore: scaffold project with uv, pyproject, CLI skeleton"
```

---

### Task 2: Define data models

**Files:**
- Create: `src/sync/models.py`

- [ ] **Step 1: Write `models.py`**

```python
# src/sync/models.py
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
```

- [ ] **Step 2: Smoke check imports**

Run: `uv run python -c "from sync.models import SourceEvent, TargetEvent, DiffResult; print('ok')"`
Expected: `ok`

- [ ] **Step 3: Commit**

```bash
git add src/sync/models.py
git commit -m "feat: add SourceEvent, TargetEvent, DiffResult dataclasses"
```

---

### Task 3: Implement diff with TDD

**Files:**
- Create: `tests/test_diff.py`
- Create: `src/sync/diff.py`

- [ ] **Step 1: Write the failing tests**

```python
# tests/test_diff.py
from sync.diff import compute_diff
from sync.models import SourceEvent, TargetEvent


def src(id_, title="Meeting", start="2026-05-01T15:00:00+00:00", end="2026-05-01T15:30:00+00:00"):
    return SourceEvent(outlook_event_id=id_, title=title, start=start, end=end)


def tgt(id_, gid="g1", title="Meeting", start="2026-05-01T15:00:00+00:00", end="2026-05-01T15:30:00+00:00"):
    return TargetEvent(outlook_event_id=id_, google_event_id=gid, title=title, start=start, end=end)


def test_all_create_when_target_empty():
    sources = [src("a"), src("b")]
    result = compute_diff(sources, [])
    assert set(result.creates) == set(sources)
    assert result.updates == ()
    assert result.deletes == ()


def test_all_delete_when_source_empty():
    targets = [tgt("a"), tgt("b")]
    result = compute_diff([], targets)
    assert result.creates == ()
    assert result.updates == ()
    assert set(result.deletes) == set(targets)


def test_noop_when_source_equals_target():
    sources = [src("a"), src("b")]
    targets = [tgt("a"), tgt("b")]
    result = compute_diff(sources, targets)
    assert result.creates == ()
    assert result.updates == ()
    assert result.deletes == ()


def test_update_when_title_differs():
    sources = [src("a", title="Renamed")]
    targets = [tgt("a", title="Original")]
    result = compute_diff(sources, targets)
    assert result.creates == ()
    assert len(result.updates) == 1
    existing, desired = result.updates[0]
    assert existing.title == "Original"
    assert desired.title == "Renamed"
    assert result.deletes == ()


def test_update_when_start_differs():
    sources = [src("a", start="2026-05-01T16:00:00+00:00")]
    targets = [tgt("a", start="2026-05-01T15:00:00+00:00")]
    result = compute_diff(sources, targets)
    assert len(result.updates) == 1


def test_update_when_end_differs():
    sources = [src("a", end="2026-05-01T16:00:00+00:00")]
    targets = [tgt("a", end="2026-05-01T15:30:00+00:00")]
    result = compute_diff(sources, targets)
    assert len(result.updates) == 1


def test_mixed_creates_updates_deletes():
    sources = [
        src("a"),                       # unchanged
        src("b", title="New name"),     # update
        src("c"),                       # create
    ]
    targets = [
        tgt("a"),
        tgt("b", title="Old name"),
        tgt("d"),                       # delete
    ]
    result = compute_diff(sources, targets)
    assert {c.outlook_event_id for c in result.creates} == {"c"}
    assert {u[1].outlook_event_id for u in result.updates} == {"b"}
    assert {d.outlook_event_id for d in result.deletes} == {"d"}


def test_recurring_event_moved_occurrence_is_create_plus_delete():
    # The original occurrence at 15:00 was moved to 16:00. The id includes the
    # start time, so the moved occurrence has a different id than the original.
    sources = [src("series123|2026-05-01T16:00:00+00:00", start="2026-05-01T16:00:00+00:00")]
    targets = [tgt("series123|2026-05-01T15:00:00+00:00", start="2026-05-01T15:00:00+00:00")]
    result = compute_diff(sources, targets)
    assert len(result.creates) == 1
    assert len(result.deletes) == 1
    assert result.updates == ()
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/test_diff.py -v`
Expected: All tests fail with `ModuleNotFoundError: No module named 'sync.diff'` or similar import error.

- [ ] **Step 3: Implement `diff.py`**

```python
# src/sync/diff.py
"""Pure diff between source events (desired) and target events (existing in Google).

No I/O. No mutation. Easy to test exhaustively.
"""
from sync.models import DiffResult, SourceEvent, TargetEvent


def compute_diff(
    sources: list[SourceEvent],
    targets: list[TargetEvent],
) -> DiffResult:
    src_by_id = {s.outlook_event_id: s for s in sources}
    tgt_by_id = {t.outlook_event_id: t for t in targets}

    creates = tuple(s for sid, s in src_by_id.items() if sid not in tgt_by_id)
    deletes = tuple(t for tid, t in tgt_by_id.items() if tid not in src_by_id)
    updates = tuple(
        (tgt_by_id[k], src_by_id[k])
        for k in src_by_id
        if k in tgt_by_id
        and (
            tgt_by_id[k].title != src_by_id[k].title
            or tgt_by_id[k].start != src_by_id[k].start
            or tgt_by_id[k].end != src_by_id[k].end
        )
    )
    return DiffResult(creates=creates, updates=updates, deletes=deletes)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/test_diff.py -v`
Expected: 8 passed.

- [ ] **Step 5: Commit**

```bash
git add tests/test_diff.py src/sync/diff.py
git commit -m "feat: add pure diff over source/target events"
```

---

### Task 4: Implement OOO filter with TDD

**Files:**
- Create: `tests/test_filters.py`
- Create: `src/sync/filters.py`

- [ ] **Step 1: Write the failing tests**

```python
# tests/test_filters.py
import pytest
from sync.filters import is_ooo


@pytest.mark.parametrize(
    "title,availability,expected",
    [
        # Availability-driven
        ("Random meeting", "Unavailable", True),
        ("Random meeting", "Busy", False),
        ("Random meeting", "Free", False),
        ("Random meeting", "Tentative", False),
        # Title-driven, case-insensitive
        ("OOO - Sarah", "Busy", True),
        ("OOF: Bob", "Busy", True),
        ("Out of Office (John)", "Busy", True),
        ("Out  of  Office", "Busy", True),     # extra spaces
        ("Vacation - Maria", "Busy", True),
        ("PTO", "Busy", True),
        ("On Leave", "Busy", True),
        ("Company Holiday", "Busy", True),
        ("Maker Blocks", "Busy", True),
        ("maker blocks", "Busy", True),
        # Should NOT match — substrings inside other words
        ("Booking review", "Busy", False),     # contains "boo" not OOO
        ("Approved decisions", "Busy", False),
        ("Important meeting", "Busy", False),
        ("Promotion talk", "Busy", False),     # contains "PTO" as substring? word boundary should reject
        # Empty title with non-OOO availability
        ("", "Busy", False),
        # Empty title with Unavailable
        ("", "Unavailable", True),
    ],
)
def test_is_ooo(title, availability, expected):
    assert is_ooo(title, availability) is expected


def test_promotion_does_not_match_pto():
    # word-boundary regex: PTO as a substring of "Promotion" should NOT match
    assert is_ooo("Promotion talk", "Busy") is False


def test_phototyping_does_not_match():
    assert is_ooo("Phototyping session", "Busy") is False
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/test_filters.py -v`
Expected: ImportError; `sync.filters` doesn't exist yet.

- [ ] **Step 3: Implement `filters.py`**

```python
# src/sync/filters.py
"""Predicate for OOO events that should not be mirrored.

An event is OOO if either:
  - availability is "Unavailable" (Exchange's OOF maps here in EventKit), or
  - title matches a case-insensitive word-bounded regex of OOO-ish phrases.

Skipping applies regardless of organizer. The user's own OOO blocks are
intentionally also skipped per the design — they don't need to appear in the
unified view.
"""
import re

OOO_TITLE_RE = re.compile(
    r"\b(OOO|OOF|out\s*of\s*office|vacation|PTO|on\s*leave|holiday|maker\s*blocks)\b",
    re.IGNORECASE,
)


def is_ooo(title: str, availability: str) -> bool:
    if availability == "Unavailable":
        return True
    if title and OOO_TITLE_RE.search(title):
        return True
    return False
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/test_filters.py -v`
Expected: all parametrized cases pass.

- [ ] **Step 5: Commit**

```bash
git add tests/test_filters.py src/sync/filters.py
git commit -m "feat: add OOO filter predicate"
```

---

### Task 5: Implement config module

**Files:**
- Create: `src/sync/config.py`

- [ ] **Step 1: Write `config.py`**

```python
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
CONFIG_DIR = HOME / ".config" / "outlook-google-sync"
CREDENTIALS_PATH = CONFIG_DIR / "credentials.json"  # OAuth client (user-supplied)
TOKEN_PATH = CONFIG_DIR / "token.json"              # OAuth refresh token (we write)
LOG_DIR = HOME / "Library" / "Logs"
LOG_PATH = LOG_DIR / "outlook-google-sync.log"
LOG_ROTATE_AT_BYTES = 5 * 1024 * 1024  # 5 MB

LAUNCHD_AGENTS_DIR = HOME / "Library" / "LaunchAgents"
LAUNCHD_PLIST_PATH = LAUNCHD_AGENTS_DIR / f"{LAUNCHD_LABEL}.plist"
```

- [ ] **Step 2: Smoke check imports**

Run: `uv run python -c "from sync.config import LAUNCHD_PLIST_PATH, GOOGLE_SCOPES; print(LAUNCHD_PLIST_PATH, GOOGLE_SCOPES)"`
Expected: prints the resolved path and scopes list.

- [ ] **Step 3: Commit**

```bash
git add src/sync/config.py
git commit -m "feat: add config constants and resolved paths"
```

---

### Task 6: Implement ISO UTC canonicalization with TDD

**Files:**
- Create: `tests/test_iso_utc.py`
- Create: `src/sync/iso_utc.py`

**Why this exists:** EventKit emits ISO 8601 strings with `+00:00` suffix and may include microseconds; Google Calendar echoes timestamps back with `Z` suffix and second-level precision. Without a canonical form, the string comparison in `compute_diff` would treat every unchanged event as "title same but start differs" and trigger an UPDATE every cycle. Both reader modules call this helper so both sides produce identical strings for the same instant.

- [ ] **Step 1: Write the failing tests**

```python
# tests/test_iso_utc.py
import pytest

from sync.iso_utc import canonicalize


@pytest.mark.parametrize(
    "raw,expected",
    [
        # Z suffix → +00:00
        ("2026-05-01T15:00:00Z", "2026-05-01T15:00:00+00:00"),
        # Already in +00:00 form
        ("2026-05-01T15:00:00+00:00", "2026-05-01T15:00:00+00:00"),
        # Microseconds dropped
        ("2026-05-01T15:00:00.123456Z", "2026-05-01T15:00:00+00:00"),
        ("2026-05-01T15:00:00.000001+00:00", "2026-05-01T15:00:00+00:00"),
        # Non-UTC offset converted to UTC
        ("2026-05-01T08:00:00-07:00", "2026-05-01T15:00:00+00:00"),
        ("2026-05-01T17:00:00+02:00", "2026-05-01T15:00:00+00:00"),
    ],
)
def test_canonicalize(raw, expected):
    assert canonicalize(raw) == expected


def test_canonicalize_raises_on_garbage():
    with pytest.raises(ValueError):
        canonicalize("not a date")
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/test_iso_utc.py -v`
Expected: ImportError; module not yet defined.

- [ ] **Step 3: Implement `iso_utc.py`**

```python
# src/sync/iso_utc.py
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/test_iso_utc.py -v`
Expected: 7 passed.

- [ ] **Step 5: Commit**

```bash
git add tests/test_iso_utc.py src/sync/iso_utc.py
git commit -m "feat: add ISO 8601 canonicalization for stable diffs"
```

---

### Task 7: Implement retry/backoff helper with TDD

**Files:**
- Create: `tests/test_retry.py`
- Create: `src/sync/retry.py`

- [ ] **Step 1: Write the failing tests**

```python
# tests/test_retry.py
import pytest
from sync.retry import retry_with_backoff


class FakeError(Exception):
    pass


def test_returns_value_on_first_success():
    calls = {"n": 0}

    def fn():
        calls["n"] += 1
        return 42

    assert retry_with_backoff(fn, retries=3, base_delay=0.0) == 42
    assert calls["n"] == 1


def test_retries_until_success():
    calls = {"n": 0}

    def fn():
        calls["n"] += 1
        if calls["n"] < 3:
            raise FakeError("transient")
        return "ok"

    assert retry_with_backoff(fn, retries=3, base_delay=0.0, retry_on=(FakeError,)) == "ok"
    assert calls["n"] == 3


def test_raises_after_exhausting_retries():
    def fn():
        raise FakeError("always")

    with pytest.raises(FakeError):
        retry_with_backoff(fn, retries=2, base_delay=0.0, retry_on=(FakeError,))


def test_does_not_retry_on_unexpected_exception():
    calls = {"n": 0}

    def fn():
        calls["n"] += 1
        raise ValueError("not retryable")

    with pytest.raises(ValueError):
        retry_with_backoff(fn, retries=3, base_delay=0.0, retry_on=(FakeError,))
    assert calls["n"] == 1
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/test_retry.py -v`
Expected: ImportError.

- [ ] **Step 3: Implement `retry.py`**

```python
# src/sync/retry.py
"""Exponential backoff for transient API failures."""
import logging
import time
from typing import Callable, TypeVar

T = TypeVar("T")
log = logging.getLogger(__name__)


def retry_with_backoff(
    fn: Callable[[], T],
    *,
    retries: int = 3,
    base_delay: float = 1.0,
    retry_on: tuple[type[BaseException], ...] = (Exception,),
) -> T:
    """Call fn(), retrying up to `retries` times on `retry_on` exceptions.

    Sleep `base_delay * 2**attempt` seconds between attempts.
    """
    last_err: BaseException | None = None
    for attempt in range(retries + 1):
        try:
            return fn()
        except retry_on as err:
            last_err = err
            if attempt >= retries:
                break
            delay = base_delay * (2 ** attempt)
            log.warning("attempt %d failed: %s; retrying in %.1fs", attempt + 1, err, delay)
            time.sleep(delay)
    assert last_err is not None
    raise last_err
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/test_retry.py -v`
Expected: 4 passed.

- [ ] **Step 5: Commit**

```bash
git add tests/test_retry.py src/sync/retry.py
git commit -m "feat: add exponential backoff helper"
```

---

### Task 8: Implement logging setup with size-based rotation

**Files:**
- Create: `src/sync/logging_setup.py`

- [ ] **Step 1: Write `logging_setup.py`**

```python
# src/sync/logging_setup.py
"""Configure root logger to write to LOG_PATH with size-based rotation.

We rotate on entry to each run rather than continuously: if the log exceeds
LOG_ROTATE_AT_BYTES at startup, rename to .1 (overwriting any prior .1).
"""
import logging
import sys
from logging import Formatter, StreamHandler

from sync.config import LOG_DIR, LOG_PATH, LOG_ROTATE_AT_BYTES


def _rotate_if_needed() -> None:
    if not LOG_PATH.exists():
        return
    if LOG_PATH.stat().st_size < LOG_ROTATE_AT_BYTES:
        return
    rotated = LOG_PATH.with_suffix(LOG_PATH.suffix + ".1")
    if rotated.exists():
        rotated.unlink()
    LOG_PATH.rename(rotated)


def configure_logging() -> None:
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    _rotate_if_needed()

    fmt = Formatter("%(asctime)s %(levelname)s %(name)s: %(message)s")

    file_handler = logging.FileHandler(LOG_PATH, mode="a", encoding="utf-8")
    file_handler.setFormatter(fmt)

    stream_handler = StreamHandler(sys.stdout)
    stream_handler.setFormatter(fmt)

    root = logging.getLogger()
    root.setLevel(logging.INFO)
    # Avoid duplicate handlers if configure_logging is called twice in tests.
    root.handlers = [file_handler, stream_handler]
```

- [ ] **Step 2: Smoke check**

Run: `uv run python -c "from sync.logging_setup import configure_logging; configure_logging(); import logging; logging.info('hello'); print('ok')"`
Expected: prints `ok`, and `~/Library/Logs/outlook-google-sync.log` now contains a line ending with `hello`.

- [ ] **Step 3: Commit**

```bash
git add src/sync/logging_setup.py
git commit -m "feat: add file logging with size-based rotation"
```

---

### Task 9: Implement Google writer with mocked tests

**Files:**
- Create: `tests/test_google_writer.py`
- Create: `src/sync/google_writer.py`

This module wraps the Google Calendar API. We test the conversion logic (Google response → `TargetEvent`, and `SourceEvent` → API request body). Network calls themselves are mocked at the `service.events()` level.

- [ ] **Step 1: Write the failing tests**

```python
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `uv run pytest tests/test_google_writer.py -v`
Expected: ImportError; module not yet defined.

- [ ] **Step 3: Implement `google_writer.py`**

```python
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `uv run pytest tests/test_google_writer.py -v`
Expected: 8 passed.

- [ ] **Step 5: Commit**

```bash
git add tests/test_google_writer.py src/sync/google_writer.py
git commit -m "feat: add Google Calendar writer with create/update/delete and pagination"
```

---

### Task 10: Implement OAuth `auth` subcommand

**Files:**
- Create: `src/sync/auth.py`

This is interactive and depends on a real OAuth client downloaded from the user's Google Cloud Console. We don't unit-test the OAuth flow itself; we smoke-test by running it once.

- [ ] **Step 1: Implement `auth.py`**

```python
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
```

- [ ] **Step 2: Smoke check imports**

Run: `uv run python -c "from sync.auth import build_calendar_service, run_auth_flow; print('ok')"`
Expected: `ok`

- [ ] **Step 3: Commit**

```bash
git add src/sync/auth.py
git commit -m "feat: add OAuth Desktop flow with token persistence"
```

---

### Task 11: Implement EventKit reader

**Files:**
- Create: `src/sync/eventkit_reader.py`

EventKit is a macOS framework. Reading from it requires (a) the user's terminal/python having Calendar TCC permission, and (b) the Exchange account being added to System Settings → Internet Accounts. This module is smoke-tested manually rather than unit-tested — the unit-testable conversion logic is the diff/filter, which already exists.

- [ ] **Step 1: Implement `eventkit_reader.py`**

```python
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
```

- [ ] **Step 2: Smoke check imports**

Run: `uv run python -c "from sync.eventkit_reader import read_source_events; print('ok')"`
Expected: `ok` (does not actually call EventKit yet — that requires `read_source_events()` to be invoked).

- [ ] **Step 3: Commit**

```bash
git add src/sync/eventkit_reader.py
git commit -m "feat: add EventKit reader with availability mapping and OOO filter"
```

---

### Task 12: Implement orchestrator (the `once` command)

**Files:**
- Create: `src/sync/orchestrator.py`

- [ ] **Step 1: Implement `orchestrator.py`**

```python
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
                lambda: writer.create_event(src),
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
                lambda: writer.update_event(
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
                lambda: writer.delete_event(google_event_id=existing.google_event_id),
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

    try:
        sources = read_source_events()
    except PermissionError as err:
        log.error("source read failed: %s", err)
        return 2
    except Exception as err:
        log.exception("source read failed: %s", err)
        return 2

    try:
        service = build_calendar_service()
    except RuntimeError as err:
        log.error("%s", err)
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
```

- [ ] **Step 2: Smoke check imports**

Run: `uv run python -c "from sync.orchestrator import run_once; print('ok')"`
Expected: `ok`

- [ ] **Step 3: Commit**

```bash
git add src/sync/orchestrator.py
git commit -m "feat: add sync orchestrator with dry-run, retry, per-event error handling"
```

---

### Task 13: Implement launchd plist + installer

**Files:**
- Create: `launchd/com.lto.outlook-google-sync.plist.template`
- Create: `src/sync/installer.py`

- [ ] **Step 1: Write the plist template**

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
        <string>__PYTHON__</string>
        <string>-m</string>
        <string>sync</string>
        <string>once</string>
    </array>

    <key>WorkingDirectory</key>
    <string>__WORKDIR__</string>

    <key>StartInterval</key>
    <integer>1200</integer>

    <key>RunAtLoad</key>
    <true/>

    <key>StandardOutPath</key>
    <string>__LOG__</string>

    <key>StandardErrorPath</key>
    <string>__LOG__</string>
</dict>
</plist>
```

Save as `launchd/com.lto.outlook-google-sync.plist.template`.

- [ ] **Step 2: Implement `installer.py`**

```python
# src/sync/installer.py
"""Render the launchd plist with absolute paths and load it."""
from __future__ import annotations

import logging
import shutil
import subprocess
import sys
from pathlib import Path

from sync.config import (
    LAUNCHD_AGENTS_DIR,
    LAUNCHD_PLIST_PATH,
    LOG_DIR,
    LOG_PATH,
)

log = logging.getLogger(__name__)


def _project_root() -> Path:
    # src/sync/installer.py -> src/sync -> src -> project root
    return Path(__file__).resolve().parents[2]


def _template_path() -> Path:
    return _project_root() / "launchd" / "com.lto.outlook-google-sync.plist.template"


def _render(template: str) -> str:
    venv_python = _project_root() / ".venv" / "bin" / "python"
    if not venv_python.exists():
        raise RuntimeError(
            f"Expected venv python at {venv_python}; run `uv sync` first."
        )
    return (
        template
        .replace("__PYTHON__", str(venv_python))
        .replace("__WORKDIR__", str(_project_root()))
        .replace("__LOG__", str(LOG_PATH))
    )


def install_launchd() -> int:
    template = _template_path().read_text()
    rendered = _render(template)

    LAUNCHD_AGENTS_DIR.mkdir(parents=True, exist_ok=True)
    LOG_DIR.mkdir(parents=True, exist_ok=True)

    # Unload any existing copy first (no-op if not loaded).
    subprocess.run(
        ["launchctl", "unload", str(LAUNCHD_PLIST_PATH)],
        check=False,
        capture_output=True,
    )

    LAUNCHD_PLIST_PATH.write_text(rendered)
    log.info("wrote %s", LAUNCHD_PLIST_PATH)

    result = subprocess.run(
        ["launchctl", "load", str(LAUNCHD_PLIST_PATH)],
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        print(f"launchctl load failed:\nstdout: {result.stdout}\nstderr: {result.stderr}",
              file=sys.stderr)
        return 1
    log.info("launchctl loaded %s", LAUNCHD_PLIST_PATH)
    print(f"Installed and loaded {LAUNCHD_PLIST_PATH}")
    return 0
```

- [ ] **Step 3: Smoke check imports**

Run: `uv run python -c "from sync.installer import install_launchd; print('ok')"`
Expected: `ok`

- [ ] **Step 4: Commit**

```bash
git add launchd/com.lto.outlook-google-sync.plist.template src/sync/installer.py
git commit -m "feat: add launchd plist template and installer"
```

---

### Task 14: Write README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write `README.md`**

```markdown
# outlook-google-sync

One-way sync from Axon work calendar (Outlook / Exchange, via macOS EventKit) to a dedicated Google Calendar account, run every 20 minutes by launchd.

Design: `docs/superpowers/specs/2026-04-30-outlook-google-calendar-sync-design.md`

## What gets mirrored

- Title + start + end only. No body, attendees, location.
- OOO events skipped (availability=Unavailable, or title matching `OOO|OOF|out of office|vacation|PTO|on leave|holiday`).
- Recurring events flattened: one mirrored event per occurrence.
- Window: past 7 days through future 30 days.

### Known limitations

- **All-day events** appear as 24-hour timed blocks in the mirror (e.g., "Conference: KubeCon" shows as a midnight-to-midnight timed event). The OOO filter usually catches the all-day events you'd want skipped (holidays, vacation), so this is rarely visible.

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
6. Download the JSON, save to `~/.config/outlook-google-sync/credentials.json`.

### 5. Run OAuth flow

```bash
mkdir -p ~/.config/outlook-google-sync
# put credentials.json there first
uv run python -m sync auth
```

A browser window opens, you sign in as the dedicated account, click through the unverified-app warning, grant Calendar access. The script saves `~/.config/outlook-google-sync/token.json`.

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
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with setup checklist and troubleshooting"
```

---

### Task 15: Full test suite + smoke test

- [ ] **Step 1: Run all unit tests**

Run: `uv run pytest -v`
Expected: all tests pass (≈30 tests across diff, filters, iso_utc, retry, google_writer).

- [ ] **Step 2: Manually run the dry-run smoke test**

Prerequisite: tasks 1–14 are complete, OAuth credentials are placed and `python -m sync auth` has been run, Outlook account is in macOS Calendar.

Run: `uv run python -m sync once --dry-run`

Expected:
- macOS Terminal prompts for Calendar access on the very first invocation. Click OK.
- Stdout shows lines like `[dry-run] CREATE <id> 'Standup' 2026-05-01T15:00:00+00:00`
- Final summary line: `synced: N created, 0 updated, 0 deleted, ...` (since target is empty on first dry-run, all source events show as creates).
- Log file `~/Library/Logs/outlook-google-sync.log` mirrors the same content.

- [ ] **Step 3: Real first-run**

Run: `uv run python -m sync once`

Expected:
- Same volume of CREATEs, this time actually applied.
- Verify in browser: open Google Calendar as the dedicated account; the mirrored events appear in the primary calendar.
- Re-run the same command; expect `0 created, 0 updated, 0 deleted, N unchanged` (idempotent).

- [ ] **Step 4: Install and verify launchd**

Run: `uv run python -m sync install-launchd`
Then: `launchctl list | grep com.lto.outlook-google-sync`
Expected: line showing the job is loaded with PID `-` (currently waiting) or a numeric PID (currently running).

After ~20 minutes, check the log for a second sync run with `synced: 0 created, 0 updated, 0 deleted, ...` confirming the schedule fires.

- [ ] **Step 5: Verify wake-from-sleep behavior**

Close the lid for at least 30 minutes, reopen. Within ~30 seconds of wake, check `~/Library/Logs/outlook-google-sync.log` — a new run should have fired immediately on wake (StartInterval coalesces missed ticks into one fire).

- [ ] **Step 6: Final commit**

If any small fix-ups were needed during smoke testing, commit them. Otherwise:

```bash
git tag v0.1.0
```

---

## Done criteria

- All unit tests pass.
- A real run populates the dedicated Google Calendar with mirror events that match Outlook events title/time-only, OOO events skipped.
- Re-running `python -m sync once` immediately after produces a no-op summary (idempotent).
- launchd job is loaded and fires every 20 minutes plus on wake from sleep.
- Primary Gmail subscribes to the dedicated account's calendar and the events appear inline in the unified view.

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

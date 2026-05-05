# src/sync/installer.py
"""Render the launchd plist with absolute paths and load it."""
from __future__ import annotations

import logging
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

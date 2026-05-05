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

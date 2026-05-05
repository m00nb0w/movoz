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

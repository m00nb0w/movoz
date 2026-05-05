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

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

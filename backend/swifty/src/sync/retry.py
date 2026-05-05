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

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

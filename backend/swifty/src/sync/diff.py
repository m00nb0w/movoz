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

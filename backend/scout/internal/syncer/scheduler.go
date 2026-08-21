package syncer

import (
	"context"
	"log"
	"time"

	"scout/internal/models"
)

// CycleLister is the subset of store.CycleStore the scheduler depends on, so
// tests can substitute a fake. Implementations must return cycles ordered
// newest-first (store.CycleStore.List orders by period_start DESC).
type CycleLister interface {
	List() ([]models.RatingCycle, error)
}

// RunSyncCycle runs one sync pass over the *current* rating cycle's window.
//
// The window is the admin-created rating cycle's own period_start/period_end
// rather than a synthetic rolling "now minus 14 days" range. metric_snapshots
// is keyed on (engineer_id, period_start, period_end) with DATE columns, so a
// rolling window silently produced a brand-new near-duplicate snapshot row on
// every new calendar day; anchoring to the cycle means every run inside one
// cycle's lifetime upserts the *same* row, and a new row only appears once the
// admin opens the next cycle. It also keeps Syncer's githubBoundaryOffset
// meaningful: consecutive windows are genuinely back-to-back when the admin
// creates cycles back-to-back, which is the expected usage pattern.
//
// If no rating cycle exists yet (fresh install, admin has not opened one),
// the run is skipped with a log line rather than falling back to some other
// window — there is deliberately only one window-computation scheme.
//
// Split from StartScheduler so the window derivation is unit-testable without
// waiting on a real ticker.
func RunSyncCycle(ctx context.Context, s *Syncer, cycles CycleLister) error {
	list, err := cycles.List()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		log.Print("scout sync: no rating cycle exists yet, skipping metrics sync")
		return nil
	}
	current := list[0]
	return s.RunOnce(ctx, current.PeriodStart, current.PeriodEnd)
}

// StartScheduler runs an immediate sync pass, then one every `interval`,
// until ctx is cancelled. Runs as an in-process goroutine started from
// cmd/server/main.go — there is no separate cmd/syncer binary (see the
// plan's judgment-calls section).
func StartScheduler(ctx context.Context, s *Syncer, cycles CycleLister, interval time.Duration) {
	go func() {
		if err := RunSyncCycle(ctx, s, cycles); err != nil {
			log.Printf("scout sync: initial run failed: %v", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := RunSyncCycle(ctx, s, cycles); err != nil {
					log.Printf("scout sync: scheduled run failed: %v", err)
				}
			}
		}
	}()
}

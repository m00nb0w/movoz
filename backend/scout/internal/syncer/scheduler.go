package syncer

import (
	"context"
	"log"
	"time"
)

// syncPeriodDays is the trailing window RunSyncCycle syncs on each run.
// Chosen to match the biweekly cadence engineer scoring cycles use
// elsewhere in Scout (store.CycleStore); see the plan's judgment-calls
// section.
const syncPeriodDays = 14

// RunSyncCycle computes the trailing biweekly period ending at `now` and
// runs one sync pass. Split from StartScheduler so the period math is
// unit-testable without waiting on a real ticker.
func RunSyncCycle(ctx context.Context, s *Syncer, now time.Time) error {
	periodEnd := now
	periodStart := now.AddDate(0, 0, -syncPeriodDays)
	return s.RunOnce(ctx, periodStart, periodEnd)
}

// StartScheduler runs an immediate sync pass, then one every `interval`,
// until ctx is cancelled. Runs as an in-process goroutine started from
// cmd/server/main.go — there is no separate cmd/syncer binary (see the
// plan's judgment-calls section).
func StartScheduler(ctx context.Context, s *Syncer, interval time.Duration) {
	go func() {
		if err := RunSyncCycle(ctx, s, time.Now()); err != nil {
			log.Printf("scout sync: initial run failed: %v", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := RunSyncCycle(ctx, s, time.Now()); err != nil {
					log.Printf("scout sync: scheduled run failed: %v", err)
				}
			}
		}
	}()
}

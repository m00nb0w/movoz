package syncer

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/store"
)

// erroringCycleLister lets the tests exercise RunSyncCycle's error path
// without needing to break the database.
type erroringCycleLister struct{ err error }

func (e *erroringCycleLister) List() ([]models.RatingCycle, error) { return nil, e.err }

func TestRunSyncCycleUsesCurrentRatingCyclePeriod(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, rating_cycles, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	cycleStore := store.NewCycleStore(db)
	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	// An older cycle plus the current one: RunSyncCycle must pick the current
	// (newest) cycle's window, not the older one and not a rolling window.
	oldStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	oldEnd := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	if _, err := cycleStore.Create(oldStart, oldEnd); err != nil {
		t.Fatalf("failed to create old cycle: %v", err)
	}
	currentStart := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	currentEnd := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if _, err := cycleStore.Create(currentStart, currentEnd); err != nil {
		t.Fatalf("failed to create current cycle: %v", err)
	}

	fakeGH := &fakeGitHub{raised: 1, reviewed: 1}
	fakeJ := &fakeJira{closed: 1, complexity: 1}
	s := NewSyncer(engineerStore, metricStore, fakeGH, fakeJ, []string{"org/repo"}, []string{"ENG"})

	if err := RunSyncCycle(context.Background(), s, cycleStore); err != nil {
		t.Fatalf("run sync cycle failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 || !snapshots[0].PeriodStart.Equal(currentStart) || !snapshots[0].PeriodEnd.Equal(currentEnd) {
		t.Fatalf("expected a single snapshot for the current cycle's period [%s, %s], got %+v", currentStart, currentEnd, snapshots)
	}

	// The values handed to RunOnce must be the cycle's own boundaries, with
	// only GitHub's inclusive `until` shifted back a day (githubBoundaryOffset).
	if fakeJ.callCount() != 1 {
		t.Fatalf("expected exactly 1 jira call, got %d", fakeJ.callCount())
	}
	if got := fakeJ.call(0); !got.since.Equal(currentStart) || !got.until.Equal(currentEnd) {
		t.Fatalf("expected jira called with the cycle's [%s, %s), got [%s, %s)", currentStart, currentEnd, got.since, got.until)
	}
	if fakeGH.callCount() != 1 {
		t.Fatalf("expected exactly 1 github call, got %d", fakeGH.callCount())
	}
	wantGitHubUntil := currentEnd.Add(githubBoundaryOffset)
	if got := fakeGH.call(0); !got.since.Equal(currentStart) || !got.until.Equal(wantGitHubUntil) {
		t.Fatalf("expected github called with [%s, %s], got [%s, %s]", currentStart, wantGitHubUntil, got.since, got.until)
	}
}

// TestRunSyncCycleRepeatRunsUpsertTheSameSnapshotRow is the point of anchoring
// the window to the rating cycle: the scheduler fires every
// SYNC_INTERVAL_HOURS (12 by default), so repeat runs across days must keep
// updating one row per (engineer, cycle) rather than accumulating a
// near-duplicate snapshot per calendar day the way the old rolling
// "now minus 14 days" window did.
func TestRunSyncCycleRepeatRunsUpsertTheSameSnapshotRow(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, rating_cycles, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	cycleStore := store.NewCycleStore(db)
	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	periodStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	if _, err := cycleStore.Create(periodStart, periodEnd); err != nil {
		t.Fatalf("failed to create cycle: %v", err)
	}

	first := &fakeGitHub{raised: 1, reviewed: 1}
	s := NewSyncer(engineerStore, metricStore, first, &fakeJira{closed: 1, complexity: 1}, nil, nil)
	if err := RunSyncCycle(context.Background(), s, cycleStore); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	second := &fakeGitHub{raised: 7, reviewed: 4}
	s2 := NewSyncer(engineerStore, metricStore, second, &fakeJira{closed: 3, complexity: 5}, nil, nil)
	if err := RunSyncCycle(context.Background(), s2, cycleStore); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected repeat runs within one cycle to upsert a single snapshot row, got %d: %+v", len(snapshots), snapshots)
	}
	if snapshots[0].PRsRaised != 7 || snapshots[0].TicketsClosed != 3 {
		t.Fatalf("expected the second run's values to overwrite the first, got %+v", snapshots[0])
	}
}

// TestRunSyncCycleSkipsWhenNoRatingCycleExists covers the fresh-install case:
// before the admin opens their first cycle there is no window to sync, and the
// scheduler must skip cleanly rather than invent one.
func TestRunSyncCycleSkipsWhenNoRatingCycleExists(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, rating_cycles, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	cycleStore := store.NewCycleStore(db)
	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	fakeGH := &fakeGitHub{raised: 1, reviewed: 1}
	fakeJ := &fakeJira{closed: 1, complexity: 1}
	s := NewSyncer(engineerStore, metricStore, fakeGH, fakeJ, nil, nil)

	if err := RunSyncCycle(context.Background(), s, cycleStore); err != nil {
		t.Fatalf("expected a clean skip with no error, got %v", err)
	}
	if fakeGH.callCount() != 0 || fakeJ.callCount() != 0 {
		t.Fatalf("expected no integration calls when no cycle exists, got github=%d jira=%d", fakeGH.callCount(), fakeJ.callCount())
	}
	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected no snapshots to be written, got %+v", snapshots)
	}
}

func TestRunSyncCycleReturnsCycleLookupError(t *testing.T) {
	db := setupTestDB(t)
	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	s := NewSyncer(engineerStore, metricStore, &fakeGitHub{}, &fakeJira{}, nil, nil)

	wantErr := errors.New("cycle lookup exploded")
	err := RunSyncCycle(context.Background(), s, &erroringCycleLister{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected the cycle lookup error to propagate, got %v", err)
	}
}

// TestStartSchedulerRunsImmediatelyThenOnEachTick verifies StartScheduler
// actually fires repeatedly on the configured interval, rather than just
// trusting the ticker loop reads correctly. It uses a short interval
// (20ms) instead of a real long sleep so the test stays fast. Rather than
// sleeping a single fixed duration and asserting a run count (flaky under
// parallel-package test load, where CPU contention can stretch scheduling
// well past a short fixed window), it polls with a generous 2s deadline:
// normally resolves in well under 100ms, and only fails if the ticker
// genuinely never fires enough times.
func TestStartSchedulerRunsImmediatelyThenOnEachTick(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, rating_cycles, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	cycleStore := store.NewCycleStore(db)
	gh := "octocat"
	jira := "abc123"
	engineerStore.Create("Alex", nil, &gh, &jira, time.Now())
	if _, err := cycleStore.Create(
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("failed to create cycle: %v", err)
	}

	fakeGH := &fakeGitHub{raised: 1, reviewed: 1}
	fakeJ := &fakeJira{closed: 1, complexity: 1}
	s := NewSyncer(engineerStore, metricStore, fakeGH, fakeJ, []string{"org/repo"}, []string{"ENG"})

	ctx, cancel := context.WithCancel(context.Background())
	StartScheduler(ctx, s, cycleStore, 20*time.Millisecond)

	const wantRuns = 2
	deadline := time.Now().Add(2 * time.Second)
	for fakeGH.callCount() < wantRuns && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	got := fakeGH.callCount()
	cancel()
	if got < wantRuns {
		t.Fatalf("expected StartScheduler to have run at least %d times (the immediate run plus >=1 tick) within 2s, got %d runs", wantRuns, got)
	}

	// The scheduler must actually stop once ctx is cancelled: give any
	// in-flight run a moment to observe cancellation, record the count,
	// then wait past several more would-be ticks and confirm no further
	// runs happened.
	time.Sleep(50 * time.Millisecond)
	afterCancel := fakeGH.callCount()
	time.Sleep(150 * time.Millisecond)
	if fakeGH.callCount() != afterCancel {
		t.Fatalf("expected StartScheduler to stop running after ctx was cancelled, but call count grew from %d to %d", afterCancel, fakeGH.callCount())
	}
}

package syncer

import (
	"context"
	"testing"
	"time"

	"scout/internal/store"
)

func TestRunSyncCycleUsesTrailingBiweeklyPeriod(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	s := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 1, reviewed: 1}, &fakeJira{closed: 1, complexity: 1}, nil, nil)

	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	if err := RunSyncCycle(context.Background(), s, now); err != nil {
		t.Fatalf("run sync cycle failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	wantStart := now.AddDate(0, 0, -14)
	if len(snapshots) != 1 || !snapshots[0].PeriodStart.Equal(wantStart) || !snapshots[0].PeriodEnd.Equal(now) {
		t.Fatalf("expected period [%s, %s], got %+v", wantStart, now, snapshots)
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
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)
	gh := "octocat"
	jira := "abc123"
	engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	fakeGH := &fakeGitHub{raised: 1, reviewed: 1}
	fakeJ := &fakeJira{closed: 1, complexity: 1}
	s := NewSyncer(engineerStore, metricStore, fakeGH, fakeJ, []string{"org/repo"}, []string{"ENG"})

	ctx, cancel := context.WithCancel(context.Background())
	StartScheduler(ctx, s, 20*time.Millisecond)

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

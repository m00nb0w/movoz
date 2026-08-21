package syncer

import (
	"context"
	"errors"
	"testing"
	"time"

	"scout/internal/store"
)

func TestSyncerRunOnceUpsertsSnapshotsForActiveEngineers(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	s := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 3, reviewed: 2}, &fakeJira{closed: 4, complexity: 9}, []string{"org/repo"}, []string{"ENG"})

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	if err := s.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("run once failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 || snapshots[0].PRsRaised != 3 || snapshots[0].TicketsClosed != 4 {
		t.Fatalf("unexpected snapshots: %+v", snapshots)
	}
}

func TestSyncerRunOnceSkipsFailingEngineerButContinues(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gh1, gh2 := "broken-user", "working-user"
	jira := "abc123"
	e1, _ := engineerStore.Create("Broken", nil, &gh1, &jira, time.Now())
	e2, _ := engineerStore.Create("Working", nil, &gh2, &jira, time.Now())

	failingGitHub := &failOnceThenSucceedGitHub{failFor: gh1}
	s := NewSyncer(engineerStore, metricStore, failingGitHub, &fakeJira{closed: 1, complexity: 1}, nil, nil)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	if err := s.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("expected RunOnce to log-and-continue past a per-engineer failure, got error: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e2.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected the working engineer to still get a snapshot despite the other's failure, got %+v", snapshots)
	}

	// The broken engineer's failed fetch must not leave behind a partial or
	// zero-valued snapshot — RunOnce should skip the upsert entirely for
	// that engineer, not write corrupted/incomplete data.
	brokenSnapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(brokenSnapshots) != 0 {
		t.Fatalf("expected the broken engineer to have no snapshot written after its fetch failed, got %+v", brokenSnapshots)
	}
}

type failOnceThenSucceedGitHub struct {
	failFor string
}

func (f *failOnceThenSucceedGitHub) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (int, int, error) {
	if username == f.failFor {
		return 0, 0, errors.New("simulated github outage")
	}
	return 1, 1, nil
}

// TestSyncerRunOnceIsIdempotentOnRerun exercises the idempotent-upsert
// behavior end-to-end through the orchestrator (not just at the
// MetricStore level, which Task 17 already covers directly): running
// RunOnce twice with the same period and updated fake stats must update the
// single existing snapshot row in place, never insert a second one, and the
// row must reflect the latest run's numbers.
func TestSyncerRunOnceIsIdempotentOnRerun(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gh := "octocat"
	jira := "abc123"
	e1, _ := engineerStore.Create("Alex", nil, &gh, &jira, time.Now())

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)

	s1 := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 3, reviewed: 2}, &fakeJira{closed: 4, complexity: 9}, []string{"org/repo"}, []string{"ENG"})
	if err := s1.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	// Second sync run for the same period, e.g. after the engineer raised
	// one more PR and closed one more ticket before the next scheduled
	// sync — the stats a real client would return for the same window can
	// change between runs (new activity, or the previous run only observed
	// a partial page). RunOnce must overwrite, not duplicate.
	s2 := NewSyncer(engineerStore, metricStore, &fakeGitHub{raised: 4, reviewed: 2}, &fakeJira{closed: 5, complexity: 9}, []string{"org/repo"}, []string{"ENG"})
	if err := s2.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected exactly one snapshot after two runs over the same period, got %d: %+v", len(snapshots), snapshots)
	}
	if snapshots[0].PRsRaised != 4 || snapshots[0].TicketsClosed != 5 {
		t.Fatalf("expected the rerun's snapshot to reflect the second run's stats, got %+v", snapshots[0])
	}
}

// TestSyncerRunOnceOffsetsGitHubUntilByOneDayForInclusiveBoundary verifies
// the fix Task 20 must apply per integrations.GitHubClient.FetchPRStats's
// doc comment: because GitHub's date range is inclusive on both ends,
// naively passing the same periodStart/periodEnd to both clients would
// double-count PRs created on the day two consecutive sync cycles share as
// one cycle's periodEnd and the next cycle's periodStart. This test doesn't
// just trust the code looks right — it captures the actual since/until
// RunOnce passed to each fake client and asserts on them directly.
func TestSyncerRunOnceOffsetsGitHubUntilByOneDayForInclusiveBoundary(t *testing.T) {
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

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC) // this cycle's periodEnd == next cycle's periodStart

	if err := s.RunOnce(context.Background(), start, end); err != nil {
		t.Fatalf("run once failed: %v", err)
	}

	if fakeGH.callCount() != 1 {
		t.Fatalf("expected exactly one github call, got %d", fakeGH.callCount())
	}
	if fakeJ.callCount() != 1 {
		t.Fatalf("expected exactly one jira call, got %d", fakeJ.callCount())
	}

	githubCall := fakeGH.call(0)
	jiraCall := fakeJ.call(0)

	if !githubCall.since.Equal(start) {
		t.Fatalf("expected github since to equal periodStart %s unchanged, got %s", start, githubCall.since)
	}
	if !jiraCall.since.Equal(start) {
		t.Fatalf("expected jira since to equal periodStart %s unchanged, got %s", start, jiraCall.since)
	}
	if !jiraCall.until.Equal(end) {
		t.Fatalf("expected jira until to equal periodEnd %s unchanged (its range is already half-open), got %s", end, jiraCall.until)
	}

	wantGithubUntil := end.AddDate(0, 0, -1)
	if !githubCall.until.Equal(wantGithubUntil) {
		t.Fatalf("expected github until to be periodEnd minus one day (%s) to avoid double-counting the boundary shared with the next sync cycle, got %s", wantGithubUntil, githubCall.until)
	}

	// Directly confirm the fix's purpose: the github until this cycle used
	// must fall strictly before the since the *next* cycle would use (which
	// is this cycle's periodEnd when the admin opens rating cycles
	// back-to-back — the window RunSyncCycle syncs),
	// so the two cycles' inclusive GitHub ranges never overlap on a shared
	// day.
	nextCycleGithubSince := end
	if !githubCall.until.Before(nextCycleGithubSince) {
		t.Fatalf("github windows would overlap across consecutive cycles: this cycle's until %s is not before next cycle's since %s", githubCall.until, nextCycleGithubSince)
	}
}

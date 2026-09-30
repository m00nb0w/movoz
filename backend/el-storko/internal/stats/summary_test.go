package stats

import (
	"testing"
	"time"

	"el-storko/internal/models"
)

func TestComputeSummary_EmptyItems(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")

	summary := ComputeSummary(nil, now)

	if summary.OpenItems != 0 {
		t.Errorf("expected OpenItems=0, got %d", summary.OpenItems)
	}
	if summary.CompletedThisWeek != 0 {
		t.Errorf("expected CompletedThisWeek=0, got %d", summary.CompletedThisWeek)
	}
	if summary.CompletionRate != 0 {
		t.Errorf("expected CompletionRate=0, got %v", summary.CompletionRate)
	}
	if summary.TotalTracked != 0 {
		t.Errorf("expected TotalTracked=0, got %d", summary.TotalTracked)
	}
	want := map[models.Status]int{
		models.StatusBacklog:        0,
		models.StatusPickedForToday: 0,
		models.StatusInProgress:     0,
		models.StatusBlocked:        0,
		models.StatusDone:           0,
	}
	for status, count := range want {
		if got := summary.StatusBreakdown[status]; got != count {
			t.Errorf("status %s: expected %d, got %d", status, count, got)
		}
	}
	if len(summary.StatusBreakdown) != 5 {
		t.Errorf("expected all five status keys present, got %d keys: %+v", len(summary.StatusBreakdown), summary.StatusBreakdown)
	}
}

func TestComputeSummary_MixOfAllStatuses(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")

	completedWithinWeek := mustTime(t, "2026-01-08T10:00:00Z")
	completedOutsideWeek := mustTime(t, "2025-12-20T10:00:00Z")

	items := []models.WorkItem{
		{Status: models.StatusBacklog, CreatedAt: mustTime(t, "2026-01-01T00:00:00Z")},
		{Status: models.StatusPickedForToday, CreatedAt: mustTime(t, "2026-01-02T00:00:00Z")},
		{Status: models.StatusInProgress, CreatedAt: mustTime(t, "2026-01-03T00:00:00Z")},
		{Status: models.StatusBlocked, CreatedAt: mustTime(t, "2026-01-04T00:00:00Z")},
		// two Done items: one completed within the trailing 7 days, one outside it
		{Status: models.StatusDone, CreatedAt: mustTime(t, "2025-12-19T00:00:00Z"), CompletedAt: &completedOutsideWeek},
		{Status: models.StatusDone, CreatedAt: mustTime(t, "2026-01-07T00:00:00Z"), CompletedAt: &completedWithinWeek},
	}

	summary := ComputeSummary(items, now)

	if summary.OpenItems != 4 {
		t.Errorf("expected OpenItems=4 (all non-done), got %d", summary.OpenItems)
	}
	if summary.CompletedThisWeek != 1 {
		t.Errorf("expected CompletedThisWeek=1, got %d", summary.CompletedThisWeek)
	}
	wantRate := 2.0 / 6.0
	if summary.CompletionRate != wantRate {
		t.Errorf("expected CompletionRate=%v, got %v", wantRate, summary.CompletionRate)
	}
	if summary.TotalTracked != 6 {
		t.Errorf("expected TotalTracked=6, got %d", summary.TotalTracked)
	}
	want := map[models.Status]int{
		models.StatusBacklog:        1,
		models.StatusPickedForToday: 1,
		models.StatusInProgress:     1,
		models.StatusBlocked:        1,
		models.StatusDone:           2,
	}
	for status, count := range want {
		if got := summary.StatusBreakdown[status]; got != count {
			t.Errorf("status %s: expected %d, got %d", status, count, got)
		}
	}
}

func TestComputeSummary_CompletedThisWeekBoundary(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")

	exactlySevenDaysAgo := now.AddDate(0, 0, -7)
	justInsideWindow := now.AddDate(0, 0, -6).Add(-time.Hour)
	justOutsideWindow := exactlySevenDaysAgo.Add(-time.Minute)

	items := []models.WorkItem{
		{Status: models.StatusDone, CreatedAt: now, CompletedAt: &justInsideWindow},
		{Status: models.StatusDone, CreatedAt: now, CompletedAt: &justOutsideWindow},
	}

	summary := ComputeSummary(items, now)

	if summary.CompletedThisWeek != 1 {
		t.Errorf("expected exactly 1 item within trailing 7 days, got %d", summary.CompletedThisWeek)
	}
}

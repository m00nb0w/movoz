package stats

import (
	"testing"
	"time"

	"el-storko/internal/models"
)

func mustTime(t *testing.T, s string) time.Time {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("bad test time %q: %v", s, err)
	}
	return tm
}

func TestComputeBurnRate(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")

	completedA := mustTime(t, "2026-01-07T10:00:00Z")
	completedC := mustTime(t, "2026-01-09T15:00:00Z")

	items := []models.WorkItem{
		{ // completed once, never reopened
			CreatedAt:   mustTime(t, "2026-01-05T00:00:00Z"),
			CompletedAt: &completedA,
		},
		{ // created mid-window, never completed
			CreatedAt:   mustTime(t, "2026-01-08T09:00:00Z"),
			CompletedAt: nil,
		},
		{ // completed, reopened, then recompleted later — only the final
			// completed_at matters to this pure function; the store layer
			// (tested separately) is what guarantees this final value is
			// correct after a reopen/recomplete cycle.
			CreatedAt:   mustTime(t, "2026-01-06T08:00:00Z"),
			CompletedAt: &completedC,
		},
	}

	points := ComputeBurnRate(items, 5, now)

	want := []Point{
		{Date: "2026-01-06", Completed: 0, Open: 2},
		{Date: "2026-01-07", Completed: 1, Open: 1},
		{Date: "2026-01-08", Completed: 0, Open: 2},
		{Date: "2026-01-09", Completed: 1, Open: 1},
		{Date: "2026-01-10", Completed: 0, Open: 1},
	}

	if len(points) != len(want) {
		t.Fatalf("expected %d points, got %d: %+v", len(want), len(points), points)
	}
	for i, w := range want {
		if points[i] != w {
			t.Errorf("point %d: expected %+v, got %+v", i, w, points[i])
		}
	}
}

func TestComputeBurnRate_EmptyItems(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")
	points := ComputeBurnRate(nil, 3, now)
	if len(points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(points))
	}
	for _, p := range points {
		if p.Completed != 0 || p.Open != 0 {
			t.Errorf("expected all-zero point for empty tracker, got %+v", p)
		}
	}
}

func TestComputeBurnRate_DefaultsDaysWhenNonPositive(t *testing.T) {
	now := mustTime(t, "2026-01-10T12:00:00Z")
	points := ComputeBurnRate(nil, 0, now)
	if len(points) != 30 {
		t.Fatalf("expected default window of 30 days, got %d", len(points))
	}
}

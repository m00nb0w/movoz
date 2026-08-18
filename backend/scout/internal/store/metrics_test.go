package store

import (
	"testing"
	"time"
)

func TestMetricStoreUpsertSnapshotIsIdempotent(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)

	if err := metricStore.UpsertSnapshot(e1.ID, start, end, 3, 5, 2, 7.5); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}
	// Re-running the sync for the same period must update in place, not
	// duplicate the row (NF4).
	if err := metricStore.UpsertSnapshot(e1.ID, start, end, 4, 6, 3, 9.0); err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected exactly one snapshot after repeated upsert, got %d", len(snapshots))
	}
	if snapshots[0].PRsRaised != 4 || snapshots[0].TicketsClosed != 3 {
		t.Fatalf("expected the second upsert's values to win, got %+v", snapshots[0])
	}
}

func TestMetricStoreMultipleEngineersDoNotCollide(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Alice", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Bob", nil, nil, nil, time.Now())
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)

	// Upsert same period for both engineers
	if err := metricStore.UpsertSnapshot(e1.ID, start, end, 3, 5, 2, 7.5); err != nil {
		t.Fatalf("alice upsert failed: %v", err)
	}
	if err := metricStore.UpsertSnapshot(e2.ID, start, end, 10, 15, 8, 9.2); err != nil {
		t.Fatalf("bob upsert failed: %v", err)
	}

	snapshotsE1, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list e1 failed: %v", err)
	}
	snapshotsE2, err := metricStore.ListByEngineer(e2.ID)
	if err != nil {
		t.Fatalf("list e2 failed: %v", err)
	}

	if len(snapshotsE1) != 1 {
		t.Fatalf("expected one snapshot for e1, got %d", len(snapshotsE1))
	}
	if len(snapshotsE2) != 1 {
		t.Fatalf("expected one snapshot for e2, got %d", len(snapshotsE2))
	}

	if snapshotsE1[0].EngineerID != e1.ID || snapshotsE1[0].PRsRaised != 3 {
		t.Fatalf("e1 snapshot has wrong values: %+v", snapshotsE1[0])
	}
	if snapshotsE2[0].EngineerID != e2.ID || snapshotsE2[0].PRsRaised != 10 {
		t.Fatalf("e2 snapshot has wrong values: %+v", snapshotsE2[0])
	}
}

func TestMetricStoreMultiplePeriodsForSameEngineer(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Charlie", nil, nil, nil, time.Now())
	period1Start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	period1End := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	period2Start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	period2End := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	// Upsert two different periods for same engineer
	if err := metricStore.UpsertSnapshot(e1.ID, period1Start, period1End, 3, 5, 2, 7.5); err != nil {
		t.Fatalf("period1 upsert failed: %v", err)
	}
	if err := metricStore.UpsertSnapshot(e1.ID, period2Start, period2End, 4, 6, 3, 8.0); err != nil {
		t.Fatalf("period2 upsert failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(snapshots) != 2 {
		t.Fatalf("expected two snapshots for same engineer with different periods, got %d", len(snapshots))
	}

	// Verify both periods exist with correct values
	period1Found := false
	period2Found := false
	for _, s := range snapshots {
		if s.PeriodStart.Equal(period1Start) && s.PeriodEnd.Equal(period1End) {
			period1Found = true
			if s.PRsRaised != 3 || s.TicketsClosed != 2 {
				t.Fatalf("period1 snapshot has wrong values: %+v", s)
			}
		}
		if s.PeriodStart.Equal(period2Start) && s.PeriodEnd.Equal(period2End) {
			period2Found = true
			if s.PRsRaised != 4 || s.TicketsClosed != 3 {
				t.Fatalf("period2 snapshot has wrong values: %+v", s)
			}
		}
	}
	if !period1Found || !period2Found {
		t.Fatalf("not all periods found in snapshots: %+v", snapshots)
	}
}

func TestMetricStoreListByEngineerOrdering(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Diana", nil, nil, nil, time.Now())

	// Create snapshots in non-chronological order
	period3Start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	period3End := time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC)
	period1Start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	period1End := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	period2Start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	period2End := time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC)

	// Insert in order: period3, period1, period2
	if err := metricStore.UpsertSnapshot(e1.ID, period3Start, period3End, 5, 7, 4, 9.0); err != nil {
		t.Fatalf("period3 upsert failed: %v", err)
	}
	if err := metricStore.UpsertSnapshot(e1.ID, period1Start, period1End, 1, 2, 1, 6.0); err != nil {
		t.Fatalf("period1 upsert failed: %v", err)
	}
	if err := metricStore.UpsertSnapshot(e1.ID, period2Start, period2End, 3, 4, 2, 7.5); err != nil {
		t.Fatalf("period2 upsert failed: %v", err)
	}

	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if len(snapshots) != 3 {
		t.Fatalf("expected three snapshots, got %d", len(snapshots))
	}

	// Verify ordering: should be DESC by period_start (newest first)
	// Expected: period3, period2, period1
	if !snapshots[0].PeriodStart.Equal(period3Start) {
		t.Fatalf("first snapshot should be period3 (DESC order), got %v", snapshots[0].PeriodStart)
	}
	if !snapshots[1].PeriodStart.Equal(period2Start) {
		t.Fatalf("second snapshot should be period2 (DESC order), got %v", snapshots[1].PeriodStart)
	}
	if !snapshots[2].PeriodStart.Equal(period1Start) {
		t.Fatalf("third snapshot should be period1 (DESC order), got %v", snapshots[2].PeriodStart)
	}
}

func TestMetricStoreListByEngineerEmpty(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	metricStore := NewMetricStore(db)

	e1, _ := engineerStore.Create("Eve", nil, nil, nil, time.Now())

	// Don't create any snapshots for this engineer
	snapshots, err := metricStore.ListByEngineer(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if snapshots == nil {
		t.Fatalf("expected empty slice, not nil")
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected zero snapshots for engineer with no data, got %d", len(snapshots))
	}
}

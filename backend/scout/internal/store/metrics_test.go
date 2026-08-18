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

package store

import (
	"testing"
	"time"
)

func TestMatchdayStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewMatchdayStore(db)

	d2026 := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	d2025 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	if _, err := s.Create(d2026); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if _, err := s.Create(d2025); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	all, err := s.List(nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 matchdays, got %d", len(all))
	}

	year := 2026
	filtered, err := s.List(&year)
	if err != nil {
		t.Fatalf("List(2026) failed: %v", err)
	}
	if len(filtered) != 1 || filtered[0].PlayedOn.Year() != 2026 {
		t.Fatalf("expected 1 matchday in 2026, got %+v", filtered)
	}
}

func TestMatchdayStoreExists(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewMatchdayStore(db)

	created, _ := s.Create(time.Now())

	exists, err := s.Exists(created.ID)
	if err != nil || !exists {
		t.Fatalf("expected matchday to exist, err=%v exists=%v", err, exists)
	}

	exists, err = s.Exists(99999)
	if err != nil || exists {
		t.Fatalf("expected missing matchday to not exist, err=%v exists=%v", err, exists)
	}
}

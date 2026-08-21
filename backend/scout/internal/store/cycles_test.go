package store

import (
	"testing"
	"time"
)

func TestCycleStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewCycleStore(db)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	created, err := s.Create(start, end)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if !created.PeriodStart.Equal(start) || !created.PeriodEnd.Equal(end) {
		t.Fatalf("unexpected created cycle: %+v", created)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected one cycle, got %+v", list)
	}
}

func TestCycleStoreGetByID(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewCycleStore(db)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	created, _ := s.Create(start, end)

	found, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Fatalf("expected to find cycle, got %+v", found)
	}

	missing, err := s.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID for missing id returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing cycle, got %+v", missing)
	}
}

func TestCycleStoreExists(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewCycleStore(db)

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	created, _ := s.Create(start, end)

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected cycle to exist")
	}

	exists, err = s.Exists(99999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected missing cycle to not exist")
	}
}

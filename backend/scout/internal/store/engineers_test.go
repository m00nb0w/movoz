package store

import (
	"testing"
	"time"
)

func TestEngineerStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	role := "Backend Engineer"
	gh := "octocat"
	created, err := s.Create("Alex Kim", &role, &gh, nil, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Name != "Alex Kim" || *created.Role != "Backend Engineer" || !created.IsActive {
		t.Fatalf("unexpected created engineer: %+v", created)
	}

	list, err := s.List(true)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 || list[0].Name != "Alex Kim" {
		t.Fatalf("expected one engineer named Alex Kim, got %+v", list)
	}
}

func TestEngineerStoreUpdate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	created, _ := s.Create("Alex Kim", nil, nil, nil, time.Now())

	newRole := "Senior Backend Engineer"
	updated, err := s.Update(created.ID, "Alexandra Kim", &newRole, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Alexandra Kim" || *updated.Role != "Senior Backend Engineer" {
		t.Fatalf("unexpected updated engineer: %+v", updated)
	}

	missing, err := s.Update(99999, "Nobody", nil, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("Update for missing id returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing engineer, got %+v", missing)
	}
}

func TestEngineerStoreDeactivate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	created, _ := s.Create("Sam Lee", nil, nil, nil, time.Now())

	ok, err := s.Deactivate(created.ID)
	if err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}
	if !ok {
		t.Fatal("expected Deactivate to report success")
	}

	active, err := s.List(true)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("expected no active engineers, got %+v", active)
	}

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected deactivated engineer row to still exist (soft delete, not hard delete)")
	}

	all, err := s.List(false)
	if err != nil {
		t.Fatalf("List(false) failed: %v", err)
	}
	found := false
	for _, e := range all {
		if e.ID == created.ID {
			found = true
			if e.IsActive {
				t.Fatalf("expected deactivated engineer to have IsActive=false, got %+v", e)
			}
		}
	}
	if !found {
		t.Fatalf("expected deactivated engineer to still be present in List(false), got %+v", all)
	}

	ok, err = s.Deactivate(99999)
	if err != nil {
		t.Fatalf("Deactivate for missing id returned error: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing engineer")
	}
}

func TestEngineerStoreReactivate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	created, _ := s.Create("Alex Kim", nil, nil, nil, time.Now())
	s.Deactivate(created.ID)

	ok, err := s.Reactivate(created.ID)
	if err != nil {
		t.Fatalf("Reactivate failed: %v", err)
	}
	if !ok {
		t.Fatal("expected Reactivate to report success")
	}

	active, err := s.List(true)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(active) != 1 || active[0].ID != created.ID {
		t.Fatalf("expected reactivated engineer back in active list, got %+v", active)
	}

	ok, err = s.Reactivate(99999)
	if err != nil {
		t.Fatalf("Reactivate for missing id returned error: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing engineer")
	}
}

func TestEngineerStoreGetByID(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	active, _ := s.Create("Alex Kim", nil, nil, nil, time.Now())
	inactive, _ := s.Create("Sam Lee", nil, nil, nil, time.Now())
	s.Deactivate(inactive.ID)

	found, err := s.GetByID(active.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil || found.ID != active.ID || found.Name != "Alex Kim" {
		t.Fatalf("expected to find active engineer, got %+v", found)
	}

	foundInactive, err := s.GetByID(inactive.ID)
	if err != nil {
		t.Fatalf("GetByID for inactive engineer returned error: %v", err)
	}
	if foundInactive == nil {
		t.Fatal("expected GetByID to find deactivated engineer, got nil")
	}
	if foundInactive.IsActive {
		t.Fatalf("expected deactivated engineer to have IsActive=false, got %+v", foundInactive)
	}

	missing, err := s.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID for missing id returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing engineer, got %+v", missing)
	}
}

func TestEngineerStoreExists(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewEngineerStore(db)

	created, _ := s.Create("Alex Kim", nil, nil, nil, time.Now())

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected engineer to exist")
	}

	exists, err = s.Exists(99999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected missing engineer to not exist")
	}
}

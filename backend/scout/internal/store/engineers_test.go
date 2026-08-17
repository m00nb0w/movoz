package store

import (
	"testing"
	"time"
)

func TestEngineerStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewEngineerStore(db)

	role := "Backend Engineer"
	gh := "octocat"
	created, err := s.Create("Alex Kim", &role, &gh, nil, time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Name != "Alex Kim" || !created.IsActive {
		t.Fatalf("unexpected created engineer: %+v", created)
	}

	list, err := s.List(true)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one active engineer, got %+v", list)
	}
}

func TestEngineerStoreDeactivateExcludesFromActiveList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := NewEngineerStore(db)

	created, _ := s.Create("Sam Lee", nil, nil, nil, time.Now())
	ok, err := s.Deactivate(created.ID)
	if err != nil || !ok {
		t.Fatalf("deactivate failed: ok=%v err=%v", ok, err)
	}

	activeIDs, err := s.ListActiveIDs()
	if err != nil {
		t.Fatalf("list active ids failed: %v", err)
	}
	for _, id := range activeIDs {
		if id == created.ID {
			t.Fatalf("deactivated engineer %d still in active list", id)
		}
	}
}

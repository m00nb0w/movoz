package store

import "testing"

func TestPlayerStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewPlayerStore(db)

	position := "Forward"
	created, err := s.Create("Alex", &position)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Name != "Alex" || *created.Position != "Forward" || !created.IsActive {
		t.Fatalf("unexpected player: %+v", created)
	}

	players, err := s.List(true)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(players) != 1 || players[0].Name != "Alex" {
		t.Fatalf("expected one player named Alex, got %+v", players)
	}
}

func TestPlayerStoreUpdate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewPlayerStore(db)

	created, _ := s.Create("Alex", nil)

	newPosition := "Midfielder"
	updated, err := s.Update(created.ID, "Alexandra", &newPosition)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Alexandra" || *updated.Position != "Midfielder" {
		t.Fatalf("unexpected updated player: %+v", updated)
	}

	missing, err := s.Update(99999, "Nobody", nil)
	if err != nil {
		t.Fatalf("Update for missing id returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing player, got %+v", missing)
	}
}

func TestPlayerStoreDeactivate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewPlayerStore(db)

	created, _ := s.Create("Alex", nil)

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
		t.Fatalf("expected no active players, got %+v", active)
	}

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected deactivated player row to still exist (soft delete, not hard delete)")
	}

	all, err := s.List(false)
	if err != nil {
		t.Fatalf("List(false) failed: %v", err)
	}
	found := false
	for _, p := range all {
		if p.ID == created.ID {
			found = true
			if p.IsActive {
				t.Fatalf("expected deactivated player to have IsActive=false, got %+v", p)
			}
		}
	}
	if !found {
		t.Fatalf("expected deactivated player to still be present in List(false), got %+v", all)
	}

	ok, err = s.Deactivate(99999)
	if err != nil {
		t.Fatalf("Deactivate for missing id returned error: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing player")
	}
}

func TestPlayerStoreExists(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewPlayerStore(db)

	created, _ := s.Create("Alex", nil)

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected player to exist")
	}

	exists, err = s.Exists(99999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected missing player to not exist")
	}
}

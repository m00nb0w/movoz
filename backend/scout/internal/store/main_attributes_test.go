package store

import (
	"database/sql"
	"testing"
)

func truncateMainAttributes(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("TRUNCATE main_attributes RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate main_attributes: %v", err)
	}
	// Reseed the initial 6 attributes after truncation
	seedMainAttributes(t, db)
}

func seedMainAttributes(t *testing.T, db *sql.DB) {
	t.Helper()
	seedData := []struct {
		key  string
		name string
	}{
		{"technical_expertise", "Technical Expertise"},
		{"critical_thinking", "Critical Thinking"},
		{"communication", "Communication"},
		{"management", "Management"},
		{"product_mindset", "Product Mindset"},
		{"force_multiplier", "Force Multiplier"},
	}

	for _, seed := range seedData {
		_, err := db.Exec("INSERT INTO main_attributes (key, name) VALUES ($1, $2) ON CONFLICT DO NOTHING", seed.key, seed.name)
		if err != nil {
			t.Fatalf("failed to seed main_attributes: %v", err)
		}
	}
}

func TestMainAttributeStoreSeedData(t *testing.T) {
	db := setupTestDB(t)
	// Note: DO NOT truncate here - we're testing that seed data exists
	s := NewMainAttributeStore(db)

	list, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) < 6 {
		t.Fatalf("expected at least 6 seeded main attributes, got %d", len(list))
	}
	found := map[string]bool{}
	for _, a := range list {
		found[a.Key] = true
	}
	for _, key := range []string{"technical_expertise", "critical_thinking", "communication", "management", "product_mindset", "force_multiplier"} {
		if !found[key] {
			t.Fatalf("expected seeded main attribute %q, not found in %+v", key, list)
		}
	}
}

func TestMainAttributeStoreCreate(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	created, err := s.Create("delivery_speed", "Delivery Speed")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Key != "delivery_speed" || created.Name != "Delivery Speed" {
		t.Fatalf("unexpected created main attribute: %+v", created)
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestMainAttributeStoreCreateDuplicateKey(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	_, err := s.Create("unique_key", "First Attribute")
	if err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	_, err = s.Create("unique_key", "Duplicate Attribute")
	if err == nil {
		t.Fatal("expected error when creating attribute with duplicate key")
	}
}

func TestMainAttributeStoreList(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	a1, _ := s.Create("attr1", "Attribute 1")
	a2, _ := s.Create("attr2", "Attribute 2")
	a3, _ := s.Create("attr3", "Attribute 3")

	list, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) < 3 {
		t.Fatalf("expected at least 3 attributes, got %d", len(list))
	}

	found := map[int]string{}
	for _, attr := range list {
		found[attr.ID] = attr.Key
	}
	if found[a1.ID] != "attr1" || found[a2.ID] != "attr2" || found[a3.ID] != "attr3" {
		t.Fatalf("expected all 3 attributes in list, got %+v", list)
	}
}

func TestMainAttributeStoreGetByID(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	created, _ := s.Create("test_attr", "Test Attribute")

	found, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil || found.ID != created.ID || found.Key != "test_attr" {
		t.Fatalf("expected to find created attribute, got %+v", found)
	}
	if found.Name != "Test Attribute" {
		t.Fatalf("expected Name to be 'Test Attribute', got %q", found.Name)
	}
}

func TestMainAttributeStoreGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	found, err := s.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID returned error for missing id: %v", err)
	}
	if found != nil {
		t.Fatalf("expected nil for missing attribute, got %+v", found)
	}
}

func TestMainAttributeStoreUpdate(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	created, _ := s.Create("attr_key", "Original Name")

	updated, err := s.Update(created.ID, "Updated Name")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated attribute, got nil")
	}
	if updated.ID != created.ID {
		t.Fatalf("expected same ID, got %d", updated.ID)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("expected Name to be 'Updated Name', got %q", updated.Name)
	}
	// Key should not change
	if updated.Key != "attr_key" {
		t.Fatalf("expected Key to remain unchanged, got %q", updated.Key)
	}

	// Verify by fetching again
	refetched, _ := s.GetByID(created.ID)
	if refetched.Name != "Updated Name" {
		t.Fatalf("expected refetched attribute to have updated name, got %q", refetched.Name)
	}
}

func TestMainAttributeStoreUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	updated, err := s.Update(99999, "New Name")
	if err != nil {
		t.Fatalf("Update for missing id returned error: %v", err)
	}
	if updated != nil {
		t.Fatalf("expected nil for missing attribute, got %+v", updated)
	}
}

func TestMainAttributeStoreExists(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	created, _ := s.Create("exist_test", "Existence Test")

	exists, err := s.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected attribute to exist")
	}
}

func TestMainAttributeStoreExistsNotFound(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	exists, err := s.Exists(99999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected missing attribute to not exist")
	}
}

func TestMainAttributeStoreMultipleOperations(t *testing.T) {
	db := setupTestDB(t)
	truncateMainAttributes(t, db)
	s := NewMainAttributeStore(db)

	// Create multiple attributes
	ids := []int{}
	for i := 1; i <= 5; i++ {
		key := "attr" + string(rune('0'+i))
		name := "Attribute " + string(rune('0'+i))
		attr, err := s.Create(key, name)
		if err != nil {
			t.Fatalf("Create %d failed: %v", i, err)
		}
		ids = append(ids, attr.ID)
	}

	// Update some
	for _, id := range ids[:3] {
		_, err := s.Update(id, "Updated Name")
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
	}

	// List and verify
	list, err := s.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) < 5 {
		t.Fatalf("expected at least 5 attributes, got %d", len(list))
	}

	// Verify exists for all
	for _, id := range ids {
		exists, err := s.Exists(id)
		if err != nil {
			t.Fatalf("Exists failed for id %d: %v", id, err)
		}
		if !exists {
			t.Fatalf("expected attribute id %d to exist", id)
		}
	}
}

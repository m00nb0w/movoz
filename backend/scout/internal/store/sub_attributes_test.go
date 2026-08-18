package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func truncateSubAttributes(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("TRUNCATE sub_attributes RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate sub_attributes: %v", err)
	}
}

func uniqueKey(base string) string {
	return base + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestSubAttributeStoreCreateAndListByMainAttribute(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	key := uniqueKey("test_main")
	main, err := mainStore.Create(key, "Test Main")
	if err != nil {
		t.Fatalf("create main attribute failed: %v", err)
	}

	desc := "Writes clean, well-tested code"
	created, err := subStore.Create(main.ID, "Code Quality", &desc)
	if err != nil {
		t.Fatalf("create sub attribute failed: %v", err)
	}
	if created.MainAttributeID != main.ID || created.Name != "Code Quality" {
		t.Fatalf("unexpected created sub attribute: %+v", created)
	}

	list, err := subStore.ListByMainAttribute(main.ID, true)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one sub attribute under main %d, got %+v", main.ID, list)
	}
}

func TestSubAttributeStoreDeactivateExcludesFromActiveList(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main2"), "Test Main 2")
	created, _ := subStore.Create(main.ID, "Ownership", nil)

	ok, err := subStore.Deactivate(created.ID)
	if err != nil || !ok {
		t.Fatalf("deactivate failed: ok=%v err=%v", ok, err)
	}

	active, err := subStore.ListAllActive()
	if err != nil {
		t.Fatalf("list all active failed: %v", err)
	}
	for _, s := range active {
		if s.ID == created.ID {
			t.Fatalf("deactivated sub attribute %d still in active list", s.ID)
		}
	}
}

func TestSubAttributeStoreCreate(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main3"), "Test Main 3")

	created, err := subStore.Create(main.ID, "Testability", nil)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.MainAttributeID != main.ID || created.Name != "Testability" {
		t.Fatalf("unexpected created sub attribute: %+v", created)
	}
	if created.Description != nil {
		t.Fatalf("expected nil description, got %q", *created.Description)
	}
	if !created.IsActive {
		t.Fatal("expected IsActive to be true for new sub attribute")
	}
	if created.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestSubAttributeStoreCreateWithDescription(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main4"), "Test Main 4")
	desc := "Excellent ability"

	created, err := subStore.Create(main.ID, "Skill", &desc)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.Description == nil || *created.Description != desc {
		t.Fatalf("expected description %q, got %v", desc, created.Description)
	}
}

func TestSubAttributeStoreGetByID(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main5"), "Test Main 5")
	created, _ := subStore.Create(main.ID, "Reliability", nil)

	found, err := subStore.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if found == nil || found.ID != created.ID {
		t.Fatalf("expected to find created sub attribute, got %v", found)
	}
	if found.Name != "Reliability" || found.MainAttributeID != main.ID {
		t.Fatalf("unexpected sub attribute: %+v", found)
	}
}

func TestSubAttributeStoreGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	subStore := NewSubAttributeStore(db)

	found, err := subStore.GetByID(99999)
	if err != nil {
		t.Fatalf("GetByID returned error for missing id: %v", err)
	}
	if found != nil {
		t.Fatalf("expected nil for missing sub attribute, got %+v", found)
	}
}

func TestSubAttributeStoreUpdate(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main6"), "Test Main 6")
	created, _ := subStore.Create(main.ID, "Original Name", nil)

	newDesc := "Updated description"
	updated, err := subStore.Update(created.ID, "Updated Name", &newDesc)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated sub attribute, got nil")
	}
	if updated.ID != created.ID {
		t.Fatalf("expected same ID, got %d", updated.ID)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("expected Name to be 'Updated Name', got %q", updated.Name)
	}
	if updated.Description == nil || *updated.Description != newDesc {
		t.Fatalf("expected description to be %q, got %v", newDesc, updated.Description)
	}
	if updated.MainAttributeID != main.ID {
		t.Fatalf("expected MainAttributeID to remain %d, got %d", main.ID, updated.MainAttributeID)
	}

	// Verify by fetching again
	refetched, _ := subStore.GetByID(created.ID)
	if refetched.Name != "Updated Name" {
		t.Fatalf("expected refetched sub attribute to have updated name, got %q", refetched.Name)
	}
}

func TestSubAttributeStoreUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	subStore := NewSubAttributeStore(db)

	updated, err := subStore.Update(99999, "New Name", nil)
	if err != nil {
		t.Fatalf("Update for missing id returned error: %v", err)
	}
	if updated != nil {
		t.Fatalf("expected nil for missing sub attribute, got %+v", updated)
	}
}

func TestSubAttributeStoreListByMainAttribute(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main1, _ := mainStore.Create(uniqueKey("test_main7a"), "Test Main 7a")
	main2, _ := mainStore.Create(uniqueKey("test_main7b"), "Test Main 7b")

	// Create subs for main1
	sub1, _ := subStore.Create(main1.ID, "Sub1", nil)
	_, _ = subStore.Create(main1.ID, "Sub2", nil)
	_, _ = subStore.Create(main1.ID, "Sub3", nil)

	// Create sub for main2
	subStore.Create(main2.ID, "Sub for main2", nil)

	// List for main1
	list, err := subStore.ListByMainAttribute(main1.ID, true)
	if err != nil {
		t.Fatalf("ListByMainAttribute failed: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 subs under main1, got %d: %+v", len(list), list)
	}

	// Verify correct subs and order
	names := map[string]bool{}
	for _, s := range list {
		names[s.Name] = true
		if s.MainAttributeID != main1.ID {
			t.Fatalf("expected MainAttributeID=%d, got %d", main1.ID, s.MainAttributeID)
		}
	}
	if !names["Sub1"] || !names["Sub2"] || !names["Sub3"] {
		t.Fatalf("expected all three subs, got %v", names)
	}

	// Deactivate one and list again
	subStore.Deactivate(sub1.ID)
	activeList, _ := subStore.ListByMainAttribute(main1.ID, true)
	if len(activeList) != 2 {
		t.Fatalf("expected 2 active subs after deactivating one, got %d", len(activeList))
	}

	// List all (including inactive)
	allList, _ := subStore.ListByMainAttribute(main1.ID, false)
	if len(allList) != 3 {
		t.Fatalf("expected 3 total subs (including inactive), got %d", len(allList))
	}
}

func TestSubAttributeStoreListAllActive(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main1, _ := mainStore.Create(uniqueKey("test_main8a"), "Test Main 8a")
	main2, _ := mainStore.Create(uniqueKey("test_main8b"), "Test Main 8b")

	sub1, _ := subStore.Create(main1.ID, "Sub1", nil)
	sub2, _ := subStore.Create(main2.ID, "Sub2", nil)
	sub3, _ := subStore.Create(main1.ID, "Sub3", nil)

	// All should be active
	active, err := subStore.ListAllActive()
	if err != nil {
		t.Fatalf("ListAllActive failed: %v", err)
	}
	// Check our subs are in the list (may have other subs from other tests)
	ids := map[int]bool{}
	for _, s := range active {
		ids[s.ID] = true
	}
	if !ids[sub1.ID] || !ids[sub2.ID] || !ids[sub3.ID] {
		t.Fatalf("expected all our subs in active list, got %v", ids)
	}

	// Deactivate one
	subStore.Deactivate(sub2.ID)
	active, _ = subStore.ListAllActive()

	ids = map[int]bool{}
	for _, s := range active {
		ids[s.ID] = true
	}
	if !ids[sub1.ID] || !ids[sub3.ID] {
		t.Fatalf("expected sub1 and sub3 in active list after deactivating sub2, got %v", ids)
	}
	if ids[sub2.ID] {
		t.Fatalf("expected deactivated sub2 not in active list")
	}
}

func TestSubAttributeStoreDeactivate(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main9"), "Test Main 9")
	created, _ := subStore.Create(main.ID, "ToDeactivate", nil)

	ok, err := subStore.Deactivate(created.ID)
	if err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}
	if !ok {
		t.Fatal("expected deactivate to return true")
	}

	// Verify still exists but is_active=false
	fetched, _ := subStore.GetByID(created.ID)
	if fetched == nil {
		t.Fatal("expected deactivated sub attribute to still exist in database")
	}
	if fetched.IsActive {
		t.Fatal("expected IsActive to be false after deactivation")
	}
	if fetched.ID != created.ID || fetched.Name != "ToDeactivate" {
		t.Fatalf("expected same ID and name, got %+v", fetched)
	}
}

func TestSubAttributeStoreDeactivateNotFound(t *testing.T) {
	db := setupTestDB(t)
	subStore := NewSubAttributeStore(db)

	ok, err := subStore.Deactivate(99999)
	if err != nil {
		t.Fatalf("Deactivate for missing id returned error: %v", err)
	}
	if ok {
		t.Fatal("expected deactivate to return false for missing id")
	}
}

func TestSubAttributeStoreExists(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_main10"), "Test Main 10")
	created, _ := subStore.Create(main.ID, "ExistenceTest", nil)

	exists, err := subStore.Exists(created.ID)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("expected sub attribute to exist")
	}
}

func TestSubAttributeStoreExistsNotFound(t *testing.T) {
	db := setupTestDB(t)
	subStore := NewSubAttributeStore(db)

	exists, err := subStore.Exists(99999)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("expected missing sub attribute to not exist")
	}
}

func TestSubAttributeStoreDeactivateExplicitlyVerify(t *testing.T) {
	db := setupTestDB(t)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)

	main, _ := mainStore.Create(uniqueKey("test_explicit"), "Test Main Explicit")
	created, _ := subStore.Create(main.ID, "ToVerify", nil)

	// Deactivate
	subStore.Deactivate(created.ID)

	// Verify the row still exists with is_active=false
	fetched, err := subStore.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected row to still exist after soft-delete (deactivation)")
	}
	if fetched.IsActive {
		t.Fatalf("expected is_active=false, got is_active=%v", fetched.IsActive)
	}
	if fetched.ID != created.ID {
		t.Fatalf("expected same ID, got %d", fetched.ID)
	}

	// Verify excluded from active-only list
	activeList, _ := subStore.ListByMainAttribute(main.ID, true)
	for _, s := range activeList {
		if s.ID == created.ID {
			t.Fatalf("deactivated sub attribute should not appear in active list")
		}
	}

	// Verify appears in all list
	allList, _ := subStore.ListByMainAttribute(main.ID, false)
	found := false
	for _, s := range allList {
		if s.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("deactivated sub attribute should appear in all list (activeOnly=false)")
	}
}

package store

import "testing"

func TestHighlightStoreCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	highlightStore := NewHighlightStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, timeNow())

	created, err := highlightStore.Create(e1.ID, "highlight", "Shipped the auth migration solo, ahead of schedule.")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Kind != "highlight" {
		t.Fatalf("unexpected kind: %s", created.Kind)
	}

	list, err := highlightStore.List(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != created.ID {
		t.Fatalf("expected one highlight entry, got %+v", list)
	}
}

func TestHighlightStoreCreateLowlight(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	highlightStore := NewHighlightStore(db)

	e1, _ := engineerStore.Create("Bob", nil, nil, nil, timeNow())

	created, err := highlightStore.Create(e1.ID, "lowlight", "Missed the sprint deadline without flagging it early.")
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.Kind != "lowlight" {
		t.Fatalf("unexpected kind: %s", created.Kind)
	}
	if created.Body != "Missed the sprint deadline without flagging it early." {
		t.Fatalf("unexpected body: %s", created.Body)
	}
}

func TestHighlightStoreListEmpty(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	highlightStore := NewHighlightStore(db)

	e1, _ := engineerStore.Create("Charlie", nil, nil, nil, timeNow())

	list, err := highlightStore.List(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %+v", list)
	}
}

func TestHighlightStoreListOrdering(t *testing.T) {
	db := setupTestDB(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := NewEngineerStore(db)
	highlightStore := NewHighlightStore(db)

	e1, _ := engineerStore.Create("Dana", nil, nil, nil, timeNow())

	e1_h1, _ := highlightStore.Create(e1.ID, "highlight", "First entry")
	e1_h2, _ := highlightStore.Create(e1.ID, "lowlight", "Second entry")

	list, err := highlightStore.List(e1.ID)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(list))
	}
	// Should be ordered by created_at DESC (newest first)
	if list[0].ID != e1_h2.ID || list[1].ID != e1_h1.ID {
		t.Fatalf("expected descending order, got %d then %d", list[0].ID, list[1].ID)
	}
}

package store

import (
	"testing"

	"el-storko/internal/models"
)

func TestWorkItemStoreCreateEpic(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	created, err := s.Create(models.WorkItem{
		Type:  models.TypeEpic,
		Title: "Home projects",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Status != models.StatusTodo {
		t.Fatalf("expected default status todo, got %s", created.Status)
	}
	if created.Source != models.SourcePersonal {
		t.Fatalf("expected default source personal, got %s", created.Source)
	}
	if created.ParentID != nil {
		t.Fatalf("expected epic to have no parent, got %+v", created.ParentID)
	}
}

func TestWorkItemStoreCreateTaskWithEpicParent(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	epic, err := s.Create(models.WorkItem{Type: models.TypeEpic, Title: "Home projects"})
	if err != nil {
		t.Fatalf("Create epic failed: %v", err)
	}

	task, err := s.Create(models.WorkItem{
		Type:     models.TypeTask,
		Title:    "Buy paint",
		ParentID: &epic.ID,
	})
	if err != nil {
		t.Fatalf("Create task failed: %v", err)
	}
	if task.ParentID == nil || *task.ParentID != epic.ID {
		t.Fatalf("expected task parent to be epic %d, got %+v", epic.ID, task.ParentID)
	}
}

func TestWorkItemStoreCreateEpicWithParentRejected(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	epic, _ := s.Create(models.WorkItem{Type: models.TypeEpic, Title: "Home projects"})

	_, err := s.Create(models.WorkItem{Type: models.TypeEpic, Title: "Nested epic", ParentID: &epic.ID})
	if err == nil {
		t.Fatal("expected error creating epic with a parent")
	}
}

func TestWorkItemStoreCreateTaskParentMustBeEpic(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	otherTask, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Standalone task"})
	if err != nil {
		t.Fatalf("Create task failed: %v", err)
	}

	_, err = s.Create(models.WorkItem{
		Type:     models.TypeTask,
		Title:    "Nested under task",
		ParentID: &otherTask.ID,
	})
	if err == nil {
		t.Fatal("expected error creating a task whose parent is not an epic")
	}
}

func TestWorkItemStoreListFilters(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	epic, _ := s.Create(models.WorkItem{Type: models.TypeEpic, Title: "Home projects"})
	task1, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", ParentID: &epic.ID})
	_, _ = s.Create(models.WorkItem{Type: models.TypeTask, Title: "Unrelated task"})

	byType, err := s.List(ListFilters{Type: ptr(models.TypeEpic)})
	if err != nil {
		t.Fatalf("List by type failed: %v", err)
	}
	if len(byType) != 1 || byType[0].ID != epic.ID {
		t.Fatalf("expected only the epic, got %+v", byType)
	}

	byParent, err := s.List(ListFilters{ParentID: &epic.ID})
	if err != nil {
		t.Fatalf("List by parent failed: %v", err)
	}
	if len(byParent) != 1 || byParent[0].ID != task1.ID {
		t.Fatalf("expected only task1, got %+v", byParent)
	}

	bySource, err := s.List(ListFilters{Source: ptr(models.SourcePersonal)})
	if err != nil {
		t.Fatalf("List by source failed: %v", err)
	}
	if len(bySource) != 3 {
		t.Fatalf("expected all 3 items to be personal source, got %d", len(bySource))
	}

	byStatus, err := s.List(ListFilters{Status: ptr(models.StatusTodo)})
	if err != nil {
		t.Fatalf("List by status failed: %v", err)
	}
	if len(byStatus) != 3 {
		t.Fatalf("expected all 3 items to default to todo, got %d", len(byStatus))
	}
}

func TestWorkItemStoreUpdate(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint"})

	newStatus := models.StatusInProgress
	newDescription := "at the hardware store"
	updated, err := s.Update(task.ID, UpdateFields{
		Status:      &newStatus,
		Description: &newDescription,
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Status != models.StatusInProgress || updated.Description != "at the hardware store" {
		t.Fatalf("unexpected updated task: %+v", updated)
	}
	if !updated.UpdatedAt.After(task.UpdatedAt) && updated.UpdatedAt != task.UpdatedAt {
		t.Fatalf("expected updated_at to move forward, got %v vs %v", updated.UpdatedAt, task.UpdatedAt)
	}
}

func TestWorkItemStoreUpdateMissing(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	newStatus := models.StatusDone
	updated, err := s.Update(99999, UpdateFields{Status: &newStatus})
	if err != nil {
		t.Fatalf("Update for missing id returned error: %v", err)
	}
	if updated != nil {
		t.Fatalf("expected nil for missing item, got %+v", updated)
	}
}

func TestWorkItemStoreDeleteEpicUnparentsTasks(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	epic, _ := s.Create(models.WorkItem{Type: models.TypeEpic, Title: "Home projects"})
	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", ParentID: &epic.ID})

	ok, err := s.Delete(epic.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !ok {
		t.Fatal("expected Delete to report success")
	}

	remaining, err := s.Get(task.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if remaining == nil {
		t.Fatal("expected task to still exist after its epic was deleted")
	}
	if remaining.ParentID != nil {
		t.Fatalf("expected task to be unparented, got parent_id=%v", *remaining.ParentID)
	}

	deletedEpic, err := s.Get(epic.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if deletedEpic != nil {
		t.Fatal("expected epic to be gone")
	}
}

func TestWorkItemStoreDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint"})

	ok, err := s.Delete(task.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !ok {
		t.Fatal("expected Delete to report success")
	}

	ok, err = s.Delete(99999)
	if err != nil {
		t.Fatalf("Delete for missing id returned error: %v", err)
	}
	if ok {
		t.Fatal("expected false for missing item")
	}
}

func TestWorkItemStoreCreateWithDoneStatusSetsCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	created, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", Status: models.StatusDone})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.CompletedAt == nil {
		t.Fatal("expected completed_at to be set when created with status=done")
	}
}

func TestWorkItemStoreCreateWithoutDoneStatusLeavesCompletedAtNil(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	created, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created.CompletedAt != nil {
		t.Fatalf("expected completed_at to be nil, got %v", *created.CompletedAt)
	}
}

func TestWorkItemStoreUpdateToDoneSetsCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint"})

	doneStatus := models.StatusDone
	updated, err := s.Update(task.ID, UpdateFields{Status: &doneStatus})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected completed_at to be set after moving to done")
	}
}

func TestWorkItemStoreEditingDoneItemDoesNotShiftCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", Status: models.StatusDone})
	firstCompletedAt := *task.CompletedAt

	newDescription := "actually two coats"
	updated, err := s.Update(task.ID, UpdateFields{Description: &newDescription})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.CompletedAt == nil || !updated.CompletedAt.Equal(firstCompletedAt) {
		t.Fatalf("expected completed_at to stay %v, got %v", firstCompletedAt, updated.CompletedAt)
	}
}

func TestWorkItemStoreMovingAwayFromDoneClearsCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", Status: models.StatusDone})

	reopenedStatus := models.StatusInProgress
	updated, err := s.Update(task.ID, UpdateFields{Status: &reopenedStatus})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.CompletedAt != nil {
		t.Fatalf("expected completed_at to be cleared, got %v", *updated.CompletedAt)
	}
}

func TestWorkItemStoreRecompletingSetsNewCompletedAt(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)
	s := NewWorkItemStore(db)

	task, _ := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", Status: models.StatusDone})
	firstCompletedAt := *task.CompletedAt

	reopenedStatus := models.StatusInProgress
	_, err := s.Update(task.ID, UpdateFields{Status: &reopenedStatus})
	if err != nil {
		t.Fatalf("Update (reopen) failed: %v", err)
	}

	doneAgain := models.StatusDone
	updated, err := s.Update(task.ID, UpdateFields{Status: &doneAgain})
	if err != nil {
		t.Fatalf("Update (recomplete) failed: %v", err)
	}
	if updated.CompletedAt == nil {
		t.Fatal("expected completed_at to be set again after recompleting")
	}
	if !updated.CompletedAt.After(firstCompletedAt) && !updated.CompletedAt.Equal(firstCompletedAt) {
		t.Fatalf("expected new completed_at %v to be >= first %v", *updated.CompletedAt, firstCompletedAt)
	}
}

func ptr[T any](v T) *T {
	return &v
}

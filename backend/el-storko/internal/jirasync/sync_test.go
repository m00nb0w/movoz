package jirasync

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"el-storko/internal/models"
	"el-storko/internal/store"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/el_storko_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available at %s: %v", url, err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err := db.Exec("TRUNCATE work_items RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	return db
}

type mockJiraClient struct {
	issues        []JiraIssue
	searchErr     error
	updatedIssues map[string]JiraIssue
	updateErr     error
}

func (m *mockJiraClient) SearchAssignedIssues() ([]JiraIssue, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	return m.issues, nil
}

func (m *mockJiraClient) UpdateIssue(key, title, description string, status models.Status) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.updatedIssues == nil {
		m.updatedIssues = map[string]JiraIssue{}
	}
	m.updatedIssues[key] = JiraIssue{Key: key, Title: title, Description: description, Status: status}
	return nil
}

func TestRunSyncCyclePullsNewAssignedIssue(t *testing.T) {
	db := setupTestDB(t)
	s := store.NewWorkItemStore(db)

	client := &mockJiraClient{
		issues: []JiraIssue{
			{
				Key:         "TCAT-1",
				Title:       "Fix the thing",
				Description: "details",
				Status:      models.StatusInProgress,
				URL:         "https://example.atlassian.net/browse/TCAT-1",
				UpdatedAt:   time.Now(),
			},
		},
	}

	if err := RunSyncCycle(s, client); err != nil {
		t.Fatalf("RunSyncCycle failed: %v", err)
	}

	item, err := s.GetByJiraKey("TCAT-1")
	if err != nil {
		t.Fatalf("GetByJiraKey failed: %v", err)
	}
	if item == nil {
		t.Fatal("expected a local item to be created for the new Jira issue")
	}
	if item.Source != models.SourceJira {
		t.Fatalf("expected source=jira, got %s", item.Source)
	}
	if item.JiraKey == nil || *item.JiraKey != "TCAT-1" {
		t.Fatalf("expected jira_key TCAT-1, got %+v", item.JiraKey)
	}
	if item.JiraURL == nil || *item.JiraURL != "https://example.atlassian.net/browse/TCAT-1" {
		t.Fatalf("expected jira_url set, got %+v", item.JiraURL)
	}
}

func TestRunSyncCycleOverwritesOlderLocalWithNewerJiraData(t *testing.T) {
	db := setupTestDB(t)
	s := store.NewWorkItemStore(db)

	local, err := s.CreateFromJira("Old title", "old description", models.StatusTodo, "TCAT-2", "https://example.atlassian.net/browse/TCAT-2")
	if err != nil {
		t.Fatalf("CreateFromJira failed: %v", err)
	}

	client := &mockJiraClient{
		issues: []JiraIssue{
			{
				Key:         "TCAT-2",
				Title:       "New title from Jira",
				Description: "new description",
				Status:      models.StatusDone,
				URL:         "https://example.atlassian.net/browse/TCAT-2",
				UpdatedAt:   local.UpdatedAt.Add(1 * time.Hour),
			},
		},
	}

	if err := RunSyncCycle(s, client); err != nil {
		t.Fatalf("RunSyncCycle failed: %v", err)
	}

	updated, err := s.GetByJiraKey("TCAT-2")
	if err != nil {
		t.Fatalf("GetByJiraKey failed: %v", err)
	}
	if updated.Title != "New title from Jira" || updated.Status != models.StatusDone {
		t.Fatalf("expected local item overwritten by newer Jira data, got %+v", updated)
	}
}

func TestRunSyncCyclePushesNewerLocalChangeToJira(t *testing.T) {
	db := setupTestDB(t)
	s := store.NewWorkItemStore(db)

	local, err := s.CreateFromJira("Original", "orig desc", models.StatusTodo, "TCAT-3", "https://example.atlassian.net/browse/TCAT-3")
	if err != nil {
		t.Fatalf("CreateFromJira failed: %v", err)
	}

	newStatus := models.StatusInProgress
	updated, err := s.Update(local.ID, store.UpdateFields{Status: &newStatus})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	client := &mockJiraClient{
		issues: []JiraIssue{
			{
				Key:         "TCAT-3",
				Title:       "Original",
				Description: "orig desc",
				Status:      models.StatusTodo,
				URL:         "https://example.atlassian.net/browse/TCAT-3",
				UpdatedAt:   updated.UpdatedAt.Add(-1 * time.Hour),
			},
		},
	}

	if err := RunSyncCycle(s, client); err != nil {
		t.Fatalf("RunSyncCycle failed: %v", err)
	}

	pushed, ok := client.updatedIssues["TCAT-3"]
	if !ok {
		t.Fatal("expected the newer local change to be pushed to Jira")
	}
	if pushed.Status != models.StatusInProgress {
		t.Fatalf("expected pushed status in_progress, got %s", pushed.Status)
	}
}

func TestRunSyncCycleDoesNotPushWhenJiraIsNewer(t *testing.T) {
	db := setupTestDB(t)
	s := store.NewWorkItemStore(db)

	local, err := s.CreateFromJira("Original", "orig desc", models.StatusTodo, "TCAT-4", "https://example.atlassian.net/browse/TCAT-4")
	if err != nil {
		t.Fatalf("CreateFromJira failed: %v", err)
	}

	client := &mockJiraClient{
		issues: []JiraIssue{
			{
				Key:         "TCAT-4",
				Title:       "Original",
				Description: "orig desc",
				Status:      models.StatusTodo,
				URL:         "https://example.atlassian.net/browse/TCAT-4",
				UpdatedAt:   local.UpdatedAt.Add(1 * time.Hour),
			},
		},
	}

	if err := RunSyncCycle(s, client); err != nil {
		t.Fatalf("RunSyncCycle failed: %v", err)
	}

	if _, ok := client.updatedIssues["TCAT-4"]; ok {
		t.Fatal("expected no push when Jira's data is newer than local")
	}
}

func TestRunSyncCycleFailureIsolation(t *testing.T) {
	db := setupTestDB(t)
	s := store.NewWorkItemStore(db)

	client := &mockJiraClient{searchErr: errors.New("network error: connection refused")}

	err := RunSyncCycle(s, client)
	if err == nil {
		t.Fatal("expected RunSyncCycle to return an error when the Jira client fails")
	}

	created, createErr := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Still works"})
	if createErr != nil {
		t.Fatalf("expected personal item creation to still succeed after a Jira sync failure, got: %v", createErr)
	}
	if created == nil || created.ID == 0 {
		t.Fatal("expected a valid created personal item")
	}

	items, listErr := s.List(store.ListFilters{})
	if listErr != nil {
		t.Fatalf("expected list to still succeed after a Jira sync failure, got: %v", listErr)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

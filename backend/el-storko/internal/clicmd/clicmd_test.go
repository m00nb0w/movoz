package clicmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"el-storko/internal/models"
)

func TestAddCreatesPersonalTask(t *testing.T) {
	var receivedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/work-items" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.WorkItem{
			ID:    1,
			Type:  models.TypeTask,
			Title: receivedBody["title"].(string),
		})
	}))
	defer server.Close()

	epicID := int64(3)
	item, err := Add(server.URL, "Buy birthday gift", &epicID, "a nice one")
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if item.Title != "Buy birthday gift" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if receivedBody["type"] != "task" {
		t.Fatalf("expected type=task, got %v", receivedBody["type"])
	}
	if receivedBody["parent_id"].(float64) != 3 {
		t.Fatalf("expected parent_id=3, got %v", receivedBody["parent_id"])
	}
	if receivedBody["description"] != "a nice one" {
		t.Fatalf("expected description set, got %v", receivedBody["description"])
	}
}

func TestListReturnsItemsWithFilters(t *testing.T) {
	var receivedQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/work-items" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		receivedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"items": []models.WorkItem{
				{ID: 1, Type: models.TypeTask, Title: "Buy paint", Status: models.StatusTodo},
			},
		})
	}))
	defer server.Close()

	items, err := List(server.URL, ListOptions{Status: "todo"})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Buy paint" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if receivedQuery != "status=todo" {
		t.Fatalf("expected status filter in query, got %q", receivedQuery)
	}
}

func TestAddPropagatesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid work item data"})
	}))
	defer server.Close()

	_, err := Add(server.URL, "", nil, "")
	if err == nil {
		t.Fatal("expected an error for a 400 response")
	}
}

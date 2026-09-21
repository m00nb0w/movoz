package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"el-storko/internal/models"
	"el-storko/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *store.WorkItemStore) {
	t.Helper()
	gin.SetMode(gin.TestMode)

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

	s := store.NewWorkItemStore(db)
	h := NewWorkItemHandler(s)

	r := gin.New()
	r.GET("/api/work-items", h.List)
	r.POST("/api/work-items", h.Create)
	r.GET("/api/work-items/:id", h.Get)
	r.PATCH("/api/work-items/:id", h.Update)
	r.DELETE("/api/work-items/:id", h.Delete)

	return r, s
}

func doRequest(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWorkItemHandlerCreateAndGet(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{
		"type":  "epic",
		"title": "Home projects",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.WorkItem
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if created.ID == 0 || created.Type != models.TypeEpic {
		t.Fatalf("unexpected created item: %+v", created)
	}

	w2 := doRequest(r, http.MethodGet, "/api/work-items/1", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestWorkItemHandlerGetNotFound(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/api/work-items/99999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWorkItemHandlerCreateEpicWithParentRejected(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{
		"type":  "epic",
		"title": "Home projects",
	})
	var epic models.WorkItem
	json.Unmarshal(w.Body.Bytes(), &epic)

	w2 := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{
		"type":      "epic",
		"title":     "Nested epic",
		"parent_id": epic.ID,
	})
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestWorkItemHandlerCreateTaskParentMustBeEpic(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{
		"type":  "task",
		"title": "Standalone task",
	})
	var task models.WorkItem
	json.Unmarshal(w.Body.Bytes(), &task)

	w2 := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{
		"type":      "task",
		"title":     "Nested under task",
		"parent_id": task.ID,
	})
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestWorkItemHandlerList(t *testing.T) {
	r, _ := setupTestRouter(t)

	doRequest(r, http.MethodPost, "/api/work-items", map[string]any{"type": "epic", "title": "Epic A"})
	doRequest(r, http.MethodPost, "/api/work-items", map[string]any{"type": "task", "title": "Task A"})

	w := doRequest(r, http.MethodGet, "/api/work-items", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Items []models.WorkItem `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Items))
	}

	wFiltered := doRequest(r, http.MethodGet, "/api/work-items?type=epic", nil)
	var respFiltered struct {
		Items []models.WorkItem `json:"items"`
	}
	json.Unmarshal(wFiltered.Body.Bytes(), &respFiltered)
	if len(respFiltered.Items) != 1 || respFiltered.Items[0].Type != models.TypeEpic {
		t.Fatalf("expected 1 epic, got %+v", respFiltered.Items)
	}
}

func TestWorkItemHandlerUpdate(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{"type": "task", "title": "Task A"})
	var task models.WorkItem
	json.Unmarshal(w.Body.Bytes(), &task)

	w2 := doRequest(r, http.MethodPatch, "/api/work-items/1", map[string]any{"status": "done"})
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	var updated models.WorkItem
	json.Unmarshal(w2.Body.Bytes(), &updated)
	if updated.Status != models.StatusDone {
		t.Fatalf("expected status done, got %s", updated.Status)
	}
}

func TestWorkItemHandlerUpdateNotFound(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPatch, "/api/work-items/99999", map[string]any{"status": "done"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestWorkItemHandlerDelete(t *testing.T) {
	r, _ := setupTestRouter(t)

	w := doRequest(r, http.MethodPost, "/api/work-items", map[string]any{"type": "task", "title": "Task A"})
	var task models.WorkItem
	json.Unmarshal(w.Body.Bytes(), &task)

	w2 := doRequest(r, http.MethodDelete, "/api/work-items/1", nil)
	if w2.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w2.Code, w2.Body.String())
	}

	w3 := doRequest(r, http.MethodDelete, "/api/work-items/99999", nil)
	if w3.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing item, got %d: %s", w3.Code, w3.Body.String())
	}
}

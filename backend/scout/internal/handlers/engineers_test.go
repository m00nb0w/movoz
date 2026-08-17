package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupEngineerTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/scout_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	if _, err := db.Exec("TRUNCATE engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	engineerStore := store.NewEngineerStore(db)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewEngineerHandler(engineerStore)
	r.GET("/api/engineers", h.List)
	r.GET("/api/engineers/:id", h.Get)
	r.POST("/api/engineers", h.Create)
	r.PUT("/api/engineers/:id", h.Update)
	r.DELETE("/api/engineers/:id", h.Deactivate)
	r.POST("/api/engineers/:id/reactivate", h.Reactivate)
	return r
}

func TestEngineerHandlerCreateAndList(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Alex Kim", "started_at": "2024-01-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/engineers", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var engineers []models.Engineer
	json.Unmarshal(w.Body.Bytes(), &engineers)
	if len(engineers) != 1 || engineers[0].Name != "Alex Kim" {
		t.Fatalf("expected one engineer Alex Kim, got %+v", engineers)
	}
}

func TestEngineerHandlerCreateMissingName(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"started_at": "2024-01-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestEngineerHandlerDeactivateNotFound(t *testing.T) {
	r := setupEngineerTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/engineers/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestEngineerHandlerGet(t *testing.T) {
	r := setupEngineerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Alex Kim", "started_at": "2024-01-15"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/engineers", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	var created models.Engineer
	json.Unmarshal(createW.Body.Bytes(), &created)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(created.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var fetched models.Engineer
	json.Unmarshal(w.Body.Bytes(), &fetched)
	if fetched.ID != created.ID || fetched.Name != "Alex Kim" {
		t.Fatalf("unexpected fetched engineer: %+v", fetched)
	}
}

func TestEngineerHandlerGetNotFound(t *testing.T) {
	r := setupEngineerTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

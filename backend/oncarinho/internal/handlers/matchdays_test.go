package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupMatchdayTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/oncarinho_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	if _, err := db.Exec("TRUNCATE matchdays RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	matchdayStore := store.NewMatchdayStore(db)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMatchdayHandler(matchdayStore)
	r.GET("/api/matchdays", h.List)
	r.POST("/api/matchdays", h.Create)

	return r
}

func TestMatchdayHandlerCreateAndList(t *testing.T) {
	r := setupMatchdayTestRouter(t)

	body, _ := json.Marshal(map[string]string{"played_on": "2026-03-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/matchdays", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/matchdays", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var matchdays []struct {
		ID       int    `json:"id"`
		PlayedOn string `json:"played_on"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &matchdays); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(matchdays) != 1 {
		t.Fatalf("expected 1 matchday, got %d", len(matchdays))
	}
}

func TestMatchdayHandlerListDateOnlyFormat(t *testing.T) {
	r := setupMatchdayTestRouter(t)

	body, _ := json.Marshal(map[string]string{"played_on": "2026-03-15"})
	req := httptest.NewRequest(http.MethodPost, "/api/matchdays", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/matchdays", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var matchdays []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &matchdays); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(matchdays) != 1 {
		t.Fatalf("expected 1 matchday, got %d", len(matchdays))
	}

	playedOn, ok := matchdays[0]["played_on"].(string)
	if !ok {
		t.Fatalf("expected played_on to be a string, got %T: %v", matchdays[0]["played_on"], matchdays[0]["played_on"])
	}
	if playedOn != "2026-03-15" {
		t.Fatalf("expected played_on %q, got %q", "2026-03-15", playedOn)
	}
	if len(playedOn) != 10 {
		t.Fatalf("expected played_on to be exactly 10 characters (date-only), got %d: %q", len(playedOn), playedOn)
	}
}

func TestMatchdayHandlerCreateInvalidDate(t *testing.T) {
	r := setupMatchdayTestRouter(t)

	body, _ := json.Marshal(map[string]string{"played_on": "not-a-date"})
	req := httptest.NewRequest(http.MethodPost, "/api/matchdays", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

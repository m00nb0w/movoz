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

	"oncarinho/internal/models"
	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupPlayerTestRouter(t *testing.T) (*gin.Engine, *store.PlayerStore) {
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
	if _, err := db.Exec("TRUNCATE players RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	playerStore := store.NewPlayerStore(db)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPlayerHandler(playerStore)
	r.GET("/api/players", h.List)
	r.POST("/api/players", h.Create)
	r.PUT("/api/players/:id", h.Update)
	r.DELETE("/api/players/:id", h.Deactivate)

	return r, playerStore
}

func TestPlayerHandlerCreateAndList(t *testing.T) {
	r, _ := setupPlayerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Alex"})
	req := httptest.NewRequest(http.MethodPost, "/api/players", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/players", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var players []models.Player
	if err := json.Unmarshal(w.Body.Bytes(), &players); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(players) != 1 || players[0].Name != "Alex" {
		t.Fatalf("expected one player Alex, got %+v", players)
	}
}

func TestPlayerHandlerCreateMissingName(t *testing.T) {
	r, _ := setupPlayerTestRouter(t)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/api/players", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPlayerHandlerUpdateNotFound(t *testing.T) {
	r, _ := setupPlayerTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Nobody"})
	req := httptest.NewRequest(http.MethodPut, "/api/players/99999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestPlayerHandlerDeactivate(t *testing.T) {
	r, playerStore := setupPlayerTestRouter(t)

	created, err := playerStore.Create("Alex", nil)
	if err != nil {
		t.Fatalf("failed to seed player: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/players/"+strconv.Itoa(created.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestPlayerHandlerDeactivateNotFound(t *testing.T) {
	r, _ := setupPlayerTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/players/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

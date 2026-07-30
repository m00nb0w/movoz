package handlers

import (
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

func setupProfileTestRouter(t *testing.T) (*gin.Engine, *store.PlayerStore) {
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
	if _, err := db.Exec("TRUNCATE match_stats, matchdays, players RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	playerStore := store.NewPlayerStore(db)
	profileStore := store.NewProfileStore(db, playerStore)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProfileHandler(profileStore)
	r.GET("/api/players/:id", h.Get)

	return r, playerStore
}

func TestProfileHandlerGet(t *testing.T) {
	r, playerStore := setupProfileTestRouter(t)
	player, _ := playerStore.Create("Alex", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/players/"+strconv.Itoa(player.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var profile models.PlayerProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	if profile.Player.Name != "Alex" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestProfileHandlerGetDeactivatedPlayer(t *testing.T) {
	r, playerStore := setupProfileTestRouter(t)
	player, _ := playerStore.Create("Alex", nil)

	if ok, err := playerStore.Deactivate(player.ID); err != nil || !ok {
		t.Fatalf("Deactivate failed: ok=%v err=%v", ok, err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/players/"+strconv.Itoa(player.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for deactivated player, got %d: %s", w.Code, w.Body.String())
	}

	var profile models.PlayerProfile
	json.Unmarshal(w.Body.Bytes(), &profile)
	if profile.Player.Name != "Alex" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if profile.Player.IsActive {
		t.Fatalf("expected player to be inactive, got %+v", profile.Player)
	}
}

func TestProfileHandlerNotFound(t *testing.T) {
	r, _ := setupProfileTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/players/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

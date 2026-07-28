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
	"time"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("invalid date %q: %v", s, err)
	}
	return d
}

func setupStatTestRouter(t *testing.T) (*gin.Engine, *store.PlayerStore, *store.MatchdayStore) {
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
	matchdayStore := store.NewMatchdayStore(db)
	statStore := store.NewStatStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStatHandler(statStore, matchdayStore, playerStore)
	r.PUT("/api/matchdays/:id/stats", h.UpsertStats)

	return r, playerStore, matchdayStore
}

func TestStatHandlerUpsertStats(t *testing.T) {
	r, playerStore, matchdayStore := setupStatTestRouter(t)

	player, _ := playerStore.Create("Alex", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))

	body, _ := json.Marshal(map[string]interface{}{
		"entries": []map[string]int{
			{"player_id": player.ID, "goals": 2, "assists": 1},
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/matchdays/"+strconv.Itoa(matchday.ID)+"/stats", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStatHandlerUnknownMatchday(t *testing.T) {
	r, playerStore, _ := setupStatTestRouter(t)
	player, _ := playerStore.Create("Alex", nil)

	body, _ := json.Marshal(map[string]interface{}{
		"entries": []map[string]int{{"player_id": player.ID, "goals": 1}},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/matchdays/99999/stats", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestStatHandlerUnknownPlayer(t *testing.T) {
	r, _, matchdayStore := setupStatTestRouter(t)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))

	body, _ := json.Marshal(map[string]interface{}{
		"entries": []map[string]int{{"player_id": 99999, "goals": 1}},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/matchdays/"+strconv.Itoa(matchday.ID)+"/stats", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

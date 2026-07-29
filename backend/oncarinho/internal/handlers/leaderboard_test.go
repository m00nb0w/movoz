package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"oncarinho/internal/models"
	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupLeaderboardTestRouter(t *testing.T) *gin.Engine {
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
	leaderboardStore := store.NewLeaderboardStore(db)

	player, _ := playerStore.Create("Alex", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	statStore.UpsertBulk(matchday.ID, []models.StatInput{{PlayerID: player.ID, Goals: 4}})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewLeaderboardHandler(leaderboardStore)
	r.GET("/api/leaderboard", h.Get)

	return r
}

func TestLeaderboardHandlerGet(t *testing.T) {
	r := setupLeaderboardTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard?year=all&stat=goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var entries []models.LeaderboardEntry
	json.Unmarshal(w.Body.Bytes(), &entries)
	if len(entries) != 1 || entries[0].Value != 4 {
		t.Fatalf("unexpected leaderboard: %+v", entries)
	}
}

func TestLeaderboardHandlerInvalidStat(t *testing.T) {
	r := setupLeaderboardTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard?stat=nonsense", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

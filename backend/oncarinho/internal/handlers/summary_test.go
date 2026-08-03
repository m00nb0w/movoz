package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"oncarinho/internal/models"
	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func TestSummaryHandlerGet(t *testing.T) {
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
	summaryStore := store.NewSummaryStore(db)

	player, _ := playerStore.Create("Alex", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	statStore.UpsertBulk(matchday.ID, []models.StatInput{{PlayerID: player.ID, Goals: 3}})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSummaryHandler(summaryStore)
	r.GET("/api/summary", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/summary?year=2026", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var summary models.Summary
	json.Unmarshal(w.Body.Bytes(), &summary)
	if summary.MatchesPlayed != 1 || summary.GoalsScored != 3 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestSummaryHandlerGetDefaultsToCurrentYear(t *testing.T) {
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

	summaryStore := store.NewSummaryStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSummaryHandler(summaryStore)
	r.GET("/api/summary", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var summary models.Summary
	json.Unmarshal(w.Body.Bytes(), &summary)
	if summary.Year != time.Now().Year() {
		t.Fatalf("expected default year %d, got %d", time.Now().Year(), summary.Year)
	}
}

func TestSummaryHandlerGetInvalidYear(t *testing.T) {
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
	t.Cleanup(func() { db.Close() })

	summaryStore := store.NewSummaryStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSummaryHandler(summaryStore)
	r.GET("/api/summary", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/summary?year=not-a-year", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

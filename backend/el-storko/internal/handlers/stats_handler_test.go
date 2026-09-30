package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"el-storko/internal/models"
	"el-storko/internal/store"

	"github.com/gin-gonic/gin"
)

func setupStatsTestRouter(t *testing.T) (*gin.Engine, *store.WorkItemStore, *sql.DB) {
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
	h := NewStatsHandler(s)

	r := gin.New()
	r.GET("/api/stats/burn-rate", h.BurnRate)
	r.GET("/api/stats/summary", h.Summary)

	return r, s, db
}

type burnRateResponse struct {
	Days   int `json:"days"`
	Points []struct {
		Date      string `json:"date"`
		Completed int    `json:"completed"`
		Open      int    `json:"open"`
	} `json:"points"`
}

func TestStatsHandlerBurnRateDefaultWindow(t *testing.T) {
	router, _, _ := setupStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats/burn-rate", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp burnRateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Days != 30 {
		t.Fatalf("expected default days=30, got %d", resp.Days)
	}
	if len(resp.Points) != 30 {
		t.Fatalf("expected 30 points, got %d", len(resp.Points))
	}
}

func TestStatsHandlerBurnRateCustomWindow(t *testing.T) {
	router, _, _ := setupStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats/burn-rate?days=7", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp burnRateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Days != 7 || len(resp.Points) != 7 {
		t.Fatalf("expected 7-day window, got days=%d points=%d", resp.Days, len(resp.Points))
	}
}

func TestStatsHandlerBurnRateEmptyTrackerIsAllZero(t *testing.T) {
	router, _, _ := setupStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats/burn-rate?days=3", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp burnRateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, p := range resp.Points {
		if p.Completed != 0 || p.Open != 0 {
			t.Fatalf("expected all-zero points for empty tracker, got %+v", p)
		}
	}
}

func TestStatsHandlerBurnRateReflectsCompletedItem(t *testing.T) {
	router, s, _ := setupStatsTestRouter(t)

	_, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Buy paint", Status: models.StatusDone})
	if err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stats/burn-rate?days=1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp burnRateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Points) != 1 || resp.Points[0].Completed != 1 {
		t.Fatalf("expected today's point to show 1 completed item, got %+v", resp.Points)
	}
}

type summaryResponse struct {
	OpenItems         int            `json:"open_items"`
	CompletedThisWeek int            `json:"completed_this_week"`
	CompletionRate    float64        `json:"completion_rate"`
	TotalTracked      int            `json:"total_tracked"`
	StatusBreakdown   map[string]int `json:"status_breakdown"`
}

func TestStatsHandlerSummaryEmptyTrackerIsAllZero(t *testing.T) {
	router, _, _ := setupStatsTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp summaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.OpenItems != 0 || resp.CompletedThisWeek != 0 || resp.CompletionRate != 0 || resp.TotalTracked != 0 {
		t.Fatalf("expected all-zero summary for empty tracker, got %+v", resp)
	}
	wantStatuses := []string{"backlog", "picked_for_today", "in_progress", "blocked", "done"}
	if len(resp.StatusBreakdown) != len(wantStatuses) {
		t.Fatalf("expected all five status keys present, got %+v", resp.StatusBreakdown)
	}
	for _, status := range wantStatuses {
		if count, ok := resp.StatusBreakdown[status]; !ok || count != 0 {
			t.Fatalf("expected status %q present with count 0, got %+v", status, resp.StatusBreakdown)
		}
	}
}

func TestStatsHandlerSummaryReflectsKnownComposition(t *testing.T) {
	router, s, db := setupStatsTestRouter(t)

	// Two open items.
	if _, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Backlog item", Status: models.StatusBacklog}); err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}
	if _, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "In progress item", Status: models.StatusInProgress}); err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}

	// A done item completed just now — inside the trailing 7 days.
	recentDone, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Recently done", Status: models.StatusDone})
	if err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}
	_ = recentDone

	// A done item completed 30 days ago — outside the trailing 7 days, but
	// still counted in total_tracked/completion_rate.
	oldDone, err := s.Create(models.WorkItem{Type: models.TypeTask, Title: "Long done", Status: models.StatusDone})
	if err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}
	if _, err := db.Exec("UPDATE work_items SET completed_at = now() - interval '30 days' WHERE id = $1", oldDone.ID); err != nil {
		t.Fatalf("failed to backdate completed_at: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/stats/summary", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp summaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.OpenItems != 2 {
		t.Errorf("expected open_items=2, got %d", resp.OpenItems)
	}
	if resp.CompletedThisWeek != 1 {
		t.Errorf("expected completed_this_week=1, got %d", resp.CompletedThisWeek)
	}
	if resp.TotalTracked != 4 {
		t.Errorf("expected total_tracked=4, got %d", resp.TotalTracked)
	}
	wantRate := 2.0 / 4.0
	if resp.CompletionRate != wantRate {
		t.Errorf("expected completion_rate=%v, got %v", wantRate, resp.CompletionRate)
	}
	wantBreakdown := map[string]int{
		"backlog":          1,
		"picked_for_today": 0,
		"in_progress":      1,
		"blocked":          0,
		"done":             2,
	}
	for status, count := range wantBreakdown {
		if got := resp.StatusBreakdown[status]; got != count {
			t.Errorf("status %q: expected %d, got %d", status, count, got)
		}
	}
}

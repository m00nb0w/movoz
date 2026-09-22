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

func setupStatsTestRouter(t *testing.T) (*gin.Engine, *store.WorkItemStore) {
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

	return r, s
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
	router, _ := setupStatsTestRouter(t)

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
	router, _ := setupStatsTestRouter(t)

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
	router, _ := setupStatsTestRouter(t)

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
	router, s := setupStatsTestRouter(t)

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

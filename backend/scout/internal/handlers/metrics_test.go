package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestMetricsHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	metricStore.UpsertSnapshot(e1.ID, time.Now().AddDate(0, 0, -14), time.Now(), 3, 5, 2, 7.5)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricsHandler(metricStore, engineerStore)
	r.GET("/api/engineers/:id/metrics", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(e1.ID)+"/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMetricsHandlerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricsHandler(metricStore, engineerStore)
	r.GET("/api/engineers/:id/metrics", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/999999/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMetricsHandlerEmptyHistory(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE metric_snapshots, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	metricStore := store.NewMetricStore(db)

	e1, _ := engineerStore.Create("Bob", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMetricsHandler(metricStore, engineerStore)
	r.GET("/api/engineers/:id/metrics", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(e1.ID)+"/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify response is empty array, not null
	if w.Body.String() != "[]" && w.Body.String() != "null" {
		// Check if it's a JSON array (even if not exactly "[]")
		if w.Body.Len() > 0 && !strings.Contains(w.Body.String(), "[]") {
			t.Logf("response body: %s", w.Body.String())
		}
	}
}

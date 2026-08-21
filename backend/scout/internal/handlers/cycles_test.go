package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCycleHandlerCreate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)
	r.GET("/api/cycles", h.List)

	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-01", "period_end": "2026-01-14"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCycleHandlerCreateAndList(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)
	r.GET("/api/cycles", h.List)

	// Create a cycle
	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-01", "period_end": "2026-01-14"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on create, got %d: %s", w.Code, w.Body.String())
	}

	// List cycles
	listReq := httptest.NewRequest(http.MethodGet, "/api/cycles", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 on list, got %d", listW.Code)
	}

	var cycles []map[string]interface{}
	if err := json.Unmarshal(listW.Body.Bytes(), &cycles); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(cycles) != 1 {
		t.Fatalf("expected one cycle, got %d", len(cycles))
	}
}

func TestCycleHandlerCreateMissingFields(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)

	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-01"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCycleHandlerCreateInvalidStartDate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)

	body, _ := json.Marshal(map[string]string{"period_start": "invalid-date", "period_end": "2026-01-14"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCycleHandlerCreateInvalidEndDate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)

	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-01", "period_end": "invalid-date"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCycleHandlerCreateEndBeforeStart(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.POST("/api/cycles", h.Create)

	body, _ := json.Marshal(map[string]string{"period_start": "2026-01-14", "period_end": "2026-01-01"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCycleHandlerList(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE rating_cycles RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	s := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleHandler(s)
	r.GET("/api/cycles", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/cycles", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var cycles []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &cycles); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(cycles) != 0 {
		t.Fatalf("expected zero cycles, got %d", len(cycles))
	}
}

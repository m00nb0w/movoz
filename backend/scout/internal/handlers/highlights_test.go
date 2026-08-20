package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestHighlightHandlerCreateAndList(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.GET("/api/engineers/:id/highlights", h.List)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "lowlight", "body": "Missed the sprint deadline without flagging it early."})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// Test List
	req = httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestHighlightHandlerCreateRejectsInvalidKind(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "sidenote", "body": "not a valid kind"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHighlightHandlerCreateWithHighlight(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Taylor", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "highlight", "body": "Shipped the auth migration solo, ahead of schedule."})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &entry); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if entry["kind"] != "highlight" {
		t.Fatalf("expected kind=highlight, got %v", entry["kind"])
	}
}

func TestHighlightHandlerCreateEngineerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "highlight", "body": "Some entry"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/99999/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHighlightHandlerListEmpty(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Jordan", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.GET("/api/engineers/:id/highlights", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if entries == nil {
		// JSON unmarshaling of empty array returns nil, which is acceptable
		return
	}
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func TestHighlightHandlerListEngineerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.GET("/api/engineers/:id/highlights", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/99999/highlights", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHighlightHandlerCreateMissingKind(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Morgan", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"body": "Some entry"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+strconv.Itoa(e1.ID)+"/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHighlightHandlerCreateInvalidEngineerId(t *testing.T) {
	db := setupTestDBForHandlers(t)

	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHighlightHandler(highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights", h.Create)

	body, _ := json.Marshal(map[string]string{"kind": "highlight", "body": "Some entry"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/abc/highlights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

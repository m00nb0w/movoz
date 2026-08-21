package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func setupMainAttributeTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "main_attributes")

	mainAttributeStore := store.NewMainAttributeStore(db)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewMainAttributeHandler(mainAttributeStore)
	r.GET("/api/main-attributes", h.List)
	r.POST("/api/main-attributes", h.Create)
	r.PUT("/api/main-attributes/:id", h.Update)
	return r
}

func TestMainAttributeHandlerCreate(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	body, _ := json.Marshal(map[string]string{"key": "delivery_speed", "name": "Delivery Speed"})
	req := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.MainAttribute
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Key != "delivery_speed" || created.Name != "Delivery Speed" {
		t.Fatalf("expected created attribute with key=delivery_speed and name=Delivery Speed, got %+v", created)
	}
}

func TestMainAttributeHandlerCreateMissingKey(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	body, _ := json.Marshal(map[string]string{"name": "Delivery Speed"})
	req := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing key, got %d", w.Code)
	}
}

func TestMainAttributeHandlerCreateMissingName(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	body, _ := json.Marshal(map[string]string{"key": "delivery_speed"})
	req := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestMainAttributeHandlerList(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	// Create first attribute
	body, _ := json.Marshal(map[string]string{"key": "delivery_speed", "name": "Delivery Speed"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	// Create second attribute
	body, _ = json.Marshal(map[string]string{"key": "code_quality", "name": "Code Quality"})
	createReq = httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW = httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)

	// List attributes
	req := httptest.NewRequest(http.MethodGet, "/api/main-attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var attrs []models.MainAttribute
	json.Unmarshal(w.Body.Bytes(), &attrs)
	if len(attrs) != 2 {
		t.Fatalf("expected 2 attributes, got %d: %+v", len(attrs), attrs)
	}
	if attrs[0].Key != "delivery_speed" || attrs[1].Key != "code_quality" {
		t.Fatalf("unexpected attribute keys: %+v", attrs)
	}
}

func TestMainAttributeHandlerListEmpty(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/main-attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var attrs []models.MainAttribute
	json.Unmarshal(w.Body.Bytes(), &attrs)
	if attrs == nil || len(attrs) != 0 {
		t.Fatalf("expected empty list, got %+v", attrs)
	}
}

func TestMainAttributeHandlerUpdate(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	// Create attribute
	body, _ := json.Marshal(map[string]string{"key": "delivery_speed", "name": "Delivery Speed"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	var created models.MainAttribute
	json.Unmarshal(createW.Body.Bytes(), &created)

	// Update attribute
	updateBody, _ := json.Marshal(map[string]string{"name": "Delivery Velocity"})
	req := httptest.NewRequest(http.MethodPut, "/api/main-attributes/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.MainAttribute
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Name != "Delivery Velocity" {
		t.Fatalf("expected name to be updated to 'Delivery Velocity', got %s", updated.Name)
	}
	if updated.Key != "delivery_speed" {
		t.Fatalf("expected key to remain 'delivery_speed', got %s", updated.Key)
	}
}

func TestMainAttributeHandlerUpdateMissingName(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	// Create attribute
	body, _ := json.Marshal(map[string]string{"key": "delivery_speed", "name": "Delivery Speed"})
	createReq := httptest.NewRequest(http.MethodPost, "/api/main-attributes", bytes.NewReader(body))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	var created models.MainAttribute
	json.Unmarshal(createW.Body.Bytes(), &created)

	// Try to update with missing name
	updateBody, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPut, "/api/main-attributes/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestMainAttributeHandlerUpdateNotFound(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	updateBody, _ := json.Marshal(map[string]string{"name": "Delivery Velocity"})
	req := httptest.NewRequest(http.MethodPut, "/api/main-attributes/99999", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMainAttributeHandlerUpdateInvalidID(t *testing.T) {
	r := setupMainAttributeTestRouter(t)

	updateBody, _ := json.Marshal(map[string]string{"name": "Delivery Velocity"})
	req := httptest.NewRequest(http.MethodPut, "/api/main-attributes/invalid", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", w.Code)
	}
}

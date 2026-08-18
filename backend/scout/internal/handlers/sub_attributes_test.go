package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func uniqueMainKey(base string) string {
	return base + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestSubAttributeHandlerCreate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_create"), "Test Create")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"main_attribute_id": main.ID, "name": "Testability"})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.MainAttributeID != main.ID || created.Name != "Testability" {
		t.Fatalf("expected created sub attribute with main_attribute_id=%d and name=Testability, got %+v", main.ID, created)
	}
}

func TestSubAttributeHandlerCreateWithDescription(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_create_desc"), "Test Create Desc")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	desc := "Writes clean, well-tested code"
	body, _ := json.Marshal(map[string]interface{}{
		"main_attribute_id": main.ID,
		"name":              "Code Quality",
		"description":       desc,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var created models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &created)
	if created.Description == nil || *created.Description != desc {
		t.Fatalf("expected description %q, got %v", desc, created.Description)
	}
}

func TestSubAttributeHandlerCreateUnknownMainAttribute(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"main_attribute_id": 999999, "name": "Testability"})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSubAttributeHandlerCreateMissingName(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_missing_name"), "Test Missing Name")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"main_attribute_id": main.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestSubAttributeHandlerCreateMissingMainAttributeID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.POST("/api/sub-attributes", h.Create)

	body, _ := json.Marshal(map[string]interface{}{"name": "Testability"})
	req := httptest.NewRequest(http.MethodPost, "/api/sub-attributes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing main_attribute_id, got %d", w.Code)
	}
}

func TestSubAttributeHandlerList(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_list"), "Test List")

	// Create multiple sub attributes
	subA, _ := subStore.Create(main.ID, "SubA", nil)
	subB, _ := subStore.Create(main.ID, "SubB", nil)
	subC, _ := subStore.Create(main.ID, "SubC", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.GET("/api/sub-attributes", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/sub-attributes?main_attribute_id="+strconv.Itoa(main.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var list []models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 3 {
		t.Fatalf("expected 3 sub attributes, got %d: %+v", len(list), list)
	}

	// Verify the correct subs are present
	ids := map[int]bool{}
	for _, s := range list {
		ids[s.ID] = true
	}
	if !ids[subA.ID] || !ids[subB.ID] || !ids[subC.ID] {
		t.Fatalf("expected all three created subs in list")
	}
}

func TestSubAttributeHandlerListActiveOnly(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_active"), "Test Active")

	sub1, _ := subStore.Create(main.ID, "Active1", nil)
	sub2, _ := subStore.Create(main.ID, "Active2", nil)
	sub3, _ := subStore.Create(main.ID, "ToDeactivate", nil)

	// Deactivate one
	subStore.Deactivate(sub3.ID)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.GET("/api/sub-attributes", h.List)

	// List active only (default)
	req := httptest.NewRequest("GET", "/api/sub-attributes?main_attribute_id="+strconv.Itoa(main.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var list []models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("expected 2 active sub attributes, got %d", len(list))
	}

	ids := map[int]bool{}
	for _, s := range list {
		ids[s.ID] = true
	}
	if !ids[sub1.ID] || !ids[sub2.ID] {
		t.Fatalf("expected active subs in list")
	}
	if ids[sub3.ID] {
		t.Fatalf("expected deactivated sub not in active list")
	}
}

func TestSubAttributeHandlerListMissingMainAttributeID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.GET("/api/sub-attributes", h.List)

	req := httptest.NewRequest(http.MethodGet, "/api/sub-attributes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing main_attribute_id, got %d", w.Code)
	}
}

func TestSubAttributeHandlerUpdate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_update"), "Test Update")

	created, _ := subStore.Create(main.ID, "Original Name", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.PUT("/api/sub-attributes/:id", h.Update)

	updateBody, _ := json.Marshal(map[string]interface{}{"name": "Updated Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/sub-attributes/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Name != "Updated Name" {
		t.Fatalf("expected name to be updated to 'Updated Name', got %q", updated.Name)
	}
	if updated.MainAttributeID != main.ID {
		t.Fatalf("expected MainAttributeID to remain unchanged, got %d", updated.MainAttributeID)
	}
}

func TestSubAttributeHandlerUpdateWithDescription(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_update_desc"), "Test Update Desc")

	created, _ := subStore.Create(main.ID, "Test Sub", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.PUT("/api/sub-attributes/:id", h.Update)

	newDesc := "New description here"
	updateBody, _ := json.Marshal(map[string]interface{}{
		"name":        "Updated Name",
		"description": newDesc,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/sub-attributes/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var updated models.SubAttribute
	json.Unmarshal(w.Body.Bytes(), &updated)
	if updated.Description == nil || *updated.Description != newDesc {
		t.Fatalf("expected description to be %q, got %v", newDesc, updated.Description)
	}
}

func TestSubAttributeHandlerUpdateNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.PUT("/api/sub-attributes/:id", h.Update)

	updateBody, _ := json.Marshal(map[string]string{"name": "New Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/sub-attributes/99999", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSubAttributeHandlerUpdateMissingName(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_missing_update"), "Test Missing Update")
	created, _ := subStore.Create(main.ID, "Test Sub", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.PUT("/api/sub-attributes/:id", h.Update)

	updateBody, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPut, "/api/sub-attributes/"+strconv.Itoa(created.ID), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestSubAttributeHandlerUpdateInvalidID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.PUT("/api/sub-attributes/:id", h.Update)

	updateBody, _ := json.Marshal(map[string]string{"name": "New Name"})
	req := httptest.NewRequest(http.MethodPut, "/api/sub-attributes/invalid", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", w.Code)
	}
}

func TestSubAttributeHandlerDeactivate(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	// Create a unique main attribute for this test
	main, _ := mainStore.Create(uniqueMainKey("test_deactivate"), "Test Deactivate")
	created, _ := subStore.Create(main.ID, "ToDeactivate", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.DELETE("/api/sub-attributes/:id", h.Deactivate)

	req := httptest.NewRequest(http.MethodDelete, "/api/sub-attributes/"+strconv.Itoa(created.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify still exists but is_active=false
	fetched, _ := subStore.GetByID(created.ID)
	if fetched == nil {
		t.Fatal("expected deactivated sub attribute to still exist in database")
	}
	if fetched.IsActive {
		t.Fatal("expected IsActive to be false after deactivation")
	}
}

func TestSubAttributeHandlerDeactivateNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.DELETE("/api/sub-attributes/:id", h.Deactivate)

	req := httptest.NewRequest(http.MethodDelete, "/api/sub-attributes/99999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSubAttributeHandlerDeactivateInvalidID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewSubAttributeHandler(subStore, mainStore)
	r.DELETE("/api/sub-attributes/:id", h.Deactivate)

	req := httptest.NewRequest(http.MethodDelete, "/api/sub-attributes/invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid ID, got %d", w.Code)
	}
}

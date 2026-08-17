package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheck(t *testing.T) {
	handler := NewHealthHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health", nil)

	handler.HealthCheck(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	expected := `{"service":"scout","status":"healthy"}`
	if w.Body.String() != expected {
		t.Fatalf("expected body %s, got %s", expected, w.Body.String())
	}
}

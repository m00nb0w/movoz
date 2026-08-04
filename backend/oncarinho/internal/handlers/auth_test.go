package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"oncarinho/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestLoginCorrectPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler("correct-password", "session-secret")
	r.POST("/api/auth/login", h.Login)

	body, _ := json.Marshal(map[string]string{"password": "correct-password"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(w.Result().Cookies()) != 1 {
		t.Fatal("expected a session cookie to be set")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler("correct-password", "session-secret")
	r.POST("/api/auth/login", h.Login)

	body, _ := json.Marshal(map[string]string{"password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAdminRejectsMissingCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin-only", RequireAdmin("session-secret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAdminAllowsValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/admin-only", RequireAdmin("session-secret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	token := auth.NewSessionToken("session-secret", time.Now())
	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.AddCookie(&http.Cookie{Name: "admin_session", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

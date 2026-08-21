package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/auth"

	"github.com/gin-gonic/gin"
)

func setupAuthTestRouter() (*gin.Engine, *AuthHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAuthHandler("correct-password", "test-secret", false)
	r.POST("/api/auth/login", h.Login)
	protected := r.Group("/api")
	protected.Use(RequireAuth("test-secret"))
	protected.GET("/whoami", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r, h
}

func TestLoginSuccessSetsCookie(t *testing.T) {
	r, _ := setupAuthTestRouter()

	body, _ := json.Marshal(map[string]string{"password": "correct-password"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(w.Result().Cookies()) == 0 {
		t.Fatal("expected a session cookie to be set")
	}
}

func TestLoginWrongPasswordRejected(t *testing.T) {
	r, _ := setupAuthTestRouter()

	body, _ := json.Marshal(map[string]string{"password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginMissingPasswordRejected(t *testing.T) {
	r, _ := setupAuthTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProtectedRouteRequiresSession(t *testing.T) {
	r, _ := setupAuthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session cookie, got %d", w.Code)
	}
}

func TestProtectedRouteAllowsValidSession(t *testing.T) {
	r, _ := setupAuthTestRouter()

	loginBody, _ := json.Marshal(map[string]string{"password": "correct-password"})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	cookies := loginW.Result().Cookies()

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid session cookie, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProtectedRouteRejectsInvalidCookie(t *testing.T) {
	r, _ := setupAuthTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "garbage-not-a-real-token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with tampered cookie, got %d", w.Code)
	}
}

func TestProtectedRouteRejectsExpiredSession(t *testing.T) {
	r, _ := setupAuthTestRouter()

	// Mint a token that "expired" in the past relative to now.
	expiredToken := auth.NewSessionToken("test-secret", time.Now().Add(-2*auth.SessionDuration))

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: expiredToken})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with expired session, got %d", w.Code)
	}
}

func TestRequestNeverReachesHandlerWithoutAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handlerReached := false
	protected := r.Group("/api")
	protected.Use(RequireAuth("test-secret"))
	protected.GET("/whoami", func(c *gin.Context) {
		handlerReached = true
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/whoami", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if handlerReached {
		t.Fatal("expected handler to never be reached without a valid session")
	}
}

package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"scout/internal/auth"
	"scout/internal/config"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const routerTestSessionSecret = "test-secret"
const routerTestSessionCookieName = "scout_session"

func setupRouterIntegrationTest(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/scout_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	if _, err := db.Exec("TRUNCATE sub_attributes, main_attributes, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AdminPassword: "test-password", SessionSecret: routerTestSessionSecret}
	router := buildRouter(db, cfg)
	server := httptest.NewServer(router)

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	t.Cleanup(func() {
		server.Close()
		db.Close()
	})

	return server, client
}

func login(t *testing.T, server *httptest.Server, client *http.Client) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"password": "test-password"})
	resp, err := client.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected login 200, got %d", resp.StatusCode)
	}
}

// TestHealthAndLoginAreExemptFromAuth confirms the only two routes reachable
// without a session are /health and POST /api/auth/login, per NF1.
func TestHealthAndLoginAreExemptFromAuth(t *testing.T) {
	server, client := setupRouterIntegrationTest(t)

	resp, err := client.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected /health to be reachable without auth, got %d", resp.StatusCode)
	}

	body, _ := json.Marshal(map[string]string{"password": "wrong"})
	resp, err = client.Post(server.URL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	resp.Body.Close()
	// A wrong password should yield 401 (invalid credentials), never 404 —
	// proving the route itself is reachable without a session cookie.
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected login route reachable and to reject bad password with 401, got %d", resp.StatusCode)
	}
}

// TestEveryApplicationRouteRequiresAuth walks every route wired in
// buildRouter (engineers, main-attributes, sub-attributes) and confirms none
// of them are reachable without a valid session — Scout has no public
// application-data route group, unlike oncarinho (NF1).
func TestEveryApplicationRouteRequiresAuth(t *testing.T) {
	server, client := setupRouterIntegrationTest(t)

	type protectedRoute struct {
		method string
		path   string
		body   string // re-wrapped into a fresh io.Reader per request
	}

	protectedRoutes := []protectedRoute{
		// engineers
		{http.MethodGet, "/api/engineers", ""},
		{http.MethodGet, "/api/engineers/1", ""},
		{http.MethodPost, "/api/engineers", `{"name":"Alex","started_at":"2024-01-15"}`},
		{http.MethodPut, "/api/engineers/1", `{"name":"Alex","started_at":"2024-01-15"}`},
		{http.MethodDelete, "/api/engineers/1", ""},
		{http.MethodPost, "/api/engineers/1/reactivate", ""},

		// main-attributes
		{http.MethodGet, "/api/main-attributes", ""},
		{http.MethodPost, "/api/main-attributes", `{"key":"k","name":"n"}`},
		{http.MethodPut, "/api/main-attributes/1", `{"key":"k","name":"n"}`},

		// sub-attributes
		{http.MethodGet, "/api/sub-attributes", ""},
		{http.MethodPost, "/api/sub-attributes", `{"main_attribute_id":1,"name":"n"}`},
		{http.MethodPut, "/api/sub-attributes/1", `{"name":"n"}`},
		{http.MethodDelete, "/api/sub-attributes/1", ""},
	}

	bodyReader := func(body string) io.Reader {
		if body == "" {
			return nil
		}
		return strings.NewReader(body)
	}

	for _, route := range protectedRoutes {
		route := route
		t.Run(fmt.Sprintf("%s %s without session", route.method, route.path), func(t *testing.T) {
			req, err := http.NewRequest(route.method, server.URL+route.path, bodyReader(route.body))
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
			}
		})
	}

	// Same requests again, but now with a valid session cookie: they should
	// no longer be rejected as unauthorized (they may still 400/404/etc. on
	// their own validation, but never 401).
	login(t, server, client)

	for _, route := range protectedRoutes {
		route := route
		t.Run(fmt.Sprintf("%s %s with session", route.method, route.path), func(t *testing.T) {
			req, err := http.NewRequest(route.method, server.URL+route.path, bodyReader(route.body))
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusUnauthorized {
				t.Fatalf("expected route to be reachable with a valid session, got 401")
			}
		})
	}
}

// TestInvalidSessionCookiesRejectedAcrossRouteGroups proves the
// tampered-cookie and expired-session rejection paths exercised in isolation
// by internal/handlers/auth_test.go (against a synthetic /api/whoami route)
// also hold against the real route table — one representative GET route from
// each of the three route groups (engineers, main-attributes,
// sub-attributes).
func TestInvalidSessionCookiesRejectedAcrossRouteGroups(t *testing.T) {
	server, _ := setupRouterIntegrationTest(t)

	expiredToken := auth.NewSessionToken(routerTestSessionSecret, time.Now().Add(-2*auth.SessionDuration))

	representativeRoutes := []string{
		"/api/engineers",
		"/api/main-attributes",
		"/api/sub-attributes",
	}

	badCookies := []struct {
		name  string
		value string
	}{
		{"tampered cookie", "garbage-not-a-real-token"},
		{"expired session", expiredToken},
	}

	// Deliberately no cookie jar here — each request sets its cookie value
	// explicitly rather than relying on one previously issued by /login.
	client := &http.Client{}

	for _, path := range representativeRoutes {
		for _, bad := range badCookies {
			t.Run(fmt.Sprintf("%s rejected on %s", bad.name, path), func(t *testing.T) {
				req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
				if err != nil {
					t.Fatalf("failed to build request: %v", err)
				}
				req.AddCookie(&http.Cookie{Name: routerTestSessionCookieName, Value: bad.value})

				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("request failed: %v", err)
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusUnauthorized {
					t.Fatalf("expected 401 for %s against %s, got %d", bad.name, path, resp.StatusCode)
				}
			})
		}
	}
}

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

	"oncarinho/internal/config"
	"oncarinho/internal/models"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func setupIntegrationTest(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/oncarinho_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	if _, err := db.Exec("TRUNCATE match_stats, matchdays, players RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}

	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AdminPassword: "test-password", SessionSecret: "test-secret"}
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

func adminLogin(t *testing.T, server *httptest.Server, client *http.Client) {
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

func TestFullMatchdayFlow(t *testing.T) {
	server, client := setupIntegrationTest(t)

	unauthedBody, _ := json.Marshal(map[string]string{"name": "Alex"})
	resp, err := client.Post(server.URL+"/api/players", "application/json", bytes.NewReader(unauthedBody))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	adminLogin(t, server, client)

	resp, err = client.Post(server.URL+"/api/players", "application/json", bytes.NewReader(unauthedBody))
	if err != nil {
		t.Fatalf("create player failed: %v", err)
	}
	var player models.Player
	json.NewDecoder(resp.Body).Decode(&player)
	resp.Body.Close()
	if player.ID == 0 {
		t.Fatalf("expected created player to have an id: %+v", player)
	}

	matchdayBody, _ := json.Marshal(map[string]string{"played_on": "2026-03-15"})
	resp, err = client.Post(server.URL+"/api/matchdays", "application/json", bytes.NewReader(matchdayBody))
	if err != nil {
		t.Fatalf("create matchday failed: %v", err)
	}
	var matchday models.Matchday
	if err := json.NewDecoder(resp.Body).Decode(&matchday); err != nil {
		t.Fatalf("failed to decode matchday: %v", err)
	}
	resp.Body.Close()

	statsBody, _ := json.Marshal(map[string]interface{}{
		"entries": []map[string]int{
			{"player_id": player.ID, "goals": 2, "assists": 1},
		},
	})
	req, _ := http.NewRequest(http.MethodPut, server.URL+fmt.Sprintf("/api/matchdays/%d/stats", matchday.ID), bytes.NewReader(statsBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("upsert stats failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, err = client.Get(server.URL + "/api/leaderboard?year=2026&stat=goals")
	if err != nil {
		t.Fatalf("leaderboard request failed: %v", err)
	}
	var entries []models.LeaderboardEntry
	json.NewDecoder(resp.Body).Decode(&entries)
	resp.Body.Close()
	if len(entries) != 1 || entries[0].Value != 2 {
		t.Fatalf("expected leaderboard value 2, got %+v", entries)
	}

	resp, err = client.Get(server.URL + fmt.Sprintf("/api/players/%d", player.ID))
	if err != nil {
		t.Fatalf("profile request failed: %v", err)
	}
	var profile models.PlayerProfile
	json.NewDecoder(resp.Body).Decode(&profile)
	resp.Body.Close()
	if profile.AllTime.Goals != 2 || profile.AllTime.MatchesPlayed != 1 {
		t.Fatalf("unexpected profile: %+v", profile)
	}

	resp, err = client.Get(server.URL + "/api/summary?year=2026")
	if err != nil {
		t.Fatalf("summary request failed: %v", err)
	}
	var summary models.Summary
	json.NewDecoder(resp.Body).Decode(&summary)
	resp.Body.Close()
	if summary.MatchesPlayed != 1 || summary.GoalsScored != 2 || summary.RosterSize != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestAdminRoutesRequireAuth(t *testing.T) {
	server, client := setupIntegrationTest(t)

	adminRoutes := []struct {
		method string
		path   string
		body   io.Reader
	}{
		{http.MethodPost, "/api/players", strings.NewReader(`{"name":"Alex"}`)},
		{http.MethodPut, "/api/players/1", strings.NewReader(`{"name":"Alex"}`)},
		{http.MethodDelete, "/api/players/1", nil},
		{http.MethodPost, "/api/players/1/reactivate", nil},
		{http.MethodPost, "/api/matchdays", strings.NewReader(`{"played_on":"2026-03-15"}`)},
		{http.MethodPut, "/api/matchdays/1/stats", strings.NewReader(`{"entries":[]}`)},
		{http.MethodDelete, "/api/matchdays/1/stats/1", nil},
	}

	for _, route := range adminRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			req, err := http.NewRequest(route.method, server.URL+route.path, route.body)
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

	t.Run("GET /api/matchdays/1/stats is public", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/matchdays/1/stats")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			t.Fatalf("expected public route to not require auth, got 401")
		}
	})
}

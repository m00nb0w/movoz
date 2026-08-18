package integrations

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGitHubClientFetchPRStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/search/issues" {
			// Raised-PR search includes "author:"; reviewed-PR search
			// includes "reviewed-by:" — respond with a distinct count for each.
			if containsSubstring(q, "author:") {
				json.NewEncoder(w).Encode(map[string]int{"total_count": 3})
				return
			}
			json.NewEncoder(w).Encode(map[string]int{"total_count": 5})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	prsRaised, prsReviewed, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}
	if prsRaised != 3 || prsReviewed != 5 {
		t.Fatalf("expected raised=3 reviewed=5, got raised=%d reviewed=%d", prsRaised, prsReviewed)
	}
}

func TestGitHubClientFetchPRStatsSendsBearerAuthAndAcceptHeaders(t *testing.T) {
	var gotAuth, gotAccept string
	var reqCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_count": 1})
	}))
	defer server.Close()

	client := NewGitHubClient("secret-token", server.Client())
	client.baseURL = server.URL

	if _, _, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}

	if reqCount == 0 {
		t.Fatalf("expected server to receive at least one request")
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("expected Authorization header %q, got %q", "Bearer secret-token", gotAuth)
	}
	if gotAccept != "application/vnd.github+json" {
		t.Fatalf("expected Accept header %q, got %q", "application/vnd.github+json", gotAccept)
	}
}

func TestGitHubClientFetchPRStatsOmitsAuthHeaderWhenTokenEmpty(t *testing.T) {
	var gotAuth string
	var authHeaderSet bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if values := r.Header["Authorization"]; len(values) > 0 {
			authHeaderSet = true
			gotAuth = values[0]
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_count": 0})
	}))
	defer server.Close()

	client := NewGitHubClient("", server.Client())
	client.baseURL = server.URL

	if _, _, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}

	if authHeaderSet {
		t.Fatalf("expected no Authorization header to be sent when token is empty, got %q", gotAuth)
	}
}

func TestGitHubClientFetchPRStatsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden) // e.g. GitHub rate-limit response
		w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	prsRaised, prsReviewed, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for HTTP 403 response, got nil (raised=%d reviewed=%d)", prsRaised, prsReviewed)
	}
	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected error to mention status code 403, got: %v", err)
	}
}

func TestGitHubClientFetchPRStatsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	_, _, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for HTTP 500 response, got nil")
	}
}

func TestGitHubClientFetchPRStatsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json{{{`))
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	_, _, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for malformed JSON response, got nil")
	}
}

func TestGitHubClientFetchPRStatsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_count": 0})
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	prsRaised, prsReviewed, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}
	if prsRaised != 0 || prsReviewed != 0 {
		t.Fatalf("expected raised=0 reviewed=0, got raised=%d reviewed=%d", prsRaised, prsReviewed)
	}
}

func TestGitHubClientFetchPRStatsEmptyReposSkipsRequest(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_count": 99})
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	prsRaised, prsReviewed, err := client.FetchPRStats(
		context.Background(), "octocat", nil,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch pr stats failed: %v", err)
	}
	if prsRaised != 0 || prsReviewed != 0 {
		t.Fatalf("expected raised=0 reviewed=0 for empty repos list, got raised=%d reviewed=%d", prsRaised, prsReviewed)
	}
	if called {
		t.Fatalf("expected no HTTP request to be made when repos list is empty")
	}
}

func TestGitHubClientFetchPRStatsNetworkFailure(t *testing.T) {
	// Point the client at an address nothing is listening on so the HTTP
	// call itself fails (connection refused), rather than returning a
	// non-200 status. This should surface as an error, not a panic.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	unreachableURL := server.URL
	server.Close() // closing immediately frees the port but leaves nothing listening

	client := NewGitHubClient("fake-token", http.DefaultClient)
	client.baseURL = unreachableURL

	_, _, err := client.FetchPRStats(
		context.Background(), "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for unreachable server, got nil")
	}
}

func TestGitHubClientFetchPRStatsRespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"total_count": 1})
	}))
	defer server.Close()

	client := NewGitHubClient("fake-token", server.Client())
	client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.FetchPRStats(
		ctx, "octocat", []string{"org/repo-a"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for already-canceled context, got nil")
	}
}

func containsSubstring(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(haystack); i++ {
			if haystack[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

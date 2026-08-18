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

func TestJiraClientFetchTicketStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total": 4,
			"issues": []map[string]interface{}{
				{"fields": map[string]interface{}{"customfield_10016": 3.0}},
				{"fields": map[string]interface{}{"customfield_10016": 5.0}},
				{"fields": map[string]interface{}{"customfield_10016": nil}},
				{"fields": map[string]interface{}{"customfield_10016": 2.0}},
			},
		})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}
	if ticketsClosed != 4 {
		t.Fatalf("expected 4 tickets closed, got %d", ticketsClosed)
	}
	if complexityScore != 10 {
		t.Fatalf("expected complexity score sum 10 (3+5+2, nil skipped), got %v", complexityScore)
	}
}

func TestJiraClientFetchTicketStatsMultiProjectQueryConstruction(t *testing.T) {
	var seenJQL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenJQL = r.URL.Query().Get("jql")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 0, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	if _, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG", "OPS", "PLAT"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}

	if !strings.Contains(seenJQL, "project in (") {
		t.Fatalf("expected JQL to contain a single \"project in (...)\" clause, got %q", seenJQL)
	}
	// All three configured projects must appear inside the "project in
	// (...)" clause. Guard against a sloppy implementation that only
	// forwards the first project or emits a separate query per project
	// (Jira's JQL "in" operator takes a single comma-separated list, unlike
	// GitHub's repo-qualifier-per-repo search syntax).
	for _, project := range []string{"ENG", "OPS", "PLAT"} {
		if !strings.Contains(seenJQL, project) {
			t.Fatalf("expected JQL %q to reference project %q", seenJQL, project)
		}
	}
	if strings.Count(seenJQL, "project in (") != 1 {
		t.Fatalf("expected exactly one \"project in (...)\" clause, got JQL %q", seenJQL)
	}
}

func TestJiraClientFetchTicketStatsSendsBasicAuthAndAcceptHeader(t *testing.T) {
	var gotAuthUser, gotAuthPass string
	var gotAuthOK bool
	var gotAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthUser, gotAuthPass, gotAuthOK = r.BasicAuth()
		gotAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 0, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "secret-token", server.Client())

	if _, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}

	if !gotAuthOK {
		t.Fatalf("expected Basic auth credentials to be sent")
	}
	if gotAuthUser != "manager@example.com" || gotAuthPass != "secret-token" {
		t.Fatalf("expected Basic auth manager@example.com/secret-token, got %s/%s", gotAuthUser, gotAuthPass)
	}
	if gotAccept != "application/json" {
		t.Fatalf("expected Accept header %q, got %q", "application/json", gotAccept)
	}
}

func TestJiraClientFetchTicketStatsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errorMessages":["You do not have permission to access this resource."]}`))
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for HTTP 401 response, got nil (tickets=%d complexity=%v)", ticketsClosed, complexityScore)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected error to mention status code 401, got: %v", err)
	}
}

func TestJiraClientFetchTicketStatsServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	_, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for HTTP 500 response, got nil")
	}
}

func TestJiraClientFetchTicketStatsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json{{{`))
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	_, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for malformed JSON response, got nil")
	}
}

func TestJiraClientFetchTicketStatsEmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 0, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}
	if ticketsClosed != 0 || complexityScore != 0 {
		t.Fatalf("expected tickets=0 complexity=0, got tickets=%d complexity=%v", ticketsClosed, complexityScore)
	}
}

func TestJiraClientFetchTicketStatsEmptyProjectsSkipsRequest(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 99, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", nil,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}
	if ticketsClosed != 0 || complexityScore != 0 {
		t.Fatalf("expected tickets=0 complexity=0 for empty projects list, got tickets=%d complexity=%v", ticketsClosed, complexityScore)
	}
	if called {
		t.Fatalf("expected no HTTP request to be made when projects list is empty")
	}
}

func TestJiraClientFetchTicketStatsNetworkFailure(t *testing.T) {
	// Point the client at an address nothing is listening on so the HTTP
	// call itself fails (connection refused), rather than returning a
	// non-200 status. This should surface as an error, not a panic.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	unreachableURL := server.URL
	server.Close() // closing immediately frees the port but leaves nothing listening

	client := NewJiraClient(unreachableURL, "manager@example.com", "fake-token", http.DefaultClient)

	_, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for unreachable server, got nil")
	}
}

func TestJiraClientFetchTicketStatsRespectsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 1, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.FetchTicketStats(
		ctx, "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error for already-canceled context, got nil")
	}
}

func TestJiraClientFetchTicketStatsDateRangeIsHalfOpen(t *testing.T) {
	var seenJQL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenJQL = r.URL.Query().Get("jql")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 0, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	if _, _, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	); err != nil {
		t.Fatalf("fetch ticket stats failed: %v", err)
	}

	if !strings.Contains(seenJQL, `resolved >= "2026-01-01"`) {
		t.Fatalf("expected JQL to include inclusive lower bound resolved >= \"2026-01-01\", got %q", seenJQL)
	}
	if !strings.Contains(seenJQL, `resolved < "2026-01-14"`) {
		t.Fatalf("expected JQL to include exclusive upper bound resolved < \"2026-01-14\", got %q", seenJQL)
	}
	if strings.Contains(seenJQL, `resolved <= `) {
		t.Fatalf("expected no inclusive upper bound (resolved <=) in JQL, got %q", seenJQL)
	}
}

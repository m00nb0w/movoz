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

func TestJiraClientFetchTicketStatsPaginatesAcrossMultiplePages(t *testing.T) {
	var requestCount int
	var seenStartAts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		startAt := r.URL.Query().Get("startAt")
		seenStartAts = append(seenStartAts, startAt)

		var issues []map[string]interface{}
		switch startAt {
		case "0":
			for i := 0; i < 50; i++ {
				issues = append(issues, map[string]interface{}{"fields": map[string]interface{}{"customfield_10016": 1.0}})
			}
		case "50":
			for i := 0; i < 25; i++ {
				issues = append(issues, map[string]interface{}{"fields": map[string]interface{}{"customfield_10016": 1.0}})
			}
		default:
			t.Errorf("unexpected startAt value %q", startAt)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total":  75,
			"issues": issues,
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
	if ticketsClosed != 75 {
		t.Fatalf("expected 75 tickets closed (from total), got %d", ticketsClosed)
	}
	// Each of the 75 issues (across both pages) contributes 1.0 to the
	// complexity score. If pagination were broken and only the first page
	// were read, this would come back as 50, not 75.
	if complexityScore != 75 {
		t.Fatalf("expected complexity score summed across both pages = 75 (not just first page's 50), got %v", complexityScore)
	}
	if requestCount != 2 {
		t.Fatalf("expected exactly 2 HTTP requests (2 pages of 50 + 25), got %d", requestCount)
	}
	if len(seenStartAts) != 2 || seenStartAts[0] != "0" || seenStartAts[1] != "50" {
		t.Fatalf("expected startAt sequence [0, 50], got %v", seenStartAts)
	}
}

func TestJiraClientFetchTicketStatsPaginationStopsOnEmptyPageWithoutInfiniteLoop(t *testing.T) {
	// Guards against an infinite loop if a server reports a total larger
	// than the issues it actually has available (e.g. total: 5 but the
	// second page comes back empty instead of erroring). FetchTicketStats
	// must terminate rather than looping forever re-requesting the same or
	// advancing startAt indefinitely.
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		startAt := r.URL.Query().Get("startAt")
		w.Header().Set("Content-Type", "application/json")
		if startAt == "0" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total": 5,
				"issues": []map[string]interface{}{
					{"fields": map[string]interface{}{"customfield_10016": 2.0}},
					{"fields": map[string]interface{}{"customfield_10016": 3.0}},
				},
			})
			return
		}
		// Any subsequent page: server has nothing more, despite total: 5.
		json.NewEncoder(w).Encode(map[string]interface{}{"total": 5, "issues": []map[string]interface{}{}})
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	done := make(chan struct{})
	var ticketsClosed int
	var complexityScore float64
	var fetchErr error
	go func() {
		ticketsClosed, complexityScore, fetchErr = client.FetchTicketStats(
			context.Background(), "abc123", []string{"ENG"},
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
		)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("FetchTicketStats did not return within 5s — likely infinite pagination loop")
	}

	if fetchErr != nil {
		t.Fatalf("fetch ticket stats failed: %v", fetchErr)
	}
	if ticketsClosed != 5 {
		t.Fatalf("expected ticketsClosed=5 (from total), got %d", ticketsClosed)
	}
	if complexityScore != 5 {
		t.Fatalf("expected complexityScore=5 (2+3 from the one non-empty page), got %v", complexityScore)
	}
	if requestCount != 2 {
		t.Fatalf("expected exactly 2 requests (first page + one empty page before stopping), got %d", requestCount)
	}
}

func TestJiraClientFetchTicketStatsPageErrorFailsWholeCall(t *testing.T) {
	// The first page succeeds; the second page (triggered by pagination)
	// fails. The whole call must return an error, not silently return
	// partial results.
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		startAt := r.URL.Query().Get("startAt")
		if startAt == "0" {
			w.Header().Set("Content-Type", "application/json")
			var issues []map[string]interface{}
			for i := 0; i < 50; i++ {
				issues = append(issues, map[string]interface{}{"fields": map[string]interface{}{"customfield_10016": 1.0}})
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"total": 75, "issues": issues})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewJiraClient(server.URL, "manager@example.com", "fake-token", server.Client())

	ticketsClosed, complexityScore, err := client.FetchTicketStats(
		context.Background(), "abc123", []string{"ENG"},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatalf("expected error when a later page fails, got nil (tickets=%d complexity=%v)", ticketsClosed, complexityScore)
	}
	if requestCount != 2 {
		t.Fatalf("expected exactly 2 requests (first page succeeds, second page fails), got %d", requestCount)
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

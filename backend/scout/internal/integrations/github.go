// Package integrations contains clients for external services (GitHub,
// Jira, ...) that Scout's sync worker pulls engineer metrics from.
package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GitHubClient fetches pull-request activity counts from the GitHub REST
// search API (F4).
type GitHubClient struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubClient builds a GitHubClient. token is sent as a Bearer token on
// every request; pass "" to make unauthenticated requests (subject to
// GitHub's much lower unauthenticated rate limits). If httpClient is nil,
// http.DefaultClient is used.
func NewGitHubClient(token string, httpClient *http.Client) *GitHubClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &GitHubClient{token: token, httpClient: httpClient, baseURL: "https://api.github.com"}
}

// searchIssuesResponse mirrors the subset of GitHub's search API response we
// need. GitHub always returns total_count as the full match count for the
// query regardless of pagination (per_page/page), so no pagination handling
// is required here — we only ever read the count, never the item list.
type searchIssuesResponse struct {
	TotalCount int `json:"total_count"`
}

// FetchPRStats returns the number of pull requests username raised and the
// number they reviewed, across repos, with created dates in [since, until]
// (GitHub's day-granularity date-range search, "created:YYYY-MM-DD..YYYY-MM-DD",
// is inclusive of both endpoints). Because both endpoints are inclusive,
// callers computing consecutive sync-cycle boundaries must not pass the same
// date as both one cycle's until and the next cycle's since — doing so will
// double-count that boundary day's PRs in both cycles' snapshots; offset by
// one day (e.g. next since = previous until + 24h) to avoid overlap.
//
// If repos is empty, FetchPRStats returns (0, 0, nil) without making any
// request, since an unscoped query would search every repo on GitHub rather
// than the configured SCOUT_GITHUB_REPOS set.
func (c *GitHubClient) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (prsRaised, prsReviewed int, err error) {
	if len(repos) == 0 {
		return 0, 0, nil
	}

	repoFilter := ""
	for _, repo := range repos {
		repoFilter += fmt.Sprintf(" repo:%s", repo)
	}
	dateRange := fmt.Sprintf("%s..%s", since.Format("2006-01-02"), until.Format("2006-01-02"))

	raisedQuery := fmt.Sprintf("is:pr author:%s created:%s%s", username, dateRange, repoFilter)
	prsRaised, err = c.searchCount(ctx, raisedQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching raised PRs for %s: %w", username, err)
	}

	reviewedQuery := fmt.Sprintf("is:pr reviewed-by:%s created:%s%s", username, dateRange, repoFilter)
	prsReviewed, err = c.searchCount(ctx, reviewedQuery)
	if err != nil {
		return 0, 0, fmt.Errorf("fetching reviewed PRs for %s: %w", username, err)
	}

	return prsRaised, prsReviewed, nil
}

// searchCount issues a single GitHub search/issues query and returns its
// total_count. Any non-200 response (including rate limiting, which GitHub
// surfaces as 403 or 422 on the search endpoint) is returned as an error
// carrying the status code and response body rather than panicking or
// silently returning a zero count.
func (c *GitHubClient) searchCount(ctx context.Context, query string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/search/issues", nil)
	if err != nil {
		return 0, fmt.Errorf("building github search request: %w", err)
	}
	q := req.URL.Query()
	q.Set("q", query)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("calling github search api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return 0, fmt.Errorf("github search returned status %d: %s", resp.StatusCode, string(body))
	}

	var result searchIssuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding github search response: %w", err)
	}
	return result.TotalCount, nil
}

package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// JiraClient fetches ticket-closed counts and complexity (story-point)
// totals from Jira's REST API search endpoint (F4).
type JiraClient struct {
	baseURL    string
	email      string
	apiToken   string
	httpClient *http.Client
}

// NewJiraClient builds a JiraClient. Requests authenticate via HTTP Basic
// auth using email and apiToken — Jira Cloud's documented API-token auth
// scheme (https://id.atlassian.com/manage-profile/security/api-tokens). If
// httpClient is nil, http.DefaultClient is used.
func NewJiraClient(baseURL, email, apiToken string, httpClient *http.Client) *JiraClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &JiraClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		email:      email,
		apiToken:   apiToken,
		httpClient: httpClient,
	}
}

// complexityFieldID is the custom field Jira Cloud uses by default for the
// "Story point estimate" on company-managed Scrum/Kanban projects. It is
// hardcoded rather than sourced from config: an instance that has renumbered
// or disabled this field would need a code change (or a future
// SCOUT_JIRA_COMPLEXITY_FIELD env var) to point elsewhere.
const complexityFieldID = "customfield_10016"

// jiraPageSize is the number of issues requested per /rest/api/3/search page.
// ticketsClosed comes directly from the response's "total" field (accurate
// regardless of page size), but complexityScore requires summing a field
// across every issue, so FetchTicketStats pages through startAt/total until
// all issues have been fetched.
const jiraPageSize = 50

// jiraSearchResponse mirrors the subset of Jira's /rest/api/3/search
// response used here. "total" is the full match count for the JQL query
// regardless of maxResults/startAt pagination.
type jiraSearchResponse struct {
	Total  int `json:"total"`
	Issues []struct {
		Fields map[string]interface{} `json:"fields"`
	} `json:"issues"`
}

// FetchTicketStats returns the number of tickets accountID resolved as Done,
// and the sum of their complexity (story-point) figures, across the given
// Jira projects, for tickets resolved in the half-open interval
// [since, until) — i.e. resolved >= since and resolved < until. Issues with
// no value set for the complexity field do not contribute to
// complexityScore (nil/missing values are treated as 0, not an error).
//
// Boundary/double-counting note: unlike GitHubClient.FetchPRStats's
// day-granularity date range (which is inclusive on *both* ends, because
// GitHub's search syntax has no time component, and therefore requires
// callers to offset consecutive sync cycles by one day to avoid double
// counting), this range is deliberately half-open. Jira's JQL treats a
// date-only literal compared against a date-time field (resolved) as
// midnight of that date, so "resolved < until" excludes everything from
// `until` onward while "resolved >= since" includes everything from
// `since`'s midnight onward. This means consecutive sync cycles can safely
// chain (next cycle's since == previous cycle's until) with no boundary
// day double-counted in both snapshots, and no one-day offset is needed
// here — in contrast to FetchPRStats.
//
// If projects is empty, FetchTicketStats returns (0, 0, nil) without making
// any request, since an unscoped query would search every project the
// account can see rather than the configured SCOUT_JIRA_PROJECTS set.
//
// Pagination: Jira's search endpoint caps each response to jiraPageSize
// issues, so FetchTicketStats pages through startAt until it has fetched
// every issue "total" reports, summing complexityScore across all of them
// (ticketsClosed is read once from "total" and does not need paging — it is
// already the full match count on every page). This matters because a
// single-page read would silently undercount complexityScore for any
// engineer who resolved more than one page's worth of complexity-scored
// tickets in a sync period, even though ticketsClosed stayed correct.
func (c *JiraClient) FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (ticketsClosed int, complexityScore float64, err error) {
	if len(projects) == 0 {
		return 0, 0, nil
	}

	quotedProjects := make([]string, len(projects))
	for i, p := range projects {
		quotedProjects[i] = fmt.Sprintf("%q", p)
	}
	// Note: "status = Done" assumes a workflow status literally named
	// "Done". Projects with renamed terminal statuses would need
	// "statusCategory = Done" instead; kept as "status = Done" here to
	// match the pinned query shape from the task brief.
	jql := fmt.Sprintf(
		`assignee = %q AND status = Done AND project in (%s) AND resolved >= %q AND resolved < %q`,
		accountID, strings.Join(quotedProjects, ","), since.Format("2006-01-02"), until.Format("2006-01-02"),
	)

	total := -1 // unknown until the first page is fetched
	startAt := 0
	for total == -1 || startAt < total {
		result, err := c.searchPage(ctx, jql, startAt)
		if err != nil {
			return 0, 0, err
		}

		if total == -1 {
			total = result.Total
			ticketsClosed = total
		}

		for _, issue := range result.Issues {
			if v, ok := issue.Fields[complexityFieldID].(float64); ok {
				complexityScore += v
			}
		}

		// If a page comes back with no issues at all (e.g. total was 0, or
		// the server returned fewer issues than it reported existed), stop
		// rather than looping forever re-requesting the same startAt.
		if len(result.Issues) == 0 {
			break
		}
		startAt += len(result.Issues)
	}

	return ticketsClosed, complexityScore, nil
}

// searchPage issues a single /rest/api/3/search request at the given
// startAt offset and decodes the response. Any transport error, non-200
// status, or JSON decode failure is returned as a wrapped error rather than
// panicking.
func (c *JiraClient) searchPage(ctx context.Context, jql string, startAt int) (jiraSearchResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/rest/api/3/search", nil)
	if err != nil {
		return jiraSearchResponse{}, fmt.Errorf("building jira search request: %w", err)
	}
	q := req.URL.Query()
	q.Set("jql", jql)
	q.Set("fields", complexityFieldID)
	q.Set("startAt", fmt.Sprintf("%d", startAt))
	q.Set("maxResults", fmt.Sprintf("%d", jiraPageSize))
	req.URL.RawQuery = q.Encode()
	req.SetBasicAuth(c.email, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return jiraSearchResponse{}, fmt.Errorf("calling jira search api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return jiraSearchResponse{}, fmt.Errorf("jira search returned status %d: %s", resp.StatusCode, string(body))
	}

	var result jiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return jiraSearchResponse{}, fmt.Errorf("decoding jira search response: %w", err)
	}
	return result, nil
}

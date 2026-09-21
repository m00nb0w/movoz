package jirasync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"el-storko/internal/models"
)

// JiraIssue is the subset of a Jira issue this sync loop cares about.
type JiraIssue struct {
	Key         string
	Title       string
	Description string
	Status      models.Status
	URL         string
	UpdatedAt   time.Time
}

// Client is implemented by both the real HTTP-backed Jira client and test
// mocks, so the sync loop never depends on net/http directly.
type Client interface {
	SearchAssignedIssues() ([]JiraIssue, error)
	UpdateIssue(key, title, description string, status models.Status) error
}

// HTTPClient talks to Jira REST API v3 using HTTP Basic auth (email + API
// token). It never logs the token.
type HTTPClient struct {
	email      string
	token      string
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(email, token, baseURL string) *HTTPClient {
	return &HTTPClient{
		email:      email,
		token:      token,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *HTTPClient) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.email, c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

type jiraSearchResponse struct {
	Issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary     string          `json:"summary"`
			Description json.RawMessage `json:"description"`
			Status      struct {
				Name string `json:"name"`
			} `json:"status"`
			Updated string `json:"updated"`
		} `json:"fields"`
	} `json:"issues"`
}

// adfNode is the subset of Atlassian Document Format this client understands:
// Jira Cloud's REST API v3 represents rich-text fields like "description" as
// an ADF document, never a plain string.
type adfNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text,omitempty"`
	Content []adfNode `json:"content,omitempty"`
}

// textFromADF extracts plain text from an ADF description, joining
// paragraphs with newlines. A null/empty field or anything that isn't a
// well-formed ADF document yields an empty string rather than an error, so
// one unusual issue never aborts the whole sync cycle.
func textFromADF(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var doc adfNode
	if err := json.Unmarshal(raw, &doc); err != nil {
		return ""
	}
	var paragraphs []string
	for _, block := range doc.Content {
		paragraphs = append(paragraphs, textFromADFNode(block))
	}
	return strings.Join(paragraphs, "\n")
}

func textFromADFNode(n adfNode) string {
	if n.Type == "text" {
		return n.Text
	}
	var parts []string
	for _, child := range n.Content {
		parts = append(parts, textFromADFNode(child))
	}
	return strings.Join(parts, "")
}

// adfFromText builds the minimal ADF document Jira Cloud's v3 API requires
// when writing a rich-text field, one paragraph per line.
func adfFromText(text string) map[string]any {
	lines := strings.Split(text, "\n")
	content := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		para := map[string]any{"type": "paragraph"}
		if line != "" {
			para["content"] = []map[string]any{{"type": "text", "text": line}}
		}
		content = append(content, para)
	}
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": content,
	}
}

// jiraStatusToLocal maps a Jira status category/name to el-storko's fixed
// four-value status set. Anything unrecognized defaults to "todo" rather
// than failing the whole sync cycle over an unmapped Jira workflow name.
func jiraStatusToLocal(name string) models.Status {
	switch name {
	case "Done", "Closed", "Resolved":
		return models.StatusDone
	case "In Progress":
		return models.StatusInProgress
	case "Blocked":
		return models.StatusBlocked
	default:
		return models.StatusTodo
	}
}

func localStatusToJiraTransitionName(status models.Status) string {
	switch status {
	case models.StatusDone:
		return "Done"
	case models.StatusInProgress:
		return "In Progress"
	case models.StatusBlocked:
		return "Blocked"
	default:
		return "To Do"
	}
}

func (c *HTTPClient) SearchAssignedIssues() ([]JiraIssue, error) {
	req, err := c.newRequest(http.MethodGet, "/rest/api/3/search?jql=assignee%20%3D%20currentUser()", nil)
	if err != nil {
		return nil, fmt.Errorf("jira: could not build search request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira: search returned status %d", resp.StatusCode)
	}

	var parsed jiraSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("jira: could not decode search response: %w", err)
	}

	issues := make([]JiraIssue, 0, len(parsed.Issues))
	for _, i := range parsed.Issues {
		updated, err := time.Parse("2006-01-02T15:04:05.000-0700", i.Fields.Updated)
		if err != nil {
			updated = time.Now()
		}
		issues = append(issues, JiraIssue{
			Key:         i.Key,
			Title:       i.Fields.Summary,
			Description: textFromADF(i.Fields.Description),
			Status:      jiraStatusToLocal(i.Fields.Status.Name),
			URL:         c.baseURL + "/browse/" + i.Key,
			UpdatedAt:   updated,
		})
	}
	return issues, nil
}

func (c *HTTPClient) UpdateIssue(key, title, description string, status models.Status) error {
	body := map[string]any{
		"fields": map[string]any{
			"summary":     title,
			"description": adfFromText(description),
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("jira: could not encode update payload: %w", err)
	}

	req, err := c.newRequest(http.MethodPut, "/rest/api/3/issue/"+key, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("jira: could not build update request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("jira: update request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jira: update returned status %d", resp.StatusCode)
	}

	return c.transitionIssue(key, status)
}

type jiraTransitionsResponse struct {
	Transitions []struct {
		ID string `json:"id"`
		To struct {
			Name string `json:"name"`
		} `json:"to"`
	} `json:"transitions"`
}

// transitionIssue looks up the issue's available transitions and applies
// the one whose target status name matches, since Jira requires a
// transition id rather than a direct status field write.
func (c *HTTPClient) transitionIssue(key string, status models.Status) error {
	req, err := c.newRequest(http.MethodGet, "/rest/api/3/issue/"+key+"/transitions", nil)
	if err != nil {
		return fmt.Errorf("jira: could not build transitions request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("jira: transitions request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jira: transitions lookup returned status %d", resp.StatusCode)
	}

	var parsed jiraTransitionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("jira: could not decode transitions response: %w", err)
	}

	target := localStatusToJiraTransitionName(status)
	var transitionID string
	for _, t := range parsed.Transitions {
		if t.To.Name == target {
			transitionID = t.ID
			break
		}
	}
	if transitionID == "" {
		return fmt.Errorf("jira: no transition found for status %q", target)
	}

	transitionBody, err := json.Marshal(map[string]any{
		"transition": map[string]string{"id": transitionID},
	})
	if err != nil {
		return fmt.Errorf("jira: could not encode transition payload: %w", err)
	}

	transitionReq, err := c.newRequest(http.MethodPost, "/rest/api/3/issue/"+key+"/transitions", bytes.NewReader(transitionBody))
	if err != nil {
		return fmt.Errorf("jira: could not build transition request: %w", err)
	}
	transitionResp, err := c.httpClient.Do(transitionReq)
	if err != nil {
		return fmt.Errorf("jira: transition request failed: %w", err)
	}
	defer transitionResp.Body.Close()
	if transitionResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("jira: transition returned status %d", transitionResp.StatusCode)
	}
	return nil
}

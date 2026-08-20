package aiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// ExistingEntry is one of an engineer's existing highlight/lowlight entries,
// passed in by the caller so this method never has to touch the DB itself.
type ExistingEntry struct {
	ID   int
	Kind string
	Body string
}

// DuplicateCheckResult is Claude's structured verdict on whether a new
// highlight/lowlight entry is a semantic duplicate of an existing one (F14).
type DuplicateCheckResult struct {
	IsDuplicate    bool   `json:"is_duplicate"`
	MatchedEntryID *int   `json:"matched_entry_id"`
	Note           string `json:"similarity_note"`
}

// CheckDuplicate asks Claude whether newBody is a likely semantic duplicate
// of any of the engineer's existing highlight/lowlight entries (F14), using
// a JSON-schema structured output so the response is always parseable —
// no code-block extraction needed here, unlike the ranking chat.
func (c *Client) CheckDuplicate(ctx context.Context, newBody string, existing []ExistingEntry) (*DuplicateCheckResult, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "New entry to check: %q\n\nExisting entries for this engineer:\n", newBody)
	for _, e := range existing {
		fmt.Fprintf(&b, "- id=%d (%s): %s\n", e.ID, e.Kind, e.Body)
	}
	b.WriteString("\nDetermine whether the new entry is a likely semantic duplicate of any existing entry — same underlying observation, not necessarily the same wording. If so, set matched_entry_id to that entry's id; otherwise null.")

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"is_duplicate": map[string]interface{}{"type": "boolean"},
			"matched_entry_id": map[string]interface{}{
				"anyOf": []map[string]interface{}{
					{"type": "integer"},
					{"type": "null"},
				},
			},
			"similarity_note": map[string]interface{}{"type": "string"},
		},
		"required":             []string{"is_duplicate", "matched_entry_id", "similarity_note"},
		"additionalProperties": false,
	}

	resp, err := c.anthropic.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-opus-5",
		MaxTokens: 512,
		OutputConfig: anthropic.OutputConfigParam{
			Format: anthropic.JSONOutputFormatParam{Schema: schema},
		},
		Messages: []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(b.String()))},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from duplicate check")
	}
	textBlock, ok := resp.Content[0].AsAny().(anthropic.TextBlock)
	if !ok {
		return nil, fmt.Errorf("unexpected response content type from duplicate check")
	}

	var result DuplicateCheckResult
	if err := json.Unmarshal([]byte(textBlock.Text), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

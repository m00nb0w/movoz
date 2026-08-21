package aiclient

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"
)

// TestCheckDuplicateParsesStructuredResponse covers the core happy path: a
// genuine duplicate is detected and the matched existing entry's id comes
// back correctly parsed from the structured JSON output.
func TestCheckDuplicateParsesStructuredResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_dup",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":true,\"matched_entry_id\":5,\"similarity_note\":\"Both describe missing the Q1 deadline\"}"}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 20, "output_tokens": 15}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Missed the Q1 deadline without flagging it early", []ExistingEntry{
		{ID: 5, Kind: "lowlight", Body: "Slipped the Q1 deadline and didn't raise it in standup"},
	})
	if err != nil {
		t.Fatalf("check duplicate failed: %v", err)
	}
	if !result.IsDuplicate {
		t.Fatal("expected IsDuplicate true")
	}
	if result.MatchedEntryID == nil || *result.MatchedEntryID != 5 {
		t.Fatalf("expected matched entry id 5, got %v", result.MatchedEntryID)
	}
	if result.Note == "" {
		t.Fatal("expected a similarity note to be populated on a genuine duplicate")
	}
}

// TestCheckDuplicateNoMatchFound covers the no-duplicate case: Claude
// determines the new entry is distinct from everything on record, so
// matched_entry_id comes back null and IsDuplicate is false.
func TestCheckDuplicateNoMatchFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_nodup",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":false,\"matched_entry_id\":null,\"similarity_note\":\"No existing entry covers this topic\"}"}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 18, "output_tokens": 12}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Shipped the new onboarding flow ahead of schedule", []ExistingEntry{
		{ID: 5, Kind: "lowlight", Body: "Slipped the Q1 deadline and didn't raise it in standup"},
	})
	if err != nil {
		t.Fatalf("check duplicate failed: %v", err)
	}
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false")
	}
	if result.MatchedEntryID != nil {
		t.Fatalf("expected no matched entry id, got %v", *result.MatchedEntryID)
	}
}

// TestCheckDuplicateZeroExistingEntries covers a first-ever entry for an
// engineer, i.e. an empty existing-entries slice. The brief does not specify
// a local short-circuit for this case (there is no "if len(existing) == 0"
// branch in the implementation), so this still round-trips through the API
// — the prompt just lists no existing entries — and should come back as
// not-a-duplicate.
func TestCheckDuplicateZeroExistingEntries(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_first",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":false,\"matched_entry_id\":null,\"similarity_note\":\"No existing entries to compare against\"}"}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 10, "output_tokens": 8}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "First highlight ever logged for this engineer", nil)
	if err != nil {
		t.Fatalf("check duplicate failed: %v", err)
	}
	if !called {
		t.Fatal("expected the API to still be called with zero existing entries (no short-circuit specified in the brief)")
	}
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false when there is nothing to compare against")
	}
	if result.MatchedEntryID != nil {
		t.Fatalf("expected no matched entry id, got %v", *result.MatchedEntryID)
	}
}

// TestCheckDuplicateAPIError covers the underlying Anthropic API returning a
// non-2xx error response. The method must surface that as a Go error rather
// than panicking or returning a zero-value result as if nothing happened.
func TestCheckDuplicateAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"type":"error","error":{"type":"api_error","message":"internal server error"}}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Some new entry", []ExistingEntry{
		{ID: 1, Kind: "highlight", Body: "Some existing entry"},
	})
	if err == nil {
		t.Fatal("expected an API error to surface as an error")
	}
	if result != nil {
		t.Fatalf("expected a nil result on API error, got %+v", result)
	}
}

// TestCheckDuplicateEmptyContent covers the API returning a well-formed
// message with an empty content array (no text block at all) — an edge case
// distinct from malformed JSON inside a text block. The method must return
// an error rather than panicking on an out-of-range index.
func TestCheckDuplicateEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_empty",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 5, "output_tokens": 0}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Some new entry", []ExistingEntry{
		{ID: 1, Kind: "highlight", Body: "Some existing entry"},
	})
	if err == nil {
		t.Fatal("expected an error when the response has no content blocks")
	}
	if result != nil {
		t.Fatalf("expected a nil result on empty content, got %+v", result)
	}
}

// TestCheckDuplicateMalformedJSONInTextBlock covers the API returning a text
// block whose contents aren't valid JSON at all (e.g. the model ignored the
// schema and returned prose). json.Unmarshal must fail gracefully with an
// error rather than the method panicking.
func TestCheckDuplicateMalformedJSONInTextBlock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_bad",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "text", "text": "Sorry, I cannot determine that."}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 12, "output_tokens": 6}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Some new entry", []ExistingEntry{
		{ID: 1, Kind: "highlight", Body: "Some existing entry"},
	})
	if err == nil {
		t.Fatal("expected an error when the text block isn't valid JSON")
	}
	if result != nil {
		t.Fatalf("expected a nil result on malformed JSON, got %+v", result)
	}
}

// TestCheckDuplicateUnexpectedContentBlockType covers the API returning a
// well-formed message whose first content block isn't a text block at all
// (e.g. a thinking block, if extended thinking were ever enabled upstream).
// The type assertion to anthropic.TextBlock must fail gracefully with an
// error rather than panicking on a bad type assertion.
func TestCheckDuplicateUnexpectedContentBlockType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"id": "msg_thinking",
			"type": "message",
			"role": "assistant",
			"model": "claude-opus-5",
			"content": [{"type": "thinking", "thinking": "pondering...", "signature": "sig"}],
			"stop_reason": "end_turn",
			"stop_sequence": null,
			"usage": {"input_tokens": 12, "output_tokens": 6}
		}`)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	result, err := client.CheckDuplicate(context.Background(), "Some new entry", []ExistingEntry{
		{ID: 1, Kind: "highlight", Body: "Some existing entry"},
	})
	if err == nil {
		t.Fatal("expected an error when the first content block isn't a text block")
	}
	if result != nil {
		t.Fatalf("expected a nil result on unexpected content block type, got %+v", result)
	}
}

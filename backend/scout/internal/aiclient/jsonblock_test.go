package aiclient

import (
	"encoding/json"
	"testing"
)

func TestExtractJSONBlockFindsFencedBlock(t *testing.T) {
	text := "Alex clearly stood out this cycle for shipping the auth migration solo.\n\n```json\n{\"ranking\":[{\"engineer_id\":1,\"rank\":1},{\"engineer_id\":2,\"rank\":2}],\"rationale\":\"Alex led the migration\"}\n```"

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected a JSON block to be found")
	}
	if string(raw) != `{"ranking":[{"engineer_id":1,"rank":1},{"engineer_id":2,"rank":2}],"rationale":"Alex led the migration"}` {
		t.Fatalf("unexpected extracted JSON: %s", raw)
	}
}

func TestExtractJSONBlockNoBlockPresent(t *testing.T) {
	_, ok := ExtractJSONBlock("Tell me more about what Sam shipped this cycle before I propose a ranking.")
	if ok {
		t.Fatal("expected no JSON block to be found in a plain clarifying question")
	}
}

func TestExtractJSONBlockRejectsInvalidJSON(t *testing.T) {
	_, ok := ExtractJSONBlock("```json\nnot valid json\n```")
	if ok {
		t.Fatal("expected invalid JSON inside the fence to be rejected")
	}
}

func TestExtractJSONBlockReturnsLastBlockWhenMultiple(t *testing.T) {
	text := "```json\n{\"draft\":1}\n```\n\nActually, here's the revised proposal:\n\n```json\n{\"draft\":2}\n```"
	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected a JSON block to be found")
	}
	if string(raw) != `{"draft":2}` {
		t.Fatalf("expected the last block to win, got %s", raw)
	}
}

// TestExtractJSONBlockEmptyString covers the trivial no-input edge case.
func TestExtractJSONBlockEmptyString(t *testing.T) {
	_, ok := ExtractJSONBlock("")
	if ok {
		t.Fatal("expected no JSON block to be found in an empty string")
	}
}

// TestExtractJSONBlockTextAfterBlock verifies that a fenced JSON block does
// not need to be the last thing in the message — trailing prose after the
// (only) block should not prevent extraction.
func TestExtractJSONBlockTextAfterBlock(t *testing.T) {
	text := "```json\n{\"ranking\":[{\"engineer_id\":1,\"rank\":1}],\"rationale\":\"ok\"}\n```\n\nLet me know if you'd like any changes to this proposal."

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected a JSON block to be found even with trailing prose after it")
	}
	if string(raw) != `{"ranking":[{"engineer_id":1,"rank":1}],"rationale":"ok"}` {
		t.Fatalf("unexpected extracted JSON: %s", raw)
	}
}

// TestExtractJSONBlockLastBlockInvalidDoesNotFallBack pins down the "last
// block wins" rule precisely: if the LAST fenced json block is malformed,
// extraction fails outright — it does not fall back to an earlier valid
// block in the same message.
func TestExtractJSONBlockLastBlockInvalidDoesNotFallBack(t *testing.T) {
	text := "```json\n{\"draft\":1}\n```\n\nWait, scratch that:\n\n```json\nthis is not json\n```"

	_, ok := ExtractJSONBlock(text)
	if ok {
		t.Fatal("expected extraction to fail when the last block is invalid, even though an earlier block was valid")
	}
}

// TestExtractJSONBlockNestedAndEscapedContent exercises nested objects,
// arrays, and escaped characters (quotes, braces-as-text, backslashes)
// within the fenced JSON to make sure the extraction doesn't get confused
// by brace/quote characters inside string values.
func TestExtractJSONBlockNestedAndEscapedContent(t *testing.T) {
	text := "```json\n{\"ranking\":[{\"engineer_id\":1,\"rank\":1,\"notes\":{\"detail\":\"uses \\\"curly\\\" braces like {this} and a backslash \\\\ in a string\"}}],\"rationale\":\"nested \\u0026 escaped\"}\n```"

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected nested/escaped JSON to be extracted successfully")
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("extracted JSON did not parse: %v", err)
	}
	if _, exists := parsed["ranking"]; !exists {
		t.Fatalf("expected parsed JSON to contain 'ranking' key, got: %v", parsed)
	}
}

// TestExtractJSONBlockIgnoresNonJSONFence ensures a fenced block tagged with
// a different language (e.g. ```python) is not mistaken for the JSON block,
// while a genuine trailing ```json block is still found.
func TestExtractJSONBlockIgnoresNonJSONFence(t *testing.T) {
	text := "Here's a snippet for reference:\n\n```python\nprint(\"ranking\")\n```\n\nAnd here's my proposal:\n\n```json\n{\"draft\":1}\n```"

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected the json-tagged block to be found")
	}
	if string(raw) != `{"draft":1}` {
		t.Fatalf("expected the json-tagged block content, got %s", raw)
	}
}

// TestExtractJSONBlockWhitespaceVariations checks that leading/trailing
// whitespace and blank lines inside the fence are trimmed correctly and
// don't affect validity.
func TestExtractJSONBlockWhitespaceVariations(t *testing.T) {
	text := "```json\n\n   {\"draft\":1}   \n\n```"

	raw, ok := ExtractJSONBlock(text)
	if !ok {
		t.Fatal("expected whitespace-padded JSON block to be found")
	}
	if string(raw) != `{"draft":1}` {
		t.Fatalf("expected trimmed JSON content, got %q", raw)
	}
}

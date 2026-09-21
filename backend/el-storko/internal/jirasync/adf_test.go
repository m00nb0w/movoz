package jirasync

import (
	"encoding/json"
	"testing"
)

func TestTextFromADF(t *testing.T) {
	raw := json.RawMessage(`{
		"type": "doc",
		"version": 1,
		"content": [
			{"type": "paragraph", "content": [{"type": "text", "text": "line one"}]},
			{"type": "paragraph", "content": [{"type": "text", "text": "line two"}]}
		]
	}`)

	got := textFromADF(raw)
	want := "line one\nline two"
	if got != want {
		t.Errorf("textFromADF() = %q, want %q", got, want)
	}
}

func TestTextFromADF_NullOrEmpty(t *testing.T) {
	if got := textFromADF(nil); got != "" {
		t.Errorf("textFromADF(nil) = %q, want empty", got)
	}
	if got := textFromADF(json.RawMessage("null")); got != "" {
		t.Errorf("textFromADF(null) = %q, want empty", got)
	}
}

func TestAdfFromText_RoundTrip(t *testing.T) {
	original := "line one\nline two"

	doc := adfFromText(original)
	encoded, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal adf doc: %v", err)
	}

	got := textFromADF(json.RawMessage(encoded))
	if got != original {
		t.Errorf("round trip = %q, want %q", got, original)
	}
}

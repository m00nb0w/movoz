package aiclient

import (
	"encoding/json"
	"regexp"
)

var jsonBlockPattern = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")

// ExtractJSONBlock finds the last ```json fenced code block in text — the
// convention this plan uses for the ranking chat to carry a proposed_ranking
// payload alongside its conversational rationale (F9) — and returns it as
// raw JSON. ok is false if no block is present or its contents aren't valid
// JSON (e.g. the assistant is still asking a clarifying question).
func ExtractJSONBlock(text string) (raw json.RawMessage, ok bool) {
	matches := jsonBlockPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, false
	}
	candidate := matches[len(matches)-1][1]
	if !json.Valid([]byte(candidate)) {
		return nil, false
	}
	return json.RawMessage(candidate), true
}

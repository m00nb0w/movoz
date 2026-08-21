// Package aiclient wraps the official Anthropic Go SDK for Scout's two AI
// flows: the conversational ranking chat (F9) and the highlight/lowlight
// semantic duplicate check (F14).
package aiclient

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the official Anthropic Go SDK for Scout's two AI flows: the
// conversational ranking chat (F9) and the highlight/lowlight semantic
// duplicate check (F14).
type Client struct {
	anthropic anthropic.Client
}

// NewClient builds a Client from an API key, plus any extra SDK request
// options (tests use this to point the client at an httptest server via
// option.WithBaseURL instead of the real Anthropic API).
func NewClient(apiKey string, opts ...option.RequestOption) *Client {
	allOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	return &Client{anthropic: anthropic.NewClient(allOpts...)}
}

// ChatMessage is one turn in the ranking chat transcript.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

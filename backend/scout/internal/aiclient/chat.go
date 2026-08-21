package aiclient

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// StreamRankingChat sends the conversation (prior transcript + new user
// message) to Claude and streams the assistant's reply as SSE "data: "
// chunks to w for the frontend chat UI to consume (F9). It returns the full
// accumulated reply text plus any proposed ranking JSON extracted from a
// trailing ```json code block in that reply — the caller (the ai-sessions
// handler) is responsible for persisting the transcript and for only ever
// writing the ranking into sub_attribute_rankings on explicit admin confirm
// (NF3); this method never touches sub_attribute_rankings.
func (c *Client) StreamRankingChat(ctx context.Context, w io.Writer, systemPrompt string, history []ChatMessage, userMessage string) (replyText string, proposedRanking json.RawMessage, err error) {
	messages := make([]anthropic.MessageParam, 0, len(history)+1)
	for _, m := range history {
		role := anthropic.MessageParamRoleUser
		if m.Role == "assistant" {
			role = anthropic.MessageParamRoleAssistant
		}
		messages = append(messages, anthropic.MessageParam{
			Role:    role,
			Content: []anthropic.ContentBlockParamUnion{anthropic.NewTextBlock(m.Content)},
		})
	}
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)))

	stream := c.anthropic.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     "claude-opus-5",
		MaxTokens: 4096,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages:  messages,
	})
	defer stream.Close()

	var builder strings.Builder
	for stream.Next() {
		event := stream.Current()
		delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent)
		if !ok {
			continue
		}
		textDelta, ok := delta.Delta.AsAny().(anthropic.TextDelta)
		if !ok {
			continue
		}
		builder.WriteString(textDelta.Text)
		if _, werr := w.Write([]byte("data: " + escapeSSE(textDelta.Text) + "\n\n")); werr != nil {
			return "", nil, werr
		}
		if flusher, ok := w.(interface{ Flush() }); ok {
			flusher.Flush()
		}
	}
	if err := stream.Err(); err != nil {
		return "", nil, err
	}

	replyText = builder.String()
	if raw, found := ExtractJSONBlock(replyText); found {
		proposedRanking = raw
	}
	return replyText, proposedRanking, nil
}

func escapeSSE(s string) string {
	return strings.ReplaceAll(s, "\n", "\\n")
}

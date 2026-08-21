package aiclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestStreamRankingChatAccumulatesTextAndExtractsProposal(t *testing.T) {
	sseBody := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_test","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Alex should rank 1st. "}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"` +
			"```json\\n{\\\"ranking\\\":[{\\\"engineer_id\\\":1,\\\"rank\\\":1}]}\\n```" +
			`"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":0}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":25}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseBody)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		context.Background(), &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err != nil {
		t.Fatalf("stream ranking chat failed: %v", err)
	}
	if !strings.Contains(replyText, "Alex should rank 1st.") {
		t.Fatalf("expected accumulated reply text to include the streamed sentence, got %q", replyText)
	}
	if proposedRanking == nil {
		t.Fatal("expected a proposed ranking to be extracted from the trailing JSON block")
	}
	if out.Len() == 0 {
		t.Fatal("expected streamed chunks to be written to the writer")
	}
}

// TestStreamRankingChatNoJSONBlock covers a reply that never emits a trailing
// ```json fenced block — e.g. the assistant is still asking a clarifying
// question rather than proposing a ranking. The full text should still come
// back; there should simply be no proposal (F9 lets the admin keep chatting
// before any ranking is proposed).
func TestStreamRankingChatNoJSONBlock(t *testing.T) {
	sseBody := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_test2","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Can you tell me more about what Sam shipped "}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"before I propose a ranking?"}}`,
		``,
		`event: content_block_stop`,
		`data: {"type":"content_block_stop","index":0}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":12}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseBody)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		context.Background(), &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err != nil {
		t.Fatalf("stream ranking chat failed: %v", err)
	}
	wantText := "Can you tell me more about what Sam shipped before I propose a ranking?"
	if replyText != wantText {
		t.Fatalf("expected accumulated reply text %q, got %q", wantText, replyText)
	}
	if proposedRanking != nil {
		t.Fatalf("expected no proposed ranking when the reply has no trailing JSON block, got %s", proposedRanking)
	}
	if out.Len() == 0 {
		t.Fatal("expected streamed chunks to still be written to the writer")
	}
}

// TestStreamRankingChatMalformedSSEEvent covers the API sending a
// content_block_delta whose data payload isn't valid JSON. The SDK's stream
// decoder surfaces this as a stream error (via stream.Err()); the method
// must return that error rather than panicking or silently returning a
// partial/garbage reply.
func TestStreamRankingChatMalformedSSEEvent(t *testing.T) {
	sseBody := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_bad","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Partial reply before things break. "}}`,
		``,
		`event: content_block_delta`,
		`data: {this is not valid json at all}`,
		``,
		// A trailing blank line is required so the SSE decoder actually
		// dispatches the malformed event above instead of leaving it
		// buffered as an incomplete trailing record.
		``,
	}, "\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, sseBody)
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		context.Background(), &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err == nil {
		t.Fatal("expected a malformed SSE event to surface as an error")
	}
	if replyText != "" {
		t.Fatalf("expected no reply text to be returned on a stream error, got %q", replyText)
	}
	if proposedRanking != nil {
		t.Fatalf("expected no proposed ranking to be returned on a stream error, got %s", proposedRanking)
	}
}

// TestStreamRankingChatNetworkError covers the underlying HTTP transport
// failing outright (connection refused) before any bytes are streamed —
// the method must return the error rather than panicking on a nil/partial
// stream.
func TestStreamRankingChatNetworkError(t *testing.T) {
	// Stand up a server just to mint a valid-looking, but now-dead, URL —
	// closing it immediately means every request to it fails to connect.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := server.URL
	server.Close()

	client := NewClient("test-key", option.WithBaseURL(deadURL))

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		context.Background(), &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err == nil {
		t.Fatal("expected a connection failure to surface as an error")
	}
	if replyText != "" {
		t.Fatalf("expected no reply text to be returned on a network error, got %q", replyText)
	}
	if proposedRanking != nil {
		t.Fatalf("expected no proposed ranking to be returned on a network error, got %s", proposedRanking)
	}
	if out.Len() != 0 {
		t.Fatalf("expected nothing written to the writer when the request never streamed, got %q", out.String())
	}
}

// TestStreamRankingChatContextCancellationMidStream covers the caller
// cancelling ctx while the SSE stream is still open (e.g. the admin closes
// the chat tab, or the HTTP handler's request context is cancelled). The
// method must return promptly with an error that reports the cancellation,
// rather than hanging or panicking.
func TestStreamRankingChatContextCancellationMidStream(t *testing.T) {
	started := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("test server ResponseWriter does not support flushing")
		}
		fmt.Fprint(w, strings.Join([]string{
			`event: message_start`,
			`data: {"type":"message_start","message":{"id":"msg_slow","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`,
			``,
		}, "\n"))
		flusher.Flush()
		close(started)

		// Hold the connection open until the client cancels (or a generous
		// timeout, so a bug can't hang the test suite indefinitely).
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	defer server.Close()

	client := NewClient("test-key", option.WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-started
		cancel()
	}()

	var out bytes.Buffer
	replyText, proposedRanking, err := client.StreamRankingChat(
		ctx, &out, "You are a ranking assistant.", nil, "Who stood out this cycle?",
	)
	if err == nil {
		t.Fatal("expected context cancellation mid-stream to surface as an error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected the error to report context cancellation, got %v (%T)", err, err)
	}
	if replyText != "" {
		t.Fatalf("expected no reply text to be returned when cancelled mid-stream, got %q", replyText)
	}
	if proposedRanking != nil {
		t.Fatalf("expected no proposed ranking to be returned when cancelled mid-stream, got %s", proposedRanking)
	}
}

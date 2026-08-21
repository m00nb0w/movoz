package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"
)

// itoa is a small local convenience wrapper matching the style used
// elsewhere in this package (strconv.Itoa) for building URL path segments.
func itoa(n int) string { return strconv.Itoa(n) }

// sseEvent renders one SSE event frame.
func sseEvent(event, data string) string {
	return "event: " + event + "\ndata: " + data + "\n\n"
}

// sseTextReply stands up an httptest server that streams a single Anthropic
// SSE message whose accumulated text is replyText, split into two deltas so
// the accumulation logic in aiclient.StreamRankingChat is genuinely
// exercised rather than trivially satisfied by one chunk.
func sseTextReply(t *testing.T, replyText string) *httptest.Server {
	t.Helper()
	escaped := strings.ReplaceAll(replyText, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	mid := len(escaped) / 2
	part1, part2 := escaped[:mid], escaped[mid:]

	body := sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"claude-opus-5","content":[],"stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":5,"output_tokens":0}}}`) +
		sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`) +
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"`+part1+`"}}`) +
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"`+part2+`"}}`) +
		sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`) +
		sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"output_tokens":8}}`) +
		sseEvent("message_stop", `{"type":"message_stop"}`)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte(body))
	}))
}

type aiChatTestFixture struct {
	db             *sql.DB
	engineerStore  *store.EngineerStore
	subAttrStore   *store.SubAttributeStore
	cycleStore     *store.CycleStore
	sessionStore   *store.AISessionStore
	metricStore    *store.MetricStore
	highlightStore *store.HighlightStore
	router         *gin.Engine
	engineerID     int
	subAttrID      int
	cycleID        int
}

// newAIChatFixture truncates the relevant tables and creates one engineer,
// one main/sub-attribute pair, and one cycle — the minimal roster every test
// below needs. It wires the router against whatever Anthropic-compatible
// SSE test server the caller passes in, so each test can control the
// simulated model reply independently.
func newAIChatFixture(t *testing.T, aiServerURL string) *aiChatTestFixture {
	t.Helper()
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "ai_ranking_sessions", "sub_attribute_rankings", "highlight_entries", "metric_snapshots", "sub_attributes", "main_attributes", "rating_cycles", "engineers")

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	sessionStore := store.NewAISessionStore(db)
	metricStore := store.NewMetricStore(db)
	highlightStore := store.NewHighlightStore(db)

	engineer, err := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("failed to create engineer: %v", err)
	}
	main, err := mainStore.Create("test_main_chat", "Test Main Chat")
	if err != nil {
		t.Fatalf("failed to create main attribute: %v", err)
	}
	sub, err := subStore.Create(main.ID, "Ownership", nil)
	if err != nil {
		t.Fatalf("failed to create sub attribute: %v", err)
	}
	cycle, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("failed to create cycle: %v", err)
	}

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(aiServerURL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIChatHandler(aiClient, sessionStore, engineerStore, metricStore, highlightStore, subStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions", h.Chat)

	return &aiChatTestFixture{
		db:             db,
		engineerStore:  engineerStore,
		subAttrStore:   subStore,
		cycleStore:     cycleStore,
		sessionStore:   sessionStore,
		metricStore:    metricStore,
		highlightStore: highlightStore,
		router:         r,
		engineerID:     engineer.ID,
		subAttrID:      sub.ID,
		cycleID:        cycle.ID,
	}
}

func (f *aiChatTestFixture) post(t *testing.T, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/"+itoa(f.cycleID)+"/ai-sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func TestAIChatHandlerStreamsAndPersistsSession(t *testing.T) {
	server := sseTextReply(t, "Tell me more about Sam's cycle.")
	defer server.Close()

	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "event: session") {
		t.Fatalf("expected the response to open with a session event carrying the session id, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Sam's cycle") {
		t.Fatalf("expected the streamed reply text in the response body, got %s", w.Body.String())
	}
}

// TestAIChatHandlerStreamingChunksReachClient asserts the reply is delivered
// as multiple SSE "data: " frames (not one buffered blob written only after
// the whole reply text is known) — the point of streaming is that the
// frontend chat UI sees the text arrive incrementally.
func TestAIChatHandlerStreamingChunksReachClient(t *testing.T) {
	server := sseTextReply(t, "First half then second half of a longer reply.")
	defer server.Close()

	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	dataFrames := strings.Count(w.Body.String(), "data: ")
	// One "data: " frame for the session event, plus at least two more for
	// the two text deltas sseTextReply splits the reply into.
	if dataFrames < 3 {
		t.Fatalf("expected at least 3 'data: ' frames (session + >=2 text deltas), got %d in body: %s", dataFrames, w.Body.String())
	}
}

// TestAIChatHandlerCreatesNewSessionRow verifies that omitting session_id on
// the first message actually inserts a new row in ai_ranking_sessions (not
// just that the handler returns 200) — and that the persisted transcript
// contains both the user's message and the assistant's reply.
func TestAIChatHandlerCreatesNewSessionRow(t *testing.T) {
	server := sseTextReply(t, "Here's what I found about the team.")
	defer server.Close()

	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	sessionID := extractSessionID(t, w.Body.String())
	session, err := f.sessionStore.GetByID(sessionID)
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}
	if session == nil {
		t.Fatalf("expected a session row to exist for id %d", sessionID)
	}
	if session.CycleID != f.cycleID || session.SubAttributeID != f.subAttrID {
		t.Fatalf("expected session to be scoped to cycle %d / sub-attribute %d, got cycle %d / sub-attribute %d", f.cycleID, f.subAttrID, session.CycleID, session.SubAttributeID)
	}

	var history []aiclient.ChatMessage
	if err := json.Unmarshal(session.Transcript, &history); err != nil {
		t.Fatalf("failed to unmarshal transcript: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected transcript to have 2 turns (user + assistant), got %d: %s", len(history), session.Transcript)
	}
	if history[0].Role != "user" || history[0].Content != "Who stood out this cycle?" {
		t.Fatalf("expected first turn to be the user's message, got %+v", history[0])
	}
	if history[1].Role != "assistant" || !strings.Contains(history[1].Content, "team") {
		t.Fatalf("expected second turn to be the assistant's reply, got %+v", history[1])
	}
}

// TestAIChatHandlerContinuesExistingSession sends a first message to create
// a session, then a second message carrying that session_id, and asserts
// the transcript is appended to (4 turns total) rather than replaced (which
// would leave only 2).
func TestAIChatHandlerContinuesExistingSession(t *testing.T) {
	server := sseTextReply(t, "First turn reply about the roster.")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w1 := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on first turn, got %d: %s", w1.Code, w1.Body.String())
	}
	sessionID := extractSessionID(t, w1.Body.String())
	server.Close()

	server2 := sseTextReply(t, "Second turn reply continuing the thread.")
	defer server2.Close()
	// Rebuild the fixture's router against the second server so the second
	// call streams from a live SSE source (the first server is now closed).
	f2 := newAIChatFixtureReusingDB(t, f, server2.URL)

	w2 := f2.post(t, map[string]interface{}{
		"session_id": sessionID,
		"message":    "What about code review turnaround?",
	})
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second turn, got %d: %s", w2.Code, w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), itoa(sessionID)) {
		t.Fatalf("expected the same session id to be echoed back on turn 2, got %s", w2.Body.String())
	}

	session, err := f.sessionStore.GetByID(sessionID)
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}
	var history []aiclient.ChatMessage
	if err := json.Unmarshal(session.Transcript, &history); err != nil {
		t.Fatalf("failed to unmarshal transcript: %v", err)
	}
	if len(history) != 4 {
		t.Fatalf("expected transcript to have 4 turns after two chat turns (appended, not replaced), got %d: %s", len(history), session.Transcript)
	}
	if history[0].Content != "Who stood out this cycle?" {
		t.Fatalf("expected turn 1 to survive in the appended transcript, got %+v", history[0])
	}
	if history[2].Content != "What about code review turnaround?" {
		t.Fatalf("expected turn 3 to be the second user message, got %+v", history[2])
	}
}

// TestAIChatHandlerPreservesProposedRankingOnConversationalTurn is the
// critical regression test called out for this task: AISessionStore.
// UpdateTranscript unconditionally overwrites proposed_ranking on every
// call, so a purely conversational turn (no trailing ```json block) MUST
// have the handler re-pass the session's existing proposed_ranking rather
// than nil — otherwise a previously proposed ranking is silently wiped out.
func TestAIChatHandlerPreservesProposedRankingOnConversationalTurn(t *testing.T) {
	proposalServer := sseTextReply(t, "Sam should rank first.\n```json\n{\"rationale\":\"Shipped the most.\",\"ranking\":[{\"engineer_id\":1,\"rank\":1}]}\n```")
	f := newAIChatFixture(t, proposalServer.URL)

	w1 := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on turn 1, got %d: %s", w1.Code, w1.Body.String())
	}
	proposalServer.Close()
	sessionID := extractSessionID(t, w1.Body.String())

	sessionAfterTurn1, err := f.sessionStore.GetByID(sessionID)
	if err != nil {
		t.Fatalf("failed to load session after turn 1: %v", err)
	}
	if len(sessionAfterTurn1.ProposedRanking) == 0 {
		t.Fatalf("expected turn 1 (which included a JSON block) to persist a proposed_ranking, got empty")
	}
	proposalAfterTurn1 := string(sessionAfterTurn1.ProposedRanking)

	conversationalServer := sseTextReply(t, "Sure — can you tell me more about Alex's contributions too?")
	defer conversationalServer.Close()
	f2 := newAIChatFixtureReusingDB(t, f, conversationalServer.URL)

	w2 := f2.post(t, map[string]interface{}{
		"session_id": sessionID,
		"message":    "What about Alex?",
	})
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on turn 2, got %d: %s", w2.Code, w2.Body.String())
	}

	sessionAfterTurn2, err := f.sessionStore.GetByID(sessionID)
	if err != nil {
		t.Fatalf("failed to load session after turn 2: %v", err)
	}
	if len(sessionAfterTurn2.ProposedRanking) == 0 {
		t.Fatalf("expected the proposed_ranking from turn 1 to survive turn 2 (a purely conversational reply with no JSON block), but it was wiped to empty/NULL")
	}
	if string(sessionAfterTurn2.ProposedRanking) != proposalAfterTurn1 {
		t.Fatalf("expected proposed_ranking to be unchanged after a conversational turn, before=%s after=%s", proposalAfterTurn1, sessionAfterTurn2.ProposedRanking)
	}

	// Belt-and-braces NF3 check: this handler must never write to
	// sub_attribute_rankings, no matter how many turns or proposals occur.
	assertNoSubAttributeRankingRows(t, f)
}

// TestAIChatHandlerNeverWritesSubAttributeRankings is a standalone NF3 check
// (in addition to the assertion embedded in the preservation test above):
// even a turn that proposes a ranking must only ever land in
// ai_ranking_sessions.proposed_ranking, never in sub_attribute_rankings —
// that table is only ever written by the separate accept endpoint (Task 27)
// on explicit admin confirmation.
func TestAIChatHandlerNeverWritesSubAttributeRankings(t *testing.T) {
	server := sseTextReply(t, "Sam should rank first.\n```json\n{\"rationale\":\"Shipped the most.\",\"ranking\":[{\"engineer_id\":1,\"rank\":1}]}\n```")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	assertNoSubAttributeRankingRows(t, f)
}

// TestAIChatHandlerAIStreamErrorSurfacesEventAndSkipsPersist covers the
// Anthropic call failing mid-flight (server closed / connection refused).
// The handler has already flushed the 200 status and the session event by
// the time this happens, so it cannot fall back to a JSON error response —
// it must instead emit an SSE error event, and it must not call
// UpdateTranscript with a bogus empty reply.
func TestAIChatHandlerAIStreamErrorSurfacesEventAndSkipsPersist(t *testing.T) {
	deadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	deadURL := deadServer.URL
	deadServer.Close()

	f := newAIChatFixture(t, deadURL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
		"message":          "Who stood out this cycle?",
	})

	// Headers/status were already committed (200) before the AI call was
	// attempted, so the failure must show up as an in-body SSE error event.
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (headers already flushed before the AI call), got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "event: error") {
		t.Fatalf("expected an SSE error event in the body when the AI call fails, got %s", w.Body.String())
	}

	sessionID := extractSessionID(t, w.Body.String())
	session, err := f.sessionStore.GetByID(sessionID)
	if err != nil {
		t.Fatalf("failed to load session: %v", err)
	}
	var history []aiclient.ChatMessage
	if err := json.Unmarshal(session.Transcript, &history); err != nil {
		t.Fatalf("failed to unmarshal transcript: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("expected the transcript to remain untouched (still '[]') when the AI call fails, got %d turns: %s", len(history), session.Transcript)
	}
}

func TestAIChatHandlerRejectsInvalidCycleID(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	body, _ := json.Marshal(map[string]interface{}{"sub_attribute_id": f.subAttrID, "message": "hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/not-a-number/ai-sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a non-numeric cycle id, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAIChatHandlerRejectsNonexistentCycle(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	body, _ := json.Marshal(map[string]interface{}{"sub_attribute_id": f.subAttrID, "message": "hi"})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/999999/ai-sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a cycle id that doesn't exist, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAIChatHandlerRejectsNonexistentSubAttribute(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": 999999,
		"message":          "hi",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a sub_attribute_id that doesn't exist, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAIChatHandlerRejectsMissingMessage(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"sub_attribute_id": f.subAttrID,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when message is missing, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAIChatHandlerRejectsMissingSubAttributeIDOnNewSession(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"message": "hi",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when sub_attribute_id is missing on a new session, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAIChatHandlerRejectsUnknownSessionID(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	w := f.post(t, map[string]interface{}{
		"session_id": 999999,
		"message":    "hi",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a session_id that doesn't exist, got %d: %s", w.Code, w.Body.String())
	}
}

// TestAIChatHandlerRejectsSessionFromDifferentCycle covers passing a
// session_id that exists but belongs to a different cycle than the one in
// the URL — this must be rejected rather than silently continuing a chat
// under the wrong cycle's route.
func TestAIChatHandlerRejectsSessionFromDifferentCycle(t *testing.T) {
	server := sseTextReply(t, "unused")
	defer server.Close()
	f := newAIChatFixture(t, server.URL)

	otherCycle, err := f.cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("failed to create second cycle: %v", err)
	}
	otherSession, err := f.sessionStore.Create(otherCycle.ID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session under other cycle: %v", err)
	}

	w := f.post(t, map[string]interface{}{
		"session_id": otherSession.ID,
		"message":    "hi",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a session_id that belongs to a different cycle, got %d: %s", w.Code, w.Body.String())
	}
}

// extractSessionID pulls the numeric session_id out of the leading
// `event: session\ndata: {"session_id":N}` frame.
func extractSessionID(t *testing.T, body string) int {
	t.Helper()
	const marker = `"session_id":`
	idx := strings.Index(body, marker)
	if idx == -1 {
		t.Fatalf("expected a session_id in the response body, got %s", body)
	}
	rest := body[idx+len(marker):]
	end := strings.IndexAny(rest, "}\n")
	if end == -1 {
		t.Fatalf("could not parse session_id out of body: %s", body)
	}
	id, err := strconv.Atoi(strings.TrimSpace(rest[:end]))
	if err != nil {
		t.Fatalf("failed to parse session_id %q: %v", rest[:end], err)
	}
	return id
}

// newAIChatFixtureReusingDB builds a fresh router (pointed at a different AI
// server URL, e.g. for a second chat turn) against the same already-seeded
// database/stores as an existing fixture, without re-truncating tables.
func newAIChatFixtureReusingDB(t *testing.T, base *aiChatTestFixture, aiServerURL string) *aiChatTestFixture {
	t.Helper()
	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(aiServerURL))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIChatHandler(aiClient, base.sessionStore, base.engineerStore, base.metricStore, base.highlightStore, base.subAttrStore, base.cycleStore)
	r.POST("/api/cycles/:id/ai-sessions", h.Chat)
	return &aiChatTestFixture{
		db:             base.db,
		engineerStore:  base.engineerStore,
		subAttrStore:   base.subAttrStore,
		cycleStore:     base.cycleStore,
		sessionStore:   base.sessionStore,
		metricStore:    base.metricStore,
		highlightStore: base.highlightStore,
		router:         r,
		engineerID:     base.engineerID,
		subAttrID:      base.subAttrID,
		cycleID:        base.cycleID,
	}
}

func assertNoSubAttributeRankingRows(t *testing.T, f *aiChatTestFixture) {
	t.Helper()
	// Reach into the underlying db directly to query sub_attribute_rankings
	// (no store in this test wraps that table) and confirm NF3 holds:
	// nothing this handler does ever lands a row there.
	rows, err := f.db.Query("SELECT count(*) FROM sub_attribute_rankings")
	if err != nil {
		t.Fatalf("failed to query sub_attribute_rankings: %v", err)
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			t.Fatalf("failed to scan count: %v", err)
		}
	}
	if count != 0 {
		t.Fatalf("expected zero rows in sub_attribute_rankings (NF3: AI proposals must never silently become official), got %d", count)
	}
}

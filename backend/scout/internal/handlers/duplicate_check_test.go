package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/gin-gonic/gin"
)

// TestDuplicateCheckHandlerReturnsFlagOnMatch covers the core happy path: a
// genuine semantic duplicate is detected against an existing entry and the
// handler surfaces the flag from aiclient.CheckDuplicate as-is over HTTP.
func TestDuplicateCheckHandlerReturnsFlagOnMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_dup2", "type": "message", "role": "assistant", "model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":true,\"matched_entry_id\":1,\"similarity_note\":\"same missed deadline\"}"}],
			"stop_reason": "end_turn", "stop_sequence": null,
			"usage": {"input_tokens": 5, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	highlightStore.Create(e1.ID, "lowlight", "Slipped the Q1 deadline and didn't raise it in standup")

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Missed the Q1 deadline without flagging it early"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if !result.IsDuplicate {
		t.Fatal("expected the duplicate flag to be true")
	}
	if result.MatchedEntryID == nil || *result.MatchedEntryID != 1 {
		t.Fatalf("expected matched entry id 1, got %v", result.MatchedEntryID)
	}
}

// TestDuplicateCheckHandlerNoMatchFound covers the case where the AI call
// succeeds but genuinely finds no duplicate among existing entries — the
// negative counterpart to the match test, distinct from the AI-failure
// degradation path below (this is a successful AI response saying "no").
func TestDuplicateCheckHandlerNoMatchFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_nodup", "type": "message", "role": "assistant", "model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":false,\"matched_entry_id\":null,\"similarity_note\":\"distinct topic\"}"}],
			"stop_reason": "end_turn", "stop_sequence": null,
			"usage": {"input_tokens": 5, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	highlightStore.Create(e1.ID, "highlight", "Shipped the onboarding flow ahead of schedule")

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Fixed a flaky CI test suite"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false for a genuinely distinct entry")
	}
	if result.MatchedEntryID != nil {
		t.Fatalf("expected no matched entry id, got %v", *result.MatchedEntryID)
	}
}

// TestDuplicateCheckHandlerZeroExistingEntries covers an engineer's very
// first highlight/lowlight entry: highlightStore.List returns an empty slice,
// which the handler must still pass through to CheckDuplicate (Task 28
// round-trips it through the API rather than short-circuiting locally) and
// return a normal 200/not-a-duplicate result, not an error.
func TestDuplicateCheckHandlerZeroExistingEntries(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"id": "msg_first", "type": "message", "role": "assistant", "model": "claude-opus-5",
			"content": [{"type": "text", "text": "{\"is_duplicate\":false,\"matched_entry_id\":null,\"similarity_note\":\"no existing entries to compare\"}"}],
			"stop_reason": "end_turn", "stop_sequence": null,
			"usage": {"input_tokens": 5, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	// No highlight entries created for e1 at all.

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "First highlight ever logged for this engineer"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !called {
		t.Fatal("expected the AI call to still be made with zero existing entries (no local short-circuit)")
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false when there is nothing to compare against")
	}
}

// TestDuplicateCheckHandlerDegradesGracefullyOnAIFailure covers the single
// most important behavior in this task (F14, NF3): if the AI call fails, the
// endpoint must still return 200 with is_duplicate=false, never a 500 or
// other error status that a frontend could mistake for the save itself
// failing. The AI failure is simulated with a server that always errors,
// exercising the real aiclient.CheckDuplicate error path end-to-end rather
// than mocking it out.
func TestDuplicateCheckHandlerDegradesGracefullyOnAIFailure(t *testing.T) {
	// A server that always errors simulates the AI call failing (F14: the
	// save must proceed without a flag rather than being blocked).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Must still be 200 with no duplicate flag — never a blocking error.
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (degraded, not blocked), got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false when the AI call fails")
	}
	if result.Note == "" {
		t.Fatal("expected a note explaining the degraded result")
	}
}

// TestDuplicateCheckHandlerAITimeout covers the timeout variant of AI-call
// failure specifically called out by NF3 ("fails/times out"): the AI server
// never responds, the handler's own context deadline should fire first, and
// the response must still degrade to 200/is_duplicate=false rather than
// hanging or erroring.
func TestDuplicateCheckHandlerAITimeout(t *testing.T) {
	// Sleep past the client's request timeout (set below) rather than
	// blocking on r.Context().Done(): whether the server observes client
	// disconnection is unspecified/racy across transports, and blocking
	// indefinitely on it would hang httptest.Server.Close() forever if the
	// context is never canceled server-side. A bounded sleep guarantees the
	// handler — and thus Close() — returns promptly either way, while still
	// safely outlasting the client's 50ms timeout below.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	aiClient := aiclient.NewClient(
		"test-key",
		option.WithBaseURL(server.URL),
		option.WithRequestTimeout(50*time.Millisecond),
	)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(w, req)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not return within 5s of the AI call timing out")
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (degraded, not blocked) on timeout, got %d: %s", w.Code, w.Body.String())
	}
	var result aiclient.DuplicateCheckResult
	json.Unmarshal(w.Body.Bytes(), &result)
	if result.IsDuplicate {
		t.Fatal("expected IsDuplicate false when the AI call times out")
	}
}

// TestDuplicateCheckHandlerEngineerNotFound covers an unknown engineer id:
// the handler must 404 rather than silently treating it as "no existing
// entries" and calling the AI anyway.
func TestDuplicateCheckHandlerEngineerNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("AI client should not be called for an unknown engineer")
	}))
	defer server.Close()

	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)

	aiClient := aiclient.NewClient("test-key", option.WithBaseURL(server.URL))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	body, _ := json.Marshal(map[string]string{"body": "Anything"})
	req := httptest.NewRequest(http.MethodPost, "/api/engineers/9999/highlights/check-duplicate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown engineer, got %d: %s", w.Code, w.Body.String())
	}
}

// TestDuplicateCheckHandlerMissingBody covers request validation: an empty
// JSON body (no "body" field) must be rejected with 400, not silently sent
// to the AI as an empty string.
func TestDuplicateCheckHandlerMissingBody(t *testing.T) {
	db := setupTestDBForHandlers(t)
	if _, err := db.Exec("TRUNCATE highlight_entries, engineers RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("failed to truncate: %v", err)
	}
	engineerStore := store.NewEngineerStore(db)
	highlightStore := store.NewHighlightStore(db)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())

	aiClient := aiclient.NewClient("test-key")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDuplicateCheckHandler(aiClient, highlightStore, engineerStore)
	r.POST("/api/engineers/:id/highlights/check-duplicate", h.Check)

	req := httptest.NewRequest(http.MethodPost, "/api/engineers/"+itoa(e1.ID)+"/highlights/check-duplicate", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing body field, got %d: %s", w.Code, w.Body.String())
	}
}

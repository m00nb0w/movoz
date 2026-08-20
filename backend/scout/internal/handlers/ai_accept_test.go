package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type aiAcceptTestFixture struct {
	db            *sql.DB
	engineerStore *store.EngineerStore
	subAttrStore  *store.SubAttributeStore
	cycleStore    *store.CycleStore
	sessionStore  *store.AISessionStore
	rankingStore  *store.RankingStore
	router        *gin.Engine
	e1, e2        int
	subAttrID     int
	cycleID       int
}

// newAIAcceptFixture truncates the relevant tables and creates a two-engineer
// active roster, one sub-attribute, and one cycle — enough to exercise both
// a valid 1..2 permutation and an invalid submission against it.
func newAIAcceptFixture(t *testing.T) *aiAcceptTestFixture {
	t.Helper()
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	sessionStore := store.NewAISessionStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, err := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("failed to create engineer 1: %v", err)
	}
	e2, err := engineerStore.Create("Bailey", nil, nil, nil, time.Now())
	if err != nil {
		t.Fatalf("failed to create engineer 2: %v", err)
	}
	main, err := mainStore.Create("test_main_accept", "Test Main Accept")
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

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIAcceptHandler(sessionStore, rankingStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions/:sessionId/accept", h.Accept)

	return &aiAcceptTestFixture{
		db:            db,
		engineerStore: engineerStore,
		subAttrStore:  subStore,
		cycleStore:    cycleStore,
		sessionStore:  sessionStore,
		rankingStore:  rankingStore,
		router:        r,
		e1:            e1.ID,
		e2:            e2.ID,
		subAttrID:     sub.ID,
		cycleID:       cycle.ID,
	}
}

func (f *aiAcceptTestFixture) accept(t *testing.T, sessionID int, rankings []map[string]int) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{"rankings": rankings})
	url := "/api/cycles/" + itoa(f.cycleID) + "/ai-sessions/" + itoa(sessionID) + "/accept"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func TestAIAcceptHandlerPersistsEditedRanking(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"ai_ranking_sessions", "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	sessionStore := store.NewAISessionStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_accept", "Test Main Accept")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	session, _ := sessionStore.Create(cycle.ID, sub.ID)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIAcceptHandler(sessionStore, rankingStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions/:sessionId/accept", h.Accept)

	// The admin may have edited the AI's proposed ranking before confirming
	// (NF3) — the request body, not session.ProposedRanking, is the source
	// of truth for what gets persisted.
	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/" + itoa(cycle.ID) + "/ai-sessions/" + itoa(session.ID) + "/accept"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	if len(rankings) != 1 || rankings[0].EngineerID != e1.ID {
		t.Fatalf("expected the accepted ranking to be persisted, got %+v", rankings)
	}
}

// TestAIAcceptHandlerPersistsRequestBodyNotStoredProposal is the core NF3
// regression test: the session's stored proposed_ranking is deliberately set
// to one ordering (as if the AI proposed engineer 1 first), and the accept
// request body carries a different ordering (as if the admin swapped the
// order before confirming). The handler must persist the body's ordering —
// silently falling back to session.ProposedRanking would be the exact bug
// this endpoint exists to prevent.
func TestAIAcceptHandlerPersistsRequestBodyNotStoredProposal(t *testing.T) {
	f := newAIAcceptFixture(t)

	session, err := f.sessionStore.Create(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	aiProposal, _ := json.Marshal([]map[string]int{
		{"engineer_id": f.e1, "rank": 1},
		{"engineer_id": f.e2, "rank": 2},
	})
	if err := f.sessionStore.UpdateTranscript(session.ID, []byte("[]"), aiProposal); err != nil {
		t.Fatalf("failed to seed proposed_ranking: %v", err)
	}

	// Admin-edited ranking: the reverse of the AI's stored proposal.
	editedRankings := []map[string]int{
		{"engineer_id": f.e2, "rank": 1},
		{"engineer_id": f.e1, "rank": 2},
	}
	w := f.accept(t, session.ID, editedRankings)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := f.rankingStore.GetByCycleAndSubAttribute(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	if len(rankings) != 2 {
		t.Fatalf("expected 2 persisted rankings, got %d: %+v", len(rankings), rankings)
	}
	byEngineer := map[int]int{}
	for _, r := range rankings {
		byEngineer[r.EngineerID] = r.Rank
	}
	if byEngineer[f.e2] != 1 || byEngineer[f.e1] != 2 {
		t.Fatalf("expected the edited (reversed) ranking to be persisted (e2=1, e1=2), got %+v — session.ProposedRanking was %s", byEngineer, aiProposal)
	}
}

// TestAIAcceptHandlerPersistsUneditedAIProposal covers the companion
// happy-path case: when the admin accepts the AI's proposal as-is (the
// request body matches what the AI proposed), it must still be persisted
// via the normal SubmitRanking path.
func TestAIAcceptHandlerPersistsUneditedAIProposal(t *testing.T) {
	f := newAIAcceptFixture(t)

	session, err := f.sessionStore.Create(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	aiProposal, _ := json.Marshal([]map[string]int{
		{"engineer_id": f.e1, "rank": 1},
		{"engineer_id": f.e2, "rank": 2},
	})
	if err := f.sessionStore.UpdateTranscript(session.ID, []byte("[]"), aiProposal); err != nil {
		t.Fatalf("failed to seed proposed_ranking: %v", err)
	}

	sameRankings := []map[string]int{
		{"engineer_id": f.e1, "rank": 1},
		{"engineer_id": f.e2, "rank": 2},
	}
	w := f.accept(t, session.ID, sameRankings)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := f.rankingStore.GetByCycleAndSubAttribute(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	byEngineer := map[int]int{}
	for _, r := range rankings {
		byEngineer[r.EngineerID] = r.Rank
	}
	if byEngineer[f.e1] != 1 || byEngineer[f.e2] != 2 {
		t.Fatalf("expected the unedited AI proposal to be persisted (e1=1, e2=2), got %+v", byEngineer)
	}
}

// TestAIAcceptHandlerRejectsInvalidPermutation confirms an invalid submission
// (here, missing one of the two active engineers) is rejected with 400 by
// delegating to RankingStore.SubmitRanking's existing validator
// (scoring.ValidatePermutation via Task 10) — this handler must not
// reimplement or bypass that check — and that nothing is persisted as a
// result of the rejected call.
func TestAIAcceptHandlerRejectsInvalidPermutation(t *testing.T) {
	f := newAIAcceptFixture(t)

	session, err := f.sessionStore.Create(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Only one of the two active engineers is ranked — not a valid 1..N
	// permutation of the active roster.
	invalidRankings := []map[string]int{
		{"engineer_id": f.e1, "rank": 1},
	}
	w := f.accept(t, session.ID, invalidRankings)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an invalid permutation, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := f.rankingStore.GetByCycleAndSubAttribute(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	if len(rankings) != 0 {
		t.Fatalf("expected no rankings persisted for a rejected submission, got %+v", rankings)
	}
}

func TestAIAcceptHandlerSessionNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	sessionStore := store.NewAISessionStore(db)
	rankingStore := store.NewRankingStore(db, store.NewEngineerStore(db))
	cycleStore := store.NewCycleStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewAIAcceptHandler(sessionStore, rankingStore, cycleStore)
	r.POST("/api/cycles/:id/ai-sessions/:sessionId/accept", h.Accept)

	body, _ := json.Marshal(map[string]interface{}{"rankings": []map[string]int{}})
	req := httptest.NewRequest(http.MethodPost, "/api/cycles/1/ai-sessions/999999/accept", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestAIAcceptHandlerRejectsSessionFromDifferentCycle covers a session that
// exists but belongs to a different cycle than the one in the URL — the
// same cross-cycle guard the chat endpoint (Task 26) enforces, needed here
// too since otherwise an accept call could persist a ranking under the
// wrong cycle using this session's sub-attribute.
func TestAIAcceptHandlerRejectsSessionFromDifferentCycle(t *testing.T) {
	f := newAIAcceptFixture(t)

	otherCycle, err := f.cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("failed to create second cycle: %v", err)
	}
	otherSession, err := f.sessionStore.Create(otherCycle.ID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session under other cycle: %v", err)
	}

	// f.cycleID is the fixture's cycle, but otherSession belongs to
	// otherCycle — accepting it through f.cycleID's URL must 404.
	w := f.accept(t, otherSession.ID, []map[string]int{
		{"engineer_id": f.e1, "rank": 1},
		{"engineer_id": f.e2, "rank": 2},
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a session belonging to a different cycle, got %d: %s", w.Code, w.Body.String())
	}

	rankings, err := f.rankingStore.GetByCycleAndSubAttribute(otherCycle.ID, f.subAttrID)
	if err != nil {
		t.Fatalf("get rankings failed: %v", err)
	}
	if len(rankings) != 0 {
		t.Fatalf("expected no rankings persisted when the cycle/session mismatch is rejected, got %+v", rankings)
	}
}

// TestAIAcceptHandlerRejectsNonexistentCycle covers a URL cycle id that
// doesn't exist at all.
func TestAIAcceptHandlerRejectsNonexistentCycle(t *testing.T) {
	f := newAIAcceptFixture(t)

	session, err := f.sessionStore.Create(f.cycleID, f.subAttrID)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{
			{"engineer_id": f.e1, "rank": 1},
			{"engineer_id": f.e2, "rank": 2},
		},
	})
	url := "/api/cycles/999999/ai-sessions/" + itoa(session.ID) + "/accept"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a nonexistent cycle id, got %d: %s", w.Code, w.Body.String())
	}
}

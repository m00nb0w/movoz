package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestRankingHandlerSubmit(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main3"), "Test Main 3")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rankings []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(rankings) != 1 {
		t.Fatalf("expected 1 ranking in response, got %d", len(rankings))
	}
	// Sole active engineer, N=1 -> RankToScore special-cases to 100.
	if score, _ := rankings[0]["score"].(float64); score != 100 {
		t.Fatalf("expected score 100 for sole ranked engineer, got %v", rankings[0]["score"])
	}
}

// TestRankingHandlerSubmitScoresMatchFormula exercises the N=2 case through
// the HTTP layer and checks both scores against Task 9's exact formula
// (rank 1 -> 100, rank N -> 50), not just "some score got saved".
func TestRankingHandlerSubmitScoresMatchFormula(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main_formula"), "Test Main Formula")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{
			{"engineer_id": e1.ID, "rank": 1},
			{"engineer_id": e2.ID, "rank": 2},
		},
	})
	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rankings []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	wantScore := map[float64]float64{float64(e1.ID): 100, float64(e2.ID): 50}
	if len(rankings) != 2 {
		t.Fatalf("expected 2 rankings, got %d", len(rankings))
	}
	for _, rk := range rankings {
		engID, _ := rk["engineer_id"].(float64)
		want, ok := wantScore[engID]
		if !ok {
			t.Fatalf("unexpected engineer_id %v in response", rk["engineer_id"])
		}
		if got, _ := rk["score"].(float64); got != want {
			t.Fatalf("engineer %v: expected score %v, got %v", engID, want, rk["score"])
		}
	}
}

// TestRankingHandlerSubmitRejectsInvalidPermutation submits tied ranks
// through the real HTTP handler (not a direct scoring.ValidatePermutation
// unit test) to prove the handler actually delegates to Task 10's validator
// and surfaces its rejection as a 400.
func TestRankingHandlerSubmitRejectsInvalidPermutation(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main_invalid"), "Test Main Invalid")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{
			{"engineer_id": e1.ID, "rank": 1},
			{"engineer_id": e2.ID, "rank": 1},
		},
	})
	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for tied ranks, got %d: %s", w.Code, w.Body.String())
	}

	// Confirm nothing was persisted despite the rejected request.
	stored, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(stored) != 0 {
		t.Fatalf("expected no rows persisted after a rejected submission, got %d", len(stored))
	}
}

// TestRankingHandlerSubmitReplacesOnResubmit submits, then resubmits a
// different ranking for the same cycle+sub-attribute through the handler,
// and confirms the old ranks are gone (not merged) via the Get endpoint.
func TestRankingHandlerSubmitReplacesOnResubmit(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main_resub_h"), "Test Main Resub H")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)
	r.GET("/api/cycles/:id/sub-attributes/:subId/ranking", h.Get)

	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"

	submit := func(rankings []map[string]int) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]interface{}{"rankings": rankings})
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	if w := submit([]map[string]int{
		{"engineer_id": e1.ID, "rank": 1},
		{"engineer_id": e2.ID, "rank": 2},
	}); w.Code != http.StatusOK {
		t.Fatalf("first submit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if w := submit([]map[string]int{
		{"engineer_id": e1.ID, "rank": 2},
		{"engineer_id": e2.ID, "rank": 1},
	}); w.Code != http.StatusOK {
		t.Fatalf("resubmit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, url, nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)
	if getW.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d: %s", getW.Code, getW.Body.String())
	}

	var rankings []map[string]interface{}
	if err := json.Unmarshal(getW.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("failed to unmarshal get response: %v", err)
	}
	if len(rankings) != 2 {
		t.Fatalf("expected exactly 2 rankings after resubmit (no leftover rows from the first submission), got %d", len(rankings))
	}
	for _, rk := range rankings {
		engID, _ := rk["engineer_id"].(float64)
		rank, _ := rk["rank"].(float64)
		if int(engID) == e1.ID && int(rank) != 2 {
			t.Fatalf("expected e1 to now be rank 2, got %v", rank)
		}
		if int(engID) == e2.ID && int(rank) != 1 {
			t.Fatalf("expected e2 to now be rank 1, got %v", rank)
		}
	}
}

func TestRankingHandlerSubmitUnknownCycle(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main_unknown_cycle"), "Test Main Unknown Cycle")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/999999/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown cycle, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRankingHandlerSubmitUnknownSubAttribute(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/999999/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown sub attribute, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRankingHandlerSubmitInvalidCycleID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{"rankings": []map[string]int{}})
	req := httptest.NewRequest(http.MethodPut, "/api/cycles/not-a-number/sub-attributes/1/ranking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-numeric cycle id, got %d: %s", w.Code, w.Body.String())
	}
}

// TestRankingHandlerSubmitInvalidSubAttributeID exercises the second,
// independent int-parsing branch in parseCycleAndSubAttribute — a valid
// numeric cycle id paired with a non-numeric subId — which
// TestRankingHandlerSubmitInvalidCycleID does not reach.
func TestRankingHandlerSubmitInvalidSubAttributeID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{"rankings": []map[string]int{}})
	req := httptest.NewRequest(http.MethodPut, "/api/cycles/1/sub-attributes/not-a-number/ranking", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-numeric sub attribute id, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["error"] != "invalid sub attribute id" {
		t.Fatalf("expected the sub-attribute-id-specific error message, got %q", resp["error"])
	}
}

func TestRankingHandlerSubmitMissingRankings(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	main, _ := mainStore.Create(uniqueMainKey("test_main_missing_body"), "Test Main Missing Body")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	body, _ := json.Marshal(map[string]interface{}{})
	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing rankings body, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRankingHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueMainKey("test_main_get"), "Test Main Get")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.GET("/api/cycles/:id/sub-attributes/:subId/ranking", h.Get)
	r.PUT("/api/cycles/:id/sub-attributes/:subId/ranking", h.Submit)

	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"

	putBody, _ := json.Marshal(map[string]interface{}{
		"rankings": []map[string]int{{"engineer_id": e1.ID, "rank": 1}},
	})
	putReq := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putW := httptest.NewRecorder()
	r.ServeHTTP(putW, putReq)
	if putW.Code != http.StatusOK {
		t.Fatalf("submit setup: expected 200, got %d: %s", putW.Code, putW.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rankings []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(rankings) != 1 {
		t.Fatalf("expected 1 ranking, got %d", len(rankings))
	}
	if engID, _ := rankings[0]["engineer_id"].(float64); int(engID) != e1.ID {
		t.Fatalf("expected engineer_id %d, got %v", e1.ID, rankings[0]["engineer_id"])
	}
}

func TestRankingHandlerGetEmpty(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	main, _ := mainStore.Create(uniqueMainKey("test_main_get_empty"), "Test Main Get Empty")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.GET("/api/cycles/:id/sub-attributes/:subId/ranking", h.Get)

	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rankings []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &rankings); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(rankings) != 0 {
		t.Fatalf("expected 0 rankings for a cycle+sub-attribute with no submission, got %d", len(rankings))
	}
}

func TestRankingHandlerGetUnknownCycle(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	main, _ := mainStore.Create(uniqueMainKey("test_main_get_unknown_cycle"), "Test Main Get Unknown Cycle")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.GET("/api/cycles/:id/sub-attributes/:subId/ranking", h.Get)

	url := "/api/cycles/999999/sub-attributes/" + strconv.Itoa(sub.ID) + "/ranking"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown cycle, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRankingHandlerGetUnknownSubAttribute(t *testing.T) {
	db := setupTestDBForHandlers(t)
	// Only truncate tables this test owns outright: engineers (so
	// ValidatePermutation's active-roster check starts from a known, exact
	// set) and sub_attribute_rankings (this test's own rows). main_attributes
	// carries migration-seeded rows nothing here reseeds, and
	// sub_attributes/rating_cycles have no seed data but no need to be
	// wiped either, since every test creates its own fresh rows via unique
	// keys — so none of those three are truncated.
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewRankingHandler(rankingStore, cycleStore, subStore)
	r.GET("/api/cycles/:id/sub-attributes/:subId/ranking", h.Get)

	url := "/api/cycles/" + strconv.Itoa(cycle.ID) + "/sub-attributes/999999/ranking"
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown sub attribute, got %d: %s", w.Code, w.Body.String())
	}
}

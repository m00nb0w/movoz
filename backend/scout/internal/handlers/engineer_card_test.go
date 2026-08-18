package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func setupEngineerCardTestRouter(t *testing.T) (*gin.Engine, *store.EngineerStore, *store.MainAttributeStore, *store.SubAttributeStore, *store.CycleStore, *store.RankingStore, *store.ScoreStore) {
	gin.SetMode(gin.TestMode)
	db := setupTestDBForHandlers(t)

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	r := gin.New()
	handler := NewEngineerCardHandler(scoreStore, engineerStore)
	r.GET("/api/engineers/:id/card", handler.Card)
	r.GET("/api/engineers/:id/trend", handler.Trend)

	return r, engineerStore, mainStore, subStore, cycleStore, rankingStore, scoreStore
}

// TestEngineerCardHandlerValid tests the happy path: engineer with rankings
// in a cycle gets a card with hand-computed expected scores.
func TestEngineerCardHandlerValid(t *testing.T) {
	r, engineerStore, mainStore, subStore, cycleStore, rankingStore, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	e1, _ := engineerStore.Create("Alice", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Bob", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("leadership", "Leadership")
	sub1, _ := subStore.Create(main.ID, "Decision Making", nil)
	sub2, _ := subStore.Create(main.ID, "Communication", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Alice ranks 1st on both sub-attributes -> both score 100 -> main avg 100 -> overall 100.
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub1.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub2.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub2 ranking failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/%d/card?cycleId=%d", e1.ID, cycle.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var card models.EngineerCard
	if err := json.Unmarshal(w.Body.Bytes(), &card); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if card.Engineer.ID != e1.ID || card.Engineer.Name != "Alice" {
		t.Fatalf("unexpected engineer in card: %+v", card.Engineer)
	}
	if card.CycleID != cycle.ID {
		t.Fatalf("unexpected cycle ID in card: %d", card.CycleID)
	}
	if card.Overall == nil || *card.Overall != 100 {
		t.Fatalf("expected overall score 100, got %v", card.Overall)
	}
	if len(card.MainAttributes) != 1 {
		t.Fatalf("expected 1 main attribute score, got %d", len(card.MainAttributes))
	}
	if card.MainAttributes[0].MainAttributeID != main.ID || card.MainAttributes[0].Score != 100 {
		t.Fatalf("expected main attribute %d with score 100, got %+v", main.ID, card.MainAttributes[0])
	}
}

// TestEngineerCardHandlerNotFound tests the not-found case: requesting a
// card for a non-existent engineer returns 404.
func TestEngineerCardHandlerNotFound(t *testing.T) {
	r, _, _, _, cycleStore, _, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/9999/card?cycleId=%d", cycle.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-existent engineer, got %d", w.Code)
	}
}

// TestEngineerCardHandlerInvalidEngineerID tests the invalid engineer ID case.
func TestEngineerCardHandlerInvalidEngineerID(t *testing.T) {
	r, _, _, _, _, _, _ := setupEngineerCardTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/not-a-number/card?cycleId=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid engineer ID, got %d", w.Code)
	}
}

// TestEngineerCardHandlerMissingCycleId tests the missing cycleId query param case.
func TestEngineerCardHandlerMissingCycleId(t *testing.T) {
	r, engineerStore, _, _, _, _, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	e1, _ := engineerStore.Create("Charlie", nil, nil, nil, time.Now())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/%d/card", e1.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing cycleId, got %d", w.Code)
	}
}

// TestEngineerCardHandlerInvalidCycleId tests the invalid cycleId query param case.
func TestEngineerCardHandlerInvalidCycleId(t *testing.T) {
	r, engineerStore, _, _, _, _, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	e1, _ := engineerStore.Create("Dave", nil, nil, nil, time.Now())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/%d/card?cycleId=not-a-number", e1.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid cycleId, got %d", w.Code)
	}
}

// TestEngineerTrendHandlerValid tests the trend endpoint: returns scores
// across all past cycles an engineer has rankings in, in chronological order.
func TestEngineerTrendHandlerValid(t *testing.T) {
	r, engineerStore, mainStore, subStore, cycleStore, rankingStore, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	e1, _ := engineerStore.Create("Emma", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Frank", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("technical", "Technical Skills")
	sub, _ := subStore.Create(main.ID, "Code Review", nil)

	// Create three cycles with different dates, in non-chronological order to test
	// that sorting happens correctly (not relying on insertion/ID order).
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	cycle1Start := base
	cycle2Start := base.Add(24 * time.Hour)
	cycle3Start := base.Add(2 * 24 * time.Hour)

	// Create in order: cycle3, cycle2, cycle1 (reverse chronological)
	cycle3, _ := cycleStore.Create(cycle3Start, cycle3Start.AddDate(0, 0, 14))
	cycle2, _ := cycleStore.Create(cycle2Start, cycle2Start.AddDate(0, 0, 14))
	cycle1, _ := cycleStore.Create(cycle1Start, cycle1Start.AddDate(0, 0, 14))

	// Submit rankings: Emma ranks 1st in cycles 1, 3; ranks 2nd in cycle 2.
	if _, err := rankingStore.SubmitRanking(cycle3.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit cycle3 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle1.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit cycle1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle2.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit cycle2 ranking failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/%d/trend", e1.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var trend []models.TrendPoint
	if err := json.Unmarshal(w.Body.Bytes(), &trend); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(trend) != 3 {
		t.Fatalf("expected 3 trend points, got %d", len(trend))
	}

	// Verify chronological ordering.
	if trend[0].CycleID != cycle1.ID {
		t.Fatalf("expected first trend point to be cycle1 (ID %d), got cycle ID %d", cycle1.ID, trend[0].CycleID)
	}
	if trend[1].CycleID != cycle2.ID {
		t.Fatalf("expected second trend point to be cycle2 (ID %d), got cycle ID %d", cycle2.ID, trend[1].CycleID)
	}
	if trend[2].CycleID != cycle3.ID {
		t.Fatalf("expected third trend point to be cycle3 (ID %d), got cycle ID %d", cycle3.ID, trend[2].CycleID)
	}

	// Verify scores: Emma was rank 1 in cycles 1 and 3 (100), rank 2 in cycle 2 (50).
	if trend[0].Overall == nil || *trend[0].Overall != 100 {
		t.Fatalf("expected cycle1 overall 100, got %v", trend[0].Overall)
	}
	if trend[1].Overall == nil || *trend[1].Overall != 50 {
		t.Fatalf("expected cycle2 overall 50, got %v", trend[1].Overall)
	}
	if trend[2].Overall == nil || *trend[2].Overall != 100 {
		t.Fatalf("expected cycle3 overall 100, got %v", trend[2].Overall)
	}
}

// TestEngineerTrendHandlerNoHistory tests the trend endpoint for an engineer
// with no rating history: returns an empty slice, not an error.
func TestEngineerTrendHandlerNoHistory(t *testing.T) {
	r, engineerStore, _, _, _, _, _ := setupEngineerCardTestRouter(t)
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	e1, _ := engineerStore.Create("Grace", nil, nil, nil, time.Now())

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/engineers/%d/trend", e1.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var trend []models.TrendPoint
	if err := json.Unmarshal(w.Body.Bytes(), &trend); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(trend) != 0 {
		t.Fatalf("expected empty trend for engineer with no history, got %d points", len(trend))
	}
}

// TestEngineerTrendHandlerInvalidEngineerID tests the invalid engineer ID case.
func TestEngineerTrendHandlerInvalidEngineerID(t *testing.T) {
	r, _, _, _, _, _, _ := setupEngineerCardTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/engineers/not-a-number/trend", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid engineer ID, got %d", w.Code)
	}
}

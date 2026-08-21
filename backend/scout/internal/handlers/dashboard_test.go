package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

// TestDashboardHandlerGet tests the basic dashboard endpoint (F11):
// returns all active engineers with their latest Overall score and most recent
// cycle date, hand-computed with exact expected values.
func TestDashboardHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers")

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	// Create 2 active engineers and rank both in a cycle.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_dash_handler", "Test Main Dashboard Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(scoreStore, engineerStore)
	r.GET("/api/dashboard", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var dashboard []models.RosterEntry
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(dashboard) != 2 {
		t.Fatalf("expected 2 engineers in dashboard, got %d", len(dashboard))
	}

	// Index by engineer name for easy verification.
	byName := make(map[string]*models.RosterEntry)
	for i := range dashboard {
		byName[dashboard[i].Engineer.Name] = &dashboard[i]
	}

	// Verify Alex (rank 1 -> score 100).
	alexEntry := byName["Alex"]
	if alexEntry == nil {
		t.Fatalf("expected Alex in dashboard")
	}
	if alexEntry.LatestOverall == nil || *alexEntry.LatestOverall != 100 {
		t.Fatalf("expected Alex latest overall 100 (rank 1 of 2), got %v", alexEntry.LatestOverall)
	}
	if alexEntry.LastCycleDate == nil {
		t.Fatalf("expected Alex last cycle date to be set")
	}

	// Verify Sam (rank 2 -> score 50).
	samEntry := byName["Sam"]
	if samEntry == nil {
		t.Fatalf("expected Sam in dashboard")
	}
	if samEntry.LatestOverall == nil || *samEntry.LatestOverall != 50 {
		t.Fatalf("expected Sam latest overall 50 (rank 2 of 2), got %v", samEntry.LatestOverall)
	}
	if samEntry.LastCycleDate == nil {
		t.Fatalf("expected Sam last cycle date to be set")
	}
}

// TestDashboardHandlerEmptyDashboard tests the empty-roster case: no engineers
// should return an empty array, not an error.
func TestDashboardHandlerEmptyDashboard(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	scoreStore := store.NewScoreStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(scoreStore, engineerStore)
	r.GET("/api/dashboard", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var dashboard []models.RosterEntry
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if dashboard == nil {
		t.Fatalf("expected empty array (not nil) for empty roster")
	}
	if len(dashboard) != 0 {
		t.Fatalf("expected empty dashboard, got %d engineers", len(dashboard))
	}
}

// TestDashboardHandlerIncludesActiveEngineerWithNoRankings verifies that an
// active engineer with no ranking history is included in the dashboard with
// nil LatestOverall and nil LastCycleDate, rather than being silently dropped.
func TestDashboardHandlerIncludesActiveEngineerWithNoRankings(t *testing.T) {
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers")

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	// Create e1 and submit ranking for the cycle.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_unranked_handler", "Test Main Unranked Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	// Create e2 AFTER ranking submission. Now e2 is active but has no rankings.
	_, _ = engineerStore.Create("Jordan", nil, nil, nil, time.Now())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(scoreStore, engineerStore)
	r.GET("/api/dashboard", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var dashboard []models.RosterEntry
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Both active engineers should appear.
	if len(dashboard) != 2 {
		t.Fatalf("expected 2 active engineers (one with rankings, one without), got %d", len(dashboard))
	}

	// Index by engineer name.
	byName := make(map[string]*models.RosterEntry)
	for i := range dashboard {
		byName[dashboard[i].Engineer.Name] = &dashboard[i]
	}

	// Verify Alex has scores.
	alexEntry := byName["Alex"]
	if alexEntry == nil {
		t.Fatalf("expected Alex in dashboard")
	}
	if alexEntry.LatestOverall == nil || *alexEntry.LatestOverall != 100 {
		t.Fatalf("expected Alex latest overall 100, got %v", alexEntry.LatestOverall)
	}

	// Verify Jordan (active, unranked) appears with nil scores.
	jordanEntry := byName["Jordan"]
	if jordanEntry == nil {
		t.Fatalf("expected Jordan (active, unranked engineer) in dashboard")
	}
	if jordanEntry.LatestOverall != nil {
		t.Fatalf("expected Jordan latest overall to be nil (no rankings), got %v", *jordanEntry.LatestOverall)
	}
	if jordanEntry.LastCycleDate != nil {
		t.Fatalf("expected Jordan last cycle date to be nil (no rankings), got %v", *jordanEntry.LastCycleDate)
	}
}

// TestDashboardHandlerExcludesDeactivatedEngineer is a regression test for the
// specific scenario flagged in code review: a deactivated engineer should NOT
// appear in the dashboard even if they have historical rankings. This verifies
// that the roster is genuinely filtered by active status, not inferred from
// ranking data.
func TestDashboardHandlerExcludesDeactivatedEngineer(t *testing.T) {
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers")

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	// Create 2 active engineers and rank both in a cycle.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_deactivated_handler", "Test Main Deactivated Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	})

	// Verify both appear in dashboard before deactivation.
	gin.SetMode(gin.TestMode)
	r1 := gin.New()
	h1 := NewDashboardHandler(scoreStore, engineerStore)
	r1.GET("/api/dashboard", h1.Get)

	req1 := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w1 := httptest.NewRecorder()
	r1.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 before deactivation, got %d", w1.Code)
	}

	var dashboardBefore []models.RosterEntry
	if err := json.Unmarshal(w1.Body.Bytes(), &dashboardBefore); err != nil {
		t.Fatalf("failed to unmarshal response before deactivation: %v", err)
	}

	if len(dashboardBefore) != 2 {
		t.Fatalf("expected 2 engineers before deactivation, got %d", len(dashboardBefore))
	}

	// Deactivate Sam (who ranked 2nd).
	if _, err := engineerStore.Deactivate(e2.ID); err != nil {
		t.Fatalf("deactivate e2 failed: %v", err)
	}

	// Query dashboard after deactivation: Sam should be excluded even though
	// the ranking record still exists in the database.
	r2 := gin.New()
	h2 := NewDashboardHandler(scoreStore, engineerStore)
	r2.GET("/api/dashboard", h2.Get)

	req2 := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w2 := httptest.NewRecorder()
	r2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 after deactivation, got %d", w2.Code)
	}

	var dashboardAfter []models.RosterEntry
	if err := json.Unmarshal(w2.Body.Bytes(), &dashboardAfter); err != nil {
		t.Fatalf("failed to unmarshal response after deactivation: %v", err)
	}

	if len(dashboardAfter) != 1 {
		t.Fatalf("expected 1 engineer after deactivation (Sam excluded), got %d", len(dashboardAfter))
	}

	// Verify the remaining engineer is Alex (not Sam).
	if dashboardAfter[0].Engineer.ID != e1.ID || dashboardAfter[0].Engineer.Name != "Alex" {
		t.Fatalf("expected only Alex in dashboard after Sam's deactivation, got %+v", dashboardAfter[0].Engineer)
	}
}

// TestDashboardHandlerComplexCycles verifies the dashboard works correctly
// with multiple active engineers and displays their latest cycle information.
func TestDashboardHandlerComplexCycles(t *testing.T) {
	db := setupTestDBForHandlers(t)
	truncateTables(t, db, "sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers")

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	// Create 2 engineers.
	e1, _ := engineerStore.Create("Alice", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Bob", nil, nil, nil, time.Now())

	main, _ := mainStore.Create("test_main_complex_handler", "Test Main Complex Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)

	// Create a cycle and submit rankings.
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	if _, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDashboardHandler(scoreStore, engineerStore)
	r.GET("/api/dashboard", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var dashboard []models.RosterEntry
	if err := json.Unmarshal(w.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(dashboard) != 2 {
		t.Fatalf("expected 2 engineers in dashboard, got %d", len(dashboard))
	}

	// Verify all engineers have their scores.
	byID := make(map[int]*models.RosterEntry)
	for i := range dashboard {
		byID[dashboard[i].Engineer.ID] = &dashboard[i]
	}

	// Verify Alice (rank 1): score 100.
	aliceEntry := byID[e1.ID]
	if aliceEntry == nil {
		t.Fatalf("expected Alice in dashboard")
	}
	if aliceEntry.LatestOverall == nil {
		t.Fatalf("expected Alice latest overall to be non-nil, got nil")
	}
	if *aliceEntry.LatestOverall != 100 {
		t.Fatalf("expected Alice score 100, got %f", *aliceEntry.LatestOverall)
	}

	// Verify Bob (rank 2): score 50.
	bobEntry := byID[e2.ID]
	if bobEntry == nil {
		t.Fatalf("expected Bob in dashboard")
	}
	if bobEntry.LatestOverall == nil {
		t.Fatalf("expected Bob latest overall to be non-nil, got nil")
	}
	if *bobEntry.LatestOverall != 50 {
		t.Fatalf("expected Bob score 50, got %f", *bobEntry.LatestOverall)
	}
}

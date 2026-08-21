package store

import (
	"testing"
	"time"

	"scout/internal/scoring"
)

// Each test below truncates only the tables it owns outright: engineers (so
// ValidatePermutation's active-roster check starts from a known, exact set)
// and sub_attribute_rankings (this test's own rows). main_attributes carries
// migration-seeded rows (see TestMainAttributeStoreSeedData) that nothing
// here reseeds, and sub_attributes/rating_cycles have no seed data but also
// no need to be wiped since every test creates its own fresh rows via
// unique keys — so none of those three are truncated.

func TestRankingStoreSubmitRankingValidPermutation(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main"), "Test Main")
	sub, _ := subStore.Create(main.ID, "Code Quality", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	rankings, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	})
	if err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}
	if len(rankings) != 2 {
		t.Fatalf("expected 2 persisted rankings, got %d", len(rankings))
	}
	for _, r := range rankings {
		if r.EngineerID == e1.ID && r.Score != 100 {
			t.Fatalf("expected rank-1 engineer to score 100, got %v", r.Score)
		}
		if r.EngineerID == e2.ID && r.Score != 50 {
			t.Fatalf("expected rank-2-of-2 engineer to score 50, got %v", r.Score)
		}
	}
}

// TestRankingStoreSubmitRankingScoreFormulaThreeEngineers exercises Task 9's
// RankToScore with N=3 (rank 1 -> 100, rank 2 -> 75, rank 3 -> 50) so the
// linear interpolation, not just the N=2 endpoints, is verified end to end.
func TestRankingStoreSubmitRankingScoreFormulaThreeEngineers(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	e3, _ := engineerStore.Create("Jo", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main_3eng"), "Test Main 3eng")
	sub, _ := subStore.Create(main.ID, "Code Quality", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	rankings, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
		{EngineerID: e3.ID, Rank: 3},
	})
	if err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}

	wantScore := map[int]float64{e1.ID: 100, e2.ID: 75, e3.ID: 50}
	if len(rankings) != 3 {
		t.Fatalf("expected 3 persisted rankings, got %d", len(rankings))
	}
	for _, r := range rankings {
		if want := wantScore[r.EngineerID]; r.Score != want {
			t.Fatalf("engineer %d: expected score %v, got %v", r.EngineerID, want, r.Score)
		}
	}
}

func TestRankingStoreSubmitRankingRejectsInvalidPermutation(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main2"), "Test Main 2")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	_, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 1},
	})
	if err == nil {
		t.Fatal("expected error for tied ranks")
	}

	// A rejected submission must not persist anything.
	rankings, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get after rejected submit failed: %v", err)
	}
	if len(rankings) != 0 {
		t.Fatalf("expected no rows persisted after a rejected submission, got %d", len(rankings))
	}
}

// TestRankingStoreSubmitRankingUsesRealActiveRoster proves the validator is
// run against the live active-engineer set from the DB (Task 2's
// ListActiveIDs), not a hardcoded list: deactivating an engineer must make a
// submission that still includes them fail, and a submission that excludes
// them (matching only the now-active roster) must succeed.
func TestRankingStoreSubmitRankingUsesRealActiveRoster(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main_roster"), "Test Main Roster")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	if _, err := engineerStore.Deactivate(e2.ID); err != nil {
		t.Fatalf("failed to deactivate e2: %v", err)
	}

	// Submitting the full original roster (including now-inactive e2) must
	// be rejected because it no longer matches the active set.
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err == nil {
		t.Fatal("expected error: submission includes a deactivated engineer")
	}

	// Submitting only the active engineer (e1) must succeed.
	rankings, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
	})
	if err != nil {
		t.Fatalf("expected submission matching the active roster to succeed, got: %v", err)
	}
	if len(rankings) != 1 || rankings[0].EngineerID != e1.ID {
		t.Fatalf("expected exactly one persisted ranking for e1, got %+v", rankings)
	}
}

func TestRankingStoreSubmitRankingReplacesOnResubmit(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main_resubmit"), "Test Main Resubmit")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	first, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	})
	if err != nil {
		t.Fatalf("first submit failed: %v", err)
	}
	firstIDs := map[int]bool{}
	for _, r := range first {
		firstIDs[r.ID] = true
	}

	// Re-submit with ranks flipped.
	second, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 2},
		{EngineerID: e2.ID, Rank: 1},
	})
	if err != nil {
		t.Fatalf("resubmit failed: %v", err)
	}
	if len(second) != 2 {
		t.Fatalf("expected 2 rows after resubmit, got %d", len(second))
	}

	// The rows from the first submission must be gone entirely (new primary
	// keys, not updated in place) — this is a delete-then-insert replace.
	for _, r := range second {
		if firstIDs[r.ID] {
			t.Fatalf("expected resubmit to replace prior rows with new ones, but found reused id %d", r.ID)
		}
	}

	stored, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get after resubmit failed: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected exactly 2 rows persisted after resubmit (no leftover rows), got %d", len(stored))
	}
	for _, r := range stored {
		if r.EngineerID == e1.ID {
			if r.Rank != 2 || r.Score != 50 {
				t.Fatalf("expected e1 to now be rank 2 / score 50, got rank=%d score=%v", r.Rank, r.Score)
			}
		}
		if r.EngineerID == e2.ID {
			if r.Rank != 1 || r.Score != 100 {
				t.Fatalf("expected e2 to now be rank 1 / score 100, got rank=%d score=%v", r.Rank, r.Score)
			}
		}
	}
}

// TestRankingStoreSubmitRankingIsolatedBySubAttribute proves SubmitRanking's
// DELETE is scoped by both cycle_id AND sub_attribute_id, not just cycle_id:
// two sub-attributes under the same cycle each get a ranking, then only one
// is resubmitted, and the other sub-attribute's rankings must be untouched.
func TestRankingStoreSubmitRankingIsolatedBySubAttribute(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create(uniqueKey("test_main_isolation"), "Test Main Isolation")
	subA, _ := subStore.Create(main.ID, "Code Quality", nil)
	subB, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	if _, err := rankingStore.SubmitRanking(cycle.ID, subA.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err != nil {
		t.Fatalf("submit for subA failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, subB.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 2},
		{EngineerID: e2.ID, Rank: 1},
	}); err != nil {
		t.Fatalf("submit for subB failed: %v", err)
	}

	// Resubmit only subA, with ranks flipped.
	if _, err := rankingStore.SubmitRanking(cycle.ID, subA.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 2},
		{EngineerID: e2.ID, Rank: 1},
	}); err != nil {
		t.Fatalf("resubmit for subA failed: %v", err)
	}

	// subA reflects the resubmit.
	gotA, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, subA.ID)
	if err != nil {
		t.Fatalf("get subA failed: %v", err)
	}
	if len(gotA) != 2 {
		t.Fatalf("expected 2 rows for subA, got %d", len(gotA))
	}
	for _, r := range gotA {
		if r.EngineerID == e1.ID && r.Rank != 2 {
			t.Fatalf("expected e1 to be rank 2 in subA after resubmit, got rank=%d", r.Rank)
		}
		if r.EngineerID == e2.ID && r.Rank != 1 {
			t.Fatalf("expected e2 to be rank 1 in subA after resubmit, got rank=%d", r.Rank)
		}
	}

	// subB must be completely unaffected by subA's resubmit.
	gotB, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, subB.ID)
	if err != nil {
		t.Fatalf("get subB failed: %v", err)
	}
	if len(gotB) != 2 {
		t.Fatalf("expected subB to still have 2 rows (untouched by subA's resubmit), got %d", len(gotB))
	}
	for _, r := range gotB {
		if r.EngineerID == e1.ID && r.Rank != 2 {
			t.Fatalf("expected e1 to remain rank 2 in subB, got rank=%d", r.Rank)
		}
		if r.EngineerID == e2.ID && r.Rank != 1 {
			t.Fatalf("expected e2 to remain rank 1 in subB, got rank=%d", r.Rank)
		}
	}
}

func TestRankingStoreGetByCycleAndSubAttributeEmpty(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)

	main, _ := mainStore.Create(uniqueKey("test_main_empty"), "Test Main Empty")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	rankings, err := rankingStore.GetByCycleAndSubAttribute(cycle.ID, sub.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if len(rankings) != 0 {
		t.Fatalf("expected no rankings for a cycle+sub-attribute with no submission, got %d", len(rankings))
	}
}

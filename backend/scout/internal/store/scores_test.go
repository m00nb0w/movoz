package store

import (
	"database/sql"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/scoring"
)

// mustSetCreatedAt pins a row's created_at to an exact timestamp so F8
// cutover tests don't depend on wall-clock ordering between statements.
func mustSetCreatedAt(t *testing.T, db *sql.DB, table string, id int, at time.Time) {
	t.Helper()
	if _, err := db.Exec("UPDATE "+table+" SET created_at = $1 WHERE id = $2", at, id); err != nil {
		t.Fatalf("failed to set %s.created_at for id %d: %v", table, id, err)
	}
}

// TestScoreStoreMainAttributeScoresAndOverall covers the simple, happy-path
// aggregation: two engineers ranked against each other on both sub-attributes
// of a single main attribute, hand-computed expected scores (F7's linear
// interpolation for n=2 gives rank 1 -> 100, rank 2 -> 50).
func TestScoreStoreMainAttributeScoresAndOverall(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_score", "Test Main Score")
	sub1, _ := subStore.Create(main.ID, "Code Quality", nil)
	sub2, _ := subStore.Create(main.ID, "Testing", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Alex ranks 1st on both sub-attributes -> both score 100 -> main attribute avg 100.
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub1.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub2.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub2 ranking failed: %v", err)
	}

	scores, err := scoreStore.MainAttributeScores(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("main attribute scores failed: %v", err)
	}
	if len(scores) != 1 || scores[0].Score != 100 {
		t.Fatalf("expected one main attribute scoring 100 for e1, got %+v", scores)
	}

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	if overall == nil || *overall != 100 {
		t.Fatalf("expected overall score 100 for e1, got %v", overall)
	}

	e2Overall, err := scoreStore.OverallScore(e2.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score for e2 failed: %v", err)
	}
	if e2Overall == nil || *e2Overall != 50 {
		t.Fatalf("expected overall score 50 for e2 (rank 2 of 2 on both sub-attributes), got %v", e2Overall)
	}
}

// TestScoreStoreOverallScoreNilWhenNoData covers the no-data edge case: an
// engineer with zero rankings anywhere for the cycle must get a nil Overall,
// not a 0 or an error (there is nothing to average).
func TestScoreStoreOverallScoreNilWhenNoData(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "engineers", "rating_cycles"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	cycleStore := NewCycleStore(db)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("NoData", nil, nil, nil, time.Now())
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	if overall != nil {
		t.Fatalf("expected nil overall score when engineer has no rankings for the cycle, got %v", *overall)
	}
}

// TestScoreStoreMainAttributeWithZeroRankingsExcludedNotZero covers a main
// attribute that exists (with a sub-attribute) but has never been ranked in
// this cycle: it must be absent from MainAttributeScores (not present with a
// score of 0), and must not drag OverallScore down toward 0 either.
func TestScoreStoreMainAttributeWithZeroRankingsExcludedNotZero(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	mainRanked, err := mainStore.Create("ranked_main", "Ranked Main")
	if err != nil {
		t.Fatalf("create ranked main failed: %v", err)
	}
	subRanked, err := subStore.Create(mainRanked.ID, "Ranked Sub", nil)
	if err != nil {
		t.Fatalf("create ranked sub failed: %v", err)
	}

	mainUnranked, err := mainStore.Create("unranked_main", "Unranked Main")
	if err != nil {
		t.Fatalf("create unranked main failed: %v", err)
	}
	// A sub-attribute exists under the second main attribute, but nobody has
	// submitted a ranking for it in this cycle.
	if _, err := subStore.Create(mainUnranked.ID, "Unranked Sub", nil); err != nil {
		t.Fatalf("create unranked sub failed: %v", err)
	}

	cycle, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("create cycle failed: %v", err)
	}

	if _, err := rankingStore.SubmitRanking(cycle.ID, subRanked.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}

	scores, err := scoreStore.MainAttributeScores(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("main attribute scores failed: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("expected only the ranked main attribute to appear, got %+v", scores)
	}
	if scores[0].MainAttributeID != mainRanked.ID || scores[0].Score != 100 {
		t.Fatalf("expected ranked main attribute scoring 100, got %+v", scores[0])
	}

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	// If the unranked main attribute were treated as a 0 instead of being
	// excluded, the overall would be (100+0)/2 = 50 instead of 100.
	if overall == nil || *overall != 100 {
		t.Fatalf("expected overall score 100 (unranked main attribute excluded, not zeroed), got %v", overall)
	}
}

// TestScoreStoreOverallScoreF8CutoverExcludesMainAttributeCreatedAfterCycle
// is the core F8 regression test: a main attribute created strictly AFTER a
// cycle was opened must never count toward that cycle's Overall, even if
// rankings were later submitted against it for that cycle (retroactive
// cutover protection). The same main attribute IS counted for a later cycle
// created after the main attribute existed. Timestamps are pinned via direct
// SQL rather than relying on wall-clock ordering, so the test is not flaky
// and the two main attributes are given different scores so an incorrectly
// inclusive/exclusive comparison changes the computed Overall, not just
// whether extra (identical) data was folded in.
func TestScoreStoreOverallScoreF8CutoverExcludesMainAttributeCreatedAfterCycle(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	mainOld, err := mainStore.Create("preexisting_main", "Preexisting Main")
	if err != nil {
		t.Fatalf("create mainOld failed: %v", err)
	}
	subOld, err := subStore.Create(mainOld.ID, "Preexisting Sub", nil)
	if err != nil {
		t.Fatalf("create subOld failed: %v", err)
	}

	mainNew, err := mainStore.Create("future_main", "Future Main")
	if err != nil {
		t.Fatalf("create mainNew failed: %v", err)
	}
	subNew, err := subStore.Create(mainNew.ID, "Future Sub", nil)
	if err != nil {
		t.Fatalf("create subNew failed: %v", err)
	}

	cycle1, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("create cycle1 failed: %v", err)
	}

	// Pin created_at timestamps explicitly rather than relying on execution
	// speed: mainOld before cycle1 (should count), mainNew strictly after
	// cycle1 was opened (should NOT count toward cycle1's overall).
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	mustSetCreatedAt(t, db, "main_attributes", mainOld.ID, base)
	mustSetCreatedAt(t, db, "rating_cycles", cycle1.ID, base.Add(24*time.Hour))
	mustSetCreatedAt(t, db, "main_attributes", mainNew.ID, base.Add(48*time.Hour))

	// e1 scores 100 on the preexisting main attribute, 50 on the future one —
	// different scores so an Overall computed with the wrong cutover
	// comparison is numerically distinguishable from the correct one.
	if _, err := rankingStore.SubmitRanking(cycle1.ID, subOld.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit subOld ranking (cycle1) failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle1.ID, subNew.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit subNew ranking (cycle1) failed: %v", err)
	}

	// MainAttributeScores is NOT gated by the cutover rule: both main
	// attributes should appear with their per-cycle scores.
	scores, err := scoreStore.MainAttributeScores(e1.ID, cycle1.ID)
	if err != nil {
		t.Fatalf("main attribute scores (cycle1) failed: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected both main attributes in the per-attribute breakdown, got %+v", scores)
	}
	byID := map[int]float64{}
	for _, s := range scores {
		byID[s.MainAttributeID] = s.Score
	}
	if byID[mainOld.ID] != 100 || byID[mainNew.ID] != 50 {
		t.Fatalf("expected mainOld=100, mainNew=50 in per-attribute breakdown, got %+v", byID)
	}

	// Overall for cycle1 must only reflect mainOld (100). If mainNew were
	// wrongly included the result would be (100+50)/2 = 75.
	overall1, err := scoreStore.OverallScore(e1.ID, cycle1.ID)
	if err != nil {
		t.Fatalf("overall score (cycle1) failed: %v", err)
	}
	if overall1 == nil || *overall1 != 100 {
		t.Fatalf("expected cycle1 overall 100 (future main attribute excluded), got %v", overall1)
	}

	// A later cycle, opened after mainNew existed, must count mainNew.
	cycle2, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("create cycle2 failed: %v", err)
	}
	mustSetCreatedAt(t, db, "rating_cycles", cycle2.ID, base.Add(72*time.Hour))

	if _, err := rankingStore.SubmitRanking(cycle2.ID, subOld.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit subOld ranking (cycle2) failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle2.ID, subNew.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit subNew ranking (cycle2) failed: %v", err)
	}

	overall2, err := scoreStore.OverallScore(e1.ID, cycle2.ID)
	if err != nil {
		t.Fatalf("overall score (cycle2) failed: %v", err)
	}
	// If mainNew were wrongly excluded from cycle2 too, this would be 100
	// instead of the correct (100+50)/2 = 75.
	if overall2 == nil || *overall2 != 75 {
		t.Fatalf("expected cycle2 overall 75 (future main attribute now counts), got %v", overall2)
	}

	// Cross-cycle isolation: re-querying cycle1's overall after submitting
	// cycle2 rankings must still be 100, unaffected by cycle2's data.
	overall1Again, err := scoreStore.OverallScore(e1.ID, cycle1.ID)
	if err != nil {
		t.Fatalf("overall score (cycle1, re-check) failed: %v", err)
	}
	if overall1Again == nil || *overall1Again != 100 {
		t.Fatalf("expected cycle1 overall to remain 100 after cycle2 activity, got %v", overall1Again)
	}
}

// TestScoreStoreOverallScoreF8CutoverIncludesMainAttributeCreatedAtSameInstantAsCycle
// pins down the exact boundary semantics: F8 says a main attribute counts if
// created_at <= the cycle's created_at. This test sets the main attribute's
// created_at to the EXACT same instant as the cycle's created_at, so it only
// passes if the implementation truly uses <= — a strict < would wrongly
// exclude it here.
func TestScoreStoreOverallScoreF8CutoverIncludesMainAttributeCreatedAtSameInstantAsCycle(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	mainA, err := mainStore.Create("main_a", "Main A")
	if err != nil {
		t.Fatalf("create mainA failed: %v", err)
	}
	subA, err := subStore.Create(mainA.ID, "Sub A", nil)
	if err != nil {
		t.Fatalf("create subA failed: %v", err)
	}
	mainB, err := mainStore.Create("main_b", "Main B (boundary)")
	if err != nil {
		t.Fatalf("create mainB failed: %v", err)
	}
	subB, err := subStore.Create(mainB.ID, "Sub B", nil)
	if err != nil {
		t.Fatalf("create subB failed: %v", err)
	}

	cycle, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("create cycle failed: %v", err)
	}

	base := time.Date(2021, 6, 1, 12, 0, 0, 0, time.UTC)
	mustSetCreatedAt(t, db, "main_attributes", mainA.ID, base.Add(-24*time.Hour))
	// Exactly the same instant as the cycle's created_at.
	mustSetCreatedAt(t, db, "main_attributes", mainB.ID, base)
	mustSetCreatedAt(t, db, "rating_cycles", cycle.ID, base)

	if _, err := rankingStore.SubmitRanking(cycle.ID, subA.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit subA ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, subB.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit subB ranking failed: %v", err)
	}

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	// (100+50)/2 = 75 only if mainB (created_at == cycle.created_at) counts.
	// A strict "<" comparison would exclude mainB and yield 100 instead.
	if overall == nil || *overall != 75 {
		t.Fatalf("expected overall 75 (main attribute created at the same instant as the cycle must count), got %v", overall)
	}
}

// TestScoreStoreOverallScoreExcludesMainAttributeWithNullCreatedAt guards
// against a latent trap in the F8 cutover predicate: main_attributes.created_at
// and rating_cycles.created_at are nullable columns (DEFAULT NOW(), no NOT
// NULL constraint). No current write path ever leaves created_at NULL, but
// if one ever did (a bulk-import path, a manual INSERT that skips the
// default, etc.), that main attribute must fail safe by being excluded from
// Overall — not silently included via undefined comparison behavior, and not
// an error. This forces created_at to NULL via raw SQL (bypassing all store
// methods, which never do this) to simulate that scenario directly.
func TestScoreStoreOverallScoreExcludesMainAttributeWithNullCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	mainNormal, err := mainStore.Create("normal_main", "Normal Main")
	if err != nil {
		t.Fatalf("create mainNormal failed: %v", err)
	}
	subNormal, err := subStore.Create(mainNormal.ID, "Normal Sub", nil)
	if err != nil {
		t.Fatalf("create subNormal failed: %v", err)
	}

	mainNullCreatedAt, err := mainStore.Create("null_created_at_main", "Null CreatedAt Main")
	if err != nil {
		t.Fatalf("create mainNullCreatedAt failed: %v", err)
	}
	subNullCreatedAt, err := subStore.Create(mainNullCreatedAt.ID, "Null CreatedAt Sub", nil)
	if err != nil {
		t.Fatalf("create subNullCreatedAt failed: %v", err)
	}

	cycle, err := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	if err != nil {
		t.Fatalf("create cycle failed: %v", err)
	}

	if _, err := db.Exec("UPDATE main_attributes SET created_at = NULL WHERE id = $1", mainNullCreatedAt.ID); err != nil {
		t.Fatalf("failed to force created_at NULL: %v", err)
	}
	// This is a real, persistent Postgres database shared across test runs
	// (not a per-test transaction rollback) — leaving a row with a NULL
	// created_at behind would break any other test that scans
	// main_attributes.created_at into a non-nullable time.Time (e.g.
	// TestMainAttributeStoreSeedData's List() call), possibly in a later,
	// unrelated `go test` invocation. Restore it once this test is done.
	t.Cleanup(func() {
		if _, err := db.Exec("UPDATE main_attributes SET created_at = NOW() WHERE id = $1", mainNullCreatedAt.ID); err != nil {
			t.Logf("cleanup: failed to restore created_at on main attribute %d: %v", mainNullCreatedAt.ID, err)
		}
	})

	// Different scores so an incorrectly-included NULL-created_at attribute
	// is numerically distinguishable from the correct (excluded) result.
	if _, err := rankingStore.SubmitRanking(cycle.ID, subNormal.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit subNormal ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, subNullCreatedAt.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit subNullCreatedAt ranking failed: %v", err)
	}

	// MainAttributeScores is not cutover-gated, so it still reports both.
	scores, err := scoreStore.MainAttributeScores(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("main attribute scores failed: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected both main attributes in the per-attribute breakdown, got %+v", scores)
	}

	overall, err := scoreStore.OverallScore(e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("overall score failed: %v", err)
	}
	// Only mainNormal (100) should count; mainNullCreatedAt (50) must be
	// excluded because its created_at is NULL. If it were wrongly included
	// the result would be (100+50)/2 = 75.
	if overall == nil || *overall != 100 {
		t.Fatalf("expected overall 100 (NULL created_at main attribute excluded), got %v", overall)
	}
}

// TestScoreStoreEngineerCard tests the engineer card endpoint (F10): returns
// the engineer's Overall + main-attribute scores for one cycle, hand-computed
// with exact expected values.
func TestScoreStoreEngineerCard(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_card", "Test Main Card")
	sub1, _ := subStore.Create(main.ID, "Ownership", nil)
	sub2, _ := subStore.Create(main.ID, "Documentation", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Alex ranks 1st on both sub-attributes -> both score 100 -> main attribute avg 100 -> overall 100.
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub1.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub2.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit sub2 ranking failed: %v", err)
	}

	card, err := scoreStore.EngineerCard(engineerStore, e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("engineer card failed: %v", err)
	}
	if card == nil {
		t.Fatalf("expected non-nil card for existing engineer")
	}
	if card.Engineer.ID != e1.ID || card.Engineer.Name != "Alex" {
		t.Fatalf("unexpected engineer in card: %+v", card.Engineer)
	}
	if card.CycleID != cycle.ID {
		t.Fatalf("unexpected cycle ID in card: %d", card.CycleID)
	}
	if card.Overall == nil || *card.Overall != 100 {
		t.Fatalf("expected overall score 100 for e1, got %v", card.Overall)
	}
	if len(card.MainAttributes) != 1 {
		t.Fatalf("expected 1 main attribute score, got %d", len(card.MainAttributes))
	}
	if card.MainAttributes[0].MainAttributeID != main.ID || card.MainAttributes[0].Score != 100 {
		t.Fatalf("expected main attribute %d with score 100, got %+v", main.ID, card.MainAttributes[0])
	}

	// Sam ranks 2nd on both sub-attributes -> both score 50 -> main attribute avg 50 -> overall 50.
	card2, err := scoreStore.EngineerCard(engineerStore, e2.ID, cycle.ID)
	if err != nil {
		t.Fatalf("engineer card for e2 failed: %v", err)
	}
	if card2.Overall == nil || *card2.Overall != 50 {
		t.Fatalf("expected overall score 50 for e2, got %v", card2.Overall)
	}
}

// TestScoreStoreEngineerCardNotFound tests the not-found case: an engineer
// that doesn't exist should return nil, not an error.
func TestScoreStoreEngineerCardNotFound(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	cycleStore := NewCycleStore(db)
	scoreStore := NewScoreStore(db)

	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	card, err := scoreStore.EngineerCard(engineerStore, 9999, cycle.ID)
	if err != nil {
		t.Fatalf("engineer card with non-existent engineer should not error: %v", err)
	}
	if card != nil {
		t.Fatalf("expected nil card for non-existent engineer, got %+v", card)
	}
}

// TestScoreStoreEngineerCardNoData tests the no-data case: an engineer with
// no rankings in a cycle gets a card with nil Overall and empty MainAttributes.
func TestScoreStoreEngineerCardNoData(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	cycleStore := NewCycleStore(db)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("NoData", nil, nil, nil, time.Now())
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	card, err := scoreStore.EngineerCard(engineerStore, e1.ID, cycle.ID)
	if err != nil {
		t.Fatalf("engineer card with no rankings should not error: %v", err)
	}
	if card == nil {
		t.Fatalf("expected non-nil card for existing engineer with no rankings")
	}
	if card.Overall != nil {
		t.Fatalf("expected nil overall for engineer with no rankings, got %v", *card.Overall)
	}
	// MainAttributeScores returns an empty slice when there are no rankings.
	if card.MainAttributes == nil {
		t.Fatalf("expected empty main attributes slice, got nil")
	}
	if len(card.MainAttributes) != 0 {
		t.Fatalf("expected no main attributes for engineer with no rankings, got %d", len(card.MainAttributes))
	}
}

// TestScoreStoreEngineerTrend tests the trend endpoint (F10): returns scores
// across all past cycles an engineer has rankings in, oldest first.
func TestScoreStoreEngineerTrend(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_trend", "Test Main Trend")
	sub, _ := subStore.Create(main.ID, "Testing", nil)

	// Create three cycles and submit rankings in non-chronological order.
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	cycle3Start := base.Add(2 * 24 * time.Hour)
	cycle2Start := base.Add(1 * 24 * time.Hour)
	cycle1Start := base

	cycle3, _ := cycleStore.Create(cycle3Start, cycle3Start.AddDate(0, 0, 14))
	cycle2, _ := cycleStore.Create(cycle2Start, cycle2Start.AddDate(0, 0, 14))
	cycle1, _ := cycleStore.Create(cycle1Start, cycle1Start.AddDate(0, 0, 14))

	// Submit rankings in non-chronological order (cycle3, cycle1, cycle2).
	// Alex ranks 1st in cycle3 (100), 1st in cycle1 (100), 2nd in cycle2 (50).
	if _, err := rankingStore.SubmitRanking(cycle3.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit cycle3 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle1.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}}); err != nil {
		t.Fatalf("submit cycle1 ranking failed: %v", err)
	}
	if _, err := rankingStore.SubmitRanking(cycle2.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 2}, {EngineerID: e2.ID, Rank: 1}}); err != nil {
		t.Fatalf("submit cycle2 ranking failed: %v", err)
	}

	trend, err := scoreStore.EngineerTrend(engineerStore, e1.ID)
	if err != nil {
		t.Fatalf("engineer trend failed: %v", err)
	}
	if len(trend) != 3 {
		t.Fatalf("expected 3 trend points (one per cycle), got %d", len(trend))
	}

	// Verify chronological ordering: cycle1, cycle2, cycle3 (period_start order).
	if trend[0].CycleID != cycle1.ID {
		t.Fatalf("expected first trend point to be cycle1 (ID %d), got cycle ID %d", cycle1.ID, trend[0].CycleID)
	}
	if trend[1].CycleID != cycle2.ID {
		t.Fatalf("expected second trend point to be cycle2 (ID %d), got cycle ID %d", cycle2.ID, trend[1].CycleID)
	}
	if trend[2].CycleID != cycle3.ID {
		t.Fatalf("expected third trend point to be cycle3 (ID %d), got cycle ID %d", cycle3.ID, trend[2].CycleID)
	}

	// Verify scores: Alex was rank 1 in cycles 1 and 3 (100), rank 2 in cycle 2 (50).
	if trend[0].Overall == nil || *trend[0].Overall != 100 {
		t.Fatalf("expected cycle1 overall 100, got %v", trend[0].Overall)
	}
	if trend[1].Overall == nil || *trend[1].Overall != 50 {
		t.Fatalf("expected cycle2 overall 50, got %v", trend[1].Overall)
	}
	if trend[2].Overall == nil || *trend[2].Overall != 100 {
		t.Fatalf("expected cycle3 overall 100, got %v", trend[2].Overall)
	}

	// Verify period bounds are populated (compare Unix time to avoid timezone issues).
	if trend[0].PeriodStart.Unix() != cycle1Start.Unix() || trend[0].PeriodEnd.Unix() != cycle1Start.AddDate(0, 0, 14).Unix() {
		t.Fatalf("expected cycle1 period bounds to match, got start=%v end=%v", trend[0].PeriodStart.Unix(), trend[0].PeriodEnd.Unix())
	}
}

// TestScoreStoreEngineerTrendNoHistory tests the no-history case: an engineer
// with no rankings in any cycle should get an empty slice, not an error.
func TestScoreStoreEngineerTrendNoHistory(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	scoreStore := NewScoreStore(db)

	e1, _ := engineerStore.Create("NoHistory", nil, nil, nil, time.Now())

	trend, err := scoreStore.EngineerTrend(engineerStore, e1.ID)
	if err != nil {
		t.Fatalf("engineer trend with no history should not error: %v", err)
	}
	if trend == nil {
		t.Fatalf("expected empty slice for engineer with no trend history, got nil")
	}
	if len(trend) != 0 {
		t.Fatalf("expected empty trend for engineer with no rankings, got %d points", len(trend))
	}
}

// TestScoreStoreCycleScores tests the cycle view endpoint (F15): returns all
// active engineers with their Overall + main-attribute scores for the cycle,
// hand-computed with exact expected values. For n=2 engineers, rank 1 scores 100,
// rank 2 scores 50 per RankToScore (linear interpolation).
func TestScoreStoreCycleScores(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	// Create 2 active engineers: Alex (rank 1, expects 100) and Sam (rank 2, expects 50).
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_cv", "Test Main CV")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}})

	scores, err := scoreStore.CycleScores(engineerStore, cycle.ID)
	if err != nil {
		t.Fatalf("cycle scores failed: %v", err)
	}
	if len(scores) != 2 {
		t.Fatalf("expected 2 engineers in cycle view, got %d", len(scores))
	}

	// Verify scores by engineer name (order may vary).
	scoresByName := make(map[string]*models.EngineerCycleScore)
	for i := range scores {
		scoresByName[scores[i].Engineer.Name] = &scores[i]
	}

	alexScore := scoresByName["Alex"]
	if alexScore == nil {
		t.Fatalf("expected Alex in cycle scores")
	}
	if alexScore.Overall == nil || *alexScore.Overall != 100 {
		t.Fatalf("expected Alex overall score 100 (rank 1 of 2), got %v", alexScore.Overall)
	}
	if len(alexScore.MainAttributes) != 1 || alexScore.MainAttributes[0].Score != 100 {
		t.Fatalf("expected Alex main attribute score 100, got %+v", alexScore.MainAttributes)
	}

	samScore := scoresByName["Sam"]
	if samScore == nil {
		t.Fatalf("expected Sam in cycle scores")
	}
	if samScore.Overall == nil || *samScore.Overall != 50 {
		t.Fatalf("expected Sam overall score 50 (rank 2 of 2), got %v", samScore.Overall)
	}
	if len(samScore.MainAttributes) != 1 || samScore.MainAttributes[0].Score != 50 {
		t.Fatalf("expected Sam main attribute score 50, got %+v", samScore.MainAttributes)
	}
}

// TestScoreStoreCycleScoresActiveEngineerWithNoRankings verifies that an active
// engineer with no rankings for the cycle appears in the cycle view with nil
// Overall and empty MainAttributes, rather than being silently dropped. This is
// tested by creating an engineer AFTER ranking submission, so the engineer is
// active but has no data for that cycle (ValidatePermutation requires all active
// engineers be ranked at submission time).
func TestScoreStoreCycleScoresActiveEngineerWithNoRankings(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	// Create e1 and submit ranking for the cycle.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_no_rank", "Test Main No Rank")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	// Create e2 AFTER ranking submission. Now e2 is active but has no rankings for this cycle.
	_, _ = engineerStore.Create("Sam", nil, nil, nil, time.Now())

	scores, err := scoreStore.CycleScores(engineerStore, cycle.ID)
	if err != nil {
		t.Fatalf("cycle scores failed: %v", err)
	}
	// Both active engineers should appear in the result.
	if len(scores) != 2 {
		t.Fatalf("expected 2 active engineers (one with rankings, one without), got %d", len(scores))
	}

	scoresByName := make(map[string]*models.EngineerCycleScore)
	for i := range scores {
		scoresByName[scores[i].Engineer.Name] = &scores[i]
	}

	// Alex should have scores.
	alexScore := scoresByName["Alex"]
	if alexScore == nil {
		t.Fatalf("expected Alex in cycle scores")
	}
	if alexScore.Overall == nil || *alexScore.Overall != 100 {
		t.Fatalf("expected Alex overall 100, got %v", alexScore.Overall)
	}

	// Sam (active but unranked) should appear with nil overall and empty main attributes.
	samScore := scoresByName["Sam"]
	if samScore == nil {
		t.Fatalf("expected Sam (active, unranked engineer) in cycle scores")
	}
	if samScore.Overall != nil {
		t.Fatalf("expected Sam overall to be nil (unranked), got %v", *samScore.Overall)
	}
	if samScore.MainAttributes == nil {
		t.Fatalf("expected empty main attributes slice, got nil")
	}
	if len(samScore.MainAttributes) != 0 {
		t.Fatalf("expected Sam main attributes to be empty (unranked), got %d", len(samScore.MainAttributes))
	}
}

// TestScoreStoreCycleScoresExcludesDeactivatedEngineer is a regression test for
// the specific scenario flagged in code review: an engineer ranked in a cycle
// who is later deactivated must NOT appear in CycleScores for that cycle. This
// verifies the filtering by ListActiveIDs() correctly excludes deactivated
// engineers even when they have historical rankings.
func TestScoreStoreCycleScoresExcludesDeactivatedEngineer(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	// Create 2 active engineers and rank both in a cycle.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_deactivated", "Test Main Deactivated")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}, {EngineerID: e2.ID, Rank: 2}})

	// Both should appear in cycle scores initially.
	scoresBeforeDeactivation, err := scoreStore.CycleScores(engineerStore, cycle.ID)
	if err != nil {
		t.Fatalf("cycle scores before deactivation failed: %v", err)
	}
	if len(scoresBeforeDeactivation) != 2 {
		t.Fatalf("expected 2 engineers before deactivation, got %d", len(scoresBeforeDeactivation))
	}

	// Deactivate Sam (who ranked 2nd).
	if _, err := engineerStore.Deactivate(e2.ID); err != nil {
		t.Fatalf("deactivate e2 failed: %v", err)
	}

	// After deactivation, Sam should be excluded from cycle scores even though
	// the ranking record still exists in the database.
	scoresAfterDeactivation, err := scoreStore.CycleScores(engineerStore, cycle.ID)
	if err != nil {
		t.Fatalf("cycle scores after deactivation failed: %v", err)
	}
	if len(scoresAfterDeactivation) != 1 {
		t.Fatalf("expected 1 engineer after deactivation (Sam excluded), got %d", len(scoresAfterDeactivation))
	}

	// Verify the remaining engineer is Alex (not Sam).
	if scoresAfterDeactivation[0].Engineer.ID != e1.ID || scoresAfterDeactivation[0].Engineer.Name != "Alex" {
		t.Fatalf("expected only Alex in cycle scores after Sam's deactivation, got %+v", scoresAfterDeactivation[0].Engineer)
	}
}

// TestScoreStoreRosterDashboard tests the roster dashboard endpoint (F11):
// returns all active engineers with their latest Overall score and most recent
// cycle date. Tests hand-computed expected values, multiple engineers, and
// correct ordering of cycles.
func TestScoreStoreRosterDashboard(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	// Create 2 engineers.
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Sam", nil, nil, nil, time.Now())

	main, _ := mainStore.Create("test_main_dashboard", "Test Main Dashboard")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)

	// Create one cycle and submit rankings.
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))

	// Submit rankings: Alex rank 1 (100), Sam rank 2 (50).
	if _, err := rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err != nil {
		t.Fatalf("submit ranking failed: %v", err)
	}

	// Query dashboard: should include both engineers with their scores.
	dashboard, err := scoreStore.RosterDashboard(engineerStore)
	if err != nil {
		t.Fatalf("roster dashboard failed: %v", err)
	}
	if len(dashboard) != 2 {
		t.Fatalf("expected 2 active engineers on the dashboard, got %d", len(dashboard))
	}

	// Index by engineer name for easy verification.
	byName := make(map[string]*models.RosterEntry)
	for i := range dashboard {
		byName[dashboard[i].Engineer.Name] = &dashboard[i]
	}

	// Verify Alex: rank 1 -> score 100.
	alexEntry := byName["Alex"]
	if alexEntry == nil {
		t.Fatalf("expected Alex in dashboard")
	}
	if alexEntry.LatestOverall == nil {
		t.Fatalf("expected Alex latest overall to be non-nil, got nil")
	}
	if *alexEntry.LatestOverall != 100 {
		t.Fatalf("expected Alex latest overall 100 (rank 1 of 2), got %f", *alexEntry.LatestOverall)
	}
	if alexEntry.LastCycleDate == nil {
		t.Fatalf("expected Alex last cycle date to be set")
	}

	// Verify Sam: rank 2 -> score 50.
	samEntry := byName["Sam"]
	if samEntry == nil {
		t.Fatalf("expected Sam in dashboard")
	}
	if samEntry.LatestOverall == nil {
		t.Fatalf("expected Sam latest overall to be non-nil, got nil")
	}
	if *samEntry.LatestOverall != 50 {
		t.Fatalf("expected Sam latest overall 50 (rank 2 of 2), got %f", *samEntry.LatestOverall)
	}

	// Deactivate Sam and re-query dashboard.
	engineerStore.Deactivate(e2.ID)

	dashboardAfterDeactivation, err := scoreStore.RosterDashboard(engineerStore)
	if err != nil {
		t.Fatalf("roster dashboard after deactivation failed: %v", err)
	}
	// Should now have only Alex (Sam deactivated).
	if len(dashboardAfterDeactivation) != 1 {
		t.Fatalf("expected 1 active engineer after deactivation (Sam excluded), got %d", len(dashboardAfterDeactivation))
	}

	// Verify the remaining engineer is Alex.
	if dashboardAfterDeactivation[0].Engineer.ID != e1.ID {
		t.Fatalf("expected only Alex in dashboard after deactivation, got %+v", dashboardAfterDeactivation[0].Engineer)
	}
}

// TestScoreStoreRosterDashboardMostRecentCycleNonChronologicalCreation verifies
// that the "most recent cycle" is determined by cycle period_start ordering, not
// by creation order. This is critical because cycles might be created out of
// order (e.g., a late submission for a past period). The test creates cycles
// in reverse chronological order (cycle3, cycle2, cycle1) to ensure that the
// dashboard correctly identifies the *most recent* period, not the most recently
// created cycle.
func TestScoreStoreRosterDashboardMostRecentCycleNonChronologicalCreation(t *testing.T) {
	db := setupTestDB(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := NewEngineerStore(db)
	mainStore := NewMainAttributeStore(db)
	subStore := NewSubAttributeStore(db)
	cycleStore := NewCycleStore(db)
	rankingStore := NewRankingStore(db, engineerStore)
	scoreStore := NewScoreStore(db)

	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Create engineers first, before cycles (so ValidatePermutation works correctly)
	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	e2, _ := engineerStore.Create("Dummy", nil, nil, nil, time.Now())

	main, _ := mainStore.Create("test_main_nonchron", "Test Main Non-Chronological")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)

	// Create cycles in reverse chronological order: cycle3 (latest), then cycle2, then cycle1 (earliest).
	cycle1Start := base
	cycle2Start := base.Add(15 * 24 * time.Hour)
	cycle3Start := base.Add(30 * 24 * time.Hour)

	// Create in reverse order: cycle3, cycle2, cycle1.
	cycle3, _ := cycleStore.Create(cycle3Start, cycle3Start.AddDate(0, 0, 14))
	cycle2, _ := cycleStore.Create(cycle2Start, cycle2Start.AddDate(0, 0, 14))
	cycle1, _ := cycleStore.Create(cycle1Start, cycle1Start.AddDate(0, 0, 14))

	// Pin created_at timestamps to ensure they're in reverse order of period_start and
	// the F8 cutover rule is satisfied (main must be created before cycles).
	mustSetCreatedAt(t, db, "main_attributes", main.ID, base.Add(-24*time.Hour))
	mustSetCreatedAt(t, db, "rating_cycles", cycle3.ID, base.Add(2*24*time.Hour))
	mustSetCreatedAt(t, db, "rating_cycles", cycle2.ID, base.Add(1*24*time.Hour))
	mustSetCreatedAt(t, db, "rating_cycles", cycle1.ID, base)

	// Submit rankings to all three cycles with both active engineers.
	// The key test: Alex gets different scores in each cycle so we can verify which one is selected.
	// cycle1: Alex rank 1 -> score 100, Dummy rank 2 -> score 50
	if _, err := rankingStore.SubmitRanking(cycle1.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err != nil {
		t.Fatalf("submit cycle1 ranking failed: %v", err)
	}
	// cycle2: Alex rank 2 -> score 50, Dummy rank 1 -> score 100
	if _, err := rankingStore.SubmitRanking(cycle2.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 2},
		{EngineerID: e2.ID, Rank: 1},
	}); err != nil {
		t.Fatalf("submit cycle2 ranking failed: %v", err)
	}
	// cycle3: Alex rank 1 -> score 100, Dummy rank 2 -> score 50
	// This is the most recent cycle, so dashboard should return this score for Alex (100).
	if _, err := rankingStore.SubmitRanking(cycle3.ID, sub.ID, []scoring.RankEntry{
		{EngineerID: e1.ID, Rank: 1},
		{EngineerID: e2.ID, Rank: 2},
	}); err != nil {
		t.Fatalf("submit cycle3 ranking failed: %v", err)
	}

	dashboard, err := scoreStore.RosterDashboard(engineerStore)
	if err != nil {
		t.Fatalf("roster dashboard failed: %v", err)
	}

	// Find Alex's entry.
	var alexEntry *models.RosterEntry
	for i := range dashboard {
		if dashboard[i].Engineer.ID == e1.ID {
			alexEntry = &dashboard[i]
			break
		}
	}
	if alexEntry == nil {
		t.Fatalf("expected to find Alex in dashboard")
	}

	// The dashboard should return Alex's latest score from cycle3 (the most recent period),
	// which is 100 (rank 1). If it incorrectly used cycle2 (created second), it would be 50.
	// If it incorrectly used cycle1 (created first), it would be 100 but the period would be wrong.
	if alexEntry.LatestOverall == nil {
		t.Fatalf("expected Alex latest overall to be non-nil from cycle3 (most recent period), got nil")
	}
	if *alexEntry.LatestOverall != 100 {
		t.Fatalf("expected Alex latest overall 100 from cycle3 (most recent period), got %f", *alexEntry.LatestOverall)
	}
	if alexEntry.LastCycleDate == nil {
		t.Fatalf("expected Alex last cycle date to be non-nil (cycle3 period end), got nil")
	}
	if alexEntry.LastCycleDate.Unix() != cycle3.PeriodEnd.Unix() {
		t.Fatalf("expected Alex last cycle date %v (cycle3 period end), got %v", cycle3.PeriodEnd, alexEntry.LastCycleDate)
	}
}

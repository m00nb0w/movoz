package store

import (
	"database/sql"
	"testing"
	"time"

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

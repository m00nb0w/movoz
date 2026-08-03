package store

import (
	"testing"

	"oncarinho/internal/models"
)

func TestSummary(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)
	summaryStore := NewSummaryStore(db)

	alex, _ := playerStore.Create("Alex", nil)
	sam, _ := playerStore.Create("Sam", nil)
	playerStore.Deactivate(sam.ID)

	m1, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	m2, _ := matchdayStore.Create(mustParseDate(t, "2025-06-01"))
	statStore.UpsertBulk(m1.ID, []models.StatInput{{PlayerID: alex.ID, Goals: 3}})
	statStore.UpsertBulk(m2.ID, []models.StatInput{{PlayerID: alex.ID, Goals: 9}})

	summary, err := summaryStore.Summary(2026)
	if err != nil {
		t.Fatalf("Summary failed: %v", err)
	}
	if summary.MatchesPlayed != 1 || summary.GoalsScored != 3 {
		t.Fatalf("unexpected 2026 summary: %+v", summary)
	}
	if summary.RosterSize != 1 {
		t.Fatalf("expected roster size 1 (Sam deactivated), got %d", summary.RosterSize)
	}
}

func TestSummaryMultiPlayerAndZeroStatsMatchdays(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)
	summaryStore := NewSummaryStore(db)

	alex, _ := playerStore.Create("Alex", nil)
	sam, _ := playerStore.Create("Sam", nil)

	// Matchday with stats recorded for two players — proves COUNT(DISTINCT m.id)
	// doesn't double-count this single matchday.
	multiStat, _ := matchdayStore.Create(mustParseDate(t, "2026-04-01"))
	statStore.UpsertBulk(multiStat.ID, []models.StatInput{
		{PlayerID: alex.ID, Goals: 2},
		{PlayerID: sam.ID, Goals: 1},
	})

	// Matchday with zero match_stats rows — proves the LEFT JOIN (not INNER JOIN)
	// still counts it, and COALESCE keeps GoalsScored unaffected by the NULL sum.
	matchdayStore.Create(mustParseDate(t, "2026-05-10"))

	summary, err := summaryStore.Summary(2026)
	if err != nil {
		t.Fatalf("Summary failed: %v", err)
	}
	if summary.MatchesPlayed != 2 {
		t.Fatalf("expected 2 matchdays (multi-stat + zero-stat), got %d", summary.MatchesPlayed)
	}
	if summary.GoalsScored != 3 {
		t.Fatalf("expected 3 goals (2+1 from multi-stat matchday, 0 from zero-stat matchday), got %d", summary.GoalsScored)
	}
}

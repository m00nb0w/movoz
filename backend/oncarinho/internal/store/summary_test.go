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

package store

import (
	"testing"

	"oncarinho/internal/models"
)

func TestLeaderboardGoals(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)
	leaderboardStore := NewLeaderboardStore(db)

	alex, _ := playerStore.Create("Alex", nil)
	sam, _ := playerStore.Create("Sam", nil)

	m2026, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	m2025, _ := matchdayStore.Create(mustParseDate(t, "2025-06-01"))

	statStore.UpsertBulk(m2026.ID, []models.StatInput{
		{PlayerID: alex.ID, Goals: 2},
		{PlayerID: sam.ID, Goals: 1},
	})
	statStore.UpsertBulk(m2025.ID, []models.StatInput{
		{PlayerID: alex.ID, Goals: 5},
	})

	allTime, err := leaderboardStore.Leaderboard(nil, "goals")
	if err != nil {
		t.Fatalf("Leaderboard failed: %v", err)
	}
	if len(allTime) != 2 || allTime[0].PlayerName != "Alex" || allTime[0].Value != 7 {
		t.Fatalf("unexpected all-time leaderboard: %+v", allTime)
	}

	year := 2026
	yearOnly, err := leaderboardStore.Leaderboard(&year, "goals")
	if err != nil {
		t.Fatalf("Leaderboard(2026) failed: %v", err)
	}
	if len(yearOnly) != 2 || yearOnly[0].PlayerName != "Alex" || yearOnly[0].Value != 2 {
		t.Fatalf("unexpected 2026 leaderboard: %+v", yearOnly)
	}
}

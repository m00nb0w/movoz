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

func TestLeaderboardCardsAndAssists(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)
	leaderboardStore := NewLeaderboardStore(db)

	alex, _ := playerStore.Create("Alex", nil)
	sam, _ := playerStore.Create("Sam", nil)

	m1, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	m2, _ := matchdayStore.Create(mustParseDate(t, "2026-06-01"))

	statStore.UpsertBulk(m1.ID, []models.StatInput{
		{PlayerID: alex.ID, Assists: 3, YellowCards: 2, RedCards: 1},
		{PlayerID: sam.ID, Assists: 1, YellowCards: 1, RedCards: 0},
	})
	statStore.UpsertBulk(m2.ID, []models.StatInput{
		{PlayerID: alex.ID, Assists: 2, YellowCards: 1, RedCards: 0},
	})

	// Alex: assists 3+2=5, cards (yellow+red) = (2+1)+(1+0) = 4
	// Sam: assists 1, cards = 1+0 = 1

	cards, err := leaderboardStore.Leaderboard(nil, "cards")
	if err != nil {
		t.Fatalf("Leaderboard(cards) failed: %v", err)
	}
	if len(cards) != 2 || cards[0].PlayerName != "Alex" || cards[0].Value != 4 {
		t.Fatalf("unexpected cards leaderboard: %+v", cards)
	}
	if cards[1].PlayerName != "Sam" || cards[1].Value != 1 {
		t.Fatalf("unexpected cards leaderboard: %+v", cards)
	}

	assists, err := leaderboardStore.Leaderboard(nil, "assists")
	if err != nil {
		t.Fatalf("Leaderboard(assists) failed: %v", err)
	}
	if len(assists) != 2 || assists[0].PlayerName != "Alex" || assists[0].Value != 5 {
		t.Fatalf("unexpected assists leaderboard: %+v", assists)
	}
	if assists[1].PlayerName != "Sam" || assists[1].Value != 1 {
		t.Fatalf("unexpected assists leaderboard: %+v", assists)
	}
}

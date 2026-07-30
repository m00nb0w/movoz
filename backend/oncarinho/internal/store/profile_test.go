package store

import (
	"testing"

	"oncarinho/internal/models"
)

func TestProfileAllTimeAndByYear(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)
	profileStore := NewProfileStore(db, playerStore)

	alex, _ := playerStore.Create("Alex", nil)
	m2026, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	m2025, _ := matchdayStore.Create(mustParseDate(t, "2025-06-01"))

	statStore.UpsertBulk(m2026.ID, []models.StatInput{{PlayerID: alex.ID, Goals: 2, Assists: 1}})
	statStore.UpsertBulk(m2025.ID, []models.StatInput{{PlayerID: alex.ID, Goals: 5, YellowCards: 1}})

	profile, err := profileStore.Profile(alex.ID)
	if err != nil {
		t.Fatalf("Profile failed: %v", err)
	}
	if profile == nil {
		t.Fatal("expected profile, got nil")
	}
	if profile.AllTime.MatchesPlayed != 2 || profile.AllTime.Goals != 7 || profile.AllTime.Assists != 1 || profile.AllTime.YellowCards != 1 {
		t.Fatalf("unexpected all-time totals: %+v", profile.AllTime)
	}
	if len(profile.ByYear) != 2 {
		t.Fatalf("expected 2 years of stats, got %+v", profile.ByYear)
	}

	missing, err := profileStore.Profile(99999)
	if err != nil {
		t.Fatalf("Profile for missing id returned error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing player, got %+v", missing)
	}
}

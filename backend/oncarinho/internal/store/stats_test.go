package store

import (
	"testing"
	"time"

	"oncarinho/internal/models"
)

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("invalid date %q: %v", s, err)
	}
	return d
}

func TestStatStoreUpsertBulk(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)

	player, _ := playerStore.Create("Alex", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))

	entries := []models.StatInput{
		{PlayerID: player.ID, Goals: 2, Assists: 1},
	}
	if err := statStore.UpsertBulk(matchday.ID, entries); err != nil {
		t.Fatalf("UpsertBulk failed: %v", err)
	}

	stats, err := statStore.ListByMatchday(matchday.ID)
	if err != nil {
		t.Fatalf("ListByMatchday failed: %v", err)
	}
	if len(stats) != 1 || stats[0].Goals != 2 || stats[0].Assists != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	entries[0].Goals = 3
	if err := statStore.UpsertBulk(matchday.ID, entries); err != nil {
		t.Fatalf("second UpsertBulk failed: %v", err)
	}
	stats, err = statStore.ListByMatchday(matchday.ID)
	if err != nil {
		t.Fatalf("ListByMatchday failed: %v", err)
	}
	if len(stats) != 1 || stats[0].Goals != 3 {
		t.Fatalf("expected updated goals=3 (no duplicate row), got %+v", stats)
	}
}

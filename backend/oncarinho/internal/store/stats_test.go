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

func TestStatStoreUpsertBulkMultipleEntries(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)

	alex, _ := playerStore.Create("Alex", nil)
	sam, _ := playerStore.Create("Sam", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))

	entries := []models.StatInput{
		{PlayerID: alex.ID, Goals: 2, Assists: 1, YellowCards: 1},
		{PlayerID: sam.ID, Goals: 0, Assists: 3, RedCards: 1},
	}
	if err := statStore.UpsertBulk(matchday.ID, entries); err != nil {
		t.Fatalf("UpsertBulk failed: %v", err)
	}

	stats, err := statStore.ListByMatchday(matchday.ID)
	if err != nil {
		t.Fatalf("ListByMatchday failed: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stat rows, got %+v", stats)
	}

	byPlayer := map[int]models.MatchStat{}
	for _, s := range stats {
		byPlayer[s.PlayerID] = s
	}

	alexStat, ok := byPlayer[alex.ID]
	if !ok || alexStat.Goals != 2 || alexStat.Assists != 1 || alexStat.YellowCards != 1 || alexStat.RedCards != 0 {
		t.Fatalf("unexpected stats for alex: %+v", alexStat)
	}

	samStat, ok := byPlayer[sam.ID]
	if !ok || samStat.Goals != 0 || samStat.Assists != 3 || samStat.YellowCards != 0 || samStat.RedCards != 1 {
		t.Fatalf("unexpected stats for sam: %+v", samStat)
	}
}

func TestStatStoreDelete(t *testing.T) {
	db := setupTestDB(t)
	truncateAll(t, db)

	playerStore := NewPlayerStore(db)
	matchdayStore := NewMatchdayStore(db)
	statStore := NewStatStore(db)

	player, _ := playerStore.Create("Alex", nil)
	matchday, _ := matchdayStore.Create(mustParseDate(t, "2026-03-15"))
	statStore.UpsertBulk(matchday.ID, []models.StatInput{{PlayerID: player.ID, Goals: 2}})

	ok, err := statStore.Delete(matchday.ID, player.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !ok {
		t.Fatal("expected Delete to report success")
	}

	stats, err := statStore.ListByMatchday(matchday.ID)
	if err != nil {
		t.Fatalf("ListByMatchday failed: %v", err)
	}
	if len(stats) != 0 {
		t.Fatalf("expected no stats after delete, got %+v", stats)
	}

	ok, err = statStore.Delete(matchday.ID, player.ID)
	if err != nil {
		t.Fatalf("Delete for missing row returned error: %v", err)
	}
	if ok {
		t.Fatal("expected false for already-deleted row")
	}
}

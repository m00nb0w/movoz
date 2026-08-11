package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMatchdayJSONRoundTrip(t *testing.T) {
	original := Matchday{
		ID:        7,
		PlayedOn:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		CreatedAt: time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded Matchday
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID: got %d, want %d", decoded.ID, original.ID)
	}
	if !decoded.PlayedOn.Equal(original.PlayedOn) {
		t.Errorf("PlayedOn: got %v, want %v", decoded.PlayedOn, original.PlayedOn)
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", decoded.CreatedAt, original.CreatedAt)
	}
}

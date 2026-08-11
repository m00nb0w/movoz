package models

import (
	"encoding/json"
	"time"
)

type Matchday struct {
	ID        int       `json:"id"`
	PlayedOn  time.Time `json:"played_on"`
	CreatedAt time.Time `json:"created_at"`
}

func (m Matchday) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID        int       `json:"id"`
		PlayedOn  string    `json:"played_on"`
		CreatedAt time.Time `json:"created_at"`
	}{
		ID:        m.ID,
		PlayedOn:  m.PlayedOn.Format("2006-01-02"),
		CreatedAt: m.CreatedAt,
	})
}

func (m *Matchday) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID        int       `json:"id"`
		PlayedOn  string    `json:"played_on"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	playedOn, err := time.Parse("2006-01-02", aux.PlayedOn)
	if err != nil {
		return err
	}
	m.ID = aux.ID
	m.PlayedOn = playedOn
	m.CreatedAt = aux.CreatedAt
	return nil
}

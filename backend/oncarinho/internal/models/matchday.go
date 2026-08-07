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

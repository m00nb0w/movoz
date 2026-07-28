package models

import "time"

type Matchday struct {
	ID        int       `json:"id"`
	PlayedOn  time.Time `json:"played_on"`
	CreatedAt time.Time `json:"created_at"`
}

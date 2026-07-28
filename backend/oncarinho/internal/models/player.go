package models

import "time"

type Player struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Position  *string   `json:"position"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

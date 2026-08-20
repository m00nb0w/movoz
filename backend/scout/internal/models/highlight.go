package models

import "time"

type HighlightEntry struct {
	ID         int       `json:"id"`
	EngineerID int       `json:"engineer_id"`
	Kind       string    `json:"kind"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

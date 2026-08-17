package models

import "time"

type MainAttribute struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type SubAttribute struct {
	ID              int       `json:"id"`
	MainAttributeID int       `json:"main_attribute_id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
}

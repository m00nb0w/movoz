package models

import "time"

type RatingCycle struct {
	ID          int       `json:"id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	CreatedAt   time.Time `json:"created_at"`
}

type SubAttributeRanking struct {
	ID             int     `json:"id"`
	CycleID        int     `json:"cycle_id"`
	SubAttributeID int     `json:"sub_attribute_id"`
	EngineerID     int     `json:"engineer_id"`
	Rank           int     `json:"rank"`
	Score          float64 `json:"score"`
}

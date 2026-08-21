package models

import "time"

type MetricSnapshot struct {
	ID              int       `json:"id"`
	EngineerID      int       `json:"engineer_id"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	PRsRaised       int       `json:"prs_raised"`
	PRsReviewed     int       `json:"prs_reviewed"`
	TicketsClosed   int       `json:"tickets_closed"`
	ComplexityScore float64   `json:"complexity_score"`
	SyncedAt        time.Time `json:"synced_at"`
}

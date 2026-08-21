package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type MetricStore struct {
	db *sql.DB
}

func NewMetricStore(db *sql.DB) *MetricStore {
	return &MetricStore{db: db}
}

// UpsertSnapshot idempotently writes a metric snapshot keyed on
// (engineer_id, period_start, period_end) — a repeated or retried sync run
// updates the existing row in place rather than duplicating it (NF4).
func (s *MetricStore) UpsertSnapshot(engineerID int, periodStart, periodEnd time.Time, prsRaised, prsReviewed, ticketsClosed int, complexityScore float64) error {
	_, err := s.db.Exec(
		`INSERT INTO metric_snapshots (engineer_id, period_start, period_end, prs_raised, prs_reviewed, tickets_closed, complexity_score, synced_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 ON CONFLICT (engineer_id, period_start, period_end)
		 DO UPDATE SET prs_raised = $4, prs_reviewed = $5, tickets_closed = $6, complexity_score = $7, synced_at = NOW()`,
		engineerID, periodStart, periodEnd, prsRaised, prsReviewed, ticketsClosed, complexityScore,
	)
	return err
}

func (s *MetricStore) ListByEngineer(engineerID int) ([]models.MetricSnapshot, error) {
	rows, err := s.db.Query(
		`SELECT id, engineer_id, period_start, period_end, prs_raised, prs_reviewed, tickets_closed, complexity_score, synced_at
		 FROM metric_snapshots WHERE engineer_id = $1 ORDER BY period_start DESC`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := []models.MetricSnapshot{}
	for rows.Next() {
		var m models.MetricSnapshot
		if err := rows.Scan(&m.ID, &m.EngineerID, &m.PeriodStart, &m.PeriodEnd, &m.PRsRaised, &m.PRsReviewed, &m.TicketsClosed, &m.ComplexityScore, &m.SyncedAt); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, m)
	}
	return snapshots, rows.Err()
}

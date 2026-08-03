package store

import (
	"database/sql"

	"oncarinho/internal/models"
)

type SummaryStore struct {
	db *sql.DB
}

func NewSummaryStore(db *sql.DB) *SummaryStore {
	return &SummaryStore{db: db}
}

func (s *SummaryStore) Summary(year int) (*models.Summary, error) {
	summary := models.Summary{Year: year}

	err := s.db.QueryRow(`
		SELECT COUNT(DISTINCT m.id), COALESCE(SUM(ms.goals), 0)
		FROM matchdays m
		LEFT JOIN match_stats ms ON ms.matchday_id = m.id
		WHERE EXTRACT(YEAR FROM m.played_on) = $1
	`, year).Scan(&summary.MatchesPlayed, &summary.GoalsScored)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM players WHERE is_active = true").Scan(&summary.RosterSize)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

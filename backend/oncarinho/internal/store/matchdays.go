package store

import (
	"database/sql"
	"time"

	"oncarinho/internal/models"
)

type MatchdayStore struct {
	db *sql.DB
}

func NewMatchdayStore(db *sql.DB) *MatchdayStore {
	return &MatchdayStore{db: db}
}

func (s *MatchdayStore) List(year *int) ([]models.Matchday, error) {
	query := "SELECT id, played_on, created_at FROM matchdays"
	args := []interface{}{}
	if year != nil {
		query += " WHERE EXTRACT(YEAR FROM played_on) = $1"
		args = append(args, *year)
	}
	query += " ORDER BY played_on DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matchdays := []models.Matchday{}
	for rows.Next() {
		var m models.Matchday
		if err := rows.Scan(&m.ID, &m.PlayedOn, &m.CreatedAt); err != nil {
			return nil, err
		}
		matchdays = append(matchdays, m)
	}
	return matchdays, rows.Err()
}

func (s *MatchdayStore) Create(playedOn time.Time) (*models.Matchday, error) {
	var m models.Matchday
	err := s.db.QueryRow(
		`INSERT INTO matchdays (played_on) VALUES ($1)
		 RETURNING id, played_on, created_at`,
		playedOn,
	).Scan(&m.ID, &m.PlayedOn, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *MatchdayStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM matchdays WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

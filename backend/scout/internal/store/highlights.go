package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

// timeNow is a thin indirection so store tests can call a package-local
// helper without importing "time" redundantly in every test file.
func timeNow() time.Time { return time.Now() }

type HighlightStore struct {
	db *sql.DB
}

func NewHighlightStore(db *sql.DB) *HighlightStore {
	return &HighlightStore{db: db}
}

func (s *HighlightStore) List(engineerID int) ([]models.HighlightEntry, error) {
	rows, err := s.db.Query(
		`SELECT id, engineer_id, kind, body, created_at
		 FROM highlight_entries WHERE engineer_id = $1 ORDER BY created_at DESC`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.HighlightEntry{}
	for rows.Next() {
		var e models.HighlightEntry
		if err := rows.Scan(&e.ID, &e.EngineerID, &e.Kind, &e.Body, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *HighlightStore) Create(engineerID int, kind, body string) (*models.HighlightEntry, error) {
	var e models.HighlightEntry
	err := s.db.QueryRow(
		`INSERT INTO highlight_entries (engineer_id, kind, body) VALUES ($1, $2, $3)
		 RETURNING id, engineer_id, kind, body, created_at`,
		engineerID, kind, body,
	).Scan(&e.ID, &e.EngineerID, &e.Kind, &e.Body, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

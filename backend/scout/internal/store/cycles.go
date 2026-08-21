package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type CycleStore struct {
	db *sql.DB
}

func NewCycleStore(db *sql.DB) *CycleStore {
	return &CycleStore{db: db}
}

func (s *CycleStore) List() ([]models.RatingCycle, error) {
	rows, err := s.db.Query("SELECT id, period_start, period_end, created_at FROM rating_cycles ORDER BY period_start DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cycles := []models.RatingCycle{}
	for rows.Next() {
		var c models.RatingCycle
		if err := rows.Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt); err != nil {
			return nil, err
		}
		cycles = append(cycles, c)
	}
	return cycles, rows.Err()
}

func (s *CycleStore) Create(periodStart, periodEnd time.Time) (*models.RatingCycle, error) {
	var c models.RatingCycle
	err := s.db.QueryRow(
		`INSERT INTO rating_cycles (period_start, period_end) VALUES ($1, $2)
		 RETURNING id, period_start, period_end, created_at`,
		periodStart, periodEnd,
	).Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CycleStore) GetByID(id int) (*models.RatingCycle, error) {
	var c models.RatingCycle
	err := s.db.QueryRow("SELECT id, period_start, period_end, created_at FROM rating_cycles WHERE id = $1", id).
		Scan(&c.ID, &c.PeriodStart, &c.PeriodEnd, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CycleStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM rating_cycles WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

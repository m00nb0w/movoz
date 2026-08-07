package store

import (
	"database/sql"

	"oncarinho/internal/models"
)

type PlayerStore struct {
	db *sql.DB
}

func NewPlayerStore(db *sql.DB) *PlayerStore {
	return &PlayerStore{db: db}
}

func (s *PlayerStore) List(activeOnly bool) ([]models.Player, error) {
	query := "SELECT id, name, position, is_active, created_at FROM players"
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var p models.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.Position, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (s *PlayerStore) GetByID(id int) (*models.Player, error) {
	var p models.Player
	err := s.db.QueryRow(
		"SELECT id, name, position, is_active, created_at FROM players WHERE id = $1",
		id,
	).Scan(&p.ID, &p.Name, &p.Position, &p.IsActive, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PlayerStore) Create(name string, position *string) (*models.Player, error) {
	var p models.Player
	err := s.db.QueryRow(
		`INSERT INTO players (name, position, is_active)
		 VALUES ($1, $2, true)
		 RETURNING id, name, position, is_active, created_at`,
		name, position,
	).Scan(&p.ID, &p.Name, &p.Position, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PlayerStore) Update(id int, name string, position *string) (*models.Player, error) {
	var p models.Player
	err := s.db.QueryRow(
		`UPDATE players SET name = $1, position = $2 WHERE id = $3
		 RETURNING id, name, position, is_active, created_at`,
		name, position, id,
	).Scan(&p.ID, &p.Name, &p.Position, &p.IsActive, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PlayerStore) Deactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE players SET is_active = false WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *PlayerStore) Reactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE players SET is_active = true WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *PlayerStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM players WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

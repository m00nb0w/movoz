package store

import (
	"database/sql"

	"scout/internal/models"
)

type MainAttributeStore struct {
	db *sql.DB
}

func NewMainAttributeStore(db *sql.DB) *MainAttributeStore {
	return &MainAttributeStore{db: db}
}

func (s *MainAttributeStore) List() ([]models.MainAttribute, error) {
	rows, err := s.db.Query("SELECT id, key, name, created_at FROM main_attributes ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attrs := []models.MainAttribute{}
	for rows.Next() {
		var a models.MainAttribute
		if err := rows.Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	return attrs, rows.Err()
}

func (s *MainAttributeStore) GetByID(id int) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow("SELECT id, key, name, created_at FROM main_attributes WHERE id = $1", id).
		Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Create(key, name string) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow(
		`INSERT INTO main_attributes (key, name) VALUES ($1, $2)
		 RETURNING id, key, name, created_at`,
		key, name,
	).Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Update(id int, name string) (*models.MainAttribute, error) {
	var a models.MainAttribute
	err := s.db.QueryRow(
		"UPDATE main_attributes SET name = $1 WHERE id = $2 RETURNING id, key, name, created_at",
		name, id,
	).Scan(&a.ID, &a.Key, &a.Name, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *MainAttributeStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM main_attributes WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

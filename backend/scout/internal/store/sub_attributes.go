package store

import (
	"database/sql"

	"scout/internal/models"
)

type SubAttributeStore struct {
	db *sql.DB
}

func NewSubAttributeStore(db *sql.DB) *SubAttributeStore {
	return &SubAttributeStore{db: db}
}

func (s *SubAttributeStore) ListByMainAttribute(mainAttributeID int, activeOnly bool) ([]models.SubAttribute, error) {
	query := `SELECT id, main_attribute_id, name, description, is_active, created_at
	          FROM sub_attributes WHERE main_attribute_id = $1`
	if activeOnly {
		query += " AND is_active = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query, mainAttributeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []models.SubAttribute{}
	for rows.Next() {
		var sa models.SubAttribute
		if err := rows.Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sa)
	}
	return subs, rows.Err()
}

func (s *SubAttributeStore) ListAllActive() ([]models.SubAttribute, error) {
	rows, err := s.db.Query(
		`SELECT id, main_attribute_id, name, description, is_active, created_at
		 FROM sub_attributes WHERE is_active = true ORDER BY main_attribute_id, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []models.SubAttribute{}
	for rows.Next() {
		var sa models.SubAttribute
		if err := rows.Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sa)
	}
	return subs, rows.Err()
}

func (s *SubAttributeStore) GetByID(id int) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`SELECT id, main_attribute_id, name, description, is_active, created_at
		 FROM sub_attributes WHERE id = $1`, id,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Create(mainAttributeID int, name string, description *string) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`INSERT INTO sub_attributes (main_attribute_id, name, description, is_active)
		 VALUES ($1, $2, $3, true)
		 RETURNING id, main_attribute_id, name, description, is_active, created_at`,
		mainAttributeID, name, description,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Update(id int, name string, description *string) (*models.SubAttribute, error) {
	var sa models.SubAttribute
	err := s.db.QueryRow(
		`UPDATE sub_attributes SET name = $1, description = $2 WHERE id = $3
		 RETURNING id, main_attribute_id, name, description, is_active, created_at`,
		name, description, id,
	).Scan(&sa.ID, &sa.MainAttributeID, &sa.Name, &sa.Description, &sa.IsActive, &sa.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

func (s *SubAttributeStore) Deactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE sub_attributes SET is_active = false WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *SubAttributeStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM sub_attributes WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

package store

import (
	"database/sql"
	"time"

	"scout/internal/models"
)

type EngineerStore struct {
	db *sql.DB
}

func NewEngineerStore(db *sql.DB) *EngineerStore {
	return &EngineerStore{db: db}
}

func (s *EngineerStore) List(activeOnly bool) ([]models.Engineer, error) {
	query := `SELECT id, name, role, github_username, jira_account_id, started_at, is_active, created_at FROM engineers`
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	engineers := []models.Engineer{}
	for rows.Next() {
		var e models.Engineer
		if err := rows.Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt); err != nil {
			return nil, err
		}
		engineers = append(engineers, e)
	}
	return engineers, rows.Err()
}

func (s *EngineerStore) GetByID(id int) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`SELECT id, name, role, github_username, jira_account_id, started_at, is_active, created_at
		 FROM engineers WHERE id = $1`, id,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Create(name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`INSERT INTO engineers (name, role, github_username, jira_account_id, started_at, is_active)
		 VALUES ($1, $2, $3, $4, $5, true)
		 RETURNING id, name, role, github_username, jira_account_id, started_at, is_active, created_at`,
		name, role, githubUsername, jiraAccountID, startedAt,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Update(id int, name string, role, githubUsername, jiraAccountID *string, startedAt time.Time) (*models.Engineer, error) {
	var e models.Engineer
	err := s.db.QueryRow(
		`UPDATE engineers SET name = $1, role = $2, github_username = $3, jira_account_id = $4, started_at = $5
		 WHERE id = $6
		 RETURNING id, name, role, github_username, jira_account_id, started_at, is_active, created_at`,
		name, role, githubUsername, jiraAccountID, startedAt, id,
	).Scan(&e.ID, &e.Name, &e.Role, &e.GitHubUsername, &e.JiraAccountID, &e.StartedAt, &e.IsActive, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *EngineerStore) Deactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE engineers SET is_active = false WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *EngineerStore) Reactivate(id int) (bool, error) {
	res, err := s.db.Exec("UPDATE engineers SET is_active = true WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (s *EngineerStore) Exists(id int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM engineers WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (s *EngineerStore) ListActiveIDs() ([]int, error) {
	rows, err := s.db.Query("SELECT id FROM engineers WHERE is_active = true ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

package store

import (
	"database/sql"
	"errors"
	"fmt"

	"el-storko/internal/models"
)

var (
	ErrEpicCannotHaveParent  = errors.New("an epic cannot have a parent")
	ErrParentMustBeEpic      = errors.New("parent_id must reference an epic")
	ErrParentNotFound        = errors.New("parent_id does not reference an existing item")
)

type WorkItemStore struct {
	db *sql.DB
}

func NewWorkItemStore(db *sql.DB) *WorkItemStore {
	return &WorkItemStore{db: db}
}

type ListFilters struct {
	Source   *models.Source
	ParentID *int64
	Type     *models.Type
	Status   *models.Status
}

type UpdateFields struct {
	Title       *string
	Description *string
	Status      *models.Status
	ParentID    *int64
	ClearParent bool
}

const workItemColumns = `id, type, parent_id, title, description, status, source, jira_key, jira_url, created_at, updated_at`

func scanWorkItem(row *sql.Row) (*models.WorkItem, error) {
	var w models.WorkItem
	err := row.Scan(&w.ID, &w.Type, &w.ParentID, &w.Title, &w.Description, &w.Status, &w.Source, &w.JiraKey, &w.JiraURL, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// validateParent enforces FR-002: an epic has no parent, a task's parent (if
// set) must reference an existing row of type epic.
func (s *WorkItemStore) validateParent(itemType models.Type, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	if itemType == models.TypeEpic {
		return ErrEpicCannotHaveParent
	}
	var parentType models.Type
	err := s.db.QueryRow("SELECT type FROM work_items WHERE id = $1", *parentID).Scan(&parentType)
	if err == sql.ErrNoRows {
		return ErrParentNotFound
	}
	if err != nil {
		return err
	}
	if parentType != models.TypeEpic {
		return ErrParentMustBeEpic
	}
	return nil
}

func (s *WorkItemStore) Create(item models.WorkItem) (*models.WorkItem, error) {
	if err := s.validateParent(item.Type, item.ParentID); err != nil {
		return nil, err
	}

	status := item.Status
	if status == "" {
		status = models.StatusTodo
	}
	source := item.Source
	if source == "" {
		source = models.SourcePersonal
	}

	row := s.db.QueryRow(
		fmt.Sprintf(`INSERT INTO work_items (type, parent_id, title, description, status, source, jira_key, jira_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING %s`, workItemColumns),
		item.Type, item.ParentID, item.Title, item.Description, status, source, item.JiraKey, item.JiraURL,
	)
	return scanWorkItem(row)
}

func (s *WorkItemStore) Get(id int64) (*models.WorkItem, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`SELECT %s FROM work_items WHERE id = $1`, workItemColumns),
		id,
	)
	return scanWorkItem(row)
}

func (s *WorkItemStore) List(filters ListFilters) ([]models.WorkItem, error) {
	query := fmt.Sprintf(`SELECT %s FROM work_items WHERE 1=1`, workItemColumns)
	args := []any{}
	argN := 1

	if filters.Source != nil {
		query += fmt.Sprintf(" AND source = $%d", argN)
		args = append(args, *filters.Source)
		argN++
	}
	if filters.ParentID != nil {
		query += fmt.Sprintf(" AND parent_id = $%d", argN)
		args = append(args, *filters.ParentID)
		argN++
	}
	if filters.Type != nil {
		query += fmt.Sprintf(" AND type = $%d", argN)
		args = append(args, *filters.Type)
		argN++
	}
	if filters.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, *filters.Status)
		argN++
	}
	query += " ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.WorkItem{}
	for rows.Next() {
		var w models.WorkItem
		if err := rows.Scan(&w.ID, &w.Type, &w.ParentID, &w.Title, &w.Description, &w.Status, &w.Source, &w.JiraKey, &w.JiraURL, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, rows.Err()
}

func (s *WorkItemStore) Update(id int64, fields UpdateFields) (*models.WorkItem, error) {
	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if fields.ParentID != nil || fields.ClearParent {
		var newParent *int64
		if !fields.ClearParent {
			newParent = fields.ParentID
		}
		if err := s.validateParent(existing.Type, newParent); err != nil {
			return nil, err
		}
	}

	title := existing.Title
	if fields.Title != nil {
		title = *fields.Title
	}
	description := existing.Description
	if fields.Description != nil {
		description = *fields.Description
	}
	status := existing.Status
	if fields.Status != nil {
		status = *fields.Status
	}
	parentID := existing.ParentID
	if fields.ClearParent {
		parentID = nil
	} else if fields.ParentID != nil {
		parentID = fields.ParentID
	}

	row := s.db.QueryRow(
		fmt.Sprintf(`UPDATE work_items SET title = $1, description = $2, status = $3, parent_id = $4, updated_at = now()
		 WHERE id = $5 RETURNING %s`, workItemColumns),
		title, description, status, parentID, id,
	)
	return scanWorkItem(row)
}

// UpdateFromSync applies a Jira-originated pull without going through the
// parent/type validation a user-facing write needs, since jira issues have
// no local parent concept.
func (s *WorkItemStore) UpdateFromSync(id int64, title, description string, status models.Status, jiraURL string) (*models.WorkItem, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`UPDATE work_items SET title = $1, description = $2, status = $3, jira_url = $4, updated_at = now()
		 WHERE id = $5 RETURNING %s`, workItemColumns),
		title, description, status, jiraURL, id,
	)
	return scanWorkItem(row)
}

func (s *WorkItemStore) GetByJiraKey(jiraKey string) (*models.WorkItem, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`SELECT %s FROM work_items WHERE jira_key = $1`, workItemColumns),
		jiraKey,
	)
	return scanWorkItem(row)
}

func (s *WorkItemStore) CreateFromJira(title, description string, status models.Status, jiraKey, jiraURL string) (*models.WorkItem, error) {
	row := s.db.QueryRow(
		fmt.Sprintf(`INSERT INTO work_items (type, title, description, status, source, jira_key, jira_url)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING %s`, workItemColumns),
		models.TypeTask, title, description, status, models.SourceJira, jiraKey, jiraURL,
	)
	return scanWorkItem(row)
}

func (s *WorkItemStore) ListBySource(source models.Source) ([]models.WorkItem, error) {
	return s.List(ListFilters{Source: &source})
}

func (s *WorkItemStore) Delete(id int64) (bool, error) {
	res, err := s.db.Exec("DELETE FROM work_items WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

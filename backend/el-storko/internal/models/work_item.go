package models

import "time"

type Type string

const (
	TypeEpic Type = "epic"
	TypeTask Type = "task"
)

func (t Type) Valid() bool {
	return t == TypeEpic || t == TypeTask
}

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDone       Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusBlocked, StatusDone:
		return true
	}
	return false
}

type Source string

const (
	SourcePersonal Source = "personal"
	SourceJira     Source = "jira"
	SourceAgent    Source = "agent"
)

func (s Source) Valid() bool {
	switch s {
	case SourcePersonal, SourceJira, SourceAgent:
		return true
	}
	return false
}

type WorkItem struct {
	ID          int64      `json:"id"`
	Type        Type       `json:"type"`
	ParentID    *int64     `json:"parent_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Source      Source     `json:"source"`
	JiraKey     *string    `json:"jira_key"`
	JiraURL     *string    `json:"jira_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

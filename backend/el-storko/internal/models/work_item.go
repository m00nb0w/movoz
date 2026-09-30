package models

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	StatusBacklog        Status = "backlog"
	StatusPickedForToday Status = "picked_for_today"
	StatusInProgress     Status = "in_progress"
	StatusBlocked        Status = "blocked"
	StatusDone           Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case StatusBacklog, StatusPickedForToday, StatusInProgress, StatusBlocked, StatusDone:
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
	ID              int64      `json:"id"`
	Type            Type       `json:"type"`
	ParentID        *int64     `json:"parent_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          Status     `json:"status"`
	Source          Source     `json:"source"`
	EstimateHours   *float64   `json:"estimate_hours"`
	DueDate         *time.Time `json:"-"`
	ReferenceNumber int        `json:"-"`
	JiraKey         *string    `json:"jira_key"`
	JiraURL         *string    `json:"jira_url"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at"`
}

// ReferenceKey is the human-facing identifier (FR-018) — EPIC-# or TASK-#,
// computed from the type-specific reference_number rather than stored as a
// string.
func (w WorkItem) ReferenceKey() string {
	prefix := "TASK"
	if w.Type == TypeEpic {
		prefix = "EPIC"
	}
	return fmt.Sprintf("%s-%d", prefix, w.ReferenceNumber)
}

// DateOnlyLayout is the wire format for DueDate — a plain calendar date, no
// time-of-day or timezone (FR-019/FR-020).
const DateOnlyLayout = "2006-01-02"

func (w WorkItem) MarshalJSON() ([]byte, error) {
	type Alias WorkItem
	var dueDate *string
	if w.DueDate != nil {
		s := w.DueDate.Format(DateOnlyLayout)
		dueDate = &s
	}
	return json.Marshal(struct {
		Alias
		ReferenceKey string  `json:"reference_key"`
		DueDate      *string `json:"due_date"`
	}{
		Alias:        Alias(w),
		ReferenceKey: w.ReferenceKey(),
		DueDate:      dueDate,
	})
}

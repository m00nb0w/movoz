package models

import "time"

type Engineer struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Role           *string   `json:"role"`
	GitHubUsername *string   `json:"github_username"`
	JiraAccountID  *string   `json:"jira_account_id"`
	StartedAt      time.Time `json:"started_at"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

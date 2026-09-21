package jirasync

import (
	"el-storko/internal/store"
)

// RunSyncCycle performs one pull-then-push poll cycle against every Jira
// issue currently assigned to the user, resolving conflicts by last-write-
// wins on updated_at (FR-007). A failure talking to Jira returns an error
// but never touches the store in a way that could break the REST API for
// personal/agent items (FR-013).
func RunSyncCycle(s *store.WorkItemStore, client Client) error {
	issues, err := client.SearchAssignedIssues()
	if err != nil {
		return err
	}

	for _, issue := range issues {
		local, err := s.GetByJiraKey(issue.Key)
		if err != nil {
			return err
		}

		if local == nil {
			if _, err := s.CreateFromJira(issue.Title, issue.Description, issue.Status, issue.Key, issue.URL); err != nil {
				return err
			}
			continue
		}

		if issue.UpdatedAt.After(local.UpdatedAt) {
			if _, err := s.UpdateFromSync(local.ID, issue.Title, issue.Description, issue.Status, issue.URL); err != nil {
				return err
			}
			continue
		}

		if local.UpdatedAt.After(issue.UpdatedAt) {
			if err := client.UpdateIssue(issue.Key, local.Title, local.Description, local.Status); err != nil {
				return err
			}
		}
	}

	return nil
}

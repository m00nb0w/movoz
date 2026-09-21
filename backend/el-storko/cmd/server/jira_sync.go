package main

import (
	"database/sql"
	"log"
	"time"

	"el-storko/internal/config"
	"el-storko/internal/jirasync"
	"el-storko/internal/store"
)

// startJiraSync runs the poll cycle on a ticker in the background. Missing
// Jira credentials disable sync entirely rather than crashing the server,
// keeping personal work items usable with no Jira configured (FR-013).
func startJiraSync(db *sql.DB, cfg *config.Config) {
	if cfg.JiraEmail == "" || cfg.JiraAPIToken == "" || cfg.JiraBaseURL == "" {
		log.Println("jira sync disabled: JIRA_EMAIL/JIRA_API_TOKEN/JIRA_BASE_URL not set")
		return
	}

	client := jirasync.NewHTTPClient(cfg.JiraEmail, cfg.JiraAPIToken, cfg.JiraBaseURL)
	workItemStore := store.NewWorkItemStore(db)

	ticker := time.NewTicker(cfg.JiraSyncInterval)
	go func() {
		for range ticker.C {
			if err := jirasync.RunSyncCycle(workItemStore, client); err != nil {
				log.Printf("jira sync cycle failed: %v", err)
			}
		}
	}()
}

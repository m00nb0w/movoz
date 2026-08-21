// Package syncer orchestrates Scout's metrics sync: for every active
// engineer, pull PR and ticket activity from GitHub/Jira for the configured
// sync period and idempotently upsert the combined result via MetricStore
// (F4).
package syncer

import (
	"context"
	"log"
	"time"

	"scout/internal/store"
)

// githubBoundaryOffset compensates for GitHubClient.FetchPRStats's inclusive
// date range (see its doc comment in internal/integrations/github.go): a
// sync period's periodEnd is shared with the *next* cycle's periodStart
// whenever the admin opens rating cycles back-to-back (RunSyncCycle syncs the
// current rating cycle's own window — see its doc comment), so calling
// GitHub with an inclusive until == periodEnd would double-count PRs
// created on that boundary day in both this cycle's and the next cycle's
// snapshot. Shifting only GitHub's until back by one day removes the
// overlap: this cycle's GitHub range becomes [periodStart, periodEnd-1],
// and the next cycle's becomes [periodEnd, nextPeriodEnd-1] — adjacent, not
// overlapping. Jira's FetchTicketStats range is already half-open
// ([since, until)) and is called with periodStart/periodEnd unchanged; no
// offset is needed there.
const githubBoundaryOffset = -24 * time.Hour

// GitHubStatsFetcher is the subset of integrations.GitHubClient the syncer
// depends on, so tests can substitute a fake.
type GitHubStatsFetcher interface {
	FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (prsRaised, prsReviewed int, err error)
}

// JiraStatsFetcher is the subset of integrations.JiraClient the syncer
// depends on, so tests can substitute a fake.
type JiraStatsFetcher interface {
	FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (ticketsClosed int, complexityScore float64, err error)
}

// Syncer pulls per-engineer metrics from GitHub and Jira and writes them to
// the metric_snapshots table via MetricStore.
type Syncer struct {
	engineerStore *store.EngineerStore
	metricStore   *store.MetricStore
	github        GitHubStatsFetcher
	jira          JiraStatsFetcher
	repos         []string
	projects      []string
}

// NewSyncer builds a Syncer. repos/projects are the configured
// SCOUT_GITHUB_REPOS/SCOUT_JIRA_PROJECTS sets, forwarded to every fetch
// call unchanged.
func NewSyncer(engineerStore *store.EngineerStore, metricStore *store.MetricStore, github GitHubStatsFetcher, jira JiraStatsFetcher, repos, projects []string) *Syncer {
	return &Syncer{
		engineerStore: engineerStore,
		metricStore:   metricStore,
		github:        github,
		jira:          jira,
		repos:         repos,
		projects:      projects,
	}
}

// RunOnce polls GitHub/Jira for every active engineer over the caller's
// sync period and idempotently upserts a metric_snapshots row per engineer
// (F4). periodStart/periodEnd are treated as half-open ([periodStart,
// periodEnd)) and passed to Jira unchanged; GitHub is called with its until
// shifted back by githubBoundaryOffset to avoid double-counting the shared
// boundary day across consecutive sync cycles (see githubBoundaryOffset's
// doc comment).
//
// A single engineer's fetch or upsert failure is logged and skipped rather
// than aborting the whole run (NF4) — the next scheduled run will retry it,
// and MetricStore.UpsertSnapshot's idempotent upsert means retries never
// duplicate or corrupt data.
func (s *Syncer) RunOnce(ctx context.Context, periodStart, periodEnd time.Time) error {
	activeIDs, err := s.engineerStore.ListActiveIDs()
	if err != nil {
		return err
	}

	githubUntil := periodEnd.Add(githubBoundaryOffset)

	for _, id := range activeIDs {
		engineer, err := s.engineerStore.GetByID(id)
		if err != nil {
			log.Printf("scout sync: failed to load engineer %d: %v", id, err)
			continue
		}
		if engineer == nil {
			continue
		}

		var prsRaised, prsReviewed int
		if engineer.GitHubUsername != nil {
			prsRaised, prsReviewed, err = s.github.FetchPRStats(ctx, *engineer.GitHubUsername, s.repos, periodStart, githubUntil)
			if err != nil {
				log.Printf("scout sync: github fetch failed for engineer %d (%s): %v", id, *engineer.GitHubUsername, err)
				continue
			}
		}

		var ticketsClosed int
		var complexityScore float64
		if engineer.JiraAccountID != nil {
			ticketsClosed, complexityScore, err = s.jira.FetchTicketStats(ctx, *engineer.JiraAccountID, s.projects, periodStart, periodEnd)
			if err != nil {
				log.Printf("scout sync: jira fetch failed for engineer %d (%s): %v", id, *engineer.JiraAccountID, err)
				continue
			}
		}

		if err := s.metricStore.UpsertSnapshot(id, periodStart, periodEnd, prsRaised, prsReviewed, ticketsClosed, complexityScore); err != nil {
			log.Printf("scout sync: upsert failed for engineer %d: %v", id, err)
			continue
		}
	}

	return nil
}

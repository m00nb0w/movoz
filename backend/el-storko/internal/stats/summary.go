package stats

import (
	"time"

	"el-storko/internal/models"
)

const completedThisWeekWindowDays = 7

var allStatuses = []models.Status{
	models.StatusBacklog,
	models.StatusPickedForToday,
	models.StatusInProgress,
	models.StatusBlocked,
	models.StatusDone,
}

// Summary is the point-in-time workload snapshot for the Stats tab (FR-017),
// separate from ComputeBurnRate's daily trend series so future summary
// metrics extend this struct/reducer without touching the trend one.
type Summary struct {
	OpenItems         int
	CompletedThisWeek int
	CompletionRate    float64
	TotalTracked      int
	StatusBreakdown   map[models.Status]int
}

// ComputeSummary reduces work items into the Stats tab's summary numbers per
// wiki/technical/001-el-storko/data-model.md's Workload & Burn-Rate Stats
// section. All five status keys are always present in StatusBreakdown, even
// when their count is zero, so the bar chart always renders five bars.
func ComputeSummary(items []models.WorkItem, now time.Time) Summary {
	breakdown := make(map[models.Status]int, len(allStatuses))
	for _, s := range allStatuses {
		breakdown[s] = 0
	}

	windowStart := now.AddDate(0, 0, -completedThisWeekWindowDays)

	var openItems, completedThisWeek, doneCount int
	for _, item := range items {
		breakdown[item.Status]++

		if item.Status != models.StatusDone {
			openItems++
		} else {
			doneCount++
		}

		if item.CompletedAt != nil {
			completedAt := *item.CompletedAt
			if !completedAt.Before(windowStart) && !completedAt.After(now) {
				completedThisWeek++
			}
		}
	}

	total := len(items)
	var completionRate float64
	if total > 0 {
		completionRate = float64(doneCount) / float64(total)
	}

	return Summary{
		OpenItems:         openItems,
		CompletedThisWeek: completedThisWeek,
		CompletionRate:    completionRate,
		TotalTracked:      total,
		StatusBreakdown:   breakdown,
	}
}

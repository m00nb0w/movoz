package stats

import (
	"time"

	"el-storko/internal/models"
)

const defaultWindowDays = 30

// Point is one calendar day's burn-rate reading (FR-016).
type Point struct {
	Date      string `json:"date"`
	Completed int    `json:"completed"`
	Open      int    `json:"open"`
}

// ComputeBurnRate reduces work items into a daily throughput/backlog series
// over the trailing `days` window ending today, per
// wiki/technical/001-el-storko/data-model.md's Burn-Rate Stats section. It
// only reads CreatedAt/CompletedAt, so it doesn't need to know an item's
// status-change history — the store layer already guarantees CompletedAt
// reflects the most recent done transition (FR-015).
func ComputeBurnRate(items []models.WorkItem, days int, now time.Time) []Point {
	if days <= 0 {
		days = defaultWindowDays
	}

	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	points := make([]Point, days)
	for i := 0; i < days; i++ {
		dayStart := today.AddDate(0, 0, -(days - 1 - i))
		dayEnd := dayStart.AddDate(0, 0, 1)

		var completed, open int
		for _, item := range items {
			if item.CompletedAt != nil {
				completedAt := item.CompletedAt.In(loc)
				if !completedAt.Before(dayStart) && completedAt.Before(dayEnd) {
					completed++
				}
			}

			createdByDayEnd := item.CreatedAt.In(loc).Before(dayEnd)
			stillOpenAtDayEnd := item.CompletedAt == nil || !item.CompletedAt.In(loc).Before(dayEnd)
			if createdByDayEnd && stillOpenAtDayEnd {
				open++
			}
		}

		points[i] = Point{Date: dayStart.Format("2006-01-02"), Completed: completed, Open: open}
	}

	return points
}

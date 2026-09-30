package handlers

import (
	"net/http"
	"strconv"
	"time"

	"el-storko/internal/stats"
	"el-storko/internal/store"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	store *store.WorkItemStore
}

func NewStatsHandler(s *store.WorkItemStore) *StatsHandler {
	return &StatsHandler{store: s}
}

// statsScopeFilters defaults to "mine" (FR-022) and rejects an unrecognized
// scope value, mirroring work_item_handler.go's List endpoint.
func statsScopeFilters(c *gin.Context) (store.ListFilters, bool) {
	scope := c.Query("scope")
	if scope == "" {
		scope = "mine"
	}
	sources, ok := scopeSources(scope)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
		return store.ListFilters{}, false
	}
	return store.ListFilters{Sources: sources}, true
}

func (h *StatsHandler) BurnRate(c *gin.Context) {
	days := 30
	if v := c.Query("days"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days"})
			return
		}
		days = parsed
	}

	filters, ok := statsScopeFilters(c)
	if !ok {
		return
	}
	items, err := h.store.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	points := stats.ComputeBurnRate(items, days, time.Now())
	c.JSON(http.StatusOK, gin.H{"days": days, "points": points})
}

func (h *StatsHandler) Summary(c *gin.Context) {
	filters, ok := statsScopeFilters(c)
	if !ok {
		return
	}
	items, err := h.store.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	summary := stats.ComputeSummary(items, time.Now())
	c.JSON(http.StatusOK, gin.H{
		"open_items":          summary.OpenItems,
		"completed_this_week": summary.CompletedThisWeek,
		"completion_rate":     summary.CompletionRate,
		"total_tracked":       summary.TotalTracked,
		"status_breakdown":    summary.StatusBreakdown,
	})
}

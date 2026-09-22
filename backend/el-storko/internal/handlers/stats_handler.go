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

	items, err := h.store.List(store.ListFilters{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	points := stats.ComputeBurnRate(items, days, time.Now())
	c.JSON(http.StatusOK, gin.H{"days": days, "points": points})
}

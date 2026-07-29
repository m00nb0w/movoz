package handlers

import (
	"net/http"
	"strconv"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	store *store.LeaderboardStore
}

func NewLeaderboardHandler(s *store.LeaderboardStore) *LeaderboardHandler {
	return &LeaderboardHandler{store: s}
}

var validLeaderboardStats = map[string]bool{"goals": true, "assists": true, "cards": true}

func (h *LeaderboardHandler) Get(c *gin.Context) {
	stat := c.DefaultQuery("stat", "goals")
	if !validLeaderboardStats[stat] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stat must be one of goals, assists, cards"})
		return
	}

	yearParam := c.DefaultQuery("year", "all")
	var year *int
	if yearParam != "all" {
		y, err := strconv.Atoi(yearParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "year must be 'all' or a 4-digit year"})
			return
		}
		year = &y
	}

	entries, err := h.store.Leaderboard(year, stat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load leaderboard"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

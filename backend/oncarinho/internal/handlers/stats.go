package handlers

import (
	"net/http"
	"strconv"

	"oncarinho/internal/models"
	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type StatHandler struct {
	statStore     *store.StatStore
	matchdayStore *store.MatchdayStore
	playerStore   *store.PlayerStore
}

func NewStatHandler(statStore *store.StatStore, matchdayStore *store.MatchdayStore, playerStore *store.PlayerStore) *StatHandler {
	return &StatHandler{statStore: statStore, matchdayStore: matchdayStore, playerStore: playerStore}
}

type upsertStatsRequest struct {
	Entries []models.StatInput `json:"entries" binding:"required,dive"`
}

func (h *StatHandler) UpsertStats(c *gin.Context) {
	matchdayID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday id"})
		return
	}

	exists, err := h.matchdayStore.Exists(matchdayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up matchday"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "matchday not found"})
		return
	}

	var req upsertStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entries are required"})
		return
	}

	for _, e := range req.Entries {
		playerExists, err := h.playerStore.Exists(e.PlayerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up player"})
			return
		}
		if !playerExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown player_id"})
			return
		}
	}

	if err := h.statStore.UpsertBulk(matchdayID, req.Entries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save stats"})
		return
	}

	stats, err := h.statStore.ListByMatchday(matchdayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load saved stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *StatHandler) GetStats(c *gin.Context) {
	matchdayID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday id"})
		return
	}

	exists, err := h.matchdayStore.Exists(matchdayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up matchday"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "matchday not found"})
		return
	}

	stats, err := h.statStore.ListByMatchday(matchdayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *StatHandler) DeleteStat(c *gin.Context) {
	matchdayID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid matchday id"})
		return
	}
	playerID, err := strconv.Atoi(c.Param("playerId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	ok, err := h.statStore.Delete(matchdayID, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete stat"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "stat not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

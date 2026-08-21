package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type EngineerCardHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
}

func NewEngineerCardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *EngineerCardHandler {
	return &EngineerCardHandler{scoreStore: s, engineerStore: engineerStore}
}

func (h *EngineerCardHandler) Card(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	cycleID, err := strconv.Atoi(c.Query("cycleId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cycleId query param is required"})
		return
	}

	card, err := h.scoreStore.EngineerCard(h.engineerStore, engineerID, cycleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute engineer card"})
		return
	}
	if card == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *EngineerCardHandler) Trend(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}

	trend, err := h.scoreStore.EngineerTrend(h.engineerStore, engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute engineer trend"})
		return
	}
	c.JSON(http.StatusOK, trend)
}

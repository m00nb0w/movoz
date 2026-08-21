package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type CycleViewHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
	cycleStore    *store.CycleStore
}

func NewCycleViewHandler(s *store.ScoreStore, engineerStore *store.EngineerStore, cycleStore *store.CycleStore) *CycleViewHandler {
	return &CycleViewHandler{scoreStore: s, engineerStore: engineerStore, cycleStore: cycleStore}
}

func (h *CycleViewHandler) Get(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	scores, err := h.scoreStore.CycleScores(h.engineerStore, cycleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute cycle scores"})
		return
	}
	c.JSON(http.StatusOK, scores)
}

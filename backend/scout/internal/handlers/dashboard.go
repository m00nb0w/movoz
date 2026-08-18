package handlers

import (
	"net/http"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	scoreStore    *store.ScoreStore
	engineerStore *store.EngineerStore
}

func NewDashboardHandler(s *store.ScoreStore, engineerStore *store.EngineerStore) *DashboardHandler {
	return &DashboardHandler{scoreStore: s, engineerStore: engineerStore}
}

func (h *DashboardHandler) Get(c *gin.Context) {
	dashboard, err := h.scoreStore.RosterDashboard(h.engineerStore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute roster dashboard"})
		return
	}
	c.JSON(http.StatusOK, dashboard)
}

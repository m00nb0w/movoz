package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	store         *store.MetricStore
	engineerStore *store.EngineerStore
}

func NewMetricsHandler(s *store.MetricStore, engineerStore *store.EngineerStore) *MetricsHandler {
	return &MetricsHandler{store: s, engineerStore: engineerStore}
}

func (h *MetricsHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	exists, err := h.engineerStore.Exists(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}

	snapshots, err := h.store.ListByEngineer(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list metrics"})
		return
	}
	c.JSON(http.StatusOK, snapshots)
}

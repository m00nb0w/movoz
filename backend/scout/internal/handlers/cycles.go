package handlers

import (
	"net/http"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type CycleHandler struct {
	store *store.CycleStore
}

func NewCycleHandler(s *store.CycleStore) *CycleHandler {
	return &CycleHandler{store: s}
}

type cycleRequest struct {
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

func (h *CycleHandler) List(c *gin.Context) {
	cycles, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list cycles"})
		return
	}
	c.JSON(http.StatusOK, cycles)
}

func (h *CycleHandler) Create(c *gin.Context) {
	var req cycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_start and period_end are required"})
		return
	}
	start, err := time.Parse("2006-01-02", req.PeriodStart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_start must be YYYY-MM-DD"})
		return
	}
	end, err := time.Parse("2006-01-02", req.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be YYYY-MM-DD"})
		return
	}
	if !end.After(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be after period_start"})
		return
	}

	cycle, err := h.store.Create(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cycle"})
		return
	}
	c.JSON(http.StatusCreated, cycle)
}

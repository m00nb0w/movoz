package handlers

import (
	"net/http"
	"strconv"
	"time"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type SummaryHandler struct {
	store *store.SummaryStore
}

func NewSummaryHandler(s *store.SummaryStore) *SummaryHandler {
	return &SummaryHandler{store: s}
}

func (h *SummaryHandler) Get(c *gin.Context) {
	yearParam := c.DefaultQuery("year", strconv.Itoa(time.Now().UTC().Year()))
	year, err := strconv.Atoi(yearParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "year must be a 4-digit year"})
		return
	}

	summary, err := h.store.Summary(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load summary"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

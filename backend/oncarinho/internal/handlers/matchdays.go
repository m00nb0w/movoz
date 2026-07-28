package handlers

import (
	"net/http"
	"strconv"
	"time"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type MatchdayHandler struct {
	store *store.MatchdayStore
}

func NewMatchdayHandler(s *store.MatchdayStore) *MatchdayHandler {
	return &MatchdayHandler{store: s}
}

func (h *MatchdayHandler) List(c *gin.Context) {
	var year *int
	if yearParam := c.Query("year"); yearParam != "" {
		y, err := strconv.Atoi(yearParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
		year = &y
	}

	matchdays, err := h.store.List(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list matchdays"})
		return
	}
	c.JSON(http.StatusOK, matchdays)
}

type matchdayRequest struct {
	PlayedOn string `json:"played_on" binding:"required"`
}

func (h *MatchdayHandler) Create(c *gin.Context) {
	var req matchdayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "played_on is required"})
		return
	}

	playedOn, err := time.Parse("2006-01-02", req.PlayedOn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "played_on must be YYYY-MM-DD"})
		return
	}

	matchday, err := h.store.Create(playedOn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create matchday"})
		return
	}
	c.JSON(http.StatusCreated, matchday)
}

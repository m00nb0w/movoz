package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type HighlightHandler struct {
	store         *store.HighlightStore
	engineerStore *store.EngineerStore
}

func NewHighlightHandler(s *store.HighlightStore, engineerStore *store.EngineerStore) *HighlightHandler {
	return &HighlightHandler{store: s, engineerStore: engineerStore}
}

type highlightRequest struct {
	Kind string `json:"kind" binding:"required"`
	Body string `json:"body" binding:"required"`
}

func (h *HighlightHandler) List(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	exists, err := h.engineerStore.Exists(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	entries, err := h.store.List(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list highlights"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *HighlightHandler) Create(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req highlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind and body are required"})
		return
	}
	if req.Kind != "highlight" && req.Kind != "lowlight" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'highlight' or 'lowlight'"})
		return
	}
	exists, err := h.engineerStore.Exists(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}

	entry, err := h.store.Create(engineerID, req.Kind, req.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create highlight entry"})
		return
	}
	c.JSON(http.StatusCreated, entry)
}

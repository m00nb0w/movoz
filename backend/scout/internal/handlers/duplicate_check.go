package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"scout/internal/aiclient"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type DuplicateCheckHandler struct {
	aiClient       *aiclient.Client
	highlightStore *store.HighlightStore
	engineerStore  *store.EngineerStore
}

func NewDuplicateCheckHandler(aiClient *aiclient.Client, highlightStore *store.HighlightStore, engineerStore *store.EngineerStore) *DuplicateCheckHandler {
	return &DuplicateCheckHandler{aiClient: aiClient, highlightStore: highlightStore, engineerStore: engineerStore}
}

type duplicateCheckRequest struct {
	Body string `json:"body" binding:"required"`
}

// Check handles POST /api/engineers/:id/highlights/check-duplicate (F14).
// It runs synchronously before save and returns a similarity flag + matched
// entry, but never blocks: if the AI call fails or times out, this returns
// 200 with is_duplicate=false rather than an error, so the admin's save can
// proceed without a flag (NF3).
func (h *DuplicateCheckHandler) Check(c *gin.Context) {
	engineerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req duplicateCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
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

	existingEntries, err := h.highlightStore.List(engineerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load existing entries"})
		return
	}
	entries := make([]aiclient.ExistingEntry, 0, len(existingEntries))
	for _, e := range existingEntries {
		entries = append(entries, aiclient.ExistingEntry{ID: e.ID, Kind: e.Kind, Body: e.Body})
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	result, err := h.aiClient.CheckDuplicate(ctx, req.Body, entries)
	if err != nil {
		// Graceful degradation: proceed without a flag rather than blocking
		// the admin's save (F14, NF3).
		c.JSON(http.StatusOK, aiclient.DuplicateCheckResult{IsDuplicate: false, Note: "duplicate check unavailable, proceeding without a flag"})
		return
	}
	c.JSON(http.StatusOK, result)
}

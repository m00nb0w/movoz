package handlers

import (
	"net/http"
	"strconv"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	store *store.ProfileStore
}

func NewProfileHandler(s *store.ProfileStore) *ProfileHandler {
	return &ProfileHandler{store: s}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	profile, err := h.store.Profile(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load profile"})
		return
	}
	if profile == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

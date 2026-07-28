package handlers

import (
	"net/http"
	"strconv"

	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

type PlayerHandler struct {
	store *store.PlayerStore
}

func NewPlayerHandler(s *store.PlayerStore) *PlayerHandler {
	return &PlayerHandler{store: s}
}

type playerRequest struct {
	Name     string  `json:"name" binding:"required"`
	Position *string `json:"position"`
}

func (h *PlayerHandler) List(c *gin.Context) {
	players, err := h.store.List(true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list players"})
		return
	}
	c.JSON(http.StatusOK, players)
}

func (h *PlayerHandler) Create(c *gin.Context) {
	var req playerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	player, err := h.store.Create(req.Name, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create player"})
		return
	}
	c.JSON(http.StatusCreated, player)
}

func (h *PlayerHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	var req playerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	player, err := h.store.Update(id, req.Name, req.Position)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update player"})
		return
	}
	if player == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	c.JSON(http.StatusOK, player)
}

func (h *PlayerHandler) Deactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	ok, err := h.store.Deactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate player"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

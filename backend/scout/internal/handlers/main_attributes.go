package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type MainAttributeHandler struct {
	store *store.MainAttributeStore
}

func NewMainAttributeHandler(s *store.MainAttributeStore) *MainAttributeHandler {
	return &MainAttributeHandler{store: s}
}

type mainAttributeRequest struct {
	Key  string `json:"key" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func (h *MainAttributeHandler) List(c *gin.Context) {
	attrs, err := h.store.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list main attributes"})
		return
	}
	c.JSON(http.StatusOK, attrs)
}

func (h *MainAttributeHandler) Create(c *gin.Context) {
	var req mainAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key and name are required"})
		return
	}
	attr, err := h.store.Create(req.Key, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create main attribute"})
		return
	}
	c.JSON(http.StatusCreated, attr)
}

func (h *MainAttributeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid main attribute id"})
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	attr, err := h.store.Update(id, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update main attribute"})
		return
	}
	if attr == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "main attribute not found"})
		return
	}
	c.JSON(http.StatusOK, attr)
}

package handlers

import (
	"net/http"
	"strconv"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type SubAttributeHandler struct {
	store         *store.SubAttributeStore
	mainAttrStore *store.MainAttributeStore
}

func NewSubAttributeHandler(s *store.SubAttributeStore, mainAttrStore *store.MainAttributeStore) *SubAttributeHandler {
	return &SubAttributeHandler{store: s, mainAttrStore: mainAttrStore}
}

type subAttributeRequest struct {
	MainAttributeID int     `json:"main_attribute_id" binding:"required"`
	Name            string  `json:"name" binding:"required"`
	Description     *string `json:"description"`
}

func (h *SubAttributeHandler) List(c *gin.Context) {
	mainAttributeID, err := strconv.Atoi(c.Query("main_attribute_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "main_attribute_id query param is required"})
		return
	}
	activeOnly := c.DefaultQuery("active", "true") != "all"
	subs, err := h.store.ListByMainAttribute(mainAttributeID, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list sub attributes"})
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *SubAttributeHandler) Create(c *gin.Context) {
	var req subAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "main_attribute_id and name are required"})
		return
	}
	exists, err := h.mainAttrStore.Exists(req.MainAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up main attribute"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown main_attribute_id"})
		return
	}

	sub, err := h.store.Create(req.MainAttributeID, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create sub attribute"})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (h *SubAttributeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	sub, err := h.store.Update(id, req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update sub attribute"})
		return
	}
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *SubAttributeHandler) Deactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return
	}
	ok, err := h.store.Deactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate sub attribute"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

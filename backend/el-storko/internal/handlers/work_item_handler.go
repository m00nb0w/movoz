package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"el-storko/internal/models"
	"el-storko/internal/store"

	"github.com/gin-gonic/gin"
)

type WorkItemHandler struct {
	store *store.WorkItemStore
}

func NewWorkItemHandler(s *store.WorkItemStore) *WorkItemHandler {
	return &WorkItemHandler{store: s}
}

type workItemCreateRequest struct {
	Type        models.Type   `json:"type" binding:"required"`
	Title       string        `json:"title" binding:"required"`
	Description string        `json:"description"`
	Status      models.Status `json:"status"`
	ParentID    *int64        `json:"parent_id"`
}

type workItemUpdateRequest struct {
	Title       *string        `json:"title"`
	Description *string        `json:"description"`
	Status      *models.Status `json:"status"`
	ParentID    *int64         `json:"parent_id"`
}

func parseStoreErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, store.ErrEpicCannotHaveParent),
		errors.Is(err, store.ErrParentMustBeEpic),
		errors.Is(err, store.ErrParentNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
	return true
}

func (h *WorkItemHandler) Create(c *gin.Context) {
	var req workItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work item data"})
		return
	}
	if !req.Type.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
		return
	}
	if req.Status != "" && !req.Status.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	item, err := h.store.Create(models.WorkItem{
		Type:        req.Type,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		ParentID:    req.ParentID,
		Source:      models.SourcePersonal,
	})
	if parseStoreErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *WorkItemHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *WorkItemHandler) List(c *gin.Context) {
	filters := store.ListFilters{}
	if v := c.Query("source"); v != "" {
		s := models.Source(v)
		filters.Source = &s
	}
	if v := c.Query("type"); v != "" {
		t := models.Type(v)
		filters.Type = &t
	}
	if v := c.Query("status"); v != "" {
		s := models.Status(v)
		filters.Status = &s
	}
	if v := c.Query("parent_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
			return
		}
		filters.ParentID = &id
	}

	items, err := h.store.List(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *WorkItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req workItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work item data"})
		return
	}
	if req.Status != nil && !req.Status.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	item, err := h.store.Update(id, store.UpdateFields{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		ParentID:    req.ParentID,
	})
	if parseStoreErr(c, err) {
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *WorkItemHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ok, err := h.store.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

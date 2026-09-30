package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

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
	Type          models.Type   `json:"type" binding:"required"`
	Title         string        `json:"title" binding:"required"`
	Description   string        `json:"description"`
	Status        models.Status `json:"status"`
	ParentID      *int64        `json:"parent_id"`
	EstimateHours *float64      `json:"estimate_hours"`
	DueDate       *string       `json:"due_date"`
}

type workItemUpdateRequest struct {
	Title         *string        `json:"title"`
	Description   *string        `json:"description"`
	Status        *models.Status `json:"status"`
	ParentID      *int64         `json:"parent_id"`
	EstimateHours *float64       `json:"estimate_hours"`
	DueDate       *string        `json:"due_date"`
}

// parseDueDate parses the wire date-only format (models.DateOnlyLayout) into a
// *time.Time, returning a 400-worthy error on an invalid date string.
func parseDueDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	if *raw == "" {
		return nil, nil
	}
	t, err := time.Parse(models.DateOnlyLayout, *raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseStoreErr(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, store.ErrEpicCannotHaveParent),
		errors.Is(err, store.ErrParentMustBeEpic),
		errors.Is(err, store.ErrParentNotFound),
		errors.Is(err, store.ErrInvalidEstimate):
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
	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date"})
		return
	}

	item, err := h.store.Create(models.WorkItem{
		Type:          req.Type,
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		ParentID:      req.ParentID,
		EstimateHours: req.EstimateHours,
		DueDate:       dueDate,
		Source:        models.SourcePersonal,
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

// scopeSources translates the FR-022 Mine/Agent scope switch into the
// underlying source values: Mine covers personal+jira, Agent covers agent.
func scopeSources(scope string) ([]models.Source, bool) {
	switch scope {
	case "mine":
		return []models.Source{models.SourcePersonal, models.SourceJira}, true
	case "agent":
		return []models.Source{models.SourceAgent}, true
	default:
		return nil, false
	}
}

func (h *WorkItemHandler) List(c *gin.Context) {
	filters := store.ListFilters{}
	if v := c.Query("scope"); v != "" {
		sources, ok := scopeSources(v)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scope"})
			return
		}
		filters.Sources = sources
	}
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

// isExplicitNull reports whether key is present in the request body with a
// literal JSON null value — distinct from the key being absent entirely.
// Plain Go pointer fields can't tell these apart (both unmarshal to nil), but
// UpdateFields.ClearParent needs exactly this distinction: "omit parent_id"
// means "don't touch it", while "parent_id: null" means "unassign the Epic"
// (FR-024's Clear action).
func isExplicitNull(body []byte, key string) bool {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return false
	}
	value, present := raw[key]
	return present && string(value) == "null"
}

func (h *WorkItemHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work item data"})
		return
	}
	var req workItemUpdateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work item data"})
		return
	}
	if req.Status != nil && !req.Status.Valid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	dueDate, err := parseDueDate(req.DueDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid due_date"})
		return
	}

	item, err := h.store.Update(id, store.UpdateFields{
		Title:         req.Title,
		Description:   req.Description,
		Status:        req.Status,
		ParentID:      req.ParentID,
		ClearParent:   isExplicitNull(body, "parent_id"),
		EstimateHours: req.EstimateHours,
		DueDate:       dueDate,
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

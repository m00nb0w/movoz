package handlers

import (
	"net/http"
	"strconv"
	"time"

	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type EngineerHandler struct {
	store *store.EngineerStore
}

func NewEngineerHandler(s *store.EngineerStore) *EngineerHandler {
	return &EngineerHandler{store: s}
}

type engineerRequest struct {
	Name           string  `json:"name" binding:"required"`
	Role           *string `json:"role"`
	GitHubUsername *string `json:"github_username"`
	JiraAccountID  *string `json:"jira_account_id"`
	StartedAt      string  `json:"started_at" binding:"required"`
}

func (h *EngineerHandler) List(c *gin.Context) {
	activeOnly := c.DefaultQuery("active", "true") != "all"
	engineers, err := h.store.List(activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list engineers"})
		return
	}
	c.JSON(http.StatusOK, engineers)
}

// Get handles GET /api/engineers/:id — a single engineer, independent of
// any cycle. The frontend engineer card page (a later task) uses this for
// the page header before any cycle-scoped data is fetched.
func (h *EngineerHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	engineer, err := h.store.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up engineer"})
		return
	}
	if engineer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, engineer)
}

func (h *EngineerHandler) Create(c *gin.Context) {
	var req engineerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and started_at are required"})
		return
	}
	startedAt, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "started_at must be YYYY-MM-DD"})
		return
	}

	engineer, err := h.store.Create(req.Name, req.Role, req.GitHubUsername, req.JiraAccountID, startedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create engineer"})
		return
	}
	c.JSON(http.StatusCreated, engineer)
}

func (h *EngineerHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	var req engineerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and started_at are required"})
		return
	}
	startedAt, err := time.Parse("2006-01-02", req.StartedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "started_at must be YYYY-MM-DD"})
		return
	}

	engineer, err := h.store.Update(id, req.Name, req.Role, req.GitHubUsername, req.JiraAccountID, startedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update engineer"})
		return
	}
	if engineer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.JSON(http.StatusOK, engineer)
}

func (h *EngineerHandler) Deactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	ok, err := h.store.Deactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate engineer"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *EngineerHandler) Reactivate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid engineer id"})
		return
	}
	ok, err := h.store.Reactivate(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reactivate engineer"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "engineer not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

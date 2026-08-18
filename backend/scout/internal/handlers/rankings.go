package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type RankingHandler struct {
	store        *store.RankingStore
	cycleStore   *store.CycleStore
	subAttrStore *store.SubAttributeStore
}

func NewRankingHandler(s *store.RankingStore, cycleStore *store.CycleStore, subAttrStore *store.SubAttributeStore) *RankingHandler {
	return &RankingHandler{store: s, cycleStore: cycleStore, subAttrStore: subAttrStore}
}

type rankingEntryRequest struct {
	EngineerID int `json:"engineer_id" binding:"required"`
	Rank       int `json:"rank" binding:"required"`
}

type submitRankingRequest struct {
	Rankings []rankingEntryRequest `json:"rankings" binding:"required,dive"`
}

// parseCycleAndSubAttribute parses the :id/:subId path params and confirms
// both rows exist, writing an error response and returning ok=false if not.
func (h *RankingHandler) parseCycleAndSubAttribute(c *gin.Context) (cycleID, subAttributeID int, ok bool) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return 0, 0, false
	}
	subAttributeID, err = strconv.Atoi(c.Param("subId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sub attribute id"})
		return 0, 0, false
	}

	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return 0, 0, false
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return 0, 0, false
	}
	if exists, err := h.subAttrStore.Exists(subAttributeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up sub attribute"})
		return 0, 0, false
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "sub attribute not found"})
		return 0, 0, false
	}

	return cycleID, subAttributeID, true
}

// Submit handles PUT /api/cycles/:id/sub-attributes/:subId/ranking (F6, F7).
// Past cycles remain editable by re-submission — Scout does not lock a
// cycle's rankings once saved (see the plan's judgment-calls section);
// resubmitting fully replaces the prior ranking for this cycle+sub-attribute.
func (h *RankingHandler) Submit(c *gin.Context) {
	cycleID, subAttributeID, ok := h.parseCycleAndSubAttribute(c)
	if !ok {
		return
	}

	var req submitRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings are required"})
		return
	}

	entries := make([]scoring.RankEntry, 0, len(req.Rankings))
	for _, r := range req.Rankings {
		entries = append(entries, scoring.RankEntry{EngineerID: r.EngineerID, Rank: r.Rank})
	}

	rankings, err := h.store.SubmitRanking(cycleID, subAttributeID, entries)
	if err != nil {
		// A rejected permutation (bad request body) is a 400; anything else
		// from the store — failing to reach the DB, a broken transaction,
		// etc. — is a server-side failure and must not be reported as if the
		// caller's input were at fault. Mirrors the 400-vs-500 split used in
		// cycles.go/sub_attributes.go for pre-store-vs-store failures.
		if errors.Is(err, store.ErrInvalidRanking) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit ranking"})
		return
	}
	c.JSON(http.StatusOK, rankings)
}

// Get handles GET /api/cycles/:id/sub-attributes/:subId/ranking — reads
// back whatever has been saved so far for this cycle+sub-attribute (an
// empty array if nothing has been submitted yet). The ranking UI (Task 35)
// uses this to pre-populate rank inputs when reopening a cycle.
func (h *RankingHandler) Get(c *gin.Context) {
	cycleID, subAttributeID, ok := h.parseCycleAndSubAttribute(c)
	if !ok {
		return
	}

	rankings, err := h.store.GetByCycleAndSubAttribute(cycleID, subAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load ranking"})
		return
	}
	c.JSON(http.StatusOK, rankings)
}

package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type AIAcceptHandler struct {
	sessionStore *store.AISessionStore
	rankingStore *store.RankingStore
	cycleStore   *store.CycleStore
}

func NewAIAcceptHandler(sessionStore *store.AISessionStore, rankingStore *store.RankingStore, cycleStore *store.CycleStore) *AIAcceptHandler {
	return &AIAcceptHandler{sessionStore: sessionStore, rankingStore: rankingStore, cycleStore: cycleStore}
}

// Accept handles POST /api/cycles/:id/ai-sessions/:sessionId/accept (F9,
// NF3). The AI's proposed_ranking is never auto-applied to
// sub_attribute_rankings — this is the only code path that writes an AI
// session's ranking there, and it only runs on this explicit call. The
// request body (the admin's possibly-edited ranking, not
// session.ProposedRanking) is the sole source of truth for what gets
// persisted, and it is handed to RankingStore.SubmitRanking unmodified — the
// exact same strict 1..N permutation validation and score computation used
// by the manual ranking endpoint (F6/F7) — rather than being re-validated or
// re-scored here.
func (h *AIAcceptHandler) Accept(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	sessionID, err := strconv.Atoi(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	session, err := h.sessionStore.GetByID(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up ai ranking session"})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ai ranking session not found"})
		return
	}
	// A session that exists but belongs to a different cycle than the one in
	// the URL must be rejected the same way the chat endpoint (Task 26)
	// rejects it — otherwise an accept call could write a ranking into the
	// wrong cycle using this session's sub-attribute.
	if session.CycleID != cycleID {
		c.JSON(http.StatusNotFound, gin.H{"error": "ai ranking session does not belong to this cycle"})
		return
	}

	// The request body is the source of truth for what gets persisted, not
	// session.ProposedRanking (NF3) — the admin may have edited the AI's
	// proposal before confirming.
	var req submitRankingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rankings are required"})
		return
	}

	entries := make([]scoring.RankEntry, 0, len(req.Rankings))
	for _, r := range req.Rankings {
		entries = append(entries, scoring.RankEntry{EngineerID: r.EngineerID, Rank: r.Rank})
	}

	rankings, err := h.rankingStore.SubmitRanking(cycleID, session.SubAttributeID, entries)
	if err != nil {
		// Same 400-vs-500 split as RankingHandler.Submit: a rejected
		// permutation is the caller's fault (bad request body); anything
		// else from the store is a server-side failure.
		if errors.Is(err, store.ErrInvalidRanking) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept ranking"})
		return
	}
	c.JSON(http.StatusOK, rankings)
}

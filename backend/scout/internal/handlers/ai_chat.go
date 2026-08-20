package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"scout/internal/aiclient"
	"scout/internal/models"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

type AIChatHandler struct {
	aiClient       *aiclient.Client
	sessionStore   *store.AISessionStore
	engineerStore  *store.EngineerStore
	metricStore    *store.MetricStore
	highlightStore *store.HighlightStore
	subAttrStore   *store.SubAttributeStore
	cycleStore     *store.CycleStore
}

func NewAIChatHandler(aiClient *aiclient.Client, sessionStore *store.AISessionStore, engineerStore *store.EngineerStore, metricStore *store.MetricStore, highlightStore *store.HighlightStore, subAttrStore *store.SubAttributeStore, cycleStore *store.CycleStore) *AIChatHandler {
	return &AIChatHandler{
		aiClient:       aiClient,
		sessionStore:   sessionStore,
		engineerStore:  engineerStore,
		metricStore:    metricStore,
		highlightStore: highlightStore,
		subAttrStore:   subAttrStore,
		cycleStore:     cycleStore,
	}
}

type aiChatRequest struct {
	SessionID      *int   `json:"session_id"`
	SubAttributeID int    `json:"sub_attribute_id"`
	Message        string `json:"message" binding:"required"`
}

// Chat handles POST /api/cycles/:id/ai-sessions (F9). It streams the
// assistant's reply as Server-Sent Events: the response opens with a single
// `event: session` frame carrying {"session_id": N} (so the frontend can
// continue the same session on the next message), followed by plain
// `data: ...` text chunks as the reply streams in. The proposed ranking (if
// any) is parsed from the reply and stored on the session — it is never
// written to sub_attribute_rankings here; only the accept endpoint (Task 27)
// does that, and only on explicit admin confirm (NF3).
func (h *AIChatHandler) Chat(c *gin.Context) {
	cycleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
		return
	}
	if exists, err := h.cycleStore.Exists(cycleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up cycle"})
		return
	} else if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "cycle not found"})
		return
	}

	var req aiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}

	if req.SessionID == nil {
		if req.SubAttributeID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sub_attribute_id is required to start a new session"})
			return
		}
		if exists, err := h.subAttrStore.Exists(req.SubAttributeID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up sub-attribute"})
			return
		} else if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "sub-attribute not found"})
			return
		}
	}

	session, status, err := h.loadOrCreateSession(cycleID, req)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	if session.CycleID != cycleID {
		c.JSON(http.StatusNotFound, gin.H{"error": "ai ranking session does not belong to this cycle"})
		return
	}

	var history []aiclient.ChatMessage
	if err := json.Unmarshal(session.Transcript, &history); err != nil {
		history = nil
	}

	systemPrompt, err := h.buildSystemPrompt(session.SubAttributeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build chat context"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	fmt.Fprintf(c.Writer, "event: session\ndata: {\"session_id\":%d}\n\n", session.ID)
	if flusher, ok := c.Writer.(interface{ Flush() }); ok {
		flusher.Flush()
	}

	replyText, proposedRanking, err := h.aiClient.StreamRankingChat(c.Request.Context(), c.Writer, systemPrompt, history, req.Message)
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\":%q}\n\n", err.Error())
		return
	}

	newHistory := append(history,
		aiclient.ChatMessage{Role: "user", Content: req.Message},
		aiclient.ChatMessage{Role: "assistant", Content: replyText},
	)
	transcriptJSON, err := json.Marshal(newHistory)
	if err != nil {
		return
	}

	// A conversational turn that produced no new JSON-block proposal must
	// not wipe out a ranking proposed in an earlier turn: UpdateTranscript
	// unconditionally overwrites proposed_ranking with whatever is passed
	// in, so fall back to the session's existing value when this turn's
	// reply had none.
	rankingToPersist := proposedRanking
	if rankingToPersist == nil {
		rankingToPersist = session.ProposedRanking
	}
	if updErr := h.sessionStore.UpdateTranscript(session.ID, transcriptJSON, rankingToPersist); updErr != nil {
		// Headers and a partial reply are already flushed to the client at
		// this point, so a persistence failure can only be surfaced as an
		// SSE error event, not a JSON error response.
		fmt.Fprintf(c.Writer, "event: error\ndata: {\"error\":%q}\n\n", "failed to save chat session: "+updErr.Error())
	}
}

func (h *AIChatHandler) loadOrCreateSession(cycleID int, req aiChatRequest) (*models.AIRankingSession, int, error) {
	if req.SessionID != nil {
		session, err := h.sessionStore.GetByID(*req.SessionID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to look up ai ranking session")
		}
		if session == nil {
			return nil, http.StatusNotFound, fmt.Errorf("ai ranking session not found")
		}
		return session, http.StatusOK, nil
	}
	session, err := h.sessionStore.Create(cycleID, req.SubAttributeID)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to create ai ranking session")
	}
	return session, http.StatusOK, nil
}

// buildSystemPrompt assembles the active roster's synced metrics and
// existing highlights/lowlights as context for the sub-attribute being
// ranked, per the architecture note in the spec.
func (h *AIChatHandler) buildSystemPrompt(subAttributeID int) (string, error) {
	subAttr, err := h.subAttrStore.GetByID(subAttributeID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("You are Scout's ranking assistant. The manager will describe observations about their engineers in natural language for the current biweekly cycle. Your job is to propose a strict 1..N rank ordering (no ties) of the active engineers for one sub-attribute, with a short rationale. ")
	if subAttr != nil {
		fmt.Fprintf(&b, "The sub-attribute being ranked is %q. ", subAttr.Name)
	}
	b.WriteString("When you are ready to propose a ranking, end your reply with a fenced ```json code block containing exactly {\"rationale\": string, \"ranking\": [{\"engineer_id\": int, \"rank\": int}, ...]}. Only include that block once you have enough information — otherwise keep asking clarifying questions.\n\n")

	engineers, err := h.engineerStore.List(true)
	if err != nil {
		return "", err
	}
	b.WriteString("Active roster, with recent synced metrics and logged highlights/lowlights:\n")
	for _, e := range engineers {
		fmt.Fprintf(&b, "- Engineer %d: %s\n", e.ID, e.Name)
		if snapshots, err := h.metricStore.ListByEngineer(e.ID); err == nil && len(snapshots) > 0 {
			latest := snapshots[0]
			fmt.Fprintf(&b, "  Metrics (%s to %s): %d PRs raised, %d PRs reviewed, %d tickets closed, complexity %.1f\n",
				latest.PeriodStart.Format("2006-01-02"), latest.PeriodEnd.Format("2006-01-02"),
				latest.PRsRaised, latest.PRsReviewed, latest.TicketsClosed, latest.ComplexityScore)
		}
		if entries, err := h.highlightStore.List(e.ID); err == nil {
			for _, entry := range entries {
				label := "Highlight"
				if entry.Kind == "lowlight" {
					label = "Lowlight"
				}
				fmt.Fprintf(&b, "  %s (%s): %s\n", label, entry.CreatedAt.Format("2006-01-02"), entry.Body)
			}
		}
	}

	return b.String(), nil
}

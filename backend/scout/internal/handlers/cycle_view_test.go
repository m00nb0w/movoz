package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"scout/internal/models"
	"scout/internal/scoring"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCycleViewHandlerGet(t *testing.T) {
	db := setupTestDBForHandlers(t)
	for _, table := range []string{"sub_attribute_rankings", "sub_attributes", "main_attributes", "rating_cycles", "engineers"} {
		if _, err := db.Exec("TRUNCATE " + table + " RESTART IDENTITY CASCADE"); err != nil {
			t.Fatalf("failed to truncate %s: %v", table, err)
		}
	}

	engineerStore := store.NewEngineerStore(db)
	mainStore := store.NewMainAttributeStore(db)
	subStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)
	scoreStore := store.NewScoreStore(db)

	e1, _ := engineerStore.Create("Alex", nil, nil, nil, time.Now())
	main, _ := mainStore.Create("test_main_cv_handler", "Test Main CV Handler")
	sub, _ := subStore.Create(main.ID, "Ownership", nil)
	cycle, _ := cycleStore.Create(time.Now(), time.Now().AddDate(0, 0, 14))
	rankingStore.SubmitRanking(cycle.ID, sub.ID, []scoring.RankEntry{{EngineerID: e1.ID, Rank: 1}})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	r.GET("/api/cycles/:id/scores", h.Get)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/cycles/%d/scores", cycle.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var scores []models.EngineerCycleScore
	if err := json.Unmarshal(w.Body.Bytes(), &scores); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(scores) != 1 {
		t.Fatalf("expected 1 engineer in cycle view, got %d", len(scores))
	}
	if scores[0].Engineer.ID != e1.ID || scores[0].Engineer.Name != "Alex" {
		t.Fatalf("unexpected engineer in cycle view: %+v", scores[0].Engineer)
	}
	if scores[0].Overall == nil || *scores[0].Overall != 100 {
		t.Fatalf("expected overall score 100 for e1, got %v", scores[0].Overall)
	}
}

func TestCycleViewHandlerNotFound(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	cycleStore := store.NewCycleStore(db)
	scoreStore := store.NewScoreStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	r.GET("/api/cycles/:id/scores", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/cycles/999999/scores", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCycleViewHandlerInvalidCycleID(t *testing.T) {
	db := setupTestDBForHandlers(t)
	engineerStore := store.NewEngineerStore(db)
	cycleStore := store.NewCycleStore(db)
	scoreStore := store.NewScoreStore(db)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	r.GET("/api/cycles/:id/scores", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/cycles/not-a-number/scores", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid cycle ID, got %d", w.Code)
	}
}

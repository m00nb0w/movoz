package main

import (
	"database/sql"

	"scout/internal/aiclient"
	"scout/internal/config"
	"scout/internal/handlers"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	engineerStore := store.NewEngineerStore(db)
	mainAttributeStore := store.NewMainAttributeStore(db)
	subAttributeStore := store.NewSubAttributeStore(db)
	cycleStore := store.NewCycleStore(db)
	rankingStore := store.NewRankingStore(db, engineerStore)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg.AdminPassword, cfg.SessionSecret, cfg.CookieSecure)
	engineerHandler := handlers.NewEngineerHandler(engineerStore)
	mainAttributeHandler := handlers.NewMainAttributeHandler(mainAttributeStore)
	subAttributeHandler := handlers.NewSubAttributeHandler(subAttributeStore, mainAttributeStore)
	cycleHandler := handlers.NewCycleHandler(cycleStore)
	rankingHandler := handlers.NewRankingHandler(rankingStore, cycleStore, subAttributeStore)
	scoreStore := store.NewScoreStore(db)
	engineerCardHandler := handlers.NewEngineerCardHandler(scoreStore, engineerStore)
	cycleViewHandler := handlers.NewCycleViewHandler(scoreStore, engineerStore, cycleStore)
	dashboardHandler := handlers.NewDashboardHandler(scoreStore, engineerStore)
	metricStore := store.NewMetricStore(db)
	metricsHandler := handlers.NewMetricsHandler(metricStore, engineerStore)
	highlightStore := store.NewHighlightStore(db)
	highlightHandler := handlers.NewHighlightHandler(highlightStore, engineerStore)

	aiClient := aiclient.NewClient(cfg.AnthropicAPIKey)
	aiSessionStore := store.NewAISessionStore(db)
	aiChatHandler := handlers.NewAIChatHandler(aiClient, aiSessionStore, engineerStore, metricStore, highlightStore, subAttributeStore, cycleStore)
	aiAcceptHandler := handlers.NewAIAcceptHandler(aiSessionStore, rankingStore, cycleStore)

	r := gin.Default()

	// Exempt from auth: infra liveness probe and the login endpoint itself
	// (which by definition cannot require a session). Every other route
	// below sits behind RequireAuth — there is no public application-data
	// group, per NF1.
	r.GET("/health", healthHandler.HealthCheck)
	r.POST("/api/auth/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(handlers.RequireAuth(cfg.SessionSecret))
	{
		api.GET("/engineers", engineerHandler.List)
		api.GET("/engineers/:id", engineerHandler.Get)
		api.POST("/engineers", engineerHandler.Create)
		api.PUT("/engineers/:id", engineerHandler.Update)
		api.DELETE("/engineers/:id", engineerHandler.Deactivate)
		api.POST("/engineers/:id/reactivate", engineerHandler.Reactivate)
		api.GET("/engineers/:id/highlights", highlightHandler.List)
		api.POST("/engineers/:id/highlights", highlightHandler.Create)

		api.GET("/main-attributes", mainAttributeHandler.List)
		api.POST("/main-attributes", mainAttributeHandler.Create)
		api.PUT("/main-attributes/:id", mainAttributeHandler.Update)

		api.GET("/sub-attributes", subAttributeHandler.List)
		api.POST("/sub-attributes", subAttributeHandler.Create)
		api.PUT("/sub-attributes/:id", subAttributeHandler.Update)
		api.DELETE("/sub-attributes/:id", subAttributeHandler.Deactivate)

		api.GET("/cycles", cycleHandler.List)
		api.POST("/cycles", cycleHandler.Create)
		api.GET("/cycles/:id/scores", cycleViewHandler.Get)
		api.GET("/dashboard", dashboardHandler.Get)

		api.PUT("/cycles/:id/sub-attributes/:subId/ranking", rankingHandler.Submit)
		api.GET("/cycles/:id/sub-attributes/:subId/ranking", rankingHandler.Get)

		api.GET("/engineers/:id/card", engineerCardHandler.Card)
		api.GET("/engineers/:id/trend", engineerCardHandler.Trend)
		api.GET("/engineers/:id/metrics", metricsHandler.Get)

		api.POST("/cycles/:id/ai-sessions", aiChatHandler.Chat)
		api.POST("/cycles/:id/ai-sessions/:sessionId/accept", aiAcceptHandler.Accept)
	}

	return r
}

package main

import (
	"database/sql"

	"oncarinho/internal/config"
	"oncarinho/internal/handlers"
	"oncarinho/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	playerStore := store.NewPlayerStore(db)
	matchdayStore := store.NewMatchdayStore(db)
	statStore := store.NewStatStore(db)
	leaderboardStore := store.NewLeaderboardStore(db)
	profileStore := store.NewProfileStore(db, playerStore)
	summaryStore := store.NewSummaryStore(db)

	healthHandler := handlers.NewHealthHandler()
	playerHandler := handlers.NewPlayerHandler(playerStore)
	matchdayHandler := handlers.NewMatchdayHandler(matchdayStore)
	statHandler := handlers.NewStatHandler(statStore, matchdayStore, playerStore)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardStore)
	profileHandler := handlers.NewProfileHandler(profileStore)
	summaryHandler := handlers.NewSummaryHandler(summaryStore)
	authHandler := handlers.NewAuthHandler(cfg.AdminPassword, cfg.SessionSecret, cfg.CookieSecure)

	r := gin.Default()

	r.GET("/health", healthHandler.HealthCheck)
	r.POST("/api/auth/login", authHandler.Login)

	r.GET("/api/players", playerHandler.List)
	r.GET("/api/players/:id", profileHandler.Get)
	r.GET("/api/matchdays", matchdayHandler.List)
	r.GET("/api/leaderboard", leaderboardHandler.Get)
	r.GET("/api/summary", summaryHandler.Get)
	r.GET("/api/matchdays/:id/stats", statHandler.GetStats)

	admin := r.Group("/api")
	admin.Use(handlers.RequireAdmin(cfg.SessionSecret))
	{
		admin.POST("/players", playerHandler.Create)
		admin.PUT("/players/:id", playerHandler.Update)
		admin.DELETE("/players/:id", playerHandler.Deactivate)
		admin.POST("/players/:id/reactivate", playerHandler.Reactivate)
		admin.POST("/matchdays", matchdayHandler.Create)
		admin.PUT("/matchdays/:id/stats", statHandler.UpsertStats)
		admin.DELETE("/matchdays/:id/stats/:playerId", statHandler.DeleteStat)
	}

	return r
}

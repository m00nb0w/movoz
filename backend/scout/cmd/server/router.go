package main

import (
	"database/sql"

	"scout/internal/config"
	"scout/internal/handlers"
	"scout/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	engineerStore := store.NewEngineerStore(db)
	mainAttributeStore := store.NewMainAttributeStore(db)
	subAttributeStore := store.NewSubAttributeStore(db)

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(cfg.AdminPassword, cfg.SessionSecret, cfg.CookieSecure)
	engineerHandler := handlers.NewEngineerHandler(engineerStore)
	mainAttributeHandler := handlers.NewMainAttributeHandler(mainAttributeStore)
	subAttributeHandler := handlers.NewSubAttributeHandler(subAttributeStore, mainAttributeStore)

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

		api.GET("/main-attributes", mainAttributeHandler.List)
		api.POST("/main-attributes", mainAttributeHandler.Create)
		api.PUT("/main-attributes/:id", mainAttributeHandler.Update)

		api.GET("/sub-attributes", subAttributeHandler.List)
		api.POST("/sub-attributes", subAttributeHandler.Create)
		api.PUT("/sub-attributes/:id", subAttributeHandler.Update)
		api.DELETE("/sub-attributes/:id", subAttributeHandler.Deactivate)
	}

	return r
}

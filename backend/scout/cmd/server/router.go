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

	healthHandler := handlers.NewHealthHandler()
	engineerHandler := handlers.NewEngineerHandler(engineerStore)

	r := gin.Default()
	r.GET("/health", healthHandler.HealthCheck)

	api := r.Group("/api")
	{
		api.GET("/engineers", engineerHandler.List)
		api.GET("/engineers/:id", engineerHandler.Get)
		api.POST("/engineers", engineerHandler.Create)
		api.PUT("/engineers/:id", engineerHandler.Update)
		api.DELETE("/engineers/:id", engineerHandler.Deactivate)
		api.POST("/engineers/:id/reactivate", engineerHandler.Reactivate)
	}

	return r
}

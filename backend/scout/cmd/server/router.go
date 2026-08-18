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
	engineerHandler := handlers.NewEngineerHandler(engineerStore)
	mainAttributeHandler := handlers.NewMainAttributeHandler(mainAttributeStore)
	subAttributeHandler := handlers.NewSubAttributeHandler(subAttributeStore, mainAttributeStore)

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

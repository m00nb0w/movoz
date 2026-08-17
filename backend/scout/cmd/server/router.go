package main

import (
	"database/sql"

	"scout/internal/config"
	"scout/internal/handlers"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	healthHandler := handlers.NewHealthHandler()

	r := gin.Default()
	r.GET("/health", healthHandler.HealthCheck)

	return r
}

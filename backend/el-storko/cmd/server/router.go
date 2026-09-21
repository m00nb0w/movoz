package main

import (
	"database/sql"

	"el-storko/internal/handlers"
	"el-storko/internal/store"

	"github.com/gin-gonic/gin"
)

func buildRouter(db *sql.DB) *gin.Engine {
	workItemStore := store.NewWorkItemStore(db)

	healthHandler := handlers.NewHealthHandler()
	workItemHandler := handlers.NewWorkItemHandler(workItemStore)

	r := gin.Default()

	r.GET("/health", healthHandler.HealthCheck)

	r.GET("/api/work-items", workItemHandler.List)
	r.POST("/api/work-items", workItemHandler.Create)
	r.GET("/api/work-items/:id", workItemHandler.Get)
	r.PATCH("/api/work-items/:id", workItemHandler.Update)
	r.DELETE("/api/work-items/:id", workItemHandler.Delete)

	return r
}

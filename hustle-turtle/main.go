package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "hustle-turtle",
		})
	})

	// Start server on port 8080
	r.Run(":8080")
}

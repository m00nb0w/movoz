package main

import (
	"flag"
	"fmt"
	"log"

	"hustle-turtle/internal/config"
	"hustle-turtle/internal/database"
	"hustle-turtle/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Define command line flags
	var (
		migrate     = flag.String("migrate", "", "Run database migrations: 'up', 'down'")
		version     = flag.Bool("version", false, "Show current migration version")
		autoMigrate = flag.Bool("auto-migrate", false, "Run up migrations on startup")
	)
	flag.Parse()

	// Create migration manager
	migrationManager := database.NewMigrationManager(cfg.DatabaseURL)

	// Handle migration commands
	if *version {
		v, dirty, err := migrationManager.Version()
		if err != nil {
			log.Fatalf("Could not get migration version: %v", err)
		}

		status := "clean"
		if dirty {
			status = "dirty"
		}

		fmt.Printf("Current migration version: %d (status: %s)\n", v, status)
		return
	}

	if *migrate != "" {
		switch *migrate {
		case "up":
			if err := migrationManager.Up(); err != nil {
				log.Fatalf("Migration up failed: %v", err)
			}
		case "down":
			if err := migrationManager.Down(); err != nil {
				log.Fatalf("Migration down failed: %v", err)
			}
		default:
			log.Fatalf("Invalid migration direction: %s (use 'up' or 'down')", *migrate)
		}
		return
	}

	// Run auto migrations if flag is set
	if *autoMigrate {
		if err := migrationManager.Up(); err != nil {
			log.Printf("Auto migration failed: %v", err)
		}
	}

	// Create Gin router
	r := gin.Default()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()

	// Setup routes
	r.GET("/health", healthHandler.HealthCheck)

	// Start server
	port := ":" + cfg.Port
	log.Printf("Starting server on port %s", cfg.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

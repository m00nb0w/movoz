package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func getMigrationInstance() (*migrate.Migrate, error) {
	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost/hustle_turtle?sslmode=disable"
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}
	defer db.Close()

	// Create migrate driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create migrate driver: %v", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %v", err)
	}

	return m, nil
}

func runMigrations(direction string) {
	m, err := getMigrationInstance()
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Could not run up migrations: %v", err)
		}
		log.Println("Up migrations completed successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Could not run down migrations: %v", err)
		}
		log.Println("Down migrations completed successfully")
	default:
		log.Fatalf("Invalid migration direction: %s (use 'up' or 'down')", direction)
	}
}

func getMigrationVersion() {
	m, err := getMigrationInstance()
	if err != nil {
		log.Fatalf("Migration setup failed: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		log.Fatalf("Could not get migration version: %v", err)
	}

	status := "clean"
	if dirty {
		status = "dirty"
	}

	fmt.Printf("Current migration version: %d (status: %s)\n", version, status)
}

func main() {
	// Define command line flags
	var (
		migrate    = flag.String("migrate", "", "Run database migrations: 'up', 'down'")
		version    = flag.Bool("version", false, "Show current migration version")
		autoMigrate = flag.Bool("auto-migrate", false, "Run up migrations on startup")
	)
	flag.Parse()

	// Handle migration commands
	if *version {
		getMigrationVersion()
		return
	}

	if *migrate != "" {
		runMigrations(*migrate)
		return
	}

	// Run auto migrations if flag is set
	if *autoMigrate {
		runMigrations("up")
	}

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
	log.Println("Starting server on :8080")
	r.Run(":8080")
}

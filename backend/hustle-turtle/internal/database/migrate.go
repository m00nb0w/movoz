package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	databaseURL string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(databaseURL string) *MigrationManager {
	return &MigrationManager{
		databaseURL: databaseURL,
	}
}

// getMigrationInstance creates a migrate instance
func (mm *MigrationManager) getMigrationInstance() (*migrate.Migrate, error) {
	// Open database connection
	db, err := sql.Open("postgres", mm.databaseURL)
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

// Up runs up migrations
func (mm *MigrationManager) Up() error {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %v", err)
	}

	log.Println("Up migrations completed successfully")
	return nil
}

// Down runs down migrations
func (mm *MigrationManager) Down() error {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run down migrations: %v", err)
	}

	log.Println("Down migrations completed successfully")
	return nil
}

// Version gets the current migration version
func (mm *MigrationManager) Version() (uint, bool, error) {
	m, err := mm.getMigrationInstance()
	if err != nil {
		return 0, false, fmt.Errorf("migration setup failed: %v", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("could not get migration version: %v", err)
	}

	return version, dirty, nil
}

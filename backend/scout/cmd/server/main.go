package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"scout/internal/config"
	"scout/internal/database"
	"scout/internal/integrations"
	"scout/internal/store"
	"scout/internal/syncer"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	var (
		migrateDir  = flag.String("migrate", "", "Run database migrations: 'up', 'down'")
		version     = flag.Bool("version", false, "Show current migration version")
		autoMigrate = flag.Bool("auto-migrate", false, "Run up migrations on startup")
	)
	flag.Parse()

	migrationManager := database.NewMigrationManager(cfg.DatabaseURL)

	if *version {
		v, dirty, err := migrationManager.Version()
		if err != nil {
			log.Fatalf("could not get migration version: %v", err)
		}
		status := "clean"
		if dirty {
			status = "dirty"
		}
		fmt.Printf("Current migration version: %d (status: %s)\n", v, status)
		return
	}

	if *migrateDir != "" {
		switch *migrateDir {
		case "up":
			if err := migrationManager.Up(); err != nil {
				log.Fatalf("migration up failed: %v", err)
			}
		case "down":
			if err := migrationManager.Down(); err != nil {
				log.Fatalf("migration down failed: %v", err)
			}
		default:
			log.Fatalf("invalid migration direction: %s", *migrateDir)
		}
		return
	}

	if *autoMigrate {
		if err := migrationManager.Up(); err != nil {
			log.Printf("auto migration failed: %v", err)
		}
	}

	if cfg.AdminPassword == "" || cfg.SessionSecret == "" {
		log.Fatal("ADMIN_PASSWORD and SESSION_SECRET must be set")
	}

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	githubClient := integrations.NewGitHubClient(cfg.GitHubToken, nil)
	jiraClient := integrations.NewJiraClient(cfg.JiraBaseURL, cfg.JiraEmail, cfg.JiraAPIToken, nil)
	engineerStoreForSync := store.NewEngineerStore(db)
	metricStoreForSync := store.NewMetricStore(db)
	cycleStoreForSync := store.NewCycleStore(db)
	syncWorker := syncer.NewSyncer(engineerStoreForSync, metricStoreForSync, githubClient, jiraClient, cfg.GitHubRepos, cfg.JiraProjects)

	syncCtx, cancelSync := context.WithCancel(context.Background())
	defer cancelSync()
	// The sync window is the current rating cycle's period, so the scheduler
	// needs to read rating_cycles on every run (see syncer.RunSyncCycle).
	syncer.StartScheduler(syncCtx, syncWorker, cycleStoreForSync, cfg.SyncInterval)

	r := buildRouter(db, cfg)

	port := ":" + cfg.Port
	log.Printf("starting server on port %s", cfg.Port)
	if err := r.Run(port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}

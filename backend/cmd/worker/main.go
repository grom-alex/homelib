package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/repository"
	"github.com/grom-alex/homelib/backend/internal/service"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	runImport := flag.Bool("import", false, "run INPX import and exit")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := repository.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := repository.RunMigrations(cfg.Database.DSN()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations applied successfully")

	if *runImport {
		bookRepo := repository.NewBookRepo(pool)
		authorRepo := repository.NewAuthorRepo(pool)
		genreRepo := repository.NewGenreRepo(pool)
		seriesRepo := repository.NewSeriesRepo(pool)
		collectionRepo := repository.NewCollectionRepo(pool)

		importSvc := service.NewImportService(pool, cfg.Import, cfg.Library, bookRepo, authorRepo, genreRepo, seriesRepo, collectionRepo)

		if err := importSvc.StartImport(ctx); err != nil {
			log.Fatalf("Failed to start import: %v", err)
		}

		// Wait for import to complete
		for {
			status := importSvc.GetStatus()
			if status.Status == "completed" || status.Status == "failed" {
				if status.Error != nil {
					log.Fatalf("Import failed: %s", *status.Error)
				}
				log.Printf("Import completed: %+v", status.Stats)
				return
			}
			select {
			case <-ctx.Done():
				log.Println("Import cancelled")
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}

	log.Println("Worker started. No tasks specified. Use --import to run INPX import.")
	<-ctx.Done()
	log.Println("Worker shutting down")
}

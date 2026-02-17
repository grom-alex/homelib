package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/grom-alex/homelib/backend/internal/api"
	"github.com/grom-alex/homelib/backend/internal/config"
	"github.com/grom-alex/homelib/backend/internal/repository"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
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

	server := api.NewServer(cfg, pool)
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

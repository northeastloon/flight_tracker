package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/northeastloon/flight_tracker/internal/domain"
	storage "github.com/northeastloon/flight_tracker/internal/postgres"
	"github.com/northeastloon/flight_tracker/internal/provider"
	"github.com/northeastloon/flight_tracker/internal/server"
)

func Run() error {

	// load environmental vars
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// initialise db
	db, err := storage.NewDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}

	if err := db.MigrateDB(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	//initialise fetcher(s)
	OpenSkyClient := provider.NewOpenSkyClient(provider.WithQueryParam("extended", "true"))

	//start ingest service
	ctx := context.Background()
	fds := domain.NewFlightDataService(OpenSkyClient, db)

	go fds.StartIngestionLoop(ctx, 10)

	//initialise server
	server, err := server.NewServer(db)
	if err != nil {
		return fmt.Errorf("failed to initialise server: %w", err)
	}

	if err := server.Echo.Start(":8080"); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil

}

func main() {

	if err := Run(); err != nil {
		slog.Error("run error", slog.Any("err", err))
		log.Panic()
	}

}

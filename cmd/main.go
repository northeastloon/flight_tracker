package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/northeastloon/flight_tracker/internal/fetch"
	storage "github.com/northeastloon/flight_tracker/internal/postgres"
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

	//fetcher
	OpenSkyClient := fetch.NewOpenSkyClient(fetch.WithQueryParam("extended", "true"))

	states, err := OpenSkyClient.FetchTelemetry(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch open sky states: %w", err)
	}

	fmt.Println(states)

	if err := db.InsertAircraftState(states); err != nil {
		return fmt.Errorf("failed to insert aircraft state: %w", err)
	}

	return nil

}

func main() {

	if err := Run(); err != nil {
		slog.Error("run error", slog.Any("err", err))
		log.Panic()
	}

}

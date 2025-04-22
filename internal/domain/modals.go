package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Telemetry struct {
	ICAO24         string
	Callsign       *string
	OriginCountry  string
	TimePosition   *time.Time
	LastContact    time.Time
	Longitude      *float64
	Latitude       *float64
	BaroAltitude   *float64
	OnGround       bool
	Velocity       *float64
	TrueTrack      *float64
	VerticalRate   *float64
	Sensors        *[]int
	GeoAltitude    *float64
	Squawk         *string
	SPI            bool
	PositionSource int
	Category       int
}

type TelemetryFilter struct {
	ICAO24        *string
	Callsign      *string
	OriginCountry *string
	TimePosition  *time.Time
	LastContact   *time.Time
	Squawk        *string
	Category      *int
	Position      *struct {
		Latitude  float64
		Longitude float64
		Radius    float64 // in kilometers
	}
}

type FlightDataProvider[T any] interface {
	FetchTelemetry(ctx context.Context) (T, error)
}

type FlightDataStore[T any] interface {
	StoreTelemetry(ctx context.Context, data T) error
	GetTelemetry(ctx context.Context, filter *TelemetryFilter) ([]Telemetry, error)
}

type FlightDataService[T any] struct {
	provider FlightDataProvider[T]
	store    FlightDataStore[T]
}

func NewFlightDataService[T any](provider FlightDataProvider[T], store FlightDataStore[T]) *FlightDataService[T] {
	return &FlightDataService[T]{
		provider: provider,
		store:    store,
	}
}

func (s *FlightDataService[T]) IngestData(ctx context.Context) error {
	data, err := s.provider.FetchTelemetry(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch telemetry: %w", err)
	}

	if err := s.store.StoreTelemetry(ctx, data); err != nil {
		return fmt.Errorf("failed to store telemetry: %w", err)
	}

	return nil
}

func (s *FlightDataService[T]) StartIngestionLoop(ctx context.Context, runsPerDay int) error {
	if runsPerDay <= 0 {
		return fmt.Errorf("runsPerDay must be a positive integer")
	}

	interval := time.Duration(1440/runsPerDay) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run the first ingestion immediately
	if err := s.IngestData(ctx); err != nil {
		// Log the error but continue the loop
		slog.Error("Error during initial data ingestion", "error", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := s.IngestData(ctx); err != nil {
				// Log the error but continue the loop
				slog.Error("Error during data ingestion", "error", err)
			}
		case <-ctx.Done():
			fmt.Println("Stopping ingestion loop due to context cancellation.")
			return ctx.Err()
		}
	}
}

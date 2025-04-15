package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/northeastloon/flight_tracker/internal/fetch"
)

type Database struct {
	Client *pgxpool.Pool
}

func NewDatabase() (*Database, error) {

	connectionString := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("POSTGRES_ADMIN_USER"),
		os.Getenv("POSTGRES_ADMIN_PASSWORD"),
		os.Getenv("SSL_MODE"),
	)

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Client: pool}, nil
}

func (d *Database) MigrateDB() error {
	schema := `
		CREATE TABLE IF NOT EXISTS aircraft_state (
			icao24            TEXT NOT NULL,
			callsign          TEXT,
			origin_country    TEXT,
			time_position     TIMESTAMP,
			last_contact      TIMESTAMP NOT NULL,
			longitude         DOUBLE PRECISION,
			latitude          DOUBLE PRECISION,
			baro_altitude     DOUBLE PRECISION,
			on_ground         BOOLEAN,
			velocity          DOUBLE PRECISION,
			true_track        DOUBLE PRECISION,
			vertical_rate     DOUBLE PRECISION,
			sensors           INTEGER[],
			geo_altitude      DOUBLE PRECISION,
			squawk            TEXT,
			spi               BOOLEAN,
			position_source   INTEGER,
			category          INTEGER
		);

		CREATE TABLE IF NOT EXISTS aircraft_category (
			id INTEGER PRIMARY KEY,
			category  TEXT
		);

		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM aircraft_category) THEN
				INSERT INTO aircraft_category (id, category) VALUES
					(0, 'No information'),
					(1, 'No ADS-B Emitter Category Information'),
					(2, 'Light (< 15500 lbs)'),
					(3, 'Small (15500 to 75000 lbs)'),
					(4, 'Large (75000 to 300000 lbs)'),
					(5, 'High Vortex Large'),
					(6, 'Heavy (> 300000 lbs)'),
					(7, 'High Performance'),
					(8, 'Rotorcraft'),
					(9, 'Glider / sailplane'),
					(10, 'Lighter-than-air'),
					(11, 'Parachutist / Skydiver'),
					(12, 'Ultralight / hang-glider / paraglider'),
					(13, 'Reserved'),
					(14, 'Unmanned Aerial Vehicle'),
					(15, 'Space / Trans-atmospheric vehicle'),
					(16, 'Surface Vehicle – Emergency Vehicle'),
					(17, 'Surface Vehicle – Service Vehicle'),
					(18, 'Point Obstacle'),
					(19, 'Cluster Obstacle'),
					(20, 'Line Obstacle');
			END IF;
		END;
		$$;

		CREATE INDEX ON aircraft_state (icao24, last_contact DESC);
		CREATE INDEX ON aircraft_state (latitude, longitude);
		CREATE INDEX ON aircraft_state (callsign);
	`

	_, err := d.Client.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

func (d *Database) InsertAircraftState(states []fetch.OpenSkyTelemetry) error {

	tx, err := d.Client.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(context.Background())

	for _, state := range states {
		// Convert Unix timestamps to time.Time
		var timePosition, lastContact *time.Time

		if state.TimePosition != nil {
			t := time.Unix(*state.TimePosition, 0)
			timePosition = &t
		}

		t := time.Unix(state.LastContact, 0)
		lastContact = &t

		_, err := tx.Exec(context.Background(), `
			INSERT INTO aircraft_state (
				icao24, callsign, origin_country, time_position, last_contact,
				longitude, latitude, baro_altitude, on_ground, velocity,
				true_track, vertical_rate, sensors, geo_altitude, squawk,
				spi, position_source, category
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		`,
			state.Icao24, state.Callsign, state.OriginCountry, timePosition, lastContact,
			state.Longitude, state.Latitude, state.BaroAltitude, state.OnGround, state.Velocity,
			state.TrueTrack, state.VerticalRate, state.Sensors, state.GeoAltitude, state.Squawk,
			state.SPI, state.PositionSource, state.Category,
		)

		if err != nil {
			return fmt.Errorf("failed to insert aircraft state: %w", err)
		}
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

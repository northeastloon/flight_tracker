package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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
	// First, enable the PostGIS extension
	enablePostgis := `
	CREATE EXTENSION IF NOT EXISTS postgis;
	`
	_, err := d.Client.Exec(context.Background(), enablePostgis)
	if err != nil {
		return fmt.Errorf("failed to enable postgis extension: %w", err)
	}

	cleanupTrigger := `
	-- First create the trigger function
	CREATE OR REPLACE FUNCTION cleanup_old_observations()
	RETURNS TRIGGER AS $$
	BEGIN
		-- Remove aircraft that have landed (on_ground changed from false to true)
		DELETE FROM opensky 
		WHERE icao24 = NEW.icao24 
		AND on_ground = false 
		AND NEW.on_ground = true;

		-- Remove observations older than 24 hours
		DELETE FROM opensky 
		WHERE last_contact < NOW() - INTERVAL '24 hours';

		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	-- Then create the trigger
	DROP TRIGGER IF EXISTS trigger_cleanup_observations ON opensky;
	CREATE TRIGGER trigger_cleanup_observations
		BEFORE INSERT ON opensky
		FOR EACH ROW
		EXECUTE FUNCTION cleanup_old_observations();
	`

	schema := `
	CREATE TABLE IF NOT EXISTS opensky (
    	icao24 TEXT NOT NULL,
    	callsign TEXT,
    	origin_country TEXT,
    	time_position TIMESTAMP,
    	last_contact TIMESTAMP NOT NULL,
    	longitude DOUBLE PRECISION,
    	latitude DOUBLE PRECISION,
    	baro_altitude DOUBLE PRECISION,
    	on_ground BOOLEAN,
    	velocity DOUBLE PRECISION,
    	true_track DOUBLE PRECISION,
    	vertical_rate DOUBLE PRECISION,
    	sensors INTEGER[],
    	geo_altitude DOUBLE PRECISION,
    	squawk TEXT,
    	spi BOOLEAN,
    	position_source INTEGER,
    	category INTEGER
	);

	CREATE TABLE IF NOT EXISTS opensky_category (
    	id INTEGER PRIMARY KEY,
    	category TEXT
	);

	INSERT INTO opensky_category (id, category) VALUES
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
    	(20, 'Line Obstacle')
	ON CONFLICT (id) DO NOTHING;  

	DROP VIEW IF EXISTS aircraft_state;
	CREATE VIEW aircraft_state AS
		SELECT 
   		 	icao24, 
   			callsign, 
   			origin_country, 
   			time_position, 
   			last_contact,
   			longitude, 
   			latitude, 
   			ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)::geography AS position,  
    		baro_altitude, 
			on_ground, 
			velocity,
			true_track, 
			vertical_rate, 
			sensors, 
			geo_altitude, 
			squawk,
			spi, 
			position_source,
			category
		FROM opensky;

	-- Optional indices
	CREATE INDEX IF NOT EXISTS idx_opensky_icao_last_contact ON opensky (icao24, last_contact DESC);
	CREATE INDEX IF NOT EXISTS idx_opensky_lat_lon ON opensky (latitude, longitude);
	CREATE INDEX IF NOT EXISTS idx_opensky_callsign ON opensky (callsign);

	-- Add the cleanup trigger function and trigger
	` + cleanupTrigger

	_, err = d.Client.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

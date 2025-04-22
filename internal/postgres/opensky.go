package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/northeastloon/flight_tracker/internal/domain"
	"github.com/northeastloon/flight_tracker/internal/provider"
)

var _ domain.FlightDataStore[[]provider.OpenSkyTelemetry] = (*Database)(nil)

func (d *Database) StoreTelemetry(ctx context.Context, data []provider.OpenSkyTelemetry) error {
	tx, err := d.Client.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, d := range data {
		var timePosition, lastContact *time.Time

		if d.TimePosition != nil {
			t := time.Unix(*d.TimePosition, 0)
			timePosition = &t
		}

		t := time.Unix(d.LastContact, 0)
		lastContact = &t

		_, err := tx.Exec(ctx, `
			INSERT INTO opensky (
				icao24, callsign, origin_country, time_position, last_contact,
				longitude, latitude, baro_altitude, on_ground, velocity,
				true_track, vertical_rate, sensors, geo_altitude, squawk,
				spi, position_source, category
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		`,
			d.Icao24, d.Callsign, d.OriginCountry, timePosition, lastContact,
			d.Longitude, d.Latitude, d.BaroAltitude, d.OnGround, d.Velocity,
			d.TrueTrack, d.VerticalRate, d.Sensors, d.GeoAltitude, d.Squawk,
			d.SPI, d.PositionSource, d.Category,
		)

		if err != nil {
			return fmt.Errorf("failed to insert opensky aircraft state: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

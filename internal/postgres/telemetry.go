package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/northeastloon/flight_tracker/internal/domain"
)

func buildTelemetryQuery(filter *domain.TelemetryFilter) (string, []any) {
	latest := filter != nil && filter.Latest != nil && *filter.Latest

	var query strings.Builder
	var params []any

	// ---------- choose SELECT ----------
	if latest {
		query.WriteString(`
            SELECT DISTINCT ON (icao24)
                icao24, callsign, origin_country, time_position, last_contact,
                longitude, latitude, baro_altitude, on_ground, velocity,
                true_track, vertical_rate, sensors, geo_altitude, squawk,
                spi, position_source, category
            FROM aircraft_state
            WHERE 1 = 1
        `)
	} else {
		query.WriteString(`
            SELECT
                icao24, callsign, origin_country, time_position, last_contact,
                longitude, latitude, baro_altitude, on_ground, velocity,
                true_track, vertical_rate, sensors, geo_altitude, squawk,
                spi, position_source, category
            FROM aircraft_state
            WHERE 1 = 1
        `)
	}

	if filter != nil {
		if filter.ICAO24 != nil {
			params = append(params, *filter.ICAO24)
			query.WriteString(fmt.Sprintf(" AND icao24 = $%d", len(params)))
		}
		if filter.Callsign != nil {
			params = append(params, *filter.Callsign)
			query.WriteString(fmt.Sprintf(" AND callsign = $%d", len(params)))
		}
		if filter.OriginCountry != nil {
			params = append(params, *filter.OriginCountry)
			query.WriteString(fmt.Sprintf(" AND origin_country = $%d", len(params)))
		}
		if filter.TimePosition != nil {
			params = append(params, *filter.TimePosition)
			query.WriteString(fmt.Sprintf(" AND time_position >= $%d", len(params)))
		}
		if filter.LastContact != nil {
			params = append(params, *filter.LastContact)
			query.WriteString(fmt.Sprintf(" AND last_contact >= $%d", len(params)))
		}
		if filter.Squawk != nil {
			params = append(params, *filter.Squawk)
			query.WriteString(fmt.Sprintf(" AND squawk = $%d", len(params)))
		}
		if filter.Category != nil {
			params = append(params, *filter.Category)
			query.WriteString(fmt.Sprintf(" AND category = $%d", len(params)))
		}
		if filter.Position != nil {
			params = append(params,
				filter.Position.Longitude,
				filter.Position.Latitude,
				filter.Position.Radius*1000,
			)
			query.WriteString(fmt.Sprintf(`
                AND ST_DWithin(
                    position,
                    ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography,
                    $%d
                )
            `, len(params)-2, len(params)-1, len(params)))
		}
	}

	if latest {
		query.WriteString(" ORDER BY icao24, last_contact DESC")
	} else {
		query.WriteString(" ORDER BY last_contact DESC")
	}

	return query.String(), params
}

func (d *Database) GetTelemetry(ctx context.Context, filter *domain.TelemetryFilter) ([]domain.Telemetry, error) {
	query, params := buildTelemetryQuery(filter)

	rows, err := d.Client.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry: %w", err)
	}
	defer rows.Close()

	var telemetry []domain.Telemetry
	for rows.Next() {
		var t domain.Telemetry
		if err := rows.Scan(
			&t.ICAO24, &t.Callsign, &t.OriginCountry, &t.TimePosition,
			&t.LastContact, &t.Longitude, &t.Latitude, &t.BaroAltitude,
			&t.OnGround, &t.Velocity, &t.TrueTrack, &t.VerticalRate,
			&t.Sensors, &t.GeoAltitude, &t.Squawk, &t.SPI,
			&t.PositionSource, &t.Category,
		); err != nil {
			return nil, fmt.Errorf("failed to scan telemetry row: %w", err)
		}
		telemetry = append(telemetry, t)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating telemetry rows: %w", err)
	}

	return telemetry, nil
}

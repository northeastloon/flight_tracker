package fetch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/northeastloon/flight_tracker/internal/domain"
)

const (
	openSkyBaseURL = "https://opensky-network.org/api/states/all"
)

type OpenSkyTelemetry struct {
	Icao24         string
	Callsign       *string
	OriginCountry  string
	TimePosition   *int64 // (unix timestamp)
	LastContact    int64  // (unix timestamp)
	Longitude      *float64
	Latitude       *float64
	BaroAltitude   *float64
	OnGround       bool
	Velocity       *float64 // (m/s)
	TrueTrack      *float64 // (degrees)
	VerticalRate   *float64
	Sensors        *[]int
	GeoAltitude    *float64
	Squawk         *string
	SPI            bool
	PositionSource int
	Category       int
}

type OpenSkyResponse struct {
	Time   int64 `json:"time"`
	States []any `json:"states"`
}

type OpenSkyClient struct {
	*Client
}

func NewOpenSkyClient(opts ...Option) *OpenSkyClient {
	baseOpts := []Option{
		WithBaseURL(openSkyBaseURL),
	}

	opts = append(baseOpts, opts...)

	return &OpenSkyClient{
		Client: NewClient(opts...),
	}
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (r *OpenSkyResponse) UnmarshalJSON(data []byte) error {

	type tempResponse struct {
		Time   int64           `json:"time"`
		States json.RawMessage `json:"states"`
	}

	var temp tempResponse
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	r.Time = temp.Time

	var states []any
	if err := json.Unmarshal(temp.States, &states); err != nil {
		return err
	}

	if len(states) == 0 {
		return fmt.Errorf("no states returned")
	}

	r.States = states
	return nil
}

func ParseOpenSkyTelemetry(r OpenSkyResponse) ([]OpenSkyTelemetry, error) {
	parsed := make([]OpenSkyTelemetry, 0)

	for _, rawState := range r.States {
		// Assert that rawState is a slice
		state, ok := rawState.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid state format: expected []any")
		}

		if len(state) < 18 {
			return nil, fmt.Errorf("invalid state length: got %d, expected 18", len(state))
		}

		// Create the telemetry struct
		telemetry := OpenSkyTelemetry{
			Icao24:         mustString(state[0]),
			Callsign:       optionalString(state[1]),
			OriginCountry:  mustString(state[2]),
			TimePosition:   optionalInt64(state[3]),
			LastContact:    mustInt64(state[4]),
			Longitude:      optionalFloat(state[5]),
			Latitude:       optionalFloat(state[6]),
			BaroAltitude:   optionalFloat(state[7]),
			OnGround:       mustBool(state[8]),
			Velocity:       optionalFloat(state[9]),
			TrueTrack:      optionalFloat(state[10]),
			VerticalRate:   optionalFloat(state[11]),
			Sensors:        optionalIntSlice(state[12]),
			GeoAltitude:    optionalFloat(state[13]),
			Squawk:         optionalString(state[14]),
			SPI:            mustBool(state[15]),
			PositionSource: mustInt(state[16]),
			Category:       mustInt(state[17]),
		}

		parsed = append(parsed, telemetry)
	}

	return parsed, nil
}

func (c *OpenSkyClient) FetchTelemetry(ctx context.Context) ([]domain.TelemetryRecord, error) {
	response, err := Fetch[OpenSkyResponse](ctx, c.Client)
	if err != nil {
		return nil, err
	}

	parsed, err := ParseOpenSkyTelemetry(response)
	if err != nil {
		return nil, err
	}

	var records []domain.TelemetryRecord
	for _, rec := range parsed {
		records = append(records, rec)
	}
	return records, nil
}

func (t OpenSkyTelemetry) ToDBRow() map[string]interface{} {
	return map[string]any{
		"icao24":          t.Icao24,
		"callsign":        t.Callsign,
		"origin_country":  t.OriginCountry,
		"time_position":   t.TimePosition,
		"last_contact":    t.LastContact,
		"longitude":       t.Longitude,
		"latitude":        t.Latitude,
		"baro_altitude":   t.BaroAltitude,
		"on_ground":       t.OnGround,
		"velocity":        t.Velocity,
		"true_track":      t.TrueTrack,
		"vertical_rate":   t.VerticalRate,
		"sensors":         t.Sensors,
		"geo_altitude":    t.GeoAltitude,
		"squawk":          t.Squawk,
		"spi":             t.SPI,
		"position_source": t.PositionSource,
		"category":        t.Category,
	}
}

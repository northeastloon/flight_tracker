package domain

type TelemetrySource interface {
	Fetch() ([]TelemetryRecord, error)
	TableName() string
}

type TelemetryRecord interface {
	ToDBRow() map[string]any
}

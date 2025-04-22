package server

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/northeastloon/flight_tracker/internal/domain"
)

type TelemetryStore interface {
	GetTelemetry(ctx context.Context, filter *domain.TelemetryFilter) ([]domain.Telemetry, error)
}

type APIHandler struct {
	store TelemetryStore
}

func NewAPIHandler(store TelemetryStore) *APIHandler {
	return &APIHandler{
		store: store,
	}
}

func (h *APIHandler) GetTelemetry(c echo.Context) error {
	filter := &domain.TelemetryFilter{}
	if err := c.Bind(filter); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	telemetry, err := h.store.GetTelemetry(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, telemetry)
}

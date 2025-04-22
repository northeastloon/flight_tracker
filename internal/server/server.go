package server

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Echo       *echo.Echo
	ApiHandler *APIHandler
	WebHandler *WebHandler
}

func NewServer(store TelemetryStore) (*Server, error) {
	e := echo.New()

	// API handler needs store for data access
	apiHandler := NewAPIHandler(store)
	// Web handler only needs templates
	webHandler, err := NewWebHandler()
	if err != nil {
		return nil, err
	}

	s := &Server{
		Echo:       e,
		ApiHandler: apiHandler,
		WebHandler: webHandler,
	}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 15 * time.Second,
	}))

	// Map routes
	s.mapRoutes()

	renderer := &Template{
		templates: webHandler.templates,
	}

	s.Echo.Renderer = renderer

	return s, nil
}

func (s *Server) mapRoutes() {

	// API routes
	api := s.Echo.Group("/api/v1")
	api.GET("/telemetry", s.ApiHandler.GetTelemetry)

	// Web routes
	s.Echo.GET("/", s.WebHandler.GlobeHandler)

	// Static files
	s.Echo.Static("/static", "internal/server/web/static")
}

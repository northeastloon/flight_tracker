package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type WebHandler struct {
	templates *template.Template
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// No store dependency needed
func NewWebHandler() (*WebHandler, error) {
	tmpl, err := template.ParseFS(templateFS,
		"templates/*.tmpl",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err) // Added error wrapping
	}

	return &WebHandler{
		templates: tmpl,
	}, nil
}

func (h *WebHandler) GlobeHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "layout.tmpl", map[string]interface{}{
		"Title": "Flight Map",
	})
}

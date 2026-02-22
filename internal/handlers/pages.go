package handlers

import (
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// PageHandler serves HTML pages for the web interface.
type PageHandler struct {
	engine      *templateengine.Engine
	authService services.AuthenticationService
}

// NewPageHandler creates a new PageHandler with the given template engine and auth service.
func NewPageHandler(engine *templateengine.Engine, authService services.AuthenticationService) *PageHandler {
	return &PageHandler{
		engine:      engine,
		authService: authService,
	}
}

// flashMessage represents a single flash message for the flash-messages partial.
type flashMessage struct {
	Type    string
	Message string
}

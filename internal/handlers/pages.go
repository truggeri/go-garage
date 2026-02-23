package handlers

import (
	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// PageHandler serves HTML pages for the web interface.
type PageHandler struct {
	engine             *templateengine.Engine
	authService        services.AuthenticationService
	tokenManager       *auth.TokenManager
	vehicleService     services.VehicleService
	maintenanceService services.MaintenanceService
}

// NewPageHandler creates a new PageHandler with the given template engine and services.
func NewPageHandler(
	engine *templateengine.Engine,
	authService services.AuthenticationService,
	tokenMgr *auth.TokenManager,
	vehicleSvc services.VehicleService,
	maintenanceSvc services.MaintenanceService,
) *PageHandler {
	return &PageHandler{
		engine:             engine,
		authService:        authService,
		tokenManager:       tokenMgr,
		vehicleService:     vehicleSvc,
		maintenanceService: maintenanceSvc,
	}
}

// flashMessage represents a single flash message for the flash-messages partial.
type flashMessage struct {
	Type    string
	Message string
}

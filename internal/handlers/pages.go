package handlers

import (
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// PageHandler serves HTML pages for the web interface.
type PageHandler struct {
	engine             *templateengine.Engine
	authService        services.AuthenticationService
	vehicleService     services.VehicleService
	maintenanceService services.MaintenanceService
	fuelService        services.FuelService
	userService        services.UserService
}

// NewPageHandler creates a new PageHandler with the given template engine and services.
func NewPageHandler(
	engine *templateengine.Engine,
	authService services.AuthenticationService,
	vehicleSvc services.VehicleService,
	maintenanceSvc services.MaintenanceService,
	fuelSvc services.FuelService,
	userSvc services.UserService,
) *PageHandler {
	return &PageHandler{
		engine:             engine,
		authService:        authService,
		vehicleService:     vehicleSvc,
		maintenanceService: maintenanceSvc,
		fuelService:        fuelSvc,
		userService:        userSvc,
	}
}

// flashMessage represents a single flash message for the flash-messages partial.
type flashMessage struct {
	Type    string
	Message string
}

// queryTrue is the literal value used when checking boolean query parameters.
const queryTrue = "true"

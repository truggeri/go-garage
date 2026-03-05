package handlers

import (
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// PageHandler serves HTML pages for the web interface.
type PageHandler struct {
	engine             *templateengine.Engine
	authService        services.AuthenticationService
	vehicleService     services.VehicleService
	maintenanceService services.MaintenanceService
	userService        services.UserService
	metricsRepo        repositories.MetricsRepository
}

// NewPageHandler creates a new PageHandler with the given template engine and services.
func NewPageHandler(
	engine *templateengine.Engine,
	authService services.AuthenticationService,
	vehicleSvc services.VehicleService,
	maintenanceSvc services.MaintenanceService,
	userSvc services.UserService,
	metricsRepo repositories.MetricsRepository,
) *PageHandler {
	return &PageHandler{
		engine:             engine,
		authService:        authService,
		vehicleService:     vehicleSvc,
		maintenanceService: maintenanceSvc,
		userService:        userSvc,
		metricsRepo:        metricsRepo,
	}
}

// flashMessage represents a single flash message for the flash-messages partial.
type flashMessage struct {
	Type    string
	Message string
}

// queryTrue is the literal value used when checking boolean query parameters.
const queryTrue = "true"

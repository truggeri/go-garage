package handlers

import (
	"context"

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

// getUserTotalSpent returns the sum of total_spent from pre-computed metrics
// for the given vehicle IDs. Returns 0 if metrics are unavailable.
func (h *PageHandler) getUserTotalSpent(ctx context.Context, vehicleIDs []string) float64 {
	if h.metricsRepo == nil || len(vehicleIDs) == 0 {
		return 0
	}
	sum, err := h.metricsRepo.SumTotalSpentByVehicleIDs(ctx, vehicleIDs)
	if err != nil {
		return 0
	}
	return sum
}

// getVehicleTotalSpent returns the total_spent from pre-computed metrics
// for a single vehicle. Returns 0 if metrics are unavailable.
func (h *PageHandler) getVehicleTotalSpent(ctx context.Context, vehicleID string) float64 {
	if h.metricsRepo == nil {
		return 0
	}
	m, err := h.metricsRepo.GetByVehicleID(ctx, vehicleID)
	if err != nil || m == nil || m.TotalSpent == nil {
		return 0
	}
	return *m.TotalSpent
}

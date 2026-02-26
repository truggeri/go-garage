package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// newVehicleLookup returns a ResourceLookup that loads a vehicle by the "id"
// path variable and returns its owner ID.
func newVehicleLookup(svc services.VehicleService) middleware.ResourceLookup {
	return func(ctx context.Context, r *http.Request) (interface{}, string, error) {
		vehicleID := mux.Vars(r)["id"]
		vehicle, err := svc.GetVehicle(ctx, vehicleID)
		if err != nil {
			var notFound *models.NotFoundError
			if errors.As(err, &notFound) {
				return nil, "", middleware.ErrResourceNotFound
			}
			return nil, "", err
		}
		return vehicle, vehicle.UserID, nil
	}
}

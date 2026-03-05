package handlers

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/models"
)

// getFuelRecordAndVehicle retrieves a fuel record and its associated vehicle.
func (h *PageHandler) getFuelRecordAndVehicle(r *http.Request) (*models.FuelRecord, *models.Vehicle, error) {
	vars := mux.Vars(r)
	recordID := vars["id"]

	record, err := h.fuelService.GetFuelRecord(r.Context(), recordID)
	if err != nil {
		return nil, nil, err
	}

	vehicle, err := h.vehicleService.GetVehicle(r.Context(), record.VehicleID)
	if err != nil {
		return nil, nil, err
	}

	return record, vehicle, nil
}

// writeFuelRecordError writes the appropriate HTTP error for fuel record lookup failures.
func writeFuelRecordError(w http.ResponseWriter, err error) {
	var notFound *models.NotFoundError
	if errors.As(err, &notFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

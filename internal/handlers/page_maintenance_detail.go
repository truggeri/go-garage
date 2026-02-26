package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// maintenanceDetailPageData holds the data passed to the maintenance detail template.
type maintenanceDetailPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Record is the maintenance record to display.
	Record *models.MaintenanceRecord
	// Vehicle is the vehicle associated with this maintenance record.
	Vehicle *models.Vehicle
	// VehicleTitle is a short human-readable title for the vehicle (e.g. "2020 Ford Focus").
	VehicleTitle string
}

// MaintenanceDetail serves the maintenance record detail page (GET /maintenance/{id}).
func (h *PageHandler) MaintenanceDetail(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	recordID := vars["id"]

	record, err := h.maintenanceService.GetMaintenance(r.Context(), recordID)
	if err != nil {
		var notFound *models.NotFoundError
		if errors.As(err, &notFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicle, err := h.vehicleService.GetVehicle(r.Context(), record.VehicleID)
	if err != nil {
		var notFound *models.NotFoundError
		if errors.As(err, &notFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	title := fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model)

	data := maintenanceDetailPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		Record:          record,
		Vehicle:         vehicle,
		VehicleTitle:    title,
	}

	if r.URL.Query().Get("updated") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Maintenance record updated successfully."},
		}
	}

	if err := h.engine.Render(w, "maintenance/detail.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

package handlers

import (
	"fmt"
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// fuelDetailPageData holds the data passed to the fuel detail template.
type fuelDetailPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// Record is the fuel record to display.
	Record *models.FuelRecord
	// Vehicle is the vehicle associated with this fuel record.
	Vehicle *models.Vehicle
	// VehicleTitle is a short human-readable title for the vehicle (e.g. "2020 Ford Focus").
	VehicleTitle string
	// CSRFToken is the CSRF protection token for the delete form.
	CSRFToken string
	// TotalCost is the computed total cost (volume * price per unit), or nil if unavailable.
	TotalCost *float64
}

// FuelDetail serves the fuel record detail page (GET /fuel/{id}).
func (h *PageHandler) FuelDetail(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getFuelRecordAndVehicle(r)
	if err != nil {
		writeFuelRecordError(w, err)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	title := fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model)

	var totalCost *float64
	if record.PricePerUnit != nil {
		tc := record.Volume * *record.PricePerUnit
		totalCost = &tc
	}

	data := fuelDetailPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		Record:          record,
		Vehicle:         vehicle,
		VehicleTitle:    title,
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
		TotalCost:       totalCost,
	}

	if r.URL.Query().Get("updated") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Fuel record updated successfully."},
		}
	}

	if err := h.engine.Render(w, "fuel/detail.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// FuelDelete handles the fuel record deletion (POST /fuel/{id}/delete).
func (h *PageHandler) FuelDelete(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	record, vehicle, err := h.getFuelRecordAndVehicle(r)
	if err != nil {
		writeFuelRecordError(w, err)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.fuelService.DeleteFuel(r.Context(), record.ID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/fuel?vehicle="+vehicle.ID, http.StatusSeeOther)
}

package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// fuelNewPageData holds the data passed to the add-fuel template.
type fuelNewPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// Vehicles is the list of the user's vehicles, used to populate the vehicle dropdown.
	Vehicles []*models.Vehicle
	// VehicleNames maps vehicle IDs to human-readable names.
	VehicleNames map[string]string
	// Errors holds field-level and general validation error messages.
	Errors    map[string]string
	CSRFToken string
	// VehicleID is the selected vehicle.
	VehicleID string
	// FuelTypes is the list of valid fuel type enum values for dropdowns.
	FuelTypes []models.FuelType
	// Form field values for repopulating the form after a failed submission.
	FillDate              string
	Mileage               string
	Volume                string
	FuelType              string
	PartialFill           bool
	PricePerUnit          string
	OctaneRating          string
	Location              string
	Brand                 string
	Notes                 string
	CityDrivingPercentage string
	VehicleReportedMPG    string
}

// FuelNew serves the add fuel record form page (GET /fuel/new).
func (h *PageHandler) FuelNew(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicleID := r.URL.Query().Get("vehicle")
	if vehicleID == "" {
		vehicleID = r.URL.Query().Get("vehicle_id")
	}

	data := fuelNewPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		Vehicles:        vehicles,
		VehicleNames:    buildVehicleNameMap(vehicles),
		VehicleID:       vehicleID,
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
		FuelTypes:       models.AllFuelTypes(),
	}

	if err := h.engine.Render(w, "fuel/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

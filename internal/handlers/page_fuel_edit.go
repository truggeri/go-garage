package handlers

import (
	"fmt"
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// fuelEditPageData holds the data passed to the edit-fuel template.
type fuelEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// RecordID is the ID of the fuel record being edited.
	RecordID string
	// VehicleTitle is a short human-readable title for the vehicle (e.g. "2020 Ford Focus").
	VehicleTitle string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// CSRFToken is the CSRF protection token to embed in the form.
	CSRFToken string
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

// fuelEditPageDataFromRecord builds an edit page data struct pre-populated
// with the current values of the given fuel record.
func fuelEditPageDataFromRecord(
	account *middleware.AccountInfo,
	record *models.FuelRecord,
	vehicle *models.Vehicle,
) fuelEditPageData {
	data := fuelEditPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "fuel",
		RecordID:        record.ID,
		VehicleTitle:    fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model),
		FillDate:        record.FillDate.Format("2006-01-02"),
		Mileage:         fmt.Sprintf("%d", record.Mileage),
		Volume:          fmt.Sprintf("%.2f", record.Volume),
		FuelType:        record.FuelType,
		PartialFill:     record.PartialFill,
		Location:        record.Location,
		Brand:           record.Brand,
		Notes:           record.Notes,
		FuelTypes:       models.AllFuelTypes(),
	}
	if record.PricePerUnit != nil {
		data.PricePerUnit = fmt.Sprintf("%.3f", *record.PricePerUnit)
	}
	if record.OctaneRating != nil {
		data.OctaneRating = fmt.Sprintf("%d", *record.OctaneRating)
	}
	if record.CityDrivingPercentage != nil {
		data.CityDrivingPercentage = fmt.Sprintf("%d", *record.CityDrivingPercentage)
	}
	if record.VehicleReportedMPG != nil {
		data.VehicleReportedMPG = fmt.Sprintf("%.1f", *record.VehicleReportedMPG)
	}
	return data
}

// FuelEdit serves the edit fuel record form page (GET /fuel/{id}/edit).
func (h *PageHandler) FuelEdit(w http.ResponseWriter, r *http.Request) {
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

	data := fuelEditPageDataFromRecord(account, record, vehicle)
	data.CSRFToken = middleware.GetCSRFToken(r.Context())

	if err := h.engine.Render(w, "fuel/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

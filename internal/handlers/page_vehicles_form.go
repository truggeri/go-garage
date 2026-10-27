package handlers

import (
	"fmt"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// vehicleEditPageData holds the data passed to the edit-vehicle template.
type vehicleEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// VehicleID is the ID of the vehicle being edited.
	VehicleID string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// CSRFToken is the CSRF protection token to embed in the form.
	CSRFToken string
	// Form field values for repopulating the form after a failed submission.
	Make            string
	Model           string
	Year            string
	VIN             string
	DisplayName     string
	Color           string
	LicensePlate    string
	PurchaseDate    string
	PurchasePrice   string
	PurchaseMileage string
	CurrentMileage  string
	Notes           string
}

// vehicleEditPageDataFromVehicle builds an edit page data struct pre-populated
// with the current values of the given vehicle.
func vehicleEditPageDataFromVehicle(account *middleware.AccountInfo, v *models.Vehicle) vehicleEditPageData {
	data := vehicleEditPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		ActiveNav:       "vehicles",
		VehicleID:       v.ID,
		Make:            v.Make,
		Model:           v.Model,
		Year:            fmt.Sprintf("%d", v.Year),
		VIN:             v.VIN,
		DisplayName:     v.DisplayName,
		Color:           v.Color,
		LicensePlate:    v.LicensePlate,
		Notes:           v.Notes,
	}
	if v.PurchaseDate != nil {
		data.PurchaseDate = v.PurchaseDate.Format("2006-01-02")
	}
	if v.PurchasePrice != nil {
		data.PurchasePrice = fmt.Sprintf("%.2f", *v.PurchasePrice)
	}
	if v.PurchaseMileage != nil {
		data.PurchaseMileage = fmt.Sprintf("%d", *v.PurchaseMileage)
	}
	if v.CurrentMileage != nil {
		data.CurrentMileage = fmt.Sprintf("%d", *v.CurrentMileage)
	}
	return data
}

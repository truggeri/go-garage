package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// vehicleNewFormResult holds the parsed results of the vehicle creation form.
type vehicleNewFormResult struct {
	Year            int
	PurchaseDate    *time.Time
	PurchasePrice   *float64
	PurchaseMileage *int
	CurrentMileage  *int
	Errors          map[string]string
}

// parseVehicleNewForm parses the raw string values submitted via the add-vehicle
// form into their typed equivalents. It returns parsed result values and a map of
// field-level error messages for any inputs that cannot be converted to the
// expected type. Business validation (e.g. required fields, value ranges) is
// handled by models.ValidateVehicleAll.
func parseVehicleNewForm(
	yearStr, purchaseDateStr, purchasePriceStr, purchaseMileageStr, currentMileageStr string,
) vehicleNewFormResult {
	result := vehicleNewFormResult{Errors: make(map[string]string)}

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err != nil {
			result.Errors["year"] = "Year must be a valid number"
		} else {
			result.Year = y
		}
	}

	if purchaseDateStr != "" {
		t, err := time.Parse("2006-01-02", purchaseDateStr)
		if err != nil {
			result.Errors["purchase_date"] = "Invalid date format"
		} else {
			result.PurchaseDate = &t
		}
	}

	if purchasePriceStr != "" {
		p, err := strconv.ParseFloat(purchasePriceStr, 64)
		if err != nil {
			result.Errors["purchase_price"] = "Purchase price must be a valid number"
		} else {
			result.PurchasePrice = &p
		}
	}

	if purchaseMileageStr != "" {
		m, err := strconv.Atoi(purchaseMileageStr)
		if err != nil {
			result.Errors["purchase_mileage"] = "Mileage at purchase must be a valid number"
		} else {
			result.PurchaseMileage = &m
		}
	}

	if currentMileageStr != "" {
		m, err := strconv.Atoi(currentMileageStr)
		if err != nil {
			result.Errors["current_mileage"] = "Current mileage must be a valid number"
		} else {
			result.CurrentMileage = &m
		}
	}

	return result
}

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

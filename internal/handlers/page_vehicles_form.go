package handlers

import (
	"strconv"
	"time"
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

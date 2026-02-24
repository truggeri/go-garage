package handlers

import (
	"strconv"
	"strings"
	"time"
)

// vehicleNewFormResult holds the parsed and validated results of the vehicle creation form.
type vehicleNewFormResult struct {
	Year            int
	PurchaseDate    *time.Time
	PurchasePrice   *float64
	PurchaseMileage *int
	CurrentMileage  *int
	Errors          map[string]string
}

// validateVehicleNewForm validates and parses the raw string values submitted via the
// add-vehicle form. It returns parsed result values and a map of field-level error
// messages for any invalid inputs.
func validateVehicleNewForm(
	vehicleMake, model, yearStr, purchaseDateStr, purchasePriceStr, purchaseMileageStr, currentMileageStr string,
) vehicleNewFormResult {
	result := vehicleNewFormResult{Errors: make(map[string]string)}

	if strings.TrimSpace(vehicleMake) == "" {
		result.Errors["make"] = "Make is required"
	}
	if strings.TrimSpace(model) == "" {
		result.Errors["model"] = "Model is required"
	}

	if yearStr == "" {
		result.Errors["year"] = "Year is required"
	} else if y, err := strconv.Atoi(yearStr); err != nil || y < 1900 || y > 2100 {
		result.Errors["year"] = "Year must be a valid year (1900-2100)"
	} else {
		result.Year = y
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
		if err != nil || p < 0 {
			result.Errors["purchase_price"] = "Purchase price must be a non-negative number"
		} else {
			result.PurchasePrice = &p
		}
	}

	if purchaseMileageStr != "" {
		m, err := strconv.Atoi(purchaseMileageStr)
		if err != nil || m < 0 {
			result.Errors["purchase_mileage"] = "Mileage at purchase must be a non-negative number"
		} else {
			result.PurchaseMileage = &m
		}
	}

	if currentMileageStr != "" {
		m, err := strconv.Atoi(currentMileageStr)
		if err != nil || m < 0 {
			result.Errors["current_mileage"] = "Current mileage must be a non-negative number"
		} else {
			result.CurrentMileage = &m
		}
	}

	return result
}

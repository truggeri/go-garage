package handlers

import (
	"math"
	"strconv"
)

// normalizeMaintenanceFormNumbers converts json-enc number inputs for the API.
func normalizeMaintenanceFormNumbers(data map[string]interface{}) []FieldError {
	var errs []FieldError
	for _, field := range []string{"mileage_at_service", "cost"} {
		value, ok := data[field].(string)
		if !ok {
			continue
		}
		if value == "" {
			data[field] = nil
			continue
		}
		number, err := strconv.ParseFloat(value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) ||
			field == "mileage_at_service" && (number != math.Trunc(number) || number > math.MaxInt32 || number < math.MinInt32) {
			errs = append(errs, FieldError{Field: field, Message: "must be a valid number"})
			continue
		}
		data[field] = number
	}
	return errs
}

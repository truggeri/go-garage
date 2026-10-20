package handlers

import (
	"math"
	"strconv"
)

var fuelFormNumberFields = []string{
	"mileage", "volume", "price_per_unit", "octane_rating",
	"city_driving_percentage", "vehicle_reported_mpg",
}

// normalizeFuelFormData converts json-enc's string-valued form inputs into the
// JSON types expected by the fuel API. Blank optional numbers are dropped and an
// unchecked partial_fill checkbox (absent from the form) becomes false.
func normalizeFuelFormData(data map[string]interface{}) []FieldError {
	var errs []FieldError
	for _, field := range fuelFormNumberFields {
		text, ok := data[field].(string)
		if !ok {
			continue
		}
		if text == "" {
			delete(data, field)
			continue
		}
		number, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			errs = append(errs, FieldError{Field: field, Message: "must be a valid number"})
			continue
		}
		data[field] = number
	}

	switch v := data["partial_fill"].(type) {
	case string:
		data["partial_fill"] = v == "on" || v == "true"
	case bool:
	default:
		data["partial_fill"] = false
	}
	return errs
}

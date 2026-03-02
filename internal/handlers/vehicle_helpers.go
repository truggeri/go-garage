package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

// buildVehicleFilterSpec creates a VehicleFilters struct from request query parameters.
func buildVehicleFilterSpec(r *http.Request, ownerID string) repositories.VehicleFilters {
	q := r.URL.Query()
	f := repositories.VehicleFilters{UserID: &ownerID}
	if v := q.Get("make"); v != "" {
		f.Make = &v
	}
	if v := q.Get("model"); v != "" {
		f.Model = &v
	}
	if v := q.Get("year"); v != "" {
		if yr, e := strconv.Atoi(v); e == nil {
			f.Year = &yr
		}
	}
	if v := q.Get("status"); v != "" {
		s := models.VehicleStatus(v)
		f.Status = &s
	}
	return f
}

// parseJSONBody decodes the request body as JSON into a map.
func parseJSONBody(r *http.Request) (map[string]interface{}, error) {
	var d map[string]interface{}
	return d, json.NewDecoder(r.Body).Decode(&d)
}

// validateRequiredKeys checks that all specified keys are present and non-empty in the data map.
// Returns a slice of FieldError for each missing or empty required field.
func validateRequiredKeys(d map[string]interface{}, keys ...string) []FieldError {
	var errs []FieldError
	for _, k := range keys {
		switch v := d[k].(type) {
		case nil:
			errs = append(errs, FieldError{Field: k, Message: "required"})
		case string:
			if v == "" {
				errs = append(errs, FieldError{Field: k, Message: "required"})
			}
		case float64:
			if k == "year" && v == 0 {
				errs = append(errs, FieldError{Field: k, Message: "required"})
			}
		}
	}
	return errs
}

// buildNewVehicleRecord creates a new Vehicle model from the input data map and owner ID.
func buildNewVehicleRecord(d map[string]interface{}, ownerID string) (*models.Vehicle, error) {
	vinStr, _ := d["vin"].(string)
	makeStr, _ := d["make"].(string)
	modelStr, _ := d["model"].(string)
	yearNum, _ := d["year"].(float64)

	st := models.VehicleStatusActive
	if sRaw, ok := d["status"].(string); ok && sRaw != "" {
		st = models.VehicleStatus(sRaw)
	}

	var pdt *time.Time
	if pdStr, ok := d["purchase_date"].(string); ok && pdStr != "" {
		t, e := time.Parse("2006-01-02", pdStr)
		if e != nil {
			return nil, e
		}
		pdt = &t
	}

	rec := &models.Vehicle{
		UserID: ownerID, VIN: strings.ToUpper(strings.TrimSpace(vinStr)),
		Make: makeStr, Model: modelStr, Year: int(yearNum), Status: st, PurchaseDate: pdt,
	}

	if dn, ok := d["display_name"].(string); ok {
		rec.DisplayName = dn
	}
	if c, ok := d["color"].(string); ok {
		rec.Color = c
	}
	if lp, ok := d["license_plate"].(string); ok {
		rec.LicensePlate = lp
	}
	if pp, ok := d["purchase_price"].(float64); ok {
		rec.PurchasePrice = &pp
	}
	if pm, ok := d["purchase_mileage"].(float64); ok {
		i := int(pm)
		rec.PurchaseMileage = &i
	}
	if cm, ok := d["current_mileage"].(float64); ok {
		i := int(cm)
		rec.CurrentMileage = &i
	}
	if n, ok := d["notes"].(string); ok {
		rec.Notes = n
	}

	return rec, nil
}

// extractVehicleChanges extracts vehicle update fields from the input data map.
func extractVehicleChanges(d map[string]interface{}) services.VehicleUpdates {
	var u services.VehicleUpdates
	if v, ok := d["display_name"].(string); ok {
		u.DisplayName = &v
	}
	if v, ok := d["vin"].(string); ok && v != "" {
		s := strings.ToUpper(strings.TrimSpace(v))
		u.VIN = &s
	}
	if v, ok := d["make"].(string); ok && v != "" {
		u.Make = &v
	}
	if v, ok := d["model"].(string); ok && v != "" {
		u.Model = &v
	}
	if v, ok := d["year"].(float64); ok && v != 0 {
		i := int(v)
		u.Year = &i
	}
	if v, ok := d["color"].(string); ok && v != "" {
		u.Color = &v
	}
	if v, ok := d["license_plate"].(string); ok && v != "" {
		u.LicensePlate = &v
	}
	if v, ok := d["current_mileage"].(float64); ok {
		i := int(v)
		u.CurrentMileage = &i
	}
	if v, ok := d["notes"].(string); ok && v != "" {
		u.Notes = &v
	}
	return u
}

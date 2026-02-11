package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// vehicleToResponseMap converts a Vehicle model to a response map for JSON encoding.
func vehicleToResponseMap(v *models.Vehicle) map[string]interface{} {
	m := map[string]interface{}{
		"id": v.ID, "user_id": v.UserID, "vin": v.VIN,
		"make": v.Make, "model": v.Model, "year": v.Year,
		"status":     string(v.Status),
		"created_at": v.CreatedAt.Format(time.RFC3339),
		"updated_at": v.UpdatedAt.Format(time.RFC3339),
	}
	if v.Color != "" {
		m["color"] = v.Color
	}
	if v.LicensePlate != "" {
		m["license_plate"] = v.LicensePlate
	}
	if v.PurchaseDate != nil {
		m["purchase_date"] = v.PurchaseDate.Format("2006-01-02")
	}
	if v.PurchasePrice != nil {
		m["purchase_price"] = *v.PurchasePrice
	}
	if v.PurchaseMileage != nil {
		m["purchase_mileage"] = *v.PurchaseMileage
	}
	if v.CurrentMileage != nil {
		m["current_mileage"] = *v.CurrentMileage
	}
	if v.Notes != "" {
		m["notes"] = v.Notes
	}
	return m
}

// buildListPayload creates a paginated list response payload.
func buildListPayload(recs []*models.Vehicle, pg, sz, total int) map[string]interface{} {
	items := make([]map[string]interface{}, len(recs))
	for i, v := range recs {
		items[i] = vehicleToResponseMap(v)
	}
	tp := 0
	if total > 0 && sz > 0 {
		tp = total / sz
		if total%sz != 0 {
			tp++
		}
	}
	return map[string]interface{}{
		"success": true, "data": items,
		"pagination": map[string]int{"page": pg, "limit": sz, "total": total, "total_pages": tp},
	}
}

// buildSinglePayload creates a single vehicle response payload.
func buildSinglePayload(v *models.Vehicle, msg string) map[string]interface{} {
	p := map[string]interface{}{"success": true, "data": vehicleToResponseMap(v)}
	if msg != "" {
		p["message"] = msg
	}
	return p
}

// handleDomainError converts domain errors to HTTP responses.
func handleDomainError(w http.ResponseWriter, err error) {
	var ve *models.ValidationError
	if models.IsValidationError(err, &ve) {
		det := []FieldError{}
		if ve.Field != "" {
			det = append(det, FieldError{Field: ve.Field, Message: ve.Message})
		}
		respondWithValidationProblems(w, ve.Message, det)
		return
	}
	var de *models.DuplicateError
	if models.IsDuplicateError(err, &de) {
		respondWithProblem(w, 409, "DUPLICATE_ERROR", de.Error())
		return
	}
	var ne *models.NotFoundError
	if models.IsNotFoundError(err, &ne) {
		respondWithProblem(w, 404, "NOT_FOUND", ne.Error())
		return
	}
	respondWithProblem(w, 500, "INTERNAL_ERROR", "Unexpected error")
}

// respondWithProblem writes a problem/error JSON response.
func respondWithProblem(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]string{"code": errCode, "message": msg},
	})
}

// respondWithValidationProblems writes a validation error JSON response with field details.
func respondWithValidationProblems(w http.ResponseWriter, msg string, details []FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]interface{}{"code": "VALIDATION_ERROR", "message": msg, "details": details},
	})
}

// respondWithPayload writes a success JSON response with the given payload.
func respondWithPayload(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
)

type vehicleAPIHandler struct {
	svc services.VehicleService
}

func MakeVehicleAPIHandler(svc services.VehicleService) *vehicleAPIHandler {
	return &vehicleAPIHandler{svc: svc}
}

func (h *vehicleAPIHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	pageIdx, pageLen := extractPaging(r)
	filterSpec := buildVehicleFilterSpec(r, caller.ID)
	offsetVal := (pageIdx - 1) * pageLen

	totalCount, countErr := h.svc.CountVehicles(ctx, filterSpec)
	if countErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed counting")
		return
	}

	records, fetchErr := h.svc.ListVehicles(ctx, filterSpec, repositories.PaginationParams{
		Limit: pageLen, Offset: offsetVal,
	})
	if fetchErr != nil {
		respondWithProblem(w, 500, "INTERNAL_ERROR", "Failed fetching")
		return
	}

	respondWithPayload(w, 200, buildListPayload(records, pageIdx, pageLen, totalCount))
}

func (h *vehicleAPIHandler) CreateOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "vin", "make", "model", "year")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	newRec, buildErr := buildNewVehicleRecord(inputData, caller.ID)
	if buildErr != nil {
		respondWithProblem(w, 400, "VALIDATION_ERROR", buildErr.Error())
		return
	}

	if svcErr := h.svc.CreateVehicle(ctx, newRec); svcErr != nil {
		handleDomainError(w, svcErr)
		return
	}

	respondWithPayload(w, 201, buildSinglePayload(newRec, "Vehicle created successfully"))
}

func (h *vehicleAPIHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	rec, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if rec.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	respondWithPayload(w, 200, buildSinglePayload(rec, ""))
}

func (h *vehicleAPIHandler) ReplaceOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	changes := extractVehicleChanges(inputData)
	updated, updateErr := h.svc.UpdateVehicle(ctx, targetID, changes)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, 200, buildSinglePayload(updated, "Vehicle updated successfully"))
}

func (h *vehicleAPIHandler) RemoveOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	if delErr := h.svc.DeleteVehicle(ctx, targetID); delErr != nil {
		handleDomainError(w, delErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true, "message": "Vehicle deleted successfully",
	})
}

func (h *vehicleAPIHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	targetID := mux.Vars(r)["id"]
	existing, lookupErr := h.svc.GetVehicle(ctx, targetID)
	if lookupErr != nil {
		handleDomainError(w, lookupErr)
		return
	}

	if existing.UserID != caller.ID {
		respondWithProblem(w, 403, "FORBIDDEN", "Not your vehicle")
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"vehicle_id": targetID, "total_maintenance_cost": 0.0,
			"maintenance_count": 0, "total_fuel_cost": 0.0,
			"fuel_fill_count": 0, "average_fuel_price": 0.0,
			"last_maintenance_date": nil, "last_fuel_date": nil,
		},
	})
}

func extractPaging(r *http.Request) (int, int) {
	q := r.URL.Query()
	pg, sz := 1, 20
	if v := q.Get("page"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			pg = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 100 {
			sz = n
		}
	}
	return pg, sz
}

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

func parseJSONBody(r *http.Request) (map[string]interface{}, error) {
	var d map[string]interface{}
	return d, json.NewDecoder(r.Body).Decode(&d)
}

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

func extractVehicleChanges(d map[string]interface{}) services.VehicleUpdates {
	var u services.VehicleUpdates
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

func buildSinglePayload(v *models.Vehicle, msg string) map[string]interface{} {
	p := map[string]interface{}{"success": true, "data": vehicleToResponseMap(v)}
	if msg != "" {
		p["message"] = msg
	}
	return p
}

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

func respondWithProblem(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]string{"code": errCode, "message": msg},
	})
}

func respondWithValidationProblems(w http.ResponseWriter, msg string, details []FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]interface{}{"code": "VALIDATION_ERROR", "message": msg, "details": details},
	})
}

func respondWithPayload(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

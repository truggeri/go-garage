package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// vehicleListPageData holds the data passed to the vehicle list template.
type vehicleListPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Vehicles is the slice of vehicles to display on the current page.
	Vehicles []*models.Vehicle
	// TotalCount is the total number of vehicles matching the current filters.
	TotalCount int
	// Page is the current page number (1-based).
	Page int
	// PageSize is the number of vehicles per page.
	PageSize int
	// TotalPages is the total number of pages.
	TotalPages int
	// FilterMake is the current make filter value.
	FilterMake string
	// FilterStatus is the current status filter value.
	FilterStatus string
	// SortBy is the current sort field.
	SortBy string
}

const vehicleListPageSize = 12

// VehicleList serves the vehicle list page (GET /vehicles).
func (h *PageHandler) VehicleList(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	page := 1
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}

	filterMake := r.URL.Query().Get("make")
	filterStatus := r.URL.Query().Get("status")

	filters := repositories.VehicleFilters{UserID: &account.ID}
	if filterMake != "" {
		filters.Make = &filterMake
	}
	if filterStatus != "" {
		s := models.VehicleStatus(filterStatus)
		filters.Status = &s
	}

	totalCount, err := h.vehicleService.CountVehicles(r.Context(), filters)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(vehicleListPageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * vehicleListPageSize
	vehicles, err := h.vehicleService.ListVehicles(r.Context(), filters, repositories.PaginationParams{
		Limit:  vehicleListPageSize,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := vehicleListPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		Vehicles:        vehicles,
		TotalCount:      totalCount,
		Page:            page,
		PageSize:        vehicleListPageSize,
		TotalPages:      totalPages,
		FilterMake:      filterMake,
		FilterStatus:    filterStatus,
	}

	if r.URL.Query().Get("added") == "true" {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Vehicle added successfully."},
		}
	}

	if err := h.engine.Render(w, "vehicles/list.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// vehicleNewPageData holds the data passed to the add-vehicle template.
type vehicleNewPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// Form field values for repopulating the form after a failed submission.
	Make            string
	Model           string
	Year            string
	VIN             string
	Color           string
	LicensePlate    string
	PurchaseDate    string
	PurchasePrice   string
	PurchaseMileage string
	CurrentMileage  string
	Notes           string
}

// VehicleNew serves the add vehicle form page (GET /vehicles/new).
func (h *PageHandler) VehicleNew(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := vehicleNewPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
	}

	if err := h.engine.Render(w, "vehicles/new.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// VehicleCreate handles the add vehicle form submission (POST /vehicles/new).
func (h *PageHandler) VehicleCreate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	vehicleMake := r.FormValue("make")
	model := r.FormValue("model")
	yearStr := r.FormValue("year")
	vin := r.FormValue("vin")
	color := r.FormValue("color")
	licensePlate := r.FormValue("license_plate")
	purchaseDateStr := r.FormValue("purchase_date")
	purchasePriceStr := r.FormValue("purchase_price")
	purchaseMileageStr := r.FormValue("purchase_mileage")
	currentMileageStr := r.FormValue("current_mileage")
	notes := r.FormValue("notes")

	parseResult := parseVehicleNewForm(yearStr, purchaseDateStr, purchasePriceStr, purchaseMileageStr, currentMileageStr)
	formErrors := parseResult.Errors

	renderForm := func(status int) {
		w.WriteHeader(status)
		data := vehicleNewPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			Errors:          formErrors,
			Make:            vehicleMake,
			Model:           model,
			Year:            yearStr,
			VIN:             vin,
			Color:           color,
			LicensePlate:    licensePlate,
			PurchaseDate:    purchaseDateStr,
			PurchasePrice:   purchasePriceStr,
			PurchaseMileage: purchaseMileageStr,
			CurrentMileage:  currentMileageStr,
			Notes:           notes,
		}
		// WriteHeader has already been called; ignore render errors since
		// sending another response is not possible at this point.
		_ = h.engine.Render(w, "vehicles/new.html", "base", data)
	}

	if len(formErrors) > 0 {
		renderForm(http.StatusBadRequest)
		return
	}

	vehicle := &models.Vehicle{
		UserID:          account.ID,
		VIN:             strings.ToUpper(strings.TrimSpace(vin)),
		Make:            strings.TrimSpace(vehicleMake),
		Model:           strings.TrimSpace(model),
		Year:            parseResult.Year,
		Color:           color,
		LicensePlate:    licensePlate,
		PurchaseDate:    parseResult.PurchaseDate,
		PurchasePrice:   parseResult.PurchasePrice,
		PurchaseMileage: parseResult.PurchaseMileage,
		CurrentMileage:  parseResult.CurrentMileage,
		Notes:           notes,
		Status:          models.VehicleStatusActive,
	}

	if validationErrs := models.ValidateVehicleAll(vehicle); len(validationErrs) > 0 {
		// Don't expose internal field validation to the form
		delete(validationErrs, "user_id")
		delete(validationErrs, "status")
		if len(validationErrs) > 0 {
			formErrors = validationErrs
			renderForm(http.StatusBadRequest)
			return
		}
	}

	if err := h.vehicleService.CreateVehicle(r.Context(), vehicle); err != nil {
		formErrors = map[string]string{"general": "Failed to add vehicle. Please try again."}
		renderForm(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles?added=true", http.StatusSeeOther)
}

// vehicleStats holds computed statistics for a vehicle's maintenance history.
type vehicleStats struct {
	// MaintenanceCount is the total number of maintenance records for the vehicle.
	MaintenanceCount int
	// TotalMaintenanceCost is the sum of all maintenance costs.
	TotalMaintenanceCost float64
	// LastMaintenanceDate is the date of the most recent maintenance record, or nil if none.
	LastMaintenanceDate *time.Time
}

// vehicleDetailPageData holds the data passed to the vehicle detail template.
type vehicleDetailPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Vehicle is the vehicle to display.
	Vehicle *models.Vehicle
	// VehicleTitle is a short human-readable title for the page (e.g. "2020 Ford Focus").
	VehicleTitle string
	// RecentMaintenance holds the five most recent maintenance records.
	RecentMaintenance []*models.MaintenanceRecord
	// Stats holds computed statistics for the vehicle.
	Stats vehicleStats
}

const vehicleDetailRecentLimit = 5

// VehicleDetail serves the vehicle detail page (GET /vehicles/{id}).
func (h *PageHandler) VehicleDetail(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	vehicleID := vars["id"]

	vehicle, err := h.vehicleService.GetVehicle(r.Context(), vehicleID)
	if err != nil {
		var notFound *models.NotFoundError
		if errors.As(err, &notFound) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if vehicle.UserID != account.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	allMaintenance, err := h.maintenanceService.GetVehicleMaintenance(r.Context(), vehicleID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Sort by most-recent service date first.
	sort.Slice(allMaintenance, func(i, j int) bool {
		return allMaintenance[i].ServiceDate.After(allMaintenance[j].ServiceDate)
	})

	// Compute statistics.
	stats := vehicleStats{MaintenanceCount: len(allMaintenance)}
	for _, rec := range allMaintenance {
		if rec.Cost != nil {
			stats.TotalMaintenanceCost += *rec.Cost
		}
	}
	if len(allMaintenance) > 0 {
		d := allMaintenance[0].ServiceDate
		stats.LastMaintenanceDate = &d
	}

	// Limit to the most recent records for the preview.
	recent := allMaintenance
	if len(recent) > vehicleDetailRecentLimit {
		recent = recent[:vehicleDetailRecentLimit]
	}

	title := fmt.Sprintf("%d %s %s", vehicle.Year, vehicle.Make, vehicle.Model)

	data := vehicleDetailPageData{
		IsAuthenticated:   true,
		UserName:          account.Name,
		Vehicle:           vehicle,
		VehicleTitle:      title,
		RecentMaintenance: recent,
		Stats:             stats,
	}

	if err := h.engine.Render(w, "vehicles/detail.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

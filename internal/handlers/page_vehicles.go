package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	formErrors := make(map[string]string)

	if strings.TrimSpace(vehicleMake) == "" {
		formErrors["make"] = "Make is required"
	}
	if strings.TrimSpace(model) == "" {
		formErrors["model"] = "Model is required"
	}

	year := 0
	if yearStr == "" {
		formErrors["year"] = "Year is required"
	} else if y, err := strconv.Atoi(yearStr); err != nil || y < 1900 || y > 2100 {
		formErrors["year"] = "Year must be a valid year (1900-2100)"
	} else {
		year = y
	}

	var purchaseDate *time.Time
	if purchaseDateStr != "" {
		t, err := time.Parse("2006-01-02", purchaseDateStr)
		if err != nil {
			formErrors["purchase_date"] = "Invalid date format"
		} else {
			purchaseDate = &t
		}
	}

	var purchasePrice *float64
	if purchasePriceStr != "" {
		p, err := strconv.ParseFloat(purchasePriceStr, 64)
		if err != nil || p < 0 {
			formErrors["purchase_price"] = "Purchase price must be a non-negative number"
		} else {
			purchasePrice = &p
		}
	}

	var purchaseMileage *int
	if purchaseMileageStr != "" {
		m, err := strconv.Atoi(purchaseMileageStr)
		if err != nil || m < 0 {
			formErrors["purchase_mileage"] = "Mileage at purchase must be a non-negative number"
		} else {
			purchaseMileage = &m
		}
	}

	var currentMileage *int
	if currentMileageStr != "" {
		m, err := strconv.Atoi(currentMileageStr)
		if err != nil || m < 0 {
			formErrors["current_mileage"] = "Current mileage must be a non-negative number"
		} else {
			currentMileage = &m
		}
	}

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
		Year:            year,
		Color:           color,
		LicensePlate:    licensePlate,
		PurchaseDate:    purchaseDate,
		PurchasePrice:   purchasePrice,
		PurchaseMileage: purchaseMileage,
		CurrentMileage:  currentMileage,
		Notes:           notes,
		Status:          models.VehicleStatusActive,
	}

	if err := h.vehicleService.CreateVehicle(r.Context(), vehicle); err != nil {
		formErrors["general"] = "Failed to add vehicle. Please try again."
		renderForm(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/vehicles?added=true", http.StatusSeeOther)
}

package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// VehicleEdit serves the edit vehicle form page (GET /vehicles/{id}/edit).
func (h *PageHandler) VehicleEdit(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resource, ok := middleware.GetLoadedResourceFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	vehicle, ok := resource.(*models.Vehicle)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := vehicleEditPageDataFromVehicle(account, vehicle)

	if err := h.engine.Render(w, "vehicles/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// VehicleUpdate handles the edit vehicle form submission (POST /vehicles/{id}/edit).
func (h *PageHandler) VehicleUpdate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resource, ok := middleware.GetLoadedResourceFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	vehicle, ok := resource.(*models.Vehicle)
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
		data := vehicleEditPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			ActiveNav:       "vehicles",
			VehicleID:       vehicle.ID,
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
		_ = h.engine.Render(w, "vehicles/edit.html", "base", data)
	}

	if len(formErrors) > 0 {
		renderForm(http.StatusBadRequest)
		return
	}

	// Apply form values to the fetched vehicle.
	vehicle.VIN = strings.ToUpper(strings.TrimSpace(vin))
	vehicle.Make = strings.TrimSpace(vehicleMake)
	vehicle.Model = strings.TrimSpace(model)
	vehicle.Year = parseResult.Year
	vehicle.Color = color
	vehicle.LicensePlate = licensePlate
	vehicle.PurchaseDate = parseResult.PurchaseDate
	vehicle.PurchasePrice = parseResult.PurchasePrice
	vehicle.PurchaseMileage = parseResult.PurchaseMileage
	vehicle.CurrentMileage = parseResult.CurrentMileage
	vehicle.Notes = notes

	if validationErrs := models.ValidateVehicleAll(vehicle); len(validationErrs) > 0 {
		formErrors = make(map[string]string)
		for field, msg := range validationErrs {
			if field != "user_id" && field != "status" {
				formErrors[field] = msg
			}
		}
		if len(formErrors) > 0 {
			renderForm(http.StatusBadRequest)
			return
		}
	}

	if err := h.vehicleService.SaveVehicle(r.Context(), vehicle); err != nil {
		formErrors = map[string]string{"general": "Failed to update vehicle. Please try again."}
		renderForm(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/vehicles/%s?updated=true", vehicle.ID), http.StatusSeeOther)
}

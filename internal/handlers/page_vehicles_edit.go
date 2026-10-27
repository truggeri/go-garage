package handlers

import (
	"net/http"

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
	data.CSRFToken = middleware.GetCSRFToken(r.Context())

	if err := h.engine.Render(w, "vehicles/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

package handlers

import (
	"net/http"
	"time"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// profilePageData holds the data passed to the profile view template.
type profilePageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// User is the full user model for the authenticated user.
	User *models.User
	// CreatedAt is the formatted account creation date.
	CreatedAt time.Time
	// VehicleCount is the total number of vehicles belonging to the user.
	VehicleCount int
	// MaintenanceCount is the total number of maintenance records across all vehicles.
	MaintenanceCount int
}

// ViewProfile serves the user profile page (GET /profile).
func (h *PageHandler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user, err := h.userService.GetUser(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	vehicles, err := h.vehicleService.GetUserVehicles(r.Context(), account.ID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	maintenanceCount := 0
	for _, v := range vehicles {
		records, svcErr := h.maintenanceService.GetVehicleMaintenance(r.Context(), v.ID)
		if svcErr != nil {
			continue
		}
		maintenanceCount += len(records)
	}

	data := profilePageData{
		IsAuthenticated:  true,
		UserName:         account.Name,
		ActiveNav:        "profile",
		User:             user,
		CreatedAt:        user.CreatedAt,
		VehicleCount:     len(vehicles),
		MaintenanceCount: maintenanceCount,
	}

	if r.URL.Query().Get("updated") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Profile updated successfully."},
		}
	}
	if r.URL.Query().Get("password_changed") == queryTrue {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Password changed successfully."},
		}
	}

	if err := h.engine.Render(w, "profile/view.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

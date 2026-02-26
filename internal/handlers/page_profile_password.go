package handlers

import (
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

// changePasswordPageData holds the data passed to the change-password template.
type changePasswordPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
}

// ChangePassword serves the change password form page (GET /profile/password).
func (h *PageHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := changePasswordPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
	}

	if err := h.engine.Render(w, "profile/password.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ChangePasswordSubmit handles the change password form submission (POST /profile/password).
func (h *PageHandler) ChangePasswordSubmit(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	currentPassword := strings.TrimSpace(r.FormValue("current_password"))
	newPassword := strings.TrimSpace(r.FormValue("new_password"))
	confirmPassword := strings.TrimSpace(r.FormValue("confirm_password"))

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := changePasswordPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			Errors:          formErrors,
		}
		_ = h.engine.Render(w, "profile/password.html", "base", data)
	}

	if currentPassword == "" {
		renderForm(http.StatusBadRequest, map[string]string{"current_password": "Current password is required."})
		return
	}
	if newPassword == "" {
		renderForm(http.StatusBadRequest, map[string]string{"new_password": "New password is required."})
		return
	}
	if newPassword != confirmPassword {
		renderForm(http.StatusBadRequest, map[string]string{"confirm_password": "Passwords do not match."})
		return
	}

	if err := h.userService.ChangePassword(r.Context(), account.ID, currentPassword, newPassword); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to change password. Please try again."})
		return
	}

	http.Redirect(w, r, "/profile?password_changed="+queryTrue, http.StatusSeeOther)
}

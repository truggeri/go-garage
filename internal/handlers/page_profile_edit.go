package handlers

import (
	"net/http"
	"strings"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// profileEditPageData holds the data passed to the edit-profile template.
type profileEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// Errors holds field-level and general validation error messages.
	Errors map[string]string
	// Form field values for repopulating the form after a failed submission.
	Username  string
	Email     string
	FirstName string
	LastName  string
}

// ProfileEdit serves the edit profile form page (GET /profile/edit).
func (h *PageHandler) ProfileEdit(w http.ResponseWriter, r *http.Request) {
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

	data := profileEditPageData{
		IsAuthenticated: true,
		UserName:        account.Name,
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
	}

	if err := h.engine.Render(w, "profile/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// ProfileUpdate handles the edit profile form submission (POST /profile/edit).
func (h *PageHandler) ProfileUpdate(w http.ResponseWriter, r *http.Request) {
	account, ok := middleware.GetAccountFromContext(r.Context())
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	firstName := strings.TrimSpace(r.FormValue("first_name"))
	lastName := strings.TrimSpace(r.FormValue("last_name"))

	renderForm := func(status int, formErrors map[string]string) {
		w.WriteHeader(status)
		data := profileEditPageData{
			IsAuthenticated: true,
			UserName:        account.Name,
			Errors:          formErrors,
			Username:        username,
			Email:           email,
			FirstName:       firstName,
			LastName:        lastName,
		}
		_ = h.engine.Render(w, "profile/edit.html", "base", data)
	}

	// Validate using a temporary User model.
	candidate := &models.User{Username: username, Email: email}
	if err := models.ValidateUser(candidate); err != nil {
		var ve *models.ValidationError
		if models.IsValidationError(err, &ve) {
			renderForm(http.StatusBadRequest, map[string]string{ve.Field: ve.Message})
			return
		}
		renderForm(http.StatusBadRequest, map[string]string{"general": "Invalid form data."})
		return
	}

	updates := services.UserUpdates{
		Username:  &username,
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
	}

	if _, err := h.userService.UpdateUser(r.Context(), account.ID, updates); err != nil {
		var de *models.DuplicateError
		if models.IsDuplicateError(err, &de) {
			renderForm(http.StatusConflict, map[string]string{de.Field: de.Field + " is already taken"})
			return
		}
		renderForm(http.StatusInternalServerError, map[string]string{"general": "Failed to update profile. Please try again."})
		return
	}

	http.Redirect(w, r, "/profile?updated="+queryTrue, http.StatusSeeOther)
}

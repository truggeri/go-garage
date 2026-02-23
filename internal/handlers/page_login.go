package handlers

import (
	"net/http"
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// loginPageData holds the data passed to the login template.
type loginPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash      interface{}
	Errors     map[string]string
	Identifier string
}

// LoginForm serves the login page (GET /login).
func (h *PageHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	data := loginPageData{}

	if r.URL.Query().Get("registered") == "true" {
		data.Flash = []flashMessage{
			{Type: "success", Message: "Account created successfully. Please log in."},
		}
	}

	if err := h.engine.Render(w, "login.html", "auth", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// LoginSubmit handles the login form submission (POST /login).
func (h *PageHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	identifier := r.FormValue("identifier")
	password := r.FormValue("password")

	errors := make(map[string]string)

	if identifier == "" {
		errors["identifier"] = "Email or username is required"
	}
	if password == "" {
		errors["password"] = "Password is required"
	}

	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		data := loginPageData{
			Errors:     errors,
			Identifier: identifier,
		}
		if renderErr := h.engine.Render(w, "login.html", "auth", data); renderErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	result, err := h.authService.Authenticate(r.Context(), identifier, password)
	if err != nil {
		var validationErr *models.ValidationError
		if models.IsValidationError(err, &validationErr) {
			errors["general"] = "Invalid email/username or password"
		} else {
			errors["general"] = "An unexpected error occurred. Please try again."
		}

		w.WriteHeader(http.StatusUnauthorized)
		data := loginPageData{
			Errors:     errors,
			Identifier: identifier,
		}
		if renderErr := h.engine.Render(w, "login.html", "auth", data); renderErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	accessMaxAge := int(time.Until(time.Unix(result.AccessExpiresAt, 0)).Seconds())
	refreshMaxAge := int(time.Until(time.Unix(result.RefreshExpiresAt, 0)).Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    result.AccessToken,
		Path:     "/",
		MaxAge:   accessMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		Path:     "/",
		MaxAge:   refreshMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

// PageHandler serves HTML pages for the web interface.
type PageHandler struct {
	engine      *templateengine.Engine
	authService services.AuthenticationService
}

// NewPageHandler creates a new PageHandler with the given template engine and auth service.
func NewPageHandler(engine *templateengine.Engine, authService services.AuthenticationService) *PageHandler {
	return &PageHandler{
		engine:      engine,
		authService: authService,
	}
}

// registerPageData holds the data passed to the registration template.
type registerPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash     interface{}
	Errors    map[string]string
	Username  string
	Email     string
	FirstName string
	LastName  string
}

// Home serves the home page.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"IsAuthenticated": false,
	}
	if err := h.engine.Render(w, "home.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RegisterForm serves the registration page (GET /register).
func (h *PageHandler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	data := registerPageData{}
	if err := h.engine.Render(w, "register.html", "auth", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RegisterSubmit handles the registration form submission (POST /register).
func (h *PageHandler) RegisterSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")

	errors := make(map[string]string)

	if username == "" {
		errors["username"] = "Username is required"
	}
	if email == "" {
		errors["email"] = "Email is required"
	}
	if password == "" {
		errors["password"] = "Password is required"
	}
	if confirmPassword == "" {
		errors["confirm_password"] = "Please confirm your password"
	} else if password != confirmPassword {
		errors["confirm_password"] = "Passwords do not match"
	}

	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		data := registerPageData{
			Errors:    errors,
			Username:  username,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}
		if renderErr := h.engine.Render(w, "register.html", "auth", data); renderErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	registration := services.RegistrationRequest{
		Username:  username,
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	_, err := h.authService.Register(r.Context(), registration)
	if err != nil {
		var validationErr *models.ValidationError
		if models.IsValidationError(err, &validationErr) {
			errors[validationErr.Field] = validationErr.Message
		}

		var duplicateErr *models.DuplicateError
		if models.IsDuplicateError(err, &duplicateErr) {
			errors[duplicateErr.Field] = duplicateErr.Error()
		}

		if len(errors) == 0 {
			errors["general"] = "An unexpected error occurred. Please try again."
		}

		w.WriteHeader(http.StatusBadRequest)
		data := registerPageData{
			Errors:    errors,
			Username:  username,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}
		if renderErr := h.engine.Render(w, "register.html", "auth", data); renderErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
}

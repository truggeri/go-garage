package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
)

// profileEditPageData holds the data passed to the edit-profile template.
type profileEditPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// CSRFToken is the CSRF protection token exposed to client-side scripts.
	CSRFToken string
	// Form field values for pre-populating the form.
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
		ActiveNav:       "profile",
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
	}

	if err := h.engine.Render(w, "profile/edit.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

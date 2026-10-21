package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
)

// changePasswordPageData holds the data passed to the change-password template.
type changePasswordPageData struct {
	// Flash holds optional flash messages rendered by the flash-messages partial template.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in (always true on this page).
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item for highlighting.
	ActiveNav string
	// CSRFToken is the CSRF protection token to embed in the layout.
	CSRFToken string
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
		ActiveNav:       "profile",
		CSRFToken:       middleware.GetCSRFToken(r.Context()),
	}

	if err := h.engine.Render(w, "profile/password.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

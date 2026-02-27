package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
)

// errorPageData holds the data passed to error page templates.
type errorPageData struct {
	// Flash holds optional flash messages for the flash-messages partial.
	Flash interface{}
	// IsAuthenticated indicates whether the user is logged in.
	IsAuthenticated bool
	// UserName is the display name of the authenticated user.
	UserName string
	// ActiveNav identifies the active navigation item (empty for error pages).
	ActiveNav string
}

// buildErrorPageData constructs template data for error pages, detecting authentication from the request context.
func (h *PageHandler) buildErrorPageData(r *http.Request) errorPageData {
	data := errorPageData{}
	if account, ok := middleware.GetAccountFromContext(r.Context()); ok {
		data.IsAuthenticated = true
		data.UserName = account.Name
	}
	return data
}

// NotFound renders the 404 Not Found error page.
func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := h.buildErrorPageData(r)
	if err := h.engine.Render(w, "errors/404.html", "base", data); err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

// Forbidden renders the 403 Forbidden error page.
func (h *PageHandler) Forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	data := h.buildErrorPageData(r)
	if err := h.engine.Render(w, "errors/403.html", "base", data); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

// ServerError renders the 500 Internal Server Error page.
func (h *PageHandler) ServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	data := h.buildErrorPageData(r)
	if err := h.engine.Render(w, "errors/500.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// RenderError renders the appropriate error page based on the HTTP status code.
// It can be passed as a middleware.PageErrorHandler.
func (h *PageHandler) RenderError(w http.ResponseWriter, r *http.Request, code int) {
	switch code {
	case http.StatusForbidden:
		h.Forbidden(w, r)
	case http.StatusNotFound:
		h.NotFound(w, r)
	default:
		h.ServerError(w, r)
	}
}

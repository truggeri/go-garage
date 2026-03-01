package handlers

import "net/http"

// Home serves the home page.
// Logged-in users (those with an access_token cookie) are redirected to the dashboard.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("access_token"); err == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	data := map[string]interface{}{
		"IsAuthenticated": false,
		"ActiveNav":       "",
	}
	if err := h.engine.Render(w, "home.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

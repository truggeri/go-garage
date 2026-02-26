package handlers

import "net/http"

// Home serves the home page.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"IsAuthenticated": false,
		"ActiveNav":       "",
	}
	if err := h.engine.Render(w, "home.html", "base", data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

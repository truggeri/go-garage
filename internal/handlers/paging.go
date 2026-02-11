package handlers

import (
	"net/http"
	"strconv"
)

// extractPaging parses page and limit query parameters from the request.
// Returns page (default 1) and limit (default 20, max 100).
func extractPaging(r *http.Request) (int, int) {
	q := r.URL.Query()
	pg, sz := 1, 20
	if v := q.Get("page"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			pg = n
		}
	}
	if v := q.Get("limit"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 100 {
			sz = n
		}
	}
	return pg, sz
}

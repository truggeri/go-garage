package handlers

import (
	"encoding/json"
	"net/http"
)

// respondWithProblem writes a problem/error JSON response.
func respondWithProblem(w http.ResponseWriter, code int, errCode, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]string{"code": errCode, "message": msg},
	})
}

// respondWithValidationProblems writes a validation error JSON response with field details.
func respondWithValidationProblems(w http.ResponseWriter, msg string, details []FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false, "error": map[string]interface{}{"code": "VALIDATION_ERROR", "message": msg, "details": details},
	})
}

// respondWithPayload writes a success JSON response with the given payload.
func respondWithPayload(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthenticationService
}

// BuildAuthHandler creates a new AuthHandler
func BuildAuthHandler(authService services.AuthenticationService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents the JSON body for user registration
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents the JSON body for user login
type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// RefreshRequest represents the JSON body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the JSON response for authentication operations
type AuthResponse struct {
	Success bool       `json:"success"`
	Data    *AuthData  `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
	Message string     `json:"message,omitempty"`
}

// AuthData contains the authentication data in the response
type AuthData struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
	AccountID        string `json:"account_id"`
	AccountName      string `json:"account_name"`
}

// ErrorInfo contains error details in the response
type ErrorInfo struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError contains field-level error details
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// HandleRegister handles POST /api/v1/auth/register
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var reqBody RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeErrorResponse(w, "INVALID_REQUEST", "Invalid JSON body", http.StatusBadRequest, nil)
		return
	}

	if reqBody.Username == "" || reqBody.Email == "" || reqBody.Password == "" {
		writeErrorResponse(w, "VALIDATION_ERROR", "Username, email, and password are required", http.StatusBadRequest, []FieldError{
			{Field: "username", Message: "required"},
			{Field: "email", Message: "required"},
			{Field: "password", Message: "required"},
		})
		return
	}

	registration := services.RegistrationRequest{
		Username:  reqBody.Username,
		Email:     reqBody.Email,
		Password:  reqBody.Password,
		FirstName: reqBody.FirstName,
		LastName:  reqBody.LastName,
	}

	result, err := h.authService.Register(r.Context(), registration)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeAuthSuccess(w, result, "Registration successful", http.StatusCreated)
}

// HandleLogin handles POST /api/v1/auth/login
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var reqBody LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeErrorResponse(w, "INVALID_REQUEST", "Invalid JSON body", http.StatusBadRequest, nil)
		return
	}

	if reqBody.Identifier == "" || reqBody.Password == "" {
		writeErrorResponse(w, "VALIDATION_ERROR", "Identifier and password are required", http.StatusBadRequest, nil)
		return
	}

	result, err := h.authService.Authenticate(r.Context(), reqBody.Identifier, reqBody.Password)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeAuthSuccess(w, result, "Login successful", http.StatusOK)
}

// HandleRefresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var reqBody RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeErrorResponse(w, "INVALID_REQUEST", "Invalid JSON body", http.StatusBadRequest, nil)
		return
	}

	if reqBody.RefreshToken == "" {
		writeErrorResponse(w, "VALIDATION_ERROR", "Refresh token is required", http.StatusBadRequest, nil)
		return
	}

	result, err := h.authService.RefreshSession(r.Context(), reqBody.RefreshToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	writeAuthSuccess(w, result, "Token refreshed successfully", http.StatusOK)
}

// HandleLogout handles POST /api/v1/auth/logout
// Note: With stateless JWT, logout is typically handled client-side by deleting the token
// This endpoint acknowledges the logout request
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "Logout successful. Please delete your tokens client-side.",
	})
}

// handleAuthError converts service errors to HTTP responses
func handleAuthError(w http.ResponseWriter, err error) {
	var validationErr *models.ValidationError
	if models.IsValidationError(err, &validationErr) {
		details := []FieldError{}
		if validationErr.Field != "" {
			details = append(details, FieldError{
				Field:   validationErr.Field,
				Message: validationErr.Message,
			})
		}
		writeErrorResponse(w, "VALIDATION_ERROR", validationErr.Message, http.StatusBadRequest, details)
		return
	}

	var duplicateErr *models.DuplicateError
	if models.IsDuplicateError(err, &duplicateErr) {
		writeErrorResponse(w, "DUPLICATE_ERROR", duplicateErr.Error(), http.StatusConflict, nil)
		return
	}

	writeErrorResponse(w, "INTERNAL_ERROR", "An unexpected error occurred", http.StatusInternalServerError, nil)
}

// writeAuthSuccess writes a successful authentication response
func writeAuthSuccess(w http.ResponseWriter, result *services.AuthenticationResult, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: message,
		Data: &AuthData{
			AccessToken:      result.AccessToken,
			RefreshToken:     result.RefreshToken,
			AccessExpiresAt:  result.AccessExpiresAt,
			RefreshExpiresAt: result.RefreshExpiresAt,
			AccountID:        result.AccountID,
			AccountName:      result.AccountName,
		},
	})
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, code, message string, statusCode int, details []FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

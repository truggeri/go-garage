package handlers

import (
	"net/http"

	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/services"
)

type userAPIHandler struct {
	svc services.UserService
}

func MakeUserAPIHandler(svc services.UserService) *userAPIHandler {
	return &userAPIHandler{svc: svc}
}

// GetMe handles GET /api/v1/users/me
func (h *userAPIHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	user, err := h.svc.GetUser(ctx, caller.ID)
	if err != nil {
		handleDomainError(w, err)
		return
	}

	respondWithPayload(w, 200, buildUserProfilePayload(user))
}

// UpdateMe handles PUT /api/v1/users/me
func (h *userAPIHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	updates := extractUserUpdates(inputData)
	updatedUser, updateErr := h.svc.UpdateUser(ctx, caller.ID, updates)
	if updateErr != nil {
		handleDomainError(w, updateErr)
		return
	}

	respondWithPayload(w, 200, buildUserProfilePayload(updatedUser, "Profile updated successfully"))
}

// ChangePassword handles PUT /api/v1/users/me/password
func (h *userAPIHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "current_password", "new_password")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	currentPassword, _ := inputData["current_password"].(string)
	newPassword, _ := inputData["new_password"].(string)

	if changeErr := h.svc.ChangePassword(ctx, caller.ID, currentPassword, newPassword); changeErr != nil {
		handleDomainError(w, changeErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	})
}

// DeleteMe handles DELETE /api/v1/users/me
func (h *userAPIHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	caller, authOK := middleware.GetAccountFromContext(ctx)
	if !authOK {
		respondWithProblem(w, 401, "AUTHENTICATION_ERROR", "Not authenticated")
		return
	}

	inputData, parseErr := parseJSONBody(r)
	if parseErr != nil {
		respondWithProblem(w, 400, "INVALID_REQUEST", "Bad JSON")
		return
	}

	valErrs := validateRequiredKeys(inputData, "password")
	if len(valErrs) > 0 {
		respondWithValidationProblems(w, "Missing fields", valErrs)
		return
	}

	password, _ := inputData["password"].(string)

	if deleteErr := h.svc.DeleteUser(ctx, caller.ID, password); deleteErr != nil {
		handleDomainError(w, deleteErr)
		return
	}

	respondWithPayload(w, 200, map[string]interface{}{
		"success": true,
		"message": "Account deleted successfully",
	})
}

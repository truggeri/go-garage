package handlers

import (
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

// buildUserProfilePayload creates a response payload for a user profile
func buildUserProfilePayload(user *models.User, message ...string) map[string]interface{} {
	payload := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	}
	
	// Add last_login_at if it exists
	if user.LastLoginAt != nil {
		payload["data"].(map[string]interface{})["last_login_at"] = user.LastLoginAt
	} else {
		payload["data"].(map[string]interface{})["last_login_at"] = nil
	}
	
	// Add optional message
	if len(message) > 0 && message[0] != "" {
		payload["message"] = message[0]
	}
	
	return payload
}

// extractUserUpdates extracts user update fields from the input data map
func extractUserUpdates(d map[string]interface{}) services.UserUpdates {
	var u services.UserUpdates
	
	if v, ok := d["username"].(string); ok && v != "" {
		u.Username = &v
	}
	if v, ok := d["email"].(string); ok && v != "" {
		u.Email = &v
	}
	if v, ok := d["first_name"].(string); ok && v != "" {
		u.FirstName = &v
	}
	if v, ok := d["last_name"].(string); ok && v != "" {
		u.LastName = &v
	}
	
	return u
}

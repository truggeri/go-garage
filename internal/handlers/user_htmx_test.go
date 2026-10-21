package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/services"
)

func TestUserHandler_UpdateMe_HTMXCookieCSRF(t *testing.T) {
	user := &models.User{ID: "u1", Username: "johndoe", Email: "john@example.com"}
	stub := &stubUserSvc{getResult: user, updateResult: user}
	page := newTestProfileEditPageHandler(t, stub)
	api := MakeUserAPIHandler(stub)
	router := mux.NewRouter()
	router.Handle("/profile/edit", middleware.CSRFProtection("secret")(http.HandlerFunc(page.ProfileEdit))).Methods(http.MethodGet)
	router.HandleFunc("/profile", page.ViewProfile).Methods(http.MethodGet)
	router.Handle("/api/v1/users/me", middleware.APICSRFProtection("secret")(http.HandlerFunc(api.UpdateMe))).Methods(http.MethodPut)
	authed := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = addAuthContext(r, "u1", "johndoe")
		r = r.WithContext(context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie))
		router.ServeHTTP(w, r)
	})
	get := httptest.NewRecorder()
	authed.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/profile/edit", nil))
	require.Equal(t, http.StatusOK, get.Code)
	token := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
	require.Len(t, token, 2)

	put := func(csrf string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", strings.NewReader(
			`{"username":" janedoe ","email":" jane@example.com ","first_name":" Jane ","last_name":""}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HX-Request", "true")
		req.Header.Set("X-CSRF-Token", csrf)
		rec := httptest.NewRecorder()
		authed.ServeHTTP(rec, req)
		return rec
	}
	for _, csrf := range []string{"", "invalid"} {
		assert.Equal(t, http.StatusForbidden, put(csrf).Code)
		assert.Nil(t, stub.updates)
	}
	success := put(token[1])
	require.Equal(t, http.StatusOK, success.Code)
	assert.Equal(t, "/profile?updated=true", success.Header().Get("HX-Redirect"))
	require.NotNil(t, stub.updates)
	assert.Equal(t, "janedoe", *stub.updates.Username)
	assert.Equal(t, "jane@example.com", *stub.updates.Email)
	assert.Equal(t, "Jane", *stub.updates.FirstName)
	require.NotNil(t, stub.updates.LastName)
	assert.Empty(t, *stub.updates.LastName)
	assert.Contains(t, success.Body.String(), `"success":true`)

	profile := httptest.NewRecorder()
	authed.ServeHTTP(profile, httptest.NewRequest(http.MethodGet, success.Header().Get("HX-Redirect"), nil))
	assert.Equal(t, http.StatusOK, profile.Code)
	assert.Contains(t, profile.Body.String(), "Profile updated successfully.")
}

func TestUserHandler_UpdateMe_HTMXErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		err    error
		status int
		field  string
	}{
		{"missing username", models.NewValidationError("username", "username is required"), http.StatusBadRequest, "username"},
		{"invalid email", models.NewValidationError("email", "invalid email"), http.StatusBadRequest, "email"},
		{"duplicate username", models.NewDuplicateError("User", "username", "taken"), http.StatusConflict, "username"},
		{"duplicate email", models.NewDuplicateError("User", "email", "taken@example.com"), http.StatusConflict, "email"},
		{"database failure", models.NewDatabaseError("update user", assert.AnError), http.StatusInternalServerError, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &stubUserSvc{updateErr: test.err}
			req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me", strings.NewReader(`{"username":"taken","email":"taken@example.com"}`))
			req = addAuthContext(req, "u1", "johndoe")
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()
			MakeUserAPIHandler(stub).UpdateMe(rec, req)

			assert.Equal(t, test.status, rec.Code)
			assert.Empty(t, rec.Header().Get("HX-Redirect"))
			var payload struct {
				Success bool `json:"success"`
				Error   struct {
					Message string       `json:"message"`
					Details []FieldError `json:"details"`
				} `json:"error"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
			assert.False(t, payload.Success)
			assert.NotEmpty(t, payload.Error.Message)
			if test.field != "" {
				require.Len(t, payload.Error.Details, 1)
				assert.Equal(t, test.field, payload.Error.Details[0].Field)
				assert.NotEmpty(t, payload.Error.Details[0].Message)
			}
		})
	}
}

func TestExtractUserUpdates_EmptyAndOmittedFields(t *testing.T) {
	assert.Equal(t, services.UserUpdates{}, extractUserUpdates(map[string]interface{}{}))
	updates := extractUserUpdates(map[string]interface{}{
		"username": " ", "email": "", "first_name": "", "last_name": " ",
	})
	for _, field := range []*string{updates.Username, updates.Email, updates.FirstName, updates.LastName} {
		require.NotNil(t, field)
		assert.Empty(t, *field)
	}
}

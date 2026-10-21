package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/auth"
	"github.com/truggeri/go-garage/internal/database"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/repositories"
	"github.com/truggeri/go-garage/internal/services"
	"github.com/truggeri/go-garage/internal/templateengine"
)

func TestUserHandler_ChangePassword_HTMXCookieSession(t *testing.T) {
	ctx := context.Background()
	db, err := database.InitializeGarage(filepath.Join(t.TempDir(), "garage.db"), database.StandardWorkerPoolSettings())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Terminate()) })
	require.NoError(t, database.BootstrapSchema(ctx, db, "../../migrations"))
	repo := repositories.NewSQLiteUserRepository(db.RawSQLConnection())
	userService := services.NewUserService(repo)
	tokens, err := auth.BuildTokenManager("test-secret", auth.StandardTokenDurations())
	require.NoError(t, err)
	authService := services.BuildAuthService(repo, tokens)
	session, err := authService.Register(ctx, services.RegistrationRequest{
		Username: "johndoe", Email: "john@example.com", Password: "OldPass123",
	})
	require.NoError(t, err)
	page := NewPageHandler(templateengine.NewEngine("../../web/templates", true),
		authService, &stubVehicleSvc{}, &stubMaintenanceSvc{}, nil, userService, nil)
	api := MakeUserAPIHandler(userService)
	router := mux.NewRouter()
	pages := router.NewRoute().Subrouter()
	pages.Use(middleware.CookieAuthGuard(tokens), middleware.CSRFProtection("csrf-secret"))
	pages.HandleFunc("/profile/password", page.ChangePassword).Methods(http.MethodGet)
	pages.HandleFunc("/profile", page.ViewProfile).Methods(http.MethodGet)
	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.HybridAuthGuard(tokens), middleware.APICSRFProtection("csrf-secret"))
	protected.HandleFunc("/users/me/password", api.ChangePassword).Methods(http.MethodPut)

	request := func(method, path, body, csrf, access string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.AddCookie(&http.Cookie{Name: "access_token", Value: access})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: session.RefreshToken})
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HX-Request", "true")
		req.Header.Set("X-CSRF-Token", csrf)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	get := request(http.MethodGet, "/profile/password", "", "", session.AccessToken)
	require.Equal(t, http.StatusOK, get.Code)
	csrf := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
	require.Len(t, csrf, 2)
	body := `{"current_password":"OldPass123","new_password":" NewPass456 ","confirm_password":" NewPass456 "}`
	for _, token := range []string{"", "invalid"} {
		assert.Equal(t, http.StatusForbidden, request(http.MethodPut, "/api/v1/users/me/password", body, token, session.AccessToken).Code)
	}
	for _, test := range []struct {
		current, password, confirmation, field string
	}{
		{"WrongPass123", "NewPass456", "NewPass456", "current_password"},
		{"OldPass123", "NewPass456", "Different456", "confirm_password"},
		{"OldPass123", "weakpass", "weakpass", "new_password"},
	} {
		data, err := json.Marshal(map[string]string{
			"current_password": test.current, "new_password": test.password, "confirm_password": test.confirmation,
		})
		require.NoError(t, err)
		rec := request(http.MethodPut, "/api/v1/users/me/password", string(data), csrf[1], session.AccessToken)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), `"field":"`+test.field+`"`)
		assert.Empty(t, rec.Header().Get("HX-Redirect"))
		for _, password := range []string{test.current, test.password, test.confirmation} {
			assert.NotContains(t, rec.Body.String(), password)
		}
	}
	success := request(http.MethodPut, "/api/v1/users/me/password", body, csrf[1], session.AccessToken)
	require.Equal(t, http.StatusOK, success.Code)
	assert.Equal(t, "/profile?password_changed=true", success.Header().Get("HX-Redirect"))
	assert.Empty(t, success.Result().Cookies())
	assert.NotContains(t, success.Body.String(), "NewPass456")

	profile := request(http.MethodGet, success.Header().Get("HX-Redirect"), "", "", session.AccessToken)
	assert.Equal(t, http.StatusOK, profile.Code)
	assert.Contains(t, profile.Body.String(), "Password changed successfully.")
	refreshed := request(http.MethodGet, "/profile", "", "", "")
	assert.Equal(t, http.StatusOK, refreshed.Code)
	assert.NotEmpty(t, refreshed.Result().Cookies())
	_, err = authService.Authenticate(ctx, "johndoe", "OldPass123")
	require.Error(t, err)
	_, err = authService.Authenticate(ctx, "johndoe", " NewPass456 ")
	require.NoError(t, err)
}

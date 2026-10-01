package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

func TestVehicleDelete_HTMXCookieCSRF(t *testing.T) {
	vehicleSvc := &stubVehicleSvc{
		getResult:   &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020},
		countResult: 1,
		listResult:  []*models.Vehicle{{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}},
	}
	page := newTestVehicleListPageHandler(t, vehicleSvc)
	api := MakeVehicleAPIHandler(vehicleSvc)
	router := mux.NewRouter()
	router.Handle("/vehicles", middleware.CSRFProtection("test-secret")(http.HandlerFunc(page.VehicleList))).Methods(http.MethodGet)
	router.Handle("/api/v1/vehicles/{id}", middleware.APICSRFProtection("test-secret")(http.HandlerFunc(api.RemoveOne))).Methods(http.MethodDelete)
	authedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = addAuthContext(r, "u1", "testuser")
		r = r.WithContext(context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie))
		router.ServeHTTP(w, r)
	})

	get := httptest.NewRecorder()
	authedRouter.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/vehicles", nil))
	require.Equal(t, http.StatusOK, get.Code)
	assert.Contains(t, get.Body.String(), `hx-delete="/api/v1/vehicles/v1"`)
	token := regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`).FindStringSubmatch(get.Body.String())
	require.Len(t, token, 2)

	deleteRequest := func(csrfToken string) *http.Request {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil)
		req.Header.Set("HX-Request", "true")
		if csrfToken != "" {
			req.Header.Set("X-CSRF-Token", csrfToken)
		}
		return req
	}

	missing := httptest.NewRecorder()
	authedRouter.ServeHTTP(missing, deleteRequest(""))
	assert.Equal(t, http.StatusForbidden, missing.Code)
	assert.Contains(t, missing.Body.String(), "CSRF_ERROR")
	assert.Empty(t, vehicleSvc.deletedID)

	invalid := httptest.NewRecorder()
	authedRouter.ServeHTTP(invalid, deleteRequest("invalid"))
	assert.Equal(t, http.StatusForbidden, invalid.Code)
	assert.Empty(t, vehicleSvc.deletedID)

	success := httptest.NewRecorder()
	authedRouter.ServeHTTP(success, deleteRequest(token[1]))
	assert.Equal(t, http.StatusOK, success.Code)
	assert.Equal(t, "/vehicles", success.Header().Get("HX-Redirect"))
	assert.Equal(t, "v1", vehicleSvc.deletedID)
}

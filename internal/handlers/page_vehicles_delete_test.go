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

var deleteCSRFToken = regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`)

func TestVehicleDeleteHTMX_CSRF(t *testing.T) {
	vehicle := &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}
	vehicleSvc := &stubVehicleSvc{getResult: vehicle, countResult: 1, listResult: []*models.Vehicle{vehicle}}
	pages := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})

	router := mux.NewRouter()
	pageRoutes := router.NewRoute().Subrouter()
	pageRoutes.Use(middleware.CSRFProtection("test-secret"))
	pageRoutes.HandleFunc("/vehicles", pages.VehicleList).Methods(http.MethodGet)
	pageRoutes.HandleFunc("/vehicles/{id}", func(w http.ResponseWriter, r *http.Request) {
		pages.VehicleDetail(w, addResourceContext(r, vehicle))
	}).Methods(http.MethodGet)
	apiRoutes := router.PathPrefix("/api/v1").Subrouter()
	apiRoutes.Use(middleware.APICSRFProtection("test-secret"))
	apiRoutes.HandleFunc("/vehicles/{id}", MakeVehicleAPIHandler(vehicleSvc).RemoveOne).Methods(http.MethodDelete)
	authedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.AuthMethodContextKey, middleware.AuthMethodCookie)
		router.ServeHTTP(w, addAuthContext(r.WithContext(ctx), "u1", "testuser"))
	})

	for _, path := range []string{"/vehicles", "/vehicles/v1"} {
		t.Run(path, func(t *testing.T) {
			get := httptest.NewRecorder()
			authedRouter.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, get.Code)
			token := deleteCSRFToken.FindStringSubmatch(get.Body.String())
			require.Len(t, token, 2)
			assert.Contains(t, get.Body.String(), `hx-delete="/api/v1/vehicles/v1"`)
			assert.Contains(t, get.Body.String(), `hx-confirm="Are you sure you want to delete this vehicle? This action cannot be undone."`)
			assert.Contains(t, get.Body.String(), `hx-swap="none"`)
			assert.NotContains(t, get.Body.String(), `data-confirm-delete=`)
			assert.NotContains(t, get.Body.String(), `id="delete-form"`)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil)
			req.Header.Set("HX-Request", "true")
			req.Header.Set("X-CSRF-Token", token[1])
			rec := httptest.NewRecorder()
			authedRouter.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "/vehicles", rec.Header().Get("HX-Redirect"))
			assert.Equal(t, "v1", vehicleSvc.deletedID)

			vehicleSvc.deletedID = ""
			rec = httptest.NewRecorder()
			authedRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil))
			assert.Equal(t, http.StatusForbidden, rec.Code)
			assert.Empty(t, vehicleSvc.deletedID)
		})
	}
}

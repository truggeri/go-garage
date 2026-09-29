package handlers

import (
	"context"
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
)

var csrfMetaToken = regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)">`)

func TestVehicleDeleteHTMX(t *testing.T) {
	vehicle := &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}
	vehicleSvc := &stubVehicleSvc{
		getResult: vehicle, countResult: 1, listResult: []*models.Vehicle{vehicle},
	}
	pageHandler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
	apiHandler := MakeVehicleAPIHandler(vehicleSvc)
	router := mux.NewRouter()
	pages := router.NewRoute().Subrouter()
	pages.Use(middleware.CSRFProtection("test-secret"))
	pages.HandleFunc("/vehicles", pageHandler.VehicleList).Methods(http.MethodGet)
	pages.HandleFunc("/vehicles/{id}", func(w http.ResponseWriter, r *http.Request) {
		pageHandler.VehicleDetail(w, addResourceContext(r, vehicle))
	}).Methods(http.MethodGet)
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.APICSRFProtection("test-secret"))
	api.HandleFunc("/vehicles/{id}", apiHandler.RemoveOne).Methods(http.MethodDelete)
	authedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.AccountContextKey, &middleware.AccountInfo{ID: "u1", Name: "testuser"})
		ctx = context.WithValue(ctx, middleware.AuthMethodContextKey, middleware.AuthMethodCookie)
		router.ServeHTTP(w, r.WithContext(ctx))
	})

	for _, page := range []string{"/vehicles", "/vehicles/v1"} {
		t.Run(page, func(t *testing.T) {
			get := httptest.NewRecorder()
			authedRouter.ServeHTTP(get, httptest.NewRequest(http.MethodGet, page, nil))
			require.Equal(t, http.StatusOK, get.Code)
			body := get.Body.String()
			require.Contains(t, body, `hx-delete="/api/v1/vehicles/v1"`)
			assert.Contains(t, body, `hx-confirm="Are you sure you want to delete this vehicle? This action cannot be undone."`)
			assert.Contains(t, body, `hx-swap="none"`)
			assert.NotContains(t, body, `data-confirm-delete=`)
			assert.NotContains(t, body, `action="/vehicles/v1/delete"`)

			matches := csrfMetaToken.FindStringSubmatch(body)
			require.Len(t, matches, 2)

			request := func(token string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil)
				req.Header.Set("HX-Request", "true")
				if token != "" {
					req.Header.Set("X-CSRF-Token", token)
				}
				rec := httptest.NewRecorder()
				authedRouter.ServeHTTP(rec, req)
				return rec
			}

			vehicleSvc.deletedID = ""
			denied := request("")
			assert.Equal(t, http.StatusForbidden, denied.Code)
			assert.Contains(t, denied.Body.String(), `"code":"CSRF_ERROR"`)
			assert.Empty(t, denied.Header().Get("HX-Redirect"))
			assert.Empty(t, vehicleSvc.deletedID)

			success := request(matches[1])
			assert.Equal(t, http.StatusOK, success.Code)
			assert.Equal(t, "/vehicles", success.Header().Get("HX-Redirect"))
			assert.Equal(t, "v1", vehicleSvc.deletedID)
		})
	}
}

func TestVehicleDeleteHTMX_Failure(t *testing.T) {
	vehicleSvc := &stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "u1"}, deleteErr: assert.AnError}
	req := mux.SetURLVars(addAuthContext(httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil), "u1", "testuser"), map[string]string{"id": "v1"})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	MakeVehicleAPIHandler(vehicleSvc).RemoveOne(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, rec.Header().Get("HX-Redirect"))
	assert.True(t, strings.Contains(rec.Body.String(), `"success":false`))
}

func TestVehicleDeleteHTMX_OtherOwner(t *testing.T) {
	vehicleSvc := &stubVehicleSvc{getResult: &models.Vehicle{ID: "v1", UserID: "other"}}
	req := mux.SetURLVars(addAuthContext(httptest.NewRequest(http.MethodDelete, "/api/v1/vehicles/v1", nil), "u1", "testuser"), map[string]string{"id": "v1"})
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()

	MakeVehicleAPIHandler(vehicleSvc).RemoveOne(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Header().Get("HX-Redirect"))
	assert.Empty(t, vehicleSvc.deletedID)
}

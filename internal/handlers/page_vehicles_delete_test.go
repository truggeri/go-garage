package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/truggeri/go-garage/internal/middleware"
	"github.com/truggeri/go-garage/internal/models"
)

var deleteFormToken = regexp.MustCompile(`<form id="delete-form"[^>]*>\s*<input type="hidden" name="csrf_token" value="([^"]+)">`)

func TestPageHandler_VehicleDelete(t *testing.T) {
	vehicle := &models.Vehicle{ID: "v1", UserID: "u1"}

	t.Run("deletes owned vehicle and redirects", func(t *testing.T) {
		vehicleSvc := &stubVehicleSvc{}
		handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
		req := addResourceContext(addAuthContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), "u1", "testuser"), vehicle)
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusSeeOther, rec.Code)
		assert.Equal(t, "/vehicles", rec.Header().Get("Location"))
		assert.Equal(t, "v1", vehicleSvc.deletedID)
	})

	t.Run("rejects another user's vehicle", func(t *testing.T) {
		vehicleSvc := &stubVehicleSvc{}
		handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
		req := addResourceContext(addAuthContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), "other", "testuser"), vehicle)
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Empty(t, vehicleSvc.deletedID)
	})

	t.Run("requires a loaded vehicle", func(t *testing.T) {
		handler := newTestVehicleDetailPageHandler(t, &stubVehicleSvc{}, &stubMaintenanceSvc{})
		req := addAuthContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), "u1", "testuser")
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("requires an account", func(t *testing.T) {
		vehicleSvc := &stubVehicleSvc{}
		handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
		req := addResourceContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), vehicle)
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Empty(t, vehicleSvc.deletedID)
	})

	t.Run("rejects a loaded resource that is not a vehicle", func(t *testing.T) {
		vehicleSvc := &stubVehicleSvc{}
		handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
		req := addResourceContext(addAuthContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), "u1", "testuser"), "v1")
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Empty(t, vehicleSvc.deletedID)
	})

	t.Run("reports deletion failure", func(t *testing.T) {
		vehicleSvc := &stubVehicleSvc{deleteErr: assert.AnError}
		handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})
		req := addResourceContext(addAuthContext(httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil), "u1", "testuser"), vehicle)
		rec := httptest.NewRecorder()

		handler.VehicleDelete(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Empty(t, rec.Header().Get("Location"))
	})
}

func TestPageHandler_VehicleDeleteFormCSRF(t *testing.T) {
	vehicle := &models.Vehicle{ID: "v1", UserID: "u1", Make: "Ford", Model: "Focus", Year: 2020}
	vehicleSvc := &stubVehicleSvc{getResult: vehicle}
	handler := newTestVehicleDetailPageHandler(t, vehicleSvc, &stubMaintenanceSvc{})

	router := mux.NewRouter()
	pages := router.NewRoute().Subrouter()
	pages.Use(middleware.CSRFProtection("test-secret"))
	vehiclePages := pages.PathPrefix("/vehicles/{id}").Subrouter()
	vehiclePages.Use(middleware.PageResourceOwnershipGuard(func(_ context.Context, r *http.Request) (interface{}, string, error) {
		if mux.Vars(r)["id"] == "v2" {
			return &models.Vehicle{ID: "v2", UserID: "other"}, "other", nil
		}
		return vehicle, vehicle.UserID, nil
	}))
	vehiclePages.HandleFunc("", handler.VehicleDetail).Methods(http.MethodGet)
	vehiclePages.HandleFunc("/delete", handler.VehicleDelete).Methods(http.MethodPost)
	authedRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		router.ServeHTTP(w, addAuthContext(r, "u1", "testuser"))
	})

	get := httptest.NewRecorder()
	authedRouter.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/vehicles/v1", nil))
	require.Equal(t, http.StatusOK, get.Code)
	matches := deleteFormToken.FindStringSubmatch(get.Body.String())
	require.Len(t, matches, 2, "vehicle delete form must contain a CSRF token")
	assert.Contains(t, get.Body.String(), `action="/vehicles/v1/delete"`)

	post := httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", strings.NewReader(url.Values{"csrf_token": {matches[1]}}.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	authedRouter.ServeHTTP(rec, post)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/vehicles", rec.Header().Get("Location"))
	assert.Equal(t, "v1", vehicleSvc.deletedID)

	vehicleSvc.deletedID = ""
	rec = httptest.NewRecorder()
	authedRouter.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/vehicles/v1/delete", nil))
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, vehicleSvc.deletedID)

	post = httptest.NewRequest(http.MethodPost, "/vehicles/v2/delete", strings.NewReader(url.Values{"csrf_token": {matches[1]}}.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	authedRouter.ServeHTTP(rec, post)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, vehicleSvc.deletedID)
}

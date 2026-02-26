package handlers

import (
	"errors"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/truggeri/go-garage/internal/models"
	"github.com/truggeri/go-garage/internal/repositories"
)

// calcTotalPages returns the total number of pages for a given total count and page size.
func calcTotalPages(total, pageSize int) int {
	if total == 0 {
		return 1
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// buildVehicleNameMap returns a map of vehicle ID to human-readable "Make Model" name.
func buildVehicleNameMap(vehicles []*models.Vehicle) map[string]string {
	names := make(map[string]string, len(vehicles))
	for _, v := range vehicles {
		names[v.ID] = vehicleName(v)
	}
	return names
}

// isOwnedVehicle returns true if vehicleID matches one of the provided vehicles.
func isOwnedVehicle(vehicleID string, vehicles []*models.Vehicle) bool {
	for _, v := range vehicles {
		if v.ID == vehicleID {
			return true
		}
	}
	return false
}

// fetchVehicleMaintenanceRecords retrieves a paginated page of maintenance records for a
// specific vehicle using DB-level filtering and pagination.
func fetchVehicleMaintenanceRecords(
	h *PageHandler, r *http.Request,
	vehicleID, serviceType string,
	page int,
) ([]*models.MaintenanceRecord, int, int, error) {
	filters := repositories.MaintenanceFilters{VehicleID: &vehicleID}
	if serviceType != "" {
		filters.ServiceType = &serviceType
	}

	totalCount, err := h.maintenanceService.CountMaintenance(r.Context(), filters)
	if err != nil {
		return nil, 0, page, err
	}

	totalPages := calcTotalPages(totalCount, maintenanceListPageSize)
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * maintenanceListPageSize
	records, err := h.maintenanceService.ListMaintenance(r.Context(), filters, repositories.PaginationParams{
		Limit:  maintenanceListPageSize,
		Offset: offset,
	})
	return records, totalCount, page, err
}

// fetchAllUserMaintenanceRecords retrieves all maintenance records across a user's vehicles,
// optionally filters by service type, sorts by most-recent service date, and paginates in memory.
func fetchAllUserMaintenanceRecords(
	h *PageHandler, r *http.Request,
	userVehicles []*models.Vehicle,
	serviceType string,
	page int,
) ([]*models.MaintenanceRecord, int, int) {
	var all []*models.MaintenanceRecord
	for _, v := range userVehicles {
		recs, err := h.maintenanceService.GetVehicleMaintenance(r.Context(), v.ID)
		if err != nil {
			continue
		}
		all = append(all, recs...)
	}

	if serviceType != "" {
		lower := strings.ToLower(serviceType)
		filtered := make([]*models.MaintenanceRecord, 0, len(all))
		for _, rec := range all {
			if strings.Contains(strings.ToLower(rec.ServiceType), lower) {
				filtered = append(filtered, rec)
			}
		}
		all = filtered
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].ServiceDate.After(all[j].ServiceDate)
	})

	totalCount := len(all)
	totalPages := calcTotalPages(totalCount, maintenanceListPageSize)
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * maintenanceListPageSize
	end := offset + maintenanceListPageSize
	if end > len(all) {
		end = len(all)
	}

	var records []*models.MaintenanceRecord
	if offset < len(all) {
		records = all[offset:end]
	}
	return records, totalCount, page
}

// getMaintenanceRecordAndVehicle retrieves a maintenance record and its associated vehicle.
func (h *PageHandler) getMaintenanceRecordAndVehicle(r *http.Request) (*models.MaintenanceRecord, *models.Vehicle, error) {
	vars := mux.Vars(r)
	recordID := vars["id"]

	record, err := h.maintenanceService.GetMaintenance(r.Context(), recordID)
	if err != nil {
		return nil, nil, err
	}

	vehicle, err := h.vehicleService.GetVehicle(r.Context(), record.VehicleID)
	if err != nil {
		return nil, nil, err
	}

	return record, vehicle, nil
}

// writeMaintenanceRecordError writes the appropriate HTTP error for maintenance record lookup failures.
func writeMaintenanceRecordError(w http.ResponseWriter, err error) {
	var notFound *models.NotFoundError
	if errors.As(err, &notFound) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

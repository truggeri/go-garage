package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/truggeri/go-garage/internal/models"
)

// SQLiteMaintenanceRepository implements MaintenanceRepository using SQLite
type SQLiteMaintenanceRepository struct {
	db *sql.DB
}

// NewSQLiteMaintenanceRepository creates a new SQLite-based maintenance repository
func NewSQLiteMaintenanceRepository(db *sql.DB) *SQLiteMaintenanceRepository {
	return &SQLiteMaintenanceRepository{db: db}
}

// Create inserts a new maintenance record into the database
func (r *SQLiteMaintenanceRepository) Create(ctx context.Context, record *models.MaintenanceRecord) error {
	if err := models.ValidateMaintenanceRecord(record); err != nil {
		return err
	}

	// Generate UUID if not provided
	if record.ID == "" {
		record.ID = uuid.New().String()
	}

	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	query := `
		INSERT INTO maintenance_records (
			id, vehicle_id, service_type, custom_service_type, service_date, mileage_at_service,
			cost, service_provider, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.VehicleID,
		record.ServiceType,
		record.CustomServiceType,
		record.ServiceDate,
		record.MileageAtService,
		record.Cost,
		record.ServiceProvider,
		record.Notes,
		record.CreatedAt,
		record.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("vehicle_id", "vehicle does not exist")
		}
		return models.NewDatabaseError("create maintenance record", err)
	}

	return nil
}

// FindByID retrieves a maintenance record by its ID
func (r *SQLiteMaintenanceRepository) FindByID(ctx context.Context, id string) (*models.MaintenanceRecord, error) {
	query := `
		SELECT id, vehicle_id, service_type, custom_service_type, service_date, mileage_at_service,
		       cost, service_provider, notes, created_at, updated_at
		FROM maintenance_records
		WHERE id = ?
	`

	record := &models.MaintenanceRecord{}
	var mileageAtService sql.NullInt64
	var cost sql.NullFloat64
	var customServiceType, serviceProvider, notes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.VehicleID,
		&record.ServiceType,
		&customServiceType,
		&record.ServiceDate,
		&mileageAtService,
		&cost,
		&serviceProvider,
		&notes,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("MaintenanceRecord", id)
		}
		return nil, models.NewDatabaseError("find maintenance record by ID", err)
	}

	// Handle nullable fields
	if customServiceType.Valid {
		record.CustomServiceType = customServiceType.String
	}
	if mileageAtService.Valid {
		m := int(mileageAtService.Int64)
		record.MileageAtService = &m
	}
	if cost.Valid {
		record.Cost = &cost.Float64
	}
	if serviceProvider.Valid {
		record.ServiceProvider = serviceProvider.String
	}
	if notes.Valid {
		record.Notes = notes.String
	}

	return record, nil
}

// FindByVehicleID retrieves all maintenance records for a specific vehicle
func (r *SQLiteMaintenanceRepository) FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.MaintenanceRecord, error) {
	query := `
		SELECT id, vehicle_id, service_type, custom_service_type, service_date, mileage_at_service,
		       cost, service_provider, notes, created_at, updated_at
		FROM maintenance_records
		WHERE vehicle_id = ?
		ORDER BY service_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, models.NewDatabaseError("find maintenance records by vehicle ID", err)
	}
	defer rows.Close()

	return r.scanMaintenanceRecords(rows)
}

// Update modifies an existing maintenance record
func (r *SQLiteMaintenanceRepository) Update(ctx context.Context, record *models.MaintenanceRecord) error {
	if err := models.ValidateMaintenanceRecord(record); err != nil {
		return err
	}

	record.UpdatedAt = time.Now()

	query := `
		UPDATE maintenance_records
		SET vehicle_id = ?, service_type = ?, custom_service_type = ?, service_date = ?,
		    mileage_at_service = ?, cost = ?, service_provider = ?,
		    notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		record.VehicleID,
		record.ServiceType,
		record.CustomServiceType,
		record.ServiceDate,
		record.MileageAtService,
		record.Cost,
		record.ServiceProvider,
		record.Notes,
		record.UpdatedAt,
		record.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("vehicle_id", "vehicle does not exist")
		}
		return models.NewDatabaseError("update maintenance record", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("update maintenance record check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("MaintenanceRecord", record.ID)
	}

	return nil
}

// Delete removes a maintenance record from the database
func (r *SQLiteMaintenanceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM maintenance_records WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return models.NewDatabaseError("delete maintenance record", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("delete maintenance record check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("MaintenanceRecord", id)
	}

	return nil
}

// List retrieves maintenance records with optional filters and pagination
func (r *SQLiteMaintenanceRepository) List(ctx context.Context, filters MaintenanceFilters, pagination PaginationParams) ([]*models.MaintenanceRecord, error) {
	query := `
		SELECT id, vehicle_id, service_type, custom_service_type, service_date, mileage_at_service,
		       cost, service_provider, notes, created_at, updated_at
		FROM maintenance_records
		WHERE 1=1
	`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlFilterVehicleID
		args = append(args, *filters.VehicleID)
	}

	if filters.ServiceType != nil {
		query += " AND service_type = ?"
		args = append(args, *filters.ServiceType)
	}

	query += " ORDER BY service_date DESC"

	// Add pagination
	if pagination.Limit > 0 {
		query += sqlLimit
		args = append(args, pagination.Limit)
	}

	if pagination.Offset > 0 {
		query += sqlOffset
		args = append(args, pagination.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, models.NewDatabaseError("list maintenance records", err)
	}
	defer rows.Close()

	return r.scanMaintenanceRecords(rows)
}

// Count returns the total number of maintenance records matching the filters
func (r *SQLiteMaintenanceRepository) Count(ctx context.Context, filters MaintenanceFilters) (int, error) {
	query := `SELECT COUNT(*) FROM maintenance_records WHERE 1=1`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlFilterVehicleID
		args = append(args, *filters.VehicleID)
	}

	if filters.ServiceType != nil {
		query += " AND service_type = ?"
		args = append(args, *filters.ServiceType)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, models.NewDatabaseError("count maintenance records", err)
	}

	return count, nil
}

// SumCostByVehicleID returns the sum of all maintenance costs for a specific vehicle.
// Returns nil if there are no records with a cost for the vehicle.
func (r *SQLiteMaintenanceRepository) SumCostByVehicleID(ctx context.Context, vehicleID string) (*float64, error) {
	query := `SELECT SUM(cost) FROM maintenance_records WHERE vehicle_id = ? AND cost IS NOT NULL`

	var total sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, vehicleID).Scan(&total)
	if err != nil {
		return nil, models.NewDatabaseError("sum cost by vehicle ID", err)
	}

	if !total.Valid {
		return nil, nil
	}

	return &total.Float64, nil
}

// scanMaintenanceRecords is a helper method to scan multiple maintenance record rows
func (r *SQLiteMaintenanceRepository) scanMaintenanceRecords(rows *sql.Rows) ([]*models.MaintenanceRecord, error) {
	records := []*models.MaintenanceRecord{}

	for rows.Next() {
		record := &models.MaintenanceRecord{}
		var mileageAtService sql.NullInt64
		var cost sql.NullFloat64
		var customServiceType, serviceProvider, notes sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.VehicleID,
			&record.ServiceType,
			&customServiceType,
			&record.ServiceDate,
			&mileageAtService,
			&cost,
			&serviceProvider,
			&notes,
			&record.CreatedAt,
			&record.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan maintenance record row: %w", err)
		}

		// Handle nullable fields
		if customServiceType.Valid {
			record.CustomServiceType = customServiceType.String
		}
		if mileageAtService.Valid {
			m := int(mileageAtService.Int64)
			record.MileageAtService = &m
		}
		if cost.Valid {
			record.Cost = &cost.Float64
		}
		if serviceProvider.Valid {
			record.ServiceProvider = serviceProvider.String
		}
		if notes.Valid {
			record.Notes = notes.String
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate maintenance record rows: %w", err)
	}

	return records, nil
}

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

// SQLiteFuelRepository implements FuelRepository using SQLite
type SQLiteFuelRepository struct {
	db *sql.DB
}

// NewSQLiteFuelRepository creates a new SQLite-based fuel repository
func NewSQLiteFuelRepository(db *sql.DB) *SQLiteFuelRepository {
	return &SQLiteFuelRepository{db: db}
}

// Create inserts a new fuel record into the database
func (r *SQLiteFuelRepository) Create(ctx context.Context, record *models.FuelRecord) error {
	if err := models.ValidateFuelRecord(record); err != nil {
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
		INSERT INTO fuel_records (
			id, vehicle_id, fill_date, odometer, cost_per_unit, volume,
			fuel_type, city_driving_pct, location, brand, notes,
			reported_mpg, partial, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.VehicleID,
		record.FillDate,
		record.Odometer,
		record.CostPerUnit,
		record.Volume,
		nullableString(record.FuelType),
		record.CityDrivingPct,
		nullableString(record.Location),
		nullableString(record.Brand),
		nullableString(record.Notes),
		record.ReportedMPG,
		record.PartialFuelUp,
		record.CreatedAt,
		record.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("vehicle_id", "vehicle does not exist")
		}
		return models.NewDatabaseError("create fuel record", err)
	}

	return nil
}

// FindByID retrieves a fuel record by its ID
func (r *SQLiteFuelRepository) FindByID(ctx context.Context, id string) (*models.FuelRecord, error) {
	query := `
		SELECT id, vehicle_id, fill_date, odometer, cost_per_unit, volume,
		       fuel_type, city_driving_pct, location, brand, notes,
		       reported_mpg, partial, created_at, updated_at
		FROM fuel_records
		WHERE id = ?
	`

	record := &models.FuelRecord{}
	var fuelType, location, brand, notes sql.NullString
	var cityDrivingPct sql.NullInt64
	var reportedMPG sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.VehicleID,
		&record.FillDate,
		&record.Odometer,
		&record.CostPerUnit,
		&record.Volume,
		&fuelType,
		&cityDrivingPct,
		&location,
		&brand,
		&notes,
		&reportedMPG,
		&record.PartialFuelUp,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("FuelRecord", id)
		}
		return nil, models.NewDatabaseError("find fuel record by ID", err)
	}

	applyFuelNullableFields(record, fuelType, cityDrivingPct, location, brand, notes, reportedMPG)

	return record, nil
}

// FindByVehicleID retrieves all fuel records for a specific vehicle
func (r *SQLiteFuelRepository) FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error) {
	query := `
		SELECT id, vehicle_id, fill_date, odometer, cost_per_unit, volume,
		       fuel_type, city_driving_pct, location, brand, notes,
		       reported_mpg, partial, created_at, updated_at
		FROM fuel_records
		WHERE vehicle_id = ?
		ORDER BY fill_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, models.NewDatabaseError("find fuel records by vehicle ID", err)
	}
	defer rows.Close()

	return r.scanFuelRecords(rows)
}

// Update modifies an existing fuel record
func (r *SQLiteFuelRepository) Update(ctx context.Context, record *models.FuelRecord) error {
	if err := models.ValidateFuelRecord(record); err != nil {
		return err
	}

	record.UpdatedAt = time.Now()

	query := `
		UPDATE fuel_records
		SET vehicle_id = ?, fill_date = ?, odometer = ?, cost_per_unit = ?, volume = ?,
		    fuel_type = ?, city_driving_pct = ?, location = ?, brand = ?, notes = ?,
		    reported_mpg = ?, partial = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		record.VehicleID,
		record.FillDate,
		record.Odometer,
		record.CostPerUnit,
		record.Volume,
		nullableString(record.FuelType),
		record.CityDrivingPct,
		nullableString(record.Location),
		nullableString(record.Brand),
		nullableString(record.Notes),
		record.ReportedMPG,
		record.PartialFuelUp,
		record.UpdatedAt,
		record.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("vehicle_id", "vehicle does not exist")
		}
		return models.NewDatabaseError("update fuel record", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("update fuel record check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("FuelRecord", record.ID)
	}

	return nil
}

// Delete removes a fuel record from the database
func (r *SQLiteFuelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM fuel_records WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return models.NewDatabaseError("delete fuel record", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("delete fuel record check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("FuelRecord", id)
	}

	return nil
}

// List retrieves fuel records with optional filters and pagination
func (r *SQLiteFuelRepository) List(ctx context.Context, filters FuelFilters, pagination PaginationParams) ([]*models.FuelRecord, error) {
	query := `
		SELECT id, vehicle_id, fill_date, odometer, cost_per_unit, volume,
		       fuel_type, city_driving_pct, location, brand, notes,
		       reported_mpg, partial, created_at, updated_at
		FROM fuel_records
		WHERE 1=1
	`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlAndVehicleID
		args = append(args, *filters.VehicleID)
	}

	if filters.FuelType != nil {
		query += " AND fuel_type = ?"
		args = append(args, *filters.FuelType)
	}

	query += " ORDER BY fill_date DESC"

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
		return nil, models.NewDatabaseError("list fuel records", err)
	}
	defer rows.Close()

	return r.scanFuelRecords(rows)
}

// Count returns the total number of fuel records matching the filters
func (r *SQLiteFuelRepository) Count(ctx context.Context, filters FuelFilters) (int, error) {
	query := `SELECT COUNT(*) FROM fuel_records WHERE 1=1`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlAndVehicleID
		args = append(args, *filters.VehicleID)
	}

	if filters.FuelType != nil {
		query += " AND fuel_type = ?"
		args = append(args, *filters.FuelType)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, models.NewDatabaseError("count fuel records", err)
	}

	return count, nil
}

// scanFuelRecords is a helper method to scan multiple fuel record rows
func (r *SQLiteFuelRepository) scanFuelRecords(rows *sql.Rows) ([]*models.FuelRecord, error) {
	records := []*models.FuelRecord{}

	for rows.Next() {
		record := &models.FuelRecord{}
		var fuelType, location, brand, notes sql.NullString
		var cityDrivingPct sql.NullInt64
		var reportedMPG sql.NullFloat64

		err := rows.Scan(
			&record.ID,
			&record.VehicleID,
			&record.FillDate,
			&record.Odometer,
			&record.CostPerUnit,
			&record.Volume,
			&fuelType,
			&cityDrivingPct,
			&location,
			&brand,
			&notes,
			&reportedMPG,
			&record.PartialFuelUp,
			&record.CreatedAt,
			&record.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan fuel record row: %w", err)
		}

		applyFuelNullableFields(record, fuelType, cityDrivingPct, location, brand, notes, reportedMPG)

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fuel record rows: %w", err)
	}

	return records, nil
}

// applyFuelNullableFields assigns nullable SQL fields to the FuelRecord model.
func applyFuelNullableFields(
	record *models.FuelRecord,
	fuelType sql.NullString,
	cityDrivingPct sql.NullInt64,
	location, brand, notes sql.NullString,
	reportedMPG sql.NullFloat64,
) {
	if fuelType.Valid {
		record.FuelType = fuelType.String
	}
	if cityDrivingPct.Valid {
		pct := int(cityDrivingPct.Int64)
		record.CityDrivingPct = &pct
	}
	if location.Valid {
		record.Location = location.String
	}
	if brand.Valid {
		record.Brand = brand.String
	}
	if notes.Valid {
		record.Notes = notes.String
	}
	if reportedMPG.Valid {
		record.ReportedMPG = &reportedMPG.Float64
	}
}

// nullableString converts an empty string to a NULL SQL value.
func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

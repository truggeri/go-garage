package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/truggeri/go-garage/internal/models"
)

const fuelRecordColumns = `id, vehicle_id, fill_date, mileage, volume, fuel_type, partial_fill,
	price_per_unit, octane_rating, location, brand, notes,
	city_driving_percentage, vehicle_reported_mpg, created_at, updated_at`

// nullableString returns a sql.NullString for empty strings
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// scanFuelRecord scans a single row into a FuelRecord
func scanFuelRecord(scanner interface{ Scan(...interface{}) error }) (*models.FuelRecord, error) {
	record := &models.FuelRecord{}
	var pricePerUnit sql.NullFloat64
	var octaneRating, cityDrivingPercentage sql.NullInt64
	var location, brand, notes sql.NullString
	var vehicleReportedMPG sql.NullFloat64

	err := scanner.Scan(
		&record.ID, &record.VehicleID, &record.FillDate, &record.Mileage,
		&record.Volume, &record.FuelType, &record.PartialFill,
		&pricePerUnit, &octaneRating,
		&location, &brand, &notes,
		&cityDrivingPercentage, &vehicleReportedMPG,
		&record.CreatedAt, &record.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if pricePerUnit.Valid {
		record.PricePerUnit = &pricePerUnit.Float64
	}
	if octaneRating.Valid {
		o := int(octaneRating.Int64)
		record.OctaneRating = &o
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
	if cityDrivingPercentage.Valid {
		c := int(cityDrivingPercentage.Int64)
		record.CityDrivingPercentage = &c
	}
	if vehicleReportedMPG.Valid {
		record.VehicleReportedMPG = &vehicleReportedMPG.Float64
	}

	return record, nil
}

// scanFuelRecords scans multiple rows into FuelRecord slices
func scanFuelRecords(rows *sql.Rows) ([]*models.FuelRecord, error) {
	records := []*models.FuelRecord{}
	for rows.Next() {
		record, err := scanFuelRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan fuel record row: %w", err)
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate fuel record rows: %w", err)
	}
	return records, nil
}

// FindByID retrieves a fuel record by its ID
func (r *SQLiteFuelRepository) FindByID(ctx context.Context, id string) (*models.FuelRecord, error) {
	query := `SELECT ` + fuelRecordColumns + ` FROM fuel_records WHERE id = ?`

	record, err := scanFuelRecord(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("FuelRecord", id)
		}
		return nil, models.NewDatabaseError("find fuel record by ID", err)
	}

	return record, nil
}

// FindByVehicleID retrieves all fuel records for a specific vehicle
func (r *SQLiteFuelRepository) FindByVehicleID(ctx context.Context, vehicleID string) ([]*models.FuelRecord, error) {
	query := `SELECT ` + fuelRecordColumns + ` FROM fuel_records WHERE vehicle_id = ? ORDER BY fill_date DESC`

	rows, err := r.db.QueryContext(ctx, query, vehicleID)
	if err != nil {
		return nil, models.NewDatabaseError("find fuel records by vehicle ID", err)
	}
	defer rows.Close()

	return scanFuelRecords(rows)
}

// List retrieves fuel records with optional filters and pagination
func (r *SQLiteFuelRepository) List(ctx context.Context, filters FuelFilters, pagination PaginationParams) ([]*models.FuelRecord, error) {
	query := `SELECT ` + fuelRecordColumns + ` FROM fuel_records WHERE 1=1`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlFilterVehicleID
		args = append(args, *filters.VehicleID)
	}

	if filters.FuelType != nil {
		query += " AND fuel_type = ?"
		args = append(args, *filters.FuelType)
	}

	query += " ORDER BY fill_date DESC"

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

	return scanFuelRecords(rows)
}

// Count returns the total number of fuel records matching the filters
func (r *SQLiteFuelRepository) Count(ctx context.Context, filters FuelFilters) (int, error) {
	query := `SELECT COUNT(*) FROM fuel_records WHERE 1=1`
	args := []interface{}{}

	if filters.VehicleID != nil {
		query += sqlFilterVehicleID
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

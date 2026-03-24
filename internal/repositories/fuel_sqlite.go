package repositories

import (
	"context"
	"database/sql"
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

	if record.ID == "" {
		record.ID = uuid.New().String()
	}

	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	query := `
		INSERT INTO fuel_records (
			id, vehicle_id, fill_date, mileage, volume, fuel_type, partial_fill,
			price_per_unit, octane_rating, location, brand, notes,
			city_driving_percentage, vehicle_reported_mpg, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		record.ID, record.VehicleID, record.FillDate, record.Mileage,
		record.Volume, record.FuelType, record.PartialFill,
		record.PricePerUnit, record.OctaneRating,
		nullableString(record.Location), nullableString(record.Brand),
		nullableString(record.Notes),
		record.CityDrivingPercentage, record.VehicleReportedMPG,
		record.CreatedAt, record.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("vehicle_id", "vehicle does not exist")
		}
		return models.NewDatabaseError("create fuel record", err)
	}

	return nil
}

// Update modifies an existing fuel record
func (r *SQLiteFuelRepository) Update(ctx context.Context, record *models.FuelRecord) error {
	if err := models.ValidateFuelRecord(record); err != nil {
		return err
	}

	record.UpdatedAt = time.Now()

	query := `
		UPDATE fuel_records
		SET vehicle_id = ?, fill_date = ?, mileage = ?, volume = ?, fuel_type = ?,
		    partial_fill = ?, price_per_unit = ?, octane_rating = ?, location = ?,
		    brand = ?, notes = ?, city_driving_percentage = ?,
		    vehicle_reported_mpg = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		record.VehicleID, record.FillDate, record.Mileage, record.Volume,
		record.FuelType, record.PartialFill,
		record.PricePerUnit, record.OctaneRating,
		nullableString(record.Location), nullableString(record.Brand),
		nullableString(record.Notes),
		record.CityDrivingPercentage, record.VehicleReportedMPG,
		record.UpdatedAt, record.ID,
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

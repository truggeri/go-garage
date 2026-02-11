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

// SQLiteVehicleRepository implements VehicleRepository using SQLite
type SQLiteVehicleRepository struct {
	db *sql.DB
}

// NewSQLiteVehicleRepository creates a new SQLite-based vehicle repository
func NewSQLiteVehicleRepository(db *sql.DB) *SQLiteVehicleRepository {
	return &SQLiteVehicleRepository{db: db}
}

// Create inserts a new vehicle into the database
func (r *SQLiteVehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	if err := models.ValidateVehicle(vehicle); err != nil {
		return err
	}

	// Generate UUID if not provided
	if vehicle.ID == "" {
		vehicle.ID = uuid.New().String()
	}

	now := time.Now()
	vehicle.CreatedAt = now
	vehicle.UpdatedAt = now

	query := `
		INSERT INTO vehicles (
			id, user_id, vin, make, model, year, color, license_plate,
			purchase_date, purchase_price, purchase_mileage, current_mileage,
			status, notes, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		vehicle.ID,
		vehicle.UserID,
		vehicle.VIN,
		vehicle.Make,
		vehicle.Model,
		vehicle.Year,
		vehicle.Color,
		vehicle.LicensePlate,
		vehicle.PurchaseDate,
		vehicle.PurchasePrice,
		vehicle.PurchaseMileage,
		vehicle.CurrentMileage,
		vehicle.Status,
		vehicle.Notes,
		vehicle.CreatedAt,
		vehicle.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "vin") {
				return models.NewDuplicateError("Vehicle", "vin", vehicle.VIN)
			}
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("user_id", "user does not exist")
		}
		return models.NewDatabaseError("create vehicle", err)
	}

	return nil
}

// FindByID retrieves a vehicle by its ID
func (r *SQLiteVehicleRepository) FindByID(ctx context.Context, id string) (*models.Vehicle, error) {
	query := `
		SELECT id, user_id, vin, make, model, year, color, license_plate,
		       purchase_date, purchase_price, purchase_mileage, current_mileage,
		       status, notes, created_at, updated_at
		FROM vehicles
		WHERE id = ?
	`

	vehicle := &models.Vehicle{}
	var purchaseDate, color, licensePlate, notes sql.NullString
	var purchasePrice sql.NullFloat64
	var purchaseMileage, currentMileage sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&vehicle.ID,
		&vehicle.UserID,
		&vehicle.VIN,
		&vehicle.Make,
		&vehicle.Model,
		&vehicle.Year,
		&color,
		&licensePlate,
		&purchaseDate,
		&purchasePrice,
		&purchaseMileage,
		&currentMileage,
		&vehicle.Status,
		&notes,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("Vehicle", id)
		}
		return nil, models.NewDatabaseError("find vehicle by ID", err)
	}

	// Handle nullable fields
	if color.Valid {
		vehicle.Color = color.String
	}
	if licensePlate.Valid {
		vehicle.LicensePlate = licensePlate.String
	}
	if purchaseDate.Valid && purchaseDate.String != "" {
		// SQLite stores dates as strings; try parsing as RFC3339 or date-only format
		t, parseErr := time.Parse(time.RFC3339, purchaseDate.String)
		if parseErr != nil {
			// Try date-only format
			t, parseErr = time.Parse("2006-01-02", purchaseDate.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse purchase date: %w", parseErr)
			}
		}
		vehicle.PurchaseDate = &t
	}
	if purchasePrice.Valid {
		vehicle.PurchasePrice = &purchasePrice.Float64
	}
	if purchaseMileage.Valid {
		m := int(purchaseMileage.Int64)
		vehicle.PurchaseMileage = &m
	}
	if currentMileage.Valid {
		m := int(currentMileage.Int64)
		vehicle.CurrentMileage = &m
	}
	if notes.Valid {
		vehicle.Notes = notes.String
	}

	return vehicle, nil
}

// FindByUserID retrieves all vehicles for a specific user
func (r *SQLiteVehicleRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Vehicle, error) {
	query := `
		SELECT id, user_id, vin, make, model, year, color, license_plate,
		       purchase_date, purchase_price, purchase_mileage, current_mileage,
		       status, notes, created_at, updated_at
		FROM vehicles
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, models.NewDatabaseError("find vehicles by user ID", err)
	}
	defer rows.Close()

	return r.scanVehicles(rows)
}

// FindByVIN retrieves a vehicle by its VIN
func (r *SQLiteVehicleRepository) FindByVIN(ctx context.Context, vin string) (*models.Vehicle, error) {
	query := `
		SELECT id, user_id, vin, make, model, year, color, license_plate,
		       purchase_date, purchase_price, purchase_mileage, current_mileage,
		       status, notes, created_at, updated_at
		FROM vehicles
		WHERE vin = ?
	`

	vehicle := &models.Vehicle{}
	var purchaseDate, color, licensePlate, notes sql.NullString
	var purchasePrice sql.NullFloat64
	var purchaseMileage, currentMileage sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, vin).Scan(
		&vehicle.ID,
		&vehicle.UserID,
		&vehicle.VIN,
		&vehicle.Make,
		&vehicle.Model,
		&vehicle.Year,
		&color,
		&licensePlate,
		&purchaseDate,
		&purchasePrice,
		&purchaseMileage,
		&currentMileage,
		&vehicle.Status,
		&notes,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.NewNotFoundError("Vehicle", vin)
		}
		return nil, models.NewDatabaseError("find vehicle by VIN", err)
	}

	// Handle nullable fields
	if color.Valid {
		vehicle.Color = color.String
	}
	if licensePlate.Valid {
		vehicle.LicensePlate = licensePlate.String
	}
	if purchaseDate.Valid && purchaseDate.String != "" {
		// SQLite stores dates as strings; try parsing as RFC3339 or date-only format
		t, parseErr := time.Parse(time.RFC3339, purchaseDate.String)
		if parseErr != nil {
			// Try date-only format
			t, parseErr = time.Parse("2006-01-02", purchaseDate.String)
			if parseErr != nil {
				return nil, fmt.Errorf("parse purchase date: %w", parseErr)
			}
		}
		vehicle.PurchaseDate = &t
	}
	if purchasePrice.Valid {
		vehicle.PurchasePrice = &purchasePrice.Float64
	}
	if purchaseMileage.Valid {
		m := int(purchaseMileage.Int64)
		vehicle.PurchaseMileage = &m
	}
	if currentMileage.Valid {
		m := int(currentMileage.Int64)
		vehicle.CurrentMileage = &m
	}
	if notes.Valid {
		vehicle.Notes = notes.String
	}

	return vehicle, nil
}

// Update modifies an existing vehicle's information
func (r *SQLiteVehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	if err := models.ValidateVehicle(vehicle); err != nil {
		return err
	}

	vehicle.UpdatedAt = time.Now()

	query := `
		UPDATE vehicles
		SET user_id = ?, vin = ?, make = ?, model = ?, year = ?, 
		    color = ?, license_plate = ?, purchase_date = ?, 
		    purchase_price = ?, purchase_mileage = ?, current_mileage = ?,
		    status = ?, notes = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		vehicle.UserID,
		vehicle.VIN,
		vehicle.Make,
		vehicle.Model,
		vehicle.Year,
		vehicle.Color,
		vehicle.LicensePlate,
		vehicle.PurchaseDate,
		vehicle.PurchasePrice,
		vehicle.PurchaseMileage,
		vehicle.CurrentMileage,
		vehicle.Status,
		vehicle.Notes,
		vehicle.UpdatedAt,
		vehicle.ID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "vin") {
				return models.NewDuplicateError("Vehicle", "vin", vehicle.VIN)
			}
		}
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return models.NewValidationError("user_id", "user does not exist")
		}
		return models.NewDatabaseError("update vehicle", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("update vehicle check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("Vehicle", vehicle.ID)
	}

	return nil
}

// Delete removes a vehicle from the database
func (r *SQLiteVehicleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM vehicles WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return models.NewDatabaseError("delete vehicle", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.NewDatabaseError("delete vehicle check", err)
	}

	if rowsAffected == 0 {
		return models.NewNotFoundError("Vehicle", id)
	}

	return nil
}

// List retrieves vehicles with optional filters and pagination
func (r *SQLiteVehicleRepository) List(ctx context.Context, filters VehicleFilters, pagination PaginationParams) ([]*models.Vehicle, error) {
	query := `
		SELECT id, user_id, vin, make, model, year, color, license_plate,
		       purchase_date, purchase_price, purchase_mileage, current_mileage,
		       status, notes, created_at, updated_at
		FROM vehicles
		WHERE 1=1
	`
	args := []interface{}{}

	if filters.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filters.UserID)
	}

	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, *filters.Status)
	}

	if filters.Make != nil {
		query += " AND make = ?"
		args = append(args, *filters.Make)
	}

	if filters.Model != nil {
		query += " AND model = ?"
		args = append(args, *filters.Model)
	}

	if filters.Year != nil {
		query += " AND year = ?"
		args = append(args, *filters.Year)
	}

	query += " ORDER BY created_at DESC"

	// Add pagination
	if pagination.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, pagination.Limit)
	}

	if pagination.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, pagination.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, models.NewDatabaseError("list vehicles", err)
	}
	defer rows.Close()

	return r.scanVehicles(rows)
}

// scanVehicles is a helper method to scan multiple vehicle rows
func (r *SQLiteVehicleRepository) scanVehicles(rows *sql.Rows) ([]*models.Vehicle, error) {
	vehicles := []*models.Vehicle{}

	for rows.Next() {
		vehicle := &models.Vehicle{}
		var purchaseDate, color, licensePlate, notes sql.NullString
		var purchasePrice sql.NullFloat64
		var purchaseMileage, currentMileage sql.NullInt64

		err := rows.Scan(
			&vehicle.ID,
			&vehicle.UserID,
			&vehicle.VIN,
			&vehicle.Make,
			&vehicle.Model,
			&vehicle.Year,
			&color,
			&licensePlate,
			&purchaseDate,
			&purchasePrice,
			&purchaseMileage,
			&currentMileage,
			&vehicle.Status,
			&notes,
			&vehicle.CreatedAt,
			&vehicle.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan vehicle row: %w", err)
		}

		// Handle nullable fields
		if color.Valid {
			vehicle.Color = color.String
		}
		if licensePlate.Valid {
			vehicle.LicensePlate = licensePlate.String
		}
		if purchaseDate.Valid && purchaseDate.String != "" {
			// SQLite stores dates as strings; try parsing as RFC3339 or date-only format
			t, parseErr := time.Parse(time.RFC3339, purchaseDate.String)
			if parseErr != nil {
				// Try date-only format
				t, parseErr = time.Parse("2006-01-02", purchaseDate.String)
				if parseErr != nil {
					return nil, fmt.Errorf("parse purchase date: %w", parseErr)
				}
			}
			vehicle.PurchaseDate = &t
		}
		if purchasePrice.Valid {
			vehicle.PurchasePrice = &purchasePrice.Float64
		}
		if purchaseMileage.Valid {
			m := int(purchaseMileage.Int64)
			vehicle.PurchaseMileage = &m
		}
		if currentMileage.Valid {
			m := int(currentMileage.Int64)
			vehicle.CurrentMileage = &m
		}
		if notes.Valid {
			vehicle.Notes = notes.String
		}

		vehicles = append(vehicles, vehicle)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle rows: %w", err)
	}

	return vehicles, nil
}

// Count returns the total number of vehicles matching the filters
func (r *SQLiteVehicleRepository) Count(ctx context.Context, filters VehicleFilters) (int, error) {
	query := `SELECT COUNT(*) FROM vehicles WHERE 1=1`
	args := []interface{}{}

	if filters.UserID != nil {
		query += " AND user_id = ?"
		args = append(args, *filters.UserID)
	}

	if filters.Status != nil {
		query += " AND status = ?"
		args = append(args, *filters.Status)
	}

	if filters.Make != nil {
		query += " AND make = ?"
		args = append(args, *filters.Make)
	}

	if filters.Model != nil {
		query += " AND model = ?"
		args = append(args, *filters.Model)
	}

	if filters.Year != nil {
		query += " AND year = ?"
		args = append(args, *filters.Year)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, models.NewDatabaseError("count vehicles", err)
	}

	return count, nil
}

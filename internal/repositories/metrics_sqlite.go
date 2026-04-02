package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/truggeri/go-garage/internal/models"
)

// SQLiteMetricsRepository implements MetricsRepository using SQLite.
type SQLiteMetricsRepository struct {
	db *sql.DB
}

// NewSQLiteMetricsRepository creates a new SQLite-based metrics repository.
func NewSQLiteMetricsRepository(db *sql.DB) *SQLiteMetricsRepository {
	return &SQLiteMetricsRepository{db: db}
}

// Upsert inserts or updates the metrics for a given vehicle.
func (r *SQLiteMetricsRepository) Upsert(ctx context.Context, metrics *models.VehicleMetrics) error {
	now := time.Now()
	metrics.UpdatedAt = now

	query := `
		INSERT INTO vehicle_metrics (vehicle_id, total_spent, total_fuel_spent, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(vehicle_id) DO UPDATE SET
			total_spent = COALESCE(excluded.total_spent, vehicle_metrics.total_spent),
			total_fuel_spent = COALESCE(excluded.total_fuel_spent, vehicle_metrics.total_fuel_spent),
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		metrics.VehicleID,
		metrics.TotalSpent,
		metrics.TotalFuelSpent,
		now,
		now,
	)
	if err != nil {
		return models.NewDatabaseError("upsert vehicle metrics", err)
	}

	return nil
}

// GetByVehicleID retrieves the metrics for a specific vehicle.
// Returns nil without error if no metrics row exists.
func (r *SQLiteMetricsRepository) GetByVehicleID(ctx context.Context, vehicleID string) (*models.VehicleMetrics, error) {
	query := `
		SELECT vehicle_id, total_spent, total_fuel_spent, created_at, updated_at
		FROM vehicle_metrics
		WHERE vehicle_id = ?
	`

	metrics := &models.VehicleMetrics{}
	var totalSpent sql.NullFloat64
	var totalFuelSpent sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, vehicleID).Scan(
		&metrics.VehicleID,
		&totalSpent,
		&totalFuelSpent,
		&metrics.CreatedAt,
		&metrics.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, models.NewDatabaseError("get vehicle metrics by vehicle ID", err)
	}

	if totalSpent.Valid {
		metrics.TotalSpent = &totalSpent.Float64
	}
	if totalFuelSpent.Valid {
		metrics.TotalFuelSpent = &totalFuelSpent.Float64
	}

	return metrics, nil
}

// SumTotalSpentByVehicleIDs returns the sum of total_spent and total_fuel_spent across the given vehicle IDs.
func (r *SQLiteMetricsRepository) SumTotalSpentByVehicleIDs(ctx context.Context, vehicleIDs []string) (float64, error) {
	if len(vehicleIDs) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(vehicleIDs))
	args := make([]interface{}, len(vehicleIDs))
	for i, id := range vehicleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(
		`SELECT COALESCE(SUM(COALESCE(total_spent, 0) + COALESCE(total_fuel_spent, 0)), 0) FROM vehicle_metrics WHERE vehicle_id IN (%s)`,
		strings.Join(placeholders, ","),
	)

	var total float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, models.NewDatabaseError("sum total spent by vehicle IDs", err)
	}

	return total, nil
}

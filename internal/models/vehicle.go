package models

import (
	"time"
)

// VehicleStatus represents the current status of a vehicle
type VehicleStatus string

const (
	// VehicleStatusActive indicates the vehicle is currently in use
	VehicleStatusActive VehicleStatus = "active"
	// VehicleStatusSold indicates the vehicle has been sold
	VehicleStatusSold VehicleStatus = "sold"
	// VehicleStatusScrapped indicates the vehicle has been scrapped
	VehicleStatusScrapped VehicleStatus = "scrapped"
)

// Vehicle represents a vehicle in the garage management system
type Vehicle struct {
	ID              string
	UserID          string
	VIN             string
	Make            string
	Model           string
	Year            int
	Color           string
	LicensePlate    string
	PurchaseDate    *time.Time
	PurchasePrice   *float64
	PurchaseMileage *int
	CurrentMileage  *int
	Status          VehicleStatus
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

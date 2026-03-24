# Database Schema

## Overview

This document defines the database schema for the Go-Garage application. The application uses SQLite for data storage across all environments.

## Database Tables

### Users Table

Stores user account information and authentication credentials.

**Fields:**

- id (primary key)
- username (unique)
- email (unique)
- password_hash
- created_at, updated_at

### Vehicles Table

Stores vehicle information for each user.

**Fields:**

- id (primary key)
- user_id (foreign key to Users)
- vin (unique, Vehicle Identification Number)
- make, model, year
- purchase_date, purchase_price
- status (active, sold, etc.)
- created_at, updated_at

### Maintenance Records Table

Stores maintenance and service records for each vehicle.

**Fields:**

- id (primary key)
- vehicle_id (foreign key to Vehicles)
- service_type
- service_date
- mileage
- cost
- provider
- notes
- created_at, updated_at

### Vehicle Metrics Table

Stores aggregated metrics for each vehicle, updated automatically when maintenance records change.

**Fields:**

- id (primary key)
- vehicle_id (foreign key to Vehicles, unique)
- total_spent (sum of maintenance costs, nullable)
- created_at, updated_at

### Fuel Records Table

Stores fuel fill-up records for each vehicle.

**Fields:**

- id (primary key)
- vehicle_id (foreign key to Vehicles)
- fill_date (date and time of fill-up)
- mileage (odometer at fill-up, required)
- volume (gallons, required)
- fuel_type (gasoline, diesel, e85; defaults to gasoline, required)
- partial_fill (boolean, defaults false, required)
- price_per_unit (price per gallon, optional)
- octane_rating (optional)
- location (optional)
- brand (optional)
- notes (optional)
- city_driving_percentage (0-100, optional)
- vehicle_reported_mpg (optional)
- created_at, updated_at

## Relationships

- **Users → Vehicles**: One-to-Many (a user can own multiple vehicles)
- **Vehicles → Maintenance Records**: One-to-Many (a vehicle can have multiple maintenance records)
- **Vehicles → Fuel Records**: One-to-Many (a vehicle can have multiple fuel records)
- **Vehicles → Vehicle Metrics**: One-to-One (each vehicle has one metrics record)

## Indexes

- Users: index on username, email
- Vehicles: index on user_id, vin
- Maintenance Records: index on vehicle_id, service_date
- Fuel Records: index on vehicle_id, fill_date
- Vehicle Metrics: index on vehicle_id

## Constraints

- All foreign keys have ON DELETE CASCADE to ensure referential integrity
- Unique constraints on username, email, and vin fields
- NOT NULL constraints on required fields

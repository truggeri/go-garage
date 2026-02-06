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

### Fuel Records Table

Stores fuel fill-up records for each vehicle.

**Fields:**
- id (primary key)
- vehicle_id (foreign key to Vehicles)
- fill_date (date and time of fill-up)
- mileage (mileage at fill-up)
- price (total price paid)
- volume (gallons)
- city_driving_percentage (0-100)
- octane_rating
- location
- brand
- notes
- created_at, updated_at

## Relationships

- **Users → Vehicles**: One-to-Many (a user can own multiple vehicles)
- **Vehicles → Maintenance Records**: One-to-Many (a vehicle can have multiple maintenance records)
- **Vehicles → Fuel Records**: One-to-Many (a vehicle can have multiple fuel records)

## Indexes

- Users: index on username, email
- Vehicles: index on user_id, vin
- Maintenance Records: index on vehicle_id, service_date
- Fuel Records: index on vehicle_id, fill_date

## Constraints

- All foreign keys have ON DELETE CASCADE to ensure referential integrity
- Unique constraints on username, email, and vin fields
- NOT NULL constraints on required fields

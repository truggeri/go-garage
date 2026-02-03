# RESTful API Endpoints

## Overview

This document defines the RESTful API endpoints for the Go-Garage application. All endpoints follow REST conventions and return JSON responses.

## Base URL

All API endpoints are prefixed with `/api/v1/`

## Endpoints

### Vehicles

#### List all vehicles
```
GET /api/v1/vehicles
```
Returns a list of all vehicles for the authenticated user.

#### Create new vehicle
```
POST /api/v1/vehicles
```
Creates a new vehicle record.

#### Get vehicle details
```
GET /api/v1/vehicles/{id}
```
Returns details for a specific vehicle.

#### Update vehicle
```
PUT /api/v1/vehicles/{id}
```
Updates an existing vehicle record.

#### Delete vehicle
```
DELETE /api/v1/vehicles/{id}
```
Deletes a vehicle record.

### Maintenance Records

#### List maintenance records for a vehicle
```
GET /api/v1/vehicles/{id}/maintenance
```
Returns all maintenance records for a specific vehicle.

#### Add maintenance record
```
POST /api/v1/vehicles/{id}/maintenance
```
Creates a new maintenance record for a vehicle.

#### Get maintenance record
```
GET /api/v1/maintenance/{id}
```
Returns details for a specific maintenance record.

#### Update maintenance record
```
PUT /api/v1/maintenance/{id}
```
Updates an existing maintenance record.

#### Delete maintenance record
```
DELETE /api/v1/maintenance/{id}
```
Deletes a maintenance record.

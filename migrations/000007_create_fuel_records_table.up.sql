CREATE TABLE IF NOT EXISTS fuel_records (
    id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL,
    fill_date DATE NOT NULL,
    mileage INTEGER NOT NULL,
    volume REAL NOT NULL,
    fuel_type TEXT NOT NULL DEFAULT 'gasoline' CHECK(fuel_type IN ('gasoline', 'diesel', 'e85')),
    partial_fill INTEGER NOT NULL DEFAULT 0,
    price_per_unit REAL,
    octane_rating INTEGER,
    location TEXT,
    brand TEXT,
    notes TEXT,
    city_driving_percentage INTEGER CHECK(city_driving_percentage IS NULL OR (city_driving_percentage >= 0 AND city_driving_percentage <= 100)),
    vehicle_reported_mpg REAL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX idx_fuel_records_vehicle_id ON fuel_records(vehicle_id);
CREATE INDEX idx_fuel_records_fill_date ON fuel_records(fill_date);

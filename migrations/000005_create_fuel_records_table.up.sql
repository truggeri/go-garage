CREATE TABLE IF NOT EXISTS fuel_records (
    id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL,
    fill_date DATETIME NOT NULL,
    odometer INTEGER NOT NULL,
    cost_per_unit REAL NOT NULL,
    volume REAL NOT NULL,
    fuel_type TEXT,
    city_driving_pct INTEGER,
    location TEXT,
    brand TEXT,
    notes TEXT,
    reported_mpg REAL,
    partial BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_fuel_vehicle_id ON fuel_records(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fuel_fill_date ON fuel_records(fill_date);

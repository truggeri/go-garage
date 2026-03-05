CREATE TABLE IF NOT EXISTS vehicle_metrics (
    id TEXT PRIMARY KEY,
    vehicle_id TEXT NOT NULL UNIQUE,
    total_spent REAL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vehicle_metrics_vehicle_id ON vehicle_metrics(vehicle_id);

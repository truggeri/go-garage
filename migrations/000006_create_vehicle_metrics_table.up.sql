CREATE TABLE IF NOT EXISTS vehicle_metrics (
    vehicle_id TEXT PRIMARY KEY,
    total_spent REAL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (vehicle_id) REFERENCES vehicles(id) ON DELETE CASCADE
);

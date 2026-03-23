-- Reverse migration for adding custom_service_type to maintenance_records.
-- Requires SQLite 3.35+ for ALTER TABLE ... DROP COLUMN support.
ALTER TABLE maintenance_records DROP COLUMN custom_service_type;

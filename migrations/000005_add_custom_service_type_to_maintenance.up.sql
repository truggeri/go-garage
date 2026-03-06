ALTER TABLE maintenance_records ADD COLUMN custom_service_type TEXT DEFAULT '';

-- Normalize existing free-text service_type values to enum-style identifiers
UPDATE maintenance_records
SET service_type = 'oil_change'
WHERE service_type = 'Oil Change';

UPDATE maintenance_records
SET service_type = 'tire_rotation'
WHERE service_type = 'Tire Rotation';

UPDATE maintenance_records
SET service_type = 'brakes'
WHERE service_type = 'Brake Service';

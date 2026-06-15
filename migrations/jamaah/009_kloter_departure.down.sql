DROP INDEX IF EXISTS idx_groups_departure;
ALTER TABLE groups
    DROP COLUMN IF EXISTS departed_at,
    DROP COLUMN IF EXISTS manifest_finalized_at,
    DROP COLUMN IF EXISTS departure_status,
    DROP COLUMN IF EXISTS departure_date,
    DROP COLUMN IF EXISTS package_id;

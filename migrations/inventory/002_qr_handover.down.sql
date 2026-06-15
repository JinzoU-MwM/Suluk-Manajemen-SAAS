DROP TABLE IF EXISTS handover_checkpoints;
DROP INDEX IF EXISTS idx_member_equipment_token;
ALTER TABLE member_equipment
    DROP COLUMN IF EXISTS luggage_checked_at,
    DROP COLUMN IF EXISTS is_luggage_checked,
    DROP COLUMN IF EXISTS handover_token;

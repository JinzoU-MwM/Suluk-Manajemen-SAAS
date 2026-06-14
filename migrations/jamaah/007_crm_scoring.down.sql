DROP INDEX IF EXISTS idx_jpr_score_updated;
DROP TABLE IF EXISTS pipeline_history;

ALTER TABLE jamaah_package_registrations
    DROP COLUMN IF EXISTS stage_entered_at,
    DROP COLUMN IF EXISTS lead_score,
    DROP COLUMN IF EXISTS lead_temp,
    DROP COLUMN IF EXISTS lost_reason,
    DROP COLUMN IF EXISTS score_updated_at;

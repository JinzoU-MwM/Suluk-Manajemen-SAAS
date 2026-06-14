-- CRM pipeline + lead scoring (Phase 3A).
-- Adds lead-scoring + stage-timing columns to registrations and a stage-change
-- history table powering the pipeline funnel (time-in-stage, conversion).

ALTER TABLE jamaah_package_registrations
    ADD COLUMN IF NOT EXISTS stage_entered_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS lead_score       SMALLINT,
    ADD COLUMN IF NOT EXISTS lead_temp        VARCHAR(8),
    ADD COLUMN IF NOT EXISTS lost_reason      VARCHAR(40),
    ADD COLUMN IF NOT EXISTS score_updated_at TIMESTAMPTZ;

-- Backfill stage_entered_at so "days in stage" is meaningful for existing rows.
UPDATE jamaah_package_registrations
    SET stage_entered_at = COALESCE(registered_at, created_at)
    WHERE stage_entered_at IS NULL;

-- Stage-transition audit/history (for funnel: avg time-in-stage + conversion).
CREATE TABLE IF NOT EXISTS pipeline_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    jamaah_id   UUID NOT NULL,
    package_id  UUID NOT NULL,
    from_status VARCHAR(20),
    to_status   VARCHAR(20) NOT NULL,
    reason      VARCHAR(255),
    changed_by  UUID,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pipeline_history_org ON pipeline_history(org_id, created_at);
CREATE INDEX IF NOT EXISTS idx_pipeline_history_jamaah ON pipeline_history(jamaah_id);

-- Lazy-refresh scan: find registrations whose score is stale.
CREATE INDEX IF NOT EXISTS idx_jpr_score_updated ON jamaah_package_registrations(org_id, score_updated_at);

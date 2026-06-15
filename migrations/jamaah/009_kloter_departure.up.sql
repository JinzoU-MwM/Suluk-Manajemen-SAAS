-- Kloter / departure operations (Phase 5A). Makes a group departure-aware: a
-- link to its package, a departure date, and an explicit status workflow
-- (draft → siap → berangkat → selesai, or batal).
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS package_id            UUID,
    ADD COLUMN IF NOT EXISTS departure_date        DATE,
    ADD COLUMN IF NOT EXISTS departure_status      VARCHAR(16) NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS manifest_finalized_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS departed_at           TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_groups_departure ON groups(org_id, departure_status, departure_date);

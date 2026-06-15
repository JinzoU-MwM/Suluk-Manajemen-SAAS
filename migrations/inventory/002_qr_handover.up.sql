-- QR-based logistics handover (Phase 4C). A per-member handover token (encoded
-- in a QR), a luggage checkpoint alongside the existing equipment one, and an
-- audit log of every scan.
ALTER TABLE member_equipment
    ADD COLUMN IF NOT EXISTS handover_token     VARCHAR(16),
    ADD COLUMN IF NOT EXISTS is_luggage_checked BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS luggage_checked_at TIMESTAMPTZ;

-- Backfill tokens for existing members (12 hex chars from a fresh uuid).
UPDATE member_equipment
    SET handover_token = substr(replace(gen_random_uuid()::text, '-', ''), 1, 12)
    WHERE handover_token IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_member_equipment_token ON member_equipment(org_id, handover_token);

CREATE TABLE IF NOT EXISTS handover_checkpoints (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    package_id  UUID NOT NULL,
    member_id   UUID NOT NULL,
    checkpoint  VARCHAR(20) NOT NULL, -- equipment|luggage
    items       TEXT[] NOT NULL DEFAULT '{}',
    scanned_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_handover_checkpoints_pkg ON handover_checkpoints(org_id, package_id, created_at);

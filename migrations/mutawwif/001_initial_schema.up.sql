-- Mutawwif / tour-guide management (Phase 5B). A guide roster plus assignments
-- of guides to departure groups (kloter). group_id references a group in
-- jamaah-service's DB, so it's an unenforced cross-service UUID.
CREATE TABLE guides (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    phone           VARCHAR(30) DEFAULT '',
    email           VARCHAR(255) DEFAULT '',
    type            VARCHAR(20) NOT NULL DEFAULT 'mutawwif', -- mutawwif|tour_leader|kesehatan
    license_no      VARCHAR(100) DEFAULT '',
    license_expiry  DATE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    notes           TEXT DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_guides_org ON guides(org_id);

CREATE TABLE guide_assignments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    guide_id    UUID NOT NULL REFERENCES guides(id) ON DELETE CASCADE,
    group_id    UUID NOT NULL, -- departure group (kloter) in jamaah-service
    role        VARCHAR(20) NOT NULL DEFAULT 'leader', -- leader|co_leader|kesehatan
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, guide_id)
);
CREATE INDEX idx_guide_assignments_group ON guide_assignments(org_id, group_id);
CREATE INDEX idx_guide_assignments_guide ON guide_assignments(guide_id);

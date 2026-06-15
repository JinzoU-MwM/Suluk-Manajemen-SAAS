-- Visa & document lifecycle (Phase 4B).
-- A per-jamaah visa application with an explicit state machine + audit history,
-- a reminder-dedup ledger for expiry notifications, and the jamaah-service's
-- first transactional outbox (visa.* events).

CREATE TABLE IF NOT EXISTS visa_applications (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL,
    jamaah_id     UUID NOT NULL REFERENCES jamaah_profiles(id) ON DELETE CASCADE,
    package_id    UUID,
    status        VARCHAR(16) NOT NULL DEFAULT 'draft', -- draft|submitted|approved|rejected|expired
    provider      VARCHAR(255) NOT NULL DEFAULT '',
    reference_no  VARCHAR(100) NOT NULL DEFAULT '',
    submitted_at  TIMESTAMPTZ,
    decided_at    TIMESTAMPTZ,
    expiry_date   DATE,
    reject_reason VARCHAR(255) NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, jamaah_id)
);

CREATE INDEX IF NOT EXISTS idx_visa_org_status ON visa_applications(org_id, status);
-- Daily expiry scan: find approved visas nearing expiry.
CREATE INDEX IF NOT EXISTS idx_visa_expiry ON visa_applications(org_id, expiry_date) WHERE status = 'approved';

CREATE TABLE IF NOT EXISTS visa_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL,
    visa_id     UUID NOT NULL,
    jamaah_id   UUID NOT NULL,
    from_status VARCHAR(16),
    to_status   VARCHAR(16) NOT NULL,
    reason      VARCHAR(255),
    changed_by  UUID,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_visa_history_visa ON visa_history(visa_id, created_at);

-- One reminder per (subject, milestone) so the daily job never double-notifies.
CREATE TABLE IF NOT EXISTS lifecycle_reminders (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL,
    subject_type VARCHAR(16) NOT NULL, -- passport|visa
    subject_id   UUID NOT NULL,        -- jamaah_id
    milestone    VARCHAR(32) NOT NULL, -- e.g. visa_30, passport_7
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, subject_type, subject_id, milestone)
);

-- Transactional outbox (shared schema across producers) — jamaah-service's first.
CREATE TABLE IF NOT EXISTS outbox (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    aggregate_type VARCHAR(40) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type     VARCHAR(60) NOT NULL,
    payload        JSONB NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ,
    attempts       INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_outbox_unpublished ON outbox(occurred_at) WHERE published_at IS NULL;

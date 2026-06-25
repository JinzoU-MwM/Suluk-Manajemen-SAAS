-- Idempotency ledger for renewal reminders (Phase 5). One row per
-- (org, expiry cycle, threshold); a renewal = new expires_at = fresh rows.
CREATE TABLE IF NOT EXISTS subscription_reminders (
    org_id     UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    threshold  INT NOT NULL,          -- 7 | 3 | 1 days before expiry
    sent_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, expires_at, threshold)
);

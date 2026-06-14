-- Kasir POS: daily cash session (open/close + reconciliation).
CREATE TABLE cash_sessions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL,
    user_id       UUID NOT NULL,
    opening_float BIGINT NOT NULL DEFAULT 0,
    expected_cash BIGINT,
    counted_cash  BIGINT,
    difference    BIGINT,
    status        VARCHAR(8) NOT NULL DEFAULT 'open', -- open | closed
    opened_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at     TIMESTAMPTZ,
    notes         TEXT NOT NULL DEFAULT ''
);
-- At most one OPEN session per (org, user).
CREATE UNIQUE INDEX uq_cash_session_open ON cash_sessions(org_id, user_id) WHERE status = 'open';
CREATE INDEX idx_cash_sessions_org ON cash_sessions(org_id, opened_at);

-- Link cash payments to the session they were taken in.
ALTER TABLE payments ADD COLUMN cash_session_id UUID;
CREATE INDEX idx_payments_session ON payments(cash_session_id) WHERE cash_session_id IS NOT NULL;

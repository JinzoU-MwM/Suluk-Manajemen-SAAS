-- Makes savings-conversion settles idempotent: a retried request (double
-- click, or the shared httpclient's own automatic retry-on-5xx) with the
-- same (invoice_id, idempotency_key) must replay the prior result instead of
-- applying the credit a second time (finding T2).
CREATE TABLE settle_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    invoice_id UUID NOT NULL,
    idempotency_key VARCHAR(100) NOT NULL,
    applied_amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (invoice_id, idempotency_key)
);

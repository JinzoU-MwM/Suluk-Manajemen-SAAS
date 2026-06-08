CREATE TABLE payment_orders (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL,
    user_id        UUID NOT NULL,
    plan_type      VARCHAR(50) NOT NULL DEFAULT 'monthly',
    amount         BIGINT NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    redirect_url   TEXT,
    gateway_ref    VARCHAR(255),
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payment_orders_org_id ON payment_orders(org_id);
CREATE INDEX idx_payment_orders_status ON payment_orders(status);

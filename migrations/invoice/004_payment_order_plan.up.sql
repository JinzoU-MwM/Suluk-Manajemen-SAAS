-- Carry the chosen subscription tier (plan) on each payment order, plus webhook
-- bookkeeping. plan_type continues to hold the billing period (monthly/yearly).
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS plan VARCHAR(50) NOT NULL DEFAULT 'pro';
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50);
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ;

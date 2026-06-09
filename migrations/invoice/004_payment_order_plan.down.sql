ALTER TABLE payment_orders DROP COLUMN IF EXISTS completed_at;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS payment_method;
ALTER TABLE payment_orders DROP COLUMN IF EXISTS plan;

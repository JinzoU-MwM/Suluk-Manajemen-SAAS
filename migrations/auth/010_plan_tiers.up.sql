-- Move subscriptions onto the 5-tier plan model (gratis/starter/pro/bisnis/enterprise).
-- Default a brand-new org to the free tier, normalize legacy names, and constrain values.

ALTER TABLE subscriptions ALTER COLUMN plan SET DEFAULT 'gratis';

UPDATE subscriptions SET plan = 'gratis' WHERE plan = 'free';
UPDATE subscriptions SET plan = 'bisnis' WHERE plan = 'business';

ALTER TABLE subscriptions
    ADD CONSTRAINT subscriptions_plan_check
    CHECK (plan IN ('gratis', 'starter', 'pro', 'bisnis', 'enterprise'));

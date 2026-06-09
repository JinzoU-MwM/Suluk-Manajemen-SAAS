ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_plan_check;
ALTER TABLE subscriptions ALTER COLUMN plan SET DEFAULT 'starter';

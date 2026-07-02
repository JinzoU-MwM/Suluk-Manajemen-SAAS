DROP INDEX IF EXISTS uq_savings_deposits_idempotency;
ALTER TABLE savings_deposits DROP COLUMN idempotency_key;

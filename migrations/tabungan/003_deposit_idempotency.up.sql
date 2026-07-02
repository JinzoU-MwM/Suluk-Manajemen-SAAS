-- Retry/double-click protection for deposits: a client-generated key, deduped
-- per account under the same row lock DepositTx already takes — no key (NULL)
-- from older/other callers is never compared against another NULL, so this
-- is opt-in and backward compatible (finding T4).
ALTER TABLE savings_deposits ADD COLUMN idempotency_key VARCHAR(64);
CREATE UNIQUE INDEX uq_savings_deposits_idempotency ON savings_deposits(account_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

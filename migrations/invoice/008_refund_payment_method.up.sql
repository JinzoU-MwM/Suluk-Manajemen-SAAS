-- Refunds need to know which cash/bank account to credit back at CompleteRefund
-- time; previously this was always missing from the accounting payload, so
-- every refund posted as a Bank outflow even when the original payment was
-- cash (tunai).
ALTER TABLE refunds ADD COLUMN payment_method VARCHAR(30) NOT NULL DEFAULT 'transfer_bank';

-- Backfill existing rows from their invoice's most recent payment, not just
-- the schema default — otherwise a refund already in flight (pending/
-- approved/processed) at deploy time would complete with the wrong account
-- once CompleteRefund starts reading this column, reproducing the very bug
-- this migration exists to fix.
UPDATE refunds SET payment_method = COALESCE(
    (SELECT p.payment_method FROM payments p WHERE p.invoice_id = refunds.invoice_id ORDER BY p.paid_at DESC LIMIT 1),
    'transfer_bank'
);

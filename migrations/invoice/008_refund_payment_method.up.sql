-- Refunds need to know which cash/bank account to credit back at CompleteRefund
-- time; previously this was always missing from the accounting payload, so
-- every refund posted as a Bank outflow even when the original payment was
-- cash (tunai).
ALTER TABLE refunds ADD COLUMN payment_method VARCHAR(30) NOT NULL DEFAULT 'transfer_bank';

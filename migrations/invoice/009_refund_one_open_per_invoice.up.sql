-- Enforce "at most one open refund per invoice" at the database, not just in
-- application code — a service-layer check-then-act guard alone can't close
-- the race window (two concurrent initiate calls, or a client retry after a
-- lost response, can both pass a pre-check before either commits its INSERT).
-- "Open" = the refund hasn't reached a terminal state yet.
CREATE UNIQUE INDEX uq_refunds_one_open_per_invoice ON refunds(invoice_id)
	WHERE status IN ('pending', 'approved', 'processed');

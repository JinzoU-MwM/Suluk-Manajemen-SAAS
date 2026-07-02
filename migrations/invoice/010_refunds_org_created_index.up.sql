-- Org-wide refund lists (GET /refunds, ORDER BY created_at DESC filtered by
-- org_id) only had single-column indexes on org_id/invoice_id/status to work
-- with — add the composite index the actual query pattern needs.
CREATE INDEX IF NOT EXISTS idx_refunds_org_created ON refunds(org_id, created_at DESC);

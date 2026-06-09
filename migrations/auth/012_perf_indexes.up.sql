-- Performance index (audit P8): GetAuditLogsByUser does
-- WHERE user_id = $1 ORDER BY created_at DESC. A composite index serves both the
-- filter and the sort, avoiding a separate sort step.
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);

ALTER TABLE payments DROP COLUMN IF EXISTS cash_session_id;
DROP TABLE IF EXISTS cash_sessions;

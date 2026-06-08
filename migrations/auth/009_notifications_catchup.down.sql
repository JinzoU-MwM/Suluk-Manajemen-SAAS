-- No-op: the notifications table is logically owned by migration 002.
-- This catch-up only ensures it exists, so rolling back one step must not drop
-- a table that 002 created. Roll back 002 to remove the table.
SELECT 1;

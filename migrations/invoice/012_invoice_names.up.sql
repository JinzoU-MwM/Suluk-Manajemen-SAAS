-- jamaah/package names live in other services' databases (this is a
-- microservices architecture — invoice-service can't JOIN across them), so
-- every screen that tried to display a jamaah/package name on an invoice was
-- silently falling back to a hardcoded placeholder. Snapshot the names at
-- invoice-creation time instead, same pattern already used for
-- price_snapshot, agent_commissions.jamaah_name, and
-- tabungan.savings_accounts.jamaah_name.
ALTER TABLE invoices ADD COLUMN jamaah_name TEXT NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN package_name TEXT NOT NULL DEFAULT '';

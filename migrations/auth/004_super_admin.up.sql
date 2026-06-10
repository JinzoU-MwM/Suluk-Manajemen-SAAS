ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE users SET is_super_admin = TRUE WHERE email = 'admin@suluk.site';

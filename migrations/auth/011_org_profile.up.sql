-- Company/legal profile fields for the Perusahaan tab (editable in Profil).
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS npwp VARCHAR(50);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS ppiu_number VARCHAR(100);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS sk_number VARCHAR(100);

-- "Sign in with Google": link a user to a Google account.
-- google_sub is the Google account's stable subject id. Nullable — only set for
-- users who signed in via Google; password users keep it NULL. Google-only users
-- carry password_hash = '' (a value bcrypt can never match), so password login is
-- effectively disabled for them without making the column nullable.
ALTER TABLE users ADD COLUMN IF NOT EXISTS google_sub TEXT;

-- Partial unique index: one user per Google account, but many NULLs allowed.
CREATE UNIQUE INDEX IF NOT EXISTS users_google_sub_key
    ON users (google_sub) WHERE google_sub IS NOT NULL;

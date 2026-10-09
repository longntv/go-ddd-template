-- The column stores a bcrypt hash, never the plain-text password.
ALTER TABLE users RENAME COLUMN password TO password_hash;

-- Rows written before this migration hold plain text. Hash them in place so
-- they can still log in once a login use case compares against the hash.
-- pgcrypto's 'bf' produces $2a$ hashes that golang.org/x/crypto/bcrypt accepts.
-- Rows that already look like a bcrypt hash ($2a$/$2b$/$2y$) are left alone.
--
-- Deploying: the old binary writes `password` and the new one `password_hash`,
-- so stop the old version before migrating (no rolling deploy across this step).
CREATE EXTENSION IF NOT EXISTS pgcrypto;

UPDATE users
SET password_hash = crypt(password_hash, gen_salt('bf', 10))
WHERE password_hash <> ''
  AND password_hash !~ '^\$2[aby]\$';

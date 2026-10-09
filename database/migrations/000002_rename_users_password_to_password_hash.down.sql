-- The hashes cannot be turned back into plain text; the column keeps them.
ALTER TABLE users RENAME COLUMN password_hash TO password;

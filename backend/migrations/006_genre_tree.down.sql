-- Revert migration 006: Genre tree support

DROP INDEX IF EXISTS idx_genres_position_pattern;
DROP INDEX IF EXISTS idx_genres_code;
DROP INDEX IF EXISTS idx_genres_position;
ALTER TABLE genres DROP COLUMN IF EXISTS is_active;
ALTER TABLE genres DROP COLUMN IF EXISTS sort_order;
ALTER TABLE genres DROP COLUMN IF EXISTS position;
ALTER TABLE genres ADD CONSTRAINT genres_code_key UNIQUE (code);
DROP TABLE IF EXISTS app_metadata;

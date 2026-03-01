-- Migration 006: Genre tree support
-- Adds hierarchical genre structure via materialized path (position column)

-- 1. Table app_metadata for storing genre tree hash and other service metadata
CREATE TABLE IF NOT EXISTS app_metadata (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Add position, sort_order, is_active columns to genres
ALTER TABLE genres ADD COLUMN position VARCHAR(50);
ALTER TABLE genres ADD COLUMN sort_order INTEGER DEFAULT 0;
ALTER TABLE genres ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- 3. Drop UNIQUE constraint on code (duplicates allowed on child levels)
ALTER TABLE genres DROP CONSTRAINT genres_code_key;

-- 4. Create UNIQUE index on position (partial: only non-null)
CREATE UNIQUE INDEX idx_genres_position ON genres(position) WHERE position IS NOT NULL;

-- 5. Create regular index on code (for fast lookup during INPX mapping)
CREATE INDEX idx_genres_code ON genres(code);

-- 6. B-tree index on position for LIKE prefix search (cascading filter)
-- PostgreSQL B-tree supports LIKE 'prefix%' with text_pattern_ops
CREATE INDEX idx_genres_position_pattern ON genres(position text_pattern_ops);

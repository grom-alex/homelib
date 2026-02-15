-- Add missing unique constraints required for ON CONFLICT upserts

-- Authors: unique on name_sort for idempotent import
CREATE UNIQUE INDEX IF NOT EXISTS idx_authors_name_sort_unique ON authors (name_sort);

-- Series: unique on name for idempotent import
CREATE UNIQUE INDEX IF NOT EXISTS idx_series_name_unique ON series (name);

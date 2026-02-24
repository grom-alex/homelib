CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_authors_name_trgm ON authors USING gin (name gin_trgm_ops);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_series_name_trgm ON series USING gin (name gin_trgm_ops);

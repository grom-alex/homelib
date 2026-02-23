CREATE TABLE reading_progress (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  chapter_id TEXT NOT NULL DEFAULT '',
  chapter_progress SMALLINT NOT NULL DEFAULT 0
    CHECK (chapter_progress BETWEEN 0 AND 100),
  total_progress SMALLINT NOT NULL DEFAULT 0
    CHECK (total_progress BETWEEN 0 AND 100),
  device TEXT DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, book_id)
);

CREATE INDEX idx_reading_progress_user ON reading_progress(user_id);

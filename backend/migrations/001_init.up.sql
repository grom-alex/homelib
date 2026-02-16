-- Extensions
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- === Catalog ===

CREATE TABLE authors (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    name_sort   TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_authors_name_sort ON authors (name_sort);
CREATE INDEX idx_authors_name_trgm ON authors USING gin (name gin_trgm_ops);

CREATE TABLE genres (
    id          SERIAL PRIMARY KEY,
    code        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    parent_id   INTEGER REFERENCES genres(id),
    meta_group  TEXT
);
CREATE INDEX idx_genres_parent ON genres (parent_id) WHERE parent_id IS NOT NULL;

CREATE TABLE collections (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    code            TEXT UNIQUE NOT NULL,
    collection_type INTEGER DEFAULT 0,
    description     TEXT,
    source_url      TEXT,
    version         TEXT,
    version_date    DATE,
    books_count     INTEGER DEFAULT 0,
    last_import_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE series (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_series_name_trgm ON series USING gin (name gin_trgm_ops);

CREATE TABLE books (
    id              BIGSERIAL PRIMARY KEY,
    collection_id   INTEGER REFERENCES collections(id),
    title           TEXT NOT NULL,
    lang            TEXT NOT NULL DEFAULT 'ru',
    year            INTEGER,
    format          TEXT NOT NULL,
    file_size       BIGINT,
    archive_name    TEXT NOT NULL,
    file_in_archive TEXT NOT NULL,
    series_id       BIGINT REFERENCES series(id),
    series_num      INTEGER,
    series_type     CHAR(1),
    lib_id          TEXT,
    lib_rate        SMALLINT,
    is_deleted      BOOLEAN DEFAULT FALSE,
    has_cover       BOOLEAN DEFAULT FALSE,
    description     TEXT,
    keywords        TEXT[],
    date_added      DATE,
    search_vector   tsvector,
    added_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (collection_id, lib_id)
);
CREATE INDEX idx_books_title_trgm ON books USING gin (title gin_trgm_ops);
CREATE INDEX idx_books_lang       ON books (lang);
CREATE INDEX idx_books_format     ON books (format);
CREATE INDEX idx_books_archive    ON books (archive_name);
CREATE INDEX idx_books_search     ON books USING gin (search_vector);
CREATE INDEX idx_books_collection ON books (collection_id);
CREATE INDEX idx_books_series     ON books (series_id) WHERE series_id IS NOT NULL;
CREATE INDEX idx_books_lib_rate   ON books (lib_rate) WHERE lib_rate IS NOT NULL;
CREATE INDEX idx_books_keywords   ON books USING gin (keywords) WHERE keywords IS NOT NULL;

-- Auto-update tsvector on book changes
CREATE OR REPLACE FUNCTION books_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('russian', coalesce(array_to_string(NEW.keywords, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_books_search_vector
    BEFORE INSERT OR UPDATE OF title, description, keywords ON books
    FOR EACH ROW EXECUTE FUNCTION books_search_vector_update();

-- M:N relationships
CREATE TABLE book_authors (
    book_id     BIGINT REFERENCES books(id) ON DELETE CASCADE,
    author_id   BIGINT REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);
CREATE INDEX idx_book_authors_author ON book_authors (author_id);

CREATE TABLE book_genres (
    book_id     BIGINT REFERENCES books(id) ON DELETE CASCADE,
    genre_id    INTEGER REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, genre_id)
);
CREATE INDEX idx_book_genres_genre ON book_genres (genre_id);

-- === Users & Auth ===

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    username        TEXT UNIQUE NOT NULL,
    display_name    TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    role            user_role NOT NULL DEFAULT 'user',
    is_active       BOOLEAN DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_username ON users (username);

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens (token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens (expires_at);

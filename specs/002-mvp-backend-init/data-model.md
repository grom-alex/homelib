# Data Model: MVP HomeLib

**Branch**: `002-mvp-backend-init` | **Date**: 2026-02-15

## Entities

### authors

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PK | |
| name | TEXT | NOT NULL | Полное имя: "Фамилия Имя Отчество" |
| name_sort | TEXT | NOT NULL | Сортируемое: "Фамилия, Имя" |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |

**Indexes**:
- `idx_authors_name_sort` — B-tree on `name_sort` (сортировка)
- `idx_authors_name_trgm` — GIN(gin_trgm_ops) on `name` (нечёткий поиск)

---

### genres

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PK | |
| code | TEXT | UNIQUE, NOT NULL | sf_fantasy, detective, etc. |
| name | TEXT | NOT NULL | Человекочитаемое название |
| parent_id | INTEGER | FK → genres(id) | Иерархия жанров |
| meta_group | TEXT | | Группировка: Фантастика, Детектив... |

---

### collections

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | SERIAL | PK | |
| name | TEXT | NOT NULL | Название коллекции |
| code | TEXT | UNIQUE, NOT NULL | Идентификатор (имя файла без .inpx) |
| collection_type | INTEGER | DEFAULT 0 | 0=fb2, 1=non-fb2, др.=флаги |
| description | TEXT | | Описание / статистика |
| source_url | TEXT | | URL источника |
| version | TEXT | | Версия из version.info (YYYYMMDD) |
| version_date | DATE | | Версия как дата |
| books_count | INTEGER | DEFAULT 0 | Кол-во книг в коллекции |
| last_import_at | TIMESTAMPTZ | | Время последнего импорта |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

---

### series

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PK | |
| name | TEXT | NOT NULL | Название серии |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |

**Indexes**:
- `idx_series_name_trgm` — GIN(gin_trgm_ops) on `name` (нечёткий поиск)

---

### books

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | BIGSERIAL | PK | |
| collection_id | INTEGER | FK → collections(id) | Принадлежность к коллекции |
| title | TEXT | NOT NULL | Название книги |
| lang | TEXT | NOT NULL, DEFAULT 'ru' | Язык (ru, en, uk...) |
| year | INTEGER | | Год издания |
| format | TEXT | NOT NULL | fb2, epub, pdf, djvu |
| file_size | BIGINT | | Размер файла в байтах |
| archive_name | TEXT | NOT NULL | Имя ZIP-архива (fb2-012345-012456.zip) |
| file_in_archive | TEXT | NOT NULL | Путь файла внутри архива (123456.fb2) |
| series_id | BIGINT | FK → series(id) | Серия |
| series_num | INTEGER | | Номер в серии |
| series_type | CHAR(1) | | 'a'=авторская, 'p'=издательская, NULL |
| lib_id | TEXT | | ID из .inpx (уникален в collection) |
| lib_rate | SMALLINT | | Рейтинг из библиотеки (1-5) |
| is_deleted | BOOLEAN | DEFAULT FALSE | Помечена как удалённая (DEL=1) |
| has_cover | BOOLEAN | DEFAULT FALSE | Есть обложка |
| description | TEXT | | Аннотация из INPX |
| keywords | TEXT[] | | Ключевые слова из INPX |
| date_added | DATE | | Дата добавления в библиотеку (из DATE) |
| search_vector | tsvector | | Полнотекстовый поиск (auto-updated) |
| added_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

**Constraints**:
- `UNIQUE (collection_id, lib_id)` — уникальность в пределах коллекции (idempotent import)

**Indexes**:
- `idx_books_title_trgm` — GIN(gin_trgm_ops) on `title` (нечёткий поиск)
- `idx_books_lang` — B-tree on `lang` (фильтрация)
- `idx_books_format` — B-tree on `format` (фильтрация)
- `idx_books_archive` — B-tree on `archive_name` (download lookup)
- `idx_books_search` — GIN on `search_vector` (полнотекстовый поиск)
- `idx_books_collection` — B-tree on `collection_id`
- `idx_books_lib_rate` — B-tree on `lib_rate` WHERE lib_rate IS NOT NULL
- `idx_books_keywords` — GIN on `keywords` WHERE keywords IS NOT NULL

**Trigger**: `trg_books_search_vector` — BEFORE INSERT OR UPDATE OF title, description, keywords
- Обновляет `search_vector` автоматически
- Weight A: title, Weight B: description, Weight C: keywords
- Конфигурация tsvector: `'russian'` (основной), с поддержкой `'english'`

---

### book_authors (M:N)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| book_id | BIGINT | FK → books(id) ON DELETE CASCADE | |
| author_id | BIGINT | FK → authors(id) ON DELETE CASCADE | |

**PK**: (book_id, author_id)
**Indexes**: `idx_book_authors_author` on `author_id`

---

### book_genres (M:N)

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| book_id | BIGINT | FK → books(id) ON DELETE CASCADE | |
| genre_id | INTEGER | FK → genres(id) ON DELETE CASCADE | |

**PK**: (book_id, genre_id)
**Indexes**: `idx_book_genres_genre` on `genre_id`

---

### users

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| email | TEXT | UNIQUE, NOT NULL | |
| username | TEXT | UNIQUE, NOT NULL | |
| display_name | TEXT | NOT NULL | Отображаемое имя |
| password_hash | TEXT | NOT NULL | bcrypt hash |
| role | user_role | NOT NULL, DEFAULT 'user' | ENUM: 'user', 'admin' |
| is_active | BOOLEAN | DEFAULT TRUE | Активен ли аккаунт |
| last_login_at | TIMESTAMPTZ | | Последний вход |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |
| updated_at | TIMESTAMPTZ | DEFAULT NOW() | |

**Enum**: `user_role` = ('user', 'admin')
**Indexes**:
- `idx_users_email` on `email`
- `idx_users_username` on `username`

**Note**: Поля `avatar_url`, `settings` (JSONB) из полной архитектуры НЕ включены в MVP.

---

### refresh_tokens

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK, DEFAULT gen_random_uuid() | |
| user_id | UUID | FK → users(id) ON DELETE CASCADE, NOT NULL | |
| token_hash | TEXT | NOT NULL | SHA-256 хеш токена |
| expires_at | TIMESTAMPTZ | NOT NULL | Время истечения (30 дней) |
| created_at | TIMESTAMPTZ | DEFAULT NOW() | |

**Indexes**:
- `idx_refresh_tokens_user` on `user_id`
- `idx_refresh_tokens_expires` on `expires_at`

---

## Relationships

```text
authors ──M:N──▶ book_authors ◀──M:N── books
genres  ──M:N──▶ book_genres  ◀──M:N── books
series  ──1:N──▶ books
collections ──1:N──▶ books
users ──1:N──▶ refresh_tokens
```

## State Transitions

### Import Process

```text
INPX file → Parse → Batch Upsert
  └─ authors: INSERT ON CONFLICT DO NOTHING → return id
  └─ genres:  INSERT ON CONFLICT DO NOTHING → return id
  └─ series:  INSERT ON CONFLICT DO NOTHING → return id
  └─ books:   INSERT ON CONFLICT (collection_id, lib_id) DO UPDATE
  └─ M:N:     DELETE + INSERT for book_authors, book_genres
```

### Book Deletion (soft)

```text
DEL=1 in INPX → books.is_deleted = true (физически НЕ удаляется)
```

## Validation Rules

- **authors.name**: NOT NULL, непустая строка
- **books.title**: NOT NULL, непустая строка
- **books.format**: NOT NULL, одно из: fb2, epub, pdf, djvu, doc, txt, rtf, htm, html, mobi, azw3
- **books.file_in_archive**: NOT NULL, непустая строка
- **books.archive_name**: NOT NULL, непустая строка
- **users.email**: UNIQUE, NOT NULL, валидный email формат
- **users.username**: UNIQUE, NOT NULL, 3-50 символов, [a-zA-Z0-9_-]
- **users.password_hash**: NOT NULL (пароль до хеширования: минимум 8 символов)
- **refresh_tokens.expires_at**: MUST быть в будущем при создании

## Extensions Required

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;  -- нечёткий поиск
-- pgvector НЕ используется в MVP
```

## Not Included in MVP

Следующие таблицы из полной архитектуры НЕ входят в MVP:
- `book_summaries` — summary embedding (pgvector)
- `llm_summary_tasks` — LLM-очередь
- `user_books` — статусы книг у пользователя
- `reading_progress` — прогресс чтения
- `shelves` — книжные полки
- `shelf_books` — книги на полках

Эти таблицы будут добавлены в последующих итерациях.

# Data Model: Древовидная структура жанров

**Feature**: 008-genre-tree | **Date**: 2026-03-01

## Migration 006: genre_tree

### Файл: `backend/migrations/006_genre_tree.up.sql`

```sql
-- 1. Таблица app_metadata для хранения хеша дерева жанров
CREATE TABLE IF NOT EXISTS app_metadata (
    key        VARCHAR(100) PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Добавить колонки position, sort_order, is_active в genres
ALTER TABLE genres ADD COLUMN position VARCHAR(50);
ALTER TABLE genres ADD COLUMN sort_order INTEGER DEFAULT 0;
ALTER TABLE genres ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- 3. Снять UNIQUE constraint с code (дубликаты допускаются на дочерних уровнях)
ALTER TABLE genres DROP CONSTRAINT genres_code_key;

-- 4. Создать UNIQUE индекс на position
CREATE UNIQUE INDEX idx_genres_position ON genres(position) WHERE position IS NOT NULL;

-- 5. Создать обычный индекс на code (для быстрого lookup при маппинге)
CREATE INDEX idx_genres_code ON genres(code);

-- 6. B-tree индекс на position для LIKE prefix search (каскадная фильтрация)
-- PostgreSQL B-tree поддерживает LIKE 'prefix%' с text_pattern_ops
CREATE INDEX idx_genres_position_pattern ON genres(position text_pattern_ops);
```

### Файл: `backend/migrations/006_genre_tree.down.sql`

```sql
DROP INDEX IF EXISTS idx_genres_position_pattern;
DROP INDEX IF EXISTS idx_genres_code;
DROP INDEX IF EXISTS idx_genres_position;
ALTER TABLE genres DROP COLUMN IF EXISTS is_active;
ALTER TABLE genres DROP COLUMN IF EXISTS sort_order;
ALTER TABLE genres DROP COLUMN IF EXISTS position;
ALTER TABLE genres ADD CONSTRAINT genres_code_key UNIQUE (code);
DROP TABLE IF EXISTS app_metadata;
```

## Обновлённые сущности

### Genre (Go model)

```go
type Genre struct {
    ID        int     `json:"id"`
    Code      string  `json:"code"`
    Name      string  `json:"name"`
    ParentID  *int    `json:"parent_id,omitempty"`
    MetaGroup string  `json:"meta_group,omitempty"`
    Position  string  `json:"position"`              // NEW: "0.1.2"
    SortOrder int     `json:"sort_order"`             // NEW: порядок в дереве
    IsActive  bool    `json:"is_active"`              // NEW: активен ли жанр (false = удалён из .glst)
}

type GenreTreeItem struct {
    ID         int             `json:"id"`
    Code       string          `json:"code"`
    Name       string          `json:"name"`
    MetaGroup  string          `json:"meta_group,omitempty"`
    Position   string          `json:"position"`        // NEW
    BooksCount int             `json:"books_count"`
    Children   []GenreTreeItem `json:"children,omitempty"`
}
```

### GenreEntry (GLST parser output)

```go
// backend/internal/glst/types.go
type GenreEntry struct {
    Position string // "0.1.2" — путь в дереве
    Code     string // "sf_action" — INPX-код (может быть пустым)
    Name     string // "Боевая фантастика"
    Level    int    // 0, 1, 2, 3 — вычисляется из position
    ParentPosition string // "0.1" — вычисляется из position
}

type ParseResult struct {
    Entries  []GenreEntry
    Warnings []string // Логи о пропущенных строках
}
```

### BookFilter (расширение)

Без изменений структуры. Поле `GenreID *int` сохраняется — логика каскадной фильтрации реализуется в repository слое (CTE с materialized path).

### GenreTreeConfig (конфигурация)

```go
// Добавляется в Config
type GenreTreeConfig struct {
    FilePath string `yaml:"file_path"` // Путь к .glst файлу (override для embedded)
}
```

### app_metadata

| key | value | Описание |
|-----|-------|----------|
| `genre_tree_hash` | SHA-256 hex string | Хеш загруженного `.glst` файла |

## Схема таблицы genres (после миграции)

```
genres
├── id          SERIAL PRIMARY KEY
├── code        TEXT NOT NULL           -- INPX-код (НЕ уникальный)
├── name        TEXT NOT NULL           -- Человекочитаемое название
├── parent_id   INTEGER → genres(id)    -- Ссылка на родителя
├── meta_group  TEXT                    -- Группа (устаревшее, заменяется position)
├── position    VARCHAR(50) UNIQUE      -- Позиция в дереве: "0.1.2"
├── sort_order  INTEGER DEFAULT 0       -- Порядок отображения
└── is_active   BOOLEAN DEFAULT TRUE    -- Активен (false = удалён из .glst, но имеет привязанные книги)
```

**Индексы**:
- `idx_genres_parent` — B-tree на `parent_id` (существующий)
- `idx_genres_position` — UNIQUE на `position` (новый)
- `idx_genres_code` — B-tree на `code` (новый, заменяет старый UNIQUE)
- `idx_genres_position_pattern` — B-tree `text_pattern_ops` для `LIKE 'prefix%'` (новый)

## Связи (без изменений)

```
book_genres
├── book_id   BIGINT → books(id)
└── genre_id  INTEGER → genres(id)
PRIMARY KEY (book_id, genre_id)
```

Изменение: одна книга теперь может быть привязана к нескольким жанрам с одинаковым кодом (разные position). Например, книга с кодом `home_cooking` будет иметь записи в `book_genres` для genre.id=X (position=0.14.1) и genre.id=Y (position=0.24.4).

## Поток данных: загрузка GLST

```
1. Старт приложения
   ↓
2. GenreTreeService.LoadIfNeeded(ctx)
   ├─ Читает embedded/file .glst
   ├─ Вычисляет SHA-256 хеш
   ├─ Сравнивает с app_metadata['genre_tree_hash']
   ├─ Если совпадает → SKIP
   └─ Если отличается:
      ↓
3. glst.Parse(reader) → ParseResult
   ├─ Парсит строки, пропускает комментарии и пустые
   ├─ Валидирует: orphaned children, root-level code duplicates
   └─ Генерирует код для записей без кода (_root_0, etc.)
      ↓
4. GenreRepo.LoadTree(ctx, entries []GenreEntry) — одна транзакция
   ├─ UPDATE genres SET is_active = FALSE (сбросить все перед загрузкой)
   ├─ Определяет parent_id по ParentPosition → position lookup
   ├─ UPSERT: INSERT ... ON CONFLICT (position) DO UPDATE SET ..., is_active = TRUE
   └─ Устанавливает sort_order по порядку в файле
   → Жанры, отсутствующие в файле, останутся is_active=FALSE
      ↓
5. GenreTreeService.RemapBooks(ctx) — batch транзакции
   ├─ SELECT DISTINCT g.code, g.id FROM genres → codeToIDs map
   ├─ Находит ID жанра «Неотсортированное» (position='0.0')
   ├─ Для каждого батча книг (3000):
   │  ├─ SELECT bg.book_id, g.code FROM book_genres bg JOIN genres g
   │  ├─ Вычисляет новые genre_ids по code → codeToIDs
   │  ├─ Книги без маппинга → Неотсортированное
   │  └─ DELETE + INSERT book_genres
   └─ Обновляет app_metadata['genre_tree_hash']
```

## Поток данных: INPX-импорт (изменённый)

```
1. ImportService.processBatch(records)
   ↓
2. Collect unique genre codes from records
   ↓
3. GenreRepo.GetIDsByCodes(ctx, codes) → map[string][]int
   ├─ SELECT code, id FROM genres WHERE code = ANY($1)
   └─ Один код может вернуть несколько ID (дубликаты)
   ↓
4. Для каждой книги:
   ├─ Собрать все genre_ids для её кодов (flatten)
   ├─ Если кодов нет или все неизвестные → добавить ID «Неотсортированное»
   └─ BatchSetBookGenres(bookGenres)
```

## Поток данных: каскадная фильтрация (GET /api/books?genre_id=X)

```
1. BookRepo.List(filter{GenreID: X})
   ↓
2. SQL с CTE:
   WITH target AS (
     SELECT id FROM genres
     WHERE id = X
        OR position LIKE (SELECT position || '.%' FROM genres WHERE id = X)
   )
   SELECT ... FROM books b
   WHERE EXISTS (
     SELECT 1 FROM book_genres bg
     WHERE bg.book_id = b.id AND bg.genre_id IN (SELECT id FROM target)
   )
```

# Research: Древовидная структура жанров

**Feature**: 008-genre-tree | **Date**: 2026-03-01

## R1: Миграция таблицы genres — снятие UNIQUE с code

### Проблема

Текущая схема: `code TEXT UNIQUE NOT NULL`. Спецификация требует дубликаты кодов на дочерних уровнях (13 кодов встречаются в нескольких позициях дерева).

### Решение: Materialized Path по position

- **Decision**: Добавить колонку `position VARCHAR(50) UNIQUE`, снять `UNIQUE` с `code`, добавить обычный индекс на `code`.
- **Rationale**: `position` (например `0.1.2`) является естественным уникальным идентификатором узла в дереве. Позволяет эффективную каскадную фильтрацию через `LIKE 'prefix.%'`.
- **Alternatives considered**:
  - **Nested sets** (lft/rgt): эффективно для чтения, но дорого при вставке/обновлении. Дерево жанров обновляется редко, но nested sets добавляют ненужную сложность.
  - **Recursive CTE** (только parent_id): корректно, но медленнее для каскадной фильтрации книг (рекурсия на каждый запрос). Не масштабируется для JOIN с book_genres.
  - **ltree extension**: PostgreSQL extension, мощный, но добавляет зависимость. Materialized path на `VARCHAR` достаточен для 4 уровней.

### Влияние на UpsertGenres

Текущий `ON CONFLICT (code)` перестанет работать. Варианты:
- **INPX-импорт**: больше не вставляет жанры (они уже загружены из GLST). Вместо upsert — lookup `GetIDsByCodes(codes) → map[string][]int`.
- **GLST-загрузка**: использует `ON CONFLICT (position) DO UPDATE`.

## R2: Каскадная фильтрация книг по жанру

### Проблема

При выборе жанра нужно показать книги этого жанра И всех потомков. Текущий запрос: `WHERE bg.genre_id = $1`.

### Решение: Фильтрация через materialized path

```sql
-- Найти все genre_id для выбранного жанра и его потомков
WITH target_genres AS (
    SELECT id FROM genres
    WHERE position = (SELECT position FROM genres WHERE id = $1)
       OR position LIKE (SELECT position || '.%' FROM genres WHERE id = $1)
)
SELECT DISTINCT b.* FROM books b
JOIN book_genres bg ON bg.book_id = b.id
WHERE bg.genre_id IN (SELECT id FROM target_genres)
```

- **Decision**: Использовать CTE с materialized path для каскадного фильтра.
- **Rationale**: Один SQL-запрос, использует B-tree индекс на `position`, нет рекурсии.
- **Performance**: Для 448 жанров даже полный скан таблицы `genres` — микросекунды. Индекс на `position` делает `LIKE 'prefix.%'` эффективным (prefix search).

### Альтернатива: Предвычисление descendant_ids

Хранить массив ID потомков для каждого жанра. Отвергнуто: добавляет сложность синхронизации при обновлении дерева.

## R3: Встраивание файла .glst через go:embed

### Решение

```go
// backend/internal/glst/embed.go
package glst

import "embed"

//go:embed genres_all.glst
var DefaultGenreFile []byte
```

- **Decision**: Встроить файл через `//go:embed` в пакет `glst`. Файл `docs/genres_all.glst` копируется (или симлинкится) в `backend/internal/glst/genres_all.glst` при сборке.
- **Rationale**: Go `embed` требует файл в том же пакете или поддереве. Копирование при сборке — стандартный паттерн.
- **Alternative**: Читать файл из файловой системы по пути из конфига. Гибче, но требует монтирования файла в Docker-контейнер. Решение: поддержать оба варианта — embedded по умолчанию, `genre_tree.file_path` в конфиге для override.

## R4: Vuetify 3 VTreeview

### Исследование

Vuetify 3.8 содержит `VTreeview` (восстановлен из labs, значительные улучшения производительности для больших деревьев — release notes октябрь/ноябрь 2025).

### Решение

- **Decision**: Использовать `v-treeview` из Vuetify 3.8 для GenresTab и фильтра на SearchTab.
- **Ключевые props**:
  - `items` — массив узлов дерева
  - `item-value` / `item-title` — маппинг полей
  - `search` — встроенный поиск по дереву (FR-009)
  - `activatable` — выбор узла для навигации
  - `open-on-click` — раскрытие по клику
  - Слот `#prepend` для кастомных иконок и счётчиков
- **Для SearchTab** (FR-010): `VTreeview` внутри `VMenu` / `VDialog` как dropdown tree selector.
- **Alternative**: Кастомный рекурсивный компонент. Отвергнуто — Vuetify VTreeview уже решает задачу, включая a11y, виртуализацию и поиск.

## R5: Идемпотентная загрузка дерева жанров

### Проблема

FR-013: повторная загрузка не должна создавать дубликатов и терять привязки книг.

### Решение: Hash-based versioning + UPSERT

1. Вычислить SHA-256 хеш содержимого `.glst` файла.
2. Сравнить с хранимым хешем в `app_metadata` (key=`genre_tree_hash`).
3. Если хеш совпадает — пропустить загрузку.
4. Если отличается или отсутствует:
   a. Парсить `.glst` → `[]GenreEntry`
   b. В одной транзакции:
      - `INSERT INTO genres ... ON CONFLICT (position) DO UPDATE SET code=, name=, parent_id=, sort_order=`
      - Жанры, удалённые из файла: НЕ удалять (у них могут быть привязанные книги), но пометить как неактивные (фильтровать в UI)
   c. Ремаппинг книг: `DELETE FROM book_genres` + batch insert через INPX-коды → genre IDs
   d. Обновить хеш в `app_metadata`

- **Decision**: Hash-based check + UPSERT по position + полный ремаппинг book_genres.
- **Rationale**: UPSERT сохраняет genre.id стабильным при обновлении. Полный ремаппинг гарантирует корректность для дубликатов кодов.
- **Risk**: Ремаппинг 600K книг может быть медленным. Mitigation: batch processing (3000-5000 за транзакцию), оптимистичный подход — ремаппинг только если хеш изменился.

## R6: Влияние на INPX-импорт

### Текущий процесс

```
INPX records → collect unique genre codes → UpsertGenres(codes) → map[code]int → BatchSetBookGenres
```

### Новый процесс

```
INPX records → collect unique genre codes → GetIDsByCodes(codes) → map[code][]int → BatchSetBookGenres (multiple IDs per code)
```

### Ключевые изменения

1. **UpsertGenres → GetIDsByCodes**: Жанры уже существуют в БД (загружены из GLST). Импорт только ищет ID по коду.
2. **Один код → несколько ID**: Для дубликатов кодов (13 шт.) одна книга привязывается ко всем позициям.
3. **Неизвестный код → «Неотсортированное»**: Если код не найден в дереве, книга привязывается к жанру с position `0.0`.
4. **Книги без жанра**: Аналогично — привязка к `0.0`.

- **Decision**: Заменить `UpsertGenres` на `GetIDsByCodes` с fallback на «Неотсортированное».
- **Backward compatibility**: Старый `UpsertGenres` можно оставить для случая, когда GLST ещё не загружен (первый импорт до автозагрузки). Но по спеке GLST загружается при старте ДО импорта, поэтому safe to remove.

## R7: Таблица app_metadata

### Решение

```sql
CREATE TABLE IF NOT EXISTS app_metadata (
    key   VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

- **Decision**: Создать таблицу `app_metadata` для хранения служебных метаданных (hash дерева жанров, etc.).
- **Rationale**: Простая key-value таблица. Соответствует §2.I (данные в PostgreSQL). Расширяема для будущих нужд.
- **Alternative**: Хранить хеш в существующей таблице (genres, collections). Отвергнуто — нет подходящей таблицы, добавление колонки в `genres` семантически некорректно.

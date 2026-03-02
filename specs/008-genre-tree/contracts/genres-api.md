# API Contracts: Genres

**Feature**: 008-genre-tree | **Date**: 2026-03-01

## Изменённые эндпоинты

### GET /api/genres

Возвращает дерево жанров с позициями и счётчиками книг.

**Response** `200 OK`:

```json
[
  {
    "id": 1,
    "code": "_root_0",
    "name": "Неотсортированное",
    "position": "0.0",
    "books_count": 1523,
    "children": []
  },
  {
    "id": 2,
    "code": "sf_all",
    "name": "Фантастика",
    "position": "0.1",
    "books_count": 45230,
    "children": [
      {
        "id": 10,
        "code": "sf_history",
        "name": "Альтернативная история",
        "position": "0.1.1",
        "books_count": 2340,
        "children": []
      },
      {
        "id": 11,
        "code": "sf_action",
        "name": "Боевая фантастика",
        "position": "0.1.2",
        "books_count": 5670,
        "children": []
      }
    ]
  }
]
```

**Изменения относительно текущего API**:
- Добавлено поле `position` (string)
- Поле `meta_group` удалено из ответа (заменено иерархией через `children`)
- Поле `books_count` теперь включает книги из всех потомков (рекурсивный подсчёт, FR-014)
- Порядок сортировки: по `sort_order` (порядок из `.glst` файла)

### GET /api/books

Фильтр `genre_id` теперь работает каскадно (FR-007).

**Query params** (без изменений структуры):

| Param | Type | Description |
|-------|------|-------------|
| `genre_id` | int | ID жанра. Возвращает книги этого жанра **и всех его потомков** |

**Поведение до**: `WHERE bg.genre_id = $1` (точное совпадение)
**Поведение после**: CTE с materialized path — включает все descendant genre IDs

### GET /api/books/:id

**Response** — поле `genres` расширено:

```json
{
  "genres": [
    {
      "id": 10,
      "code": "sf_history",
      "name": "Альтернативная история",
      "position": "0.1.1"
    }
  ]
}
```

Добавлено поле `position` в `BookGenreDetailRef`.

## Новые эндпоинты

### POST /api/admin/genres/reload

Перезагрузка дерева жанров из `.glst` файла. Доступен только администраторам.

**Auth**: JWT с ролью `admin`

**Request**: пустое тело

**Response** `200 OK`:

```json
{
  "genres_loaded": 448,
  "books_remapped": 623456,
  "warnings": [
    "line 15: skipped invalid line",
    "line 230: orphaned child 0.27.3 (parent 0.27 not found)"
  ]
}
```

**Response** `409 Conflict` (если уже выполняется):

```json
{
  "error": "genre reload already in progress"
}
```

**Response** `500 Internal Server Error`:

```json
{
  "error": "failed to load genre tree: ..."
}
```

## TypeScript типы (frontend)

```typescript
// Обновлённый GenreTreeItem
interface GenreTreeItem {
  id: number
  code: string
  name: string
  position: string        // NEW
  books_count: number
  children?: GenreTreeItem[]
}

// Для BookDetail
interface BookGenreDetailRef {
  id: number
  code: string
  name: string
  position: string        // NEW
}

// Response от admin reload
interface GenreReloadResult {
  genres_loaded: number
  books_remapped: number
  warnings: string[]
}
```

## API клиент (frontend)

```typescript
// books.ts — новые/изменённые методы

// Существующий (без изменений сигнатуры, response type обновлён)
export async function getGenres(): Promise<GenreTreeItem[]>

// Новый
export async function reloadGenres(): Promise<GenreReloadResult>
```

# Data Model: Редизайн каталога в стиле MyHomeLib

**Date**: 2026-02-23
**Feature**: [spec.md](spec.md)

## Обзор

Фича преимущественно фронтендовая. Новых таблиц и миграций БД не требуется. Изменения в модели данных касаются:
1. Расширения JSONB-структуры `users.settings` (ключ `catalog`)
2. Новых TypeScript-типов для фронтенда

## Существующие сущности (без изменений)

### Book (books)
Основная сущность каталога. Используется в таблице книг и панели деталей.

| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL | PK |
| title | TEXT | Название книги |
| lang | VARCHAR(3) | Язык (ru, en, ...) |
| year | SMALLINT | Год издания |
| format | VARCHAR(10) | Формат файла (FB2, EPUB, ...) |
| file_size | BIGINT | Размер файла в байтах |
| lib_rate | SMALLINT | Рейтинг библиотеки (не используется в UI) |
| description | TEXT | Аннотация книги |
| keywords | TEXT | Ключевые слова |
| date_added | TIMESTAMPTZ | Дата добавления |
| series_id | BIGINT FK | Серия (nullable) |
| series_num | SMALLINT | Номер в серии |
| collection_id | BIGINT FK | Коллекция |
| lib_id | TEXT | ID в библиотеке |

**Связи:**
- `book_authors` (M:N) → `authors`
- `book_genres` (M:N) → `genres`
- `series` (M:1) → `series`

### Author (authors)
| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL | PK |
| name | TEXT | ФИО автора |

### Genre (genres)
| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL | PK |
| code | VARCHAR(50) | Код жанра (уникальный) |
| name | TEXT | Название жанра |
| meta_group | TEXT | Родительская категория |

### Series (series)
| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL | PK |
| name | TEXT | Название серии |

### User Settings (users.settings JSONB)
Существующая колонка, расширяется новым ключом `catalog`.

## Изменения в модели данных

### Расширение users.settings JSONB

**Текущая структура:**
```json
{
  "reader": {
    "theme": "light",
    "fontSize": 18,
    "fontFamily": "Georgia",
    ...
  }
}
```

**Новая структура (добавляется `catalog`):**
```json
{
  "reader": {
    "theme": null,
    "fontSize": 18,
    ...
  },
  "catalog": {
    "theme": "light",
    "panelSizes": {
      "leftWidth": 25,
      "tableHeight": 60
    },
    "activeTab": "authors",
    "tableSort": {
      "field": "title",
      "order": "asc"
    }
  }
}
```

**Поля `catalog`:**

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| theme | string | "light" | Тема каталога: light / dark / sepia / night |
| panelSizes.leftWidth | number (0-100) | 25 | Ширина левой панели в % |
| panelSizes.tableHeight | number (0-100) | 60 | Высота таблицы в % от правой области |
| activeTab | string | "authors" | Активная вкладка: authors / series / genres / search |
| tableSort.field | string | "title" | Поле сортировки: title / year / file_size |
| tableSort.order | string | "asc" | Направление: asc / desc |

**Изменение `reader.theme`:**

| Значение | Поведение |
|----------|-----------|
| null (по умолчанию) | Наследует `catalog.theme` |
| "light" / "dark" / "sepia" / "night" | Независимый override |

**Миграция не требуется** — JSONB-колонка уже существует (migration 004). Новые ключи добавляются через `settings || $patch` (JSONB merge operator). Бэкенд: добавить `"catalog"` в whitelist `settings.go`.

## Новые TypeScript-типы (фронтенд)

### CatalogTheme

```typescript
type CatalogThemeName = 'light' | 'dark' | 'sepia' | 'night'

interface CatalogThemeDefinition {
  name: CatalogThemeName
  label: string           // «Светлая», «Тёмная», «Сепия», «Ночная»
  dark: boolean           // Vuetify dark flag
  colors: Record<string, string>
  variables?: Record<string, string | number>
}
```

### CatalogSettings

```typescript
interface CatalogSettings {
  theme: CatalogThemeName
  panelSizes: {
    leftWidth: number     // 0-100, по умолчанию 25
    tableHeight: number   // 0-100, по умолчанию 60
  }
  activeTab: TabType
  tableSort: {
    field: SortField
    order: SortOrder
  }
}

type TabType = 'authors' | 'series' | 'genres' | 'search'
type SortField = 'title' | 'year' | 'file_size'
type SortOrder = 'asc' | 'desc'
```

### NavigationItem (для левой панели)

```typescript
interface AuthorItem {
  id: number
  name: string
  booksCount: number
}

interface SeriesItem {
  id: number
  name: string
  booksCount: number
}

interface GenreTreeItem {
  id: number
  code: string
  name: string
  metaGroup?: string
  booksCount: number
  children?: GenreTreeItem[]
}
```

### BookTableRow (для таблицы)

```typescript
interface BookTableRow {
  id: number
  title: string
  authorName: string     // Отформатированное имя (первый автор + «и др.»)
  seriesName?: string    // «Серия #N» или null
  genreName: string      // Первый жанр
  fileSize: string       // Отформатированный размер (например, «1.2 MB»)
}
```

## Диаграмма состояний

### Тема каталога
```
[Light] --переключение--> [Dark] --переключение--> [Sepia] --переключение--> [Night] --переключение--> [Light]
```

### Тема читалки (наследование)
```
[null — наследует каталог]
   |-- пользователь явно выбирает --> [override: light|dark|sepia|night]
   |-- «Сбросить к теме каталога» --> [null — наследует каталог]
```

### Активная вкладка навигации
```
[authors] <--> [series] <--> [genres] <--> [search]
```
При переключении вкладки:
- Сбрасывается выбранный элемент в левой панели
- Таблица книг очищается (до выбора нового элемента)
- Панель деталей показывает placeholder

## Валидация

| Поле | Правило | Источник |
|------|---------|----------|
| catalog.theme | enum: light, dark, sepia, night | Фронтенд |
| catalog.panelSizes.leftWidth | 10 ≤ x ≤ 50 | Фронтенд (min-size/max-size splitpanes) |
| catalog.panelSizes.tableHeight | 20 ≤ x ≤ 80 | Фронтенд (min-size/max-size splitpanes) |
| catalog.activeTab | enum: authors, series, genres, search | Фронтенд |
| reader.theme | null OR enum: light, dark, sepia, night | Фронтенд |
| settings total size | ≤ 64KB | Бэкенд (существующая проверка) |

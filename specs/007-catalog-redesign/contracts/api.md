# API Contracts: Редизайн каталога в стиле MyHomeLib

**Date**: 2026-02-23
**Feature**: [spec.md](../spec.md)

## Обзор

Все необходимые API-эндпоинты **уже существуют**. Единственное изменение — расширение whitelist настроек для ключа `catalog`. Ниже — полный список используемых эндпоинтов с маппингом на функциональные требования.

## Существующие эндпоинты (без изменений)

### GET /api/books

**Используется для**: Таблица книг (FR-007, FR-008, FR-016)

**Query parameters:**

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|-------------|----------|
| q | string | — | Полнотекстовый поиск |
| author_id | int64 | — | Фильтр по автору |
| genre_id | int64 | — | Фильтр по жанру |
| series_id | int64 | — | Фильтр по серии |
| lang | string | — | Фильтр по языку |
| format | string | — | Фильтр по формату |
| page | int | 1 | Страница |
| limit | int | 20 | Записей на странице (max 100) |
| sort | string | title | Поле сортировки: title, year, added_at, lib_rate |
| order | string | asc | Направление: asc, desc |

**Response 200:**
```json
{
  "items": [
    {
      "id": 123,
      "title": "Основание",
      "lang": "ru",
      "year": 1951,
      "format": "FB2",
      "file_size": 524288,
      "authors": [{ "id": 1, "name": "Азимов, Айзек" }],
      "genres": [{ "id": 5, "code": "sf", "name": "Фантастика" }],
      "series": { "id": 10, "name": "Основание", "num": 1 }
    }
  ],
  "total": 1500,
  "page": 1,
  "limit": 20
}
```

---

### GET /api/books/:id

**Используется для**: Панель деталей книги (FR-009, FR-010)

**Response 200:**
```json
{
  "id": 123,
  "title": "Основание",
  "lang": "ru",
  "year": 1951,
  "format": "FB2",
  "file_size": 524288,
  "description": "Аннотация книги...",
  "keywords": ["фантастика", "космос"],
  "date_added": "2026-01-15T12:00:00Z",
  "authors": [{ "id": 1, "name": "Азимов, Айзек" }],
  "genres": [{ "id": 5, "code": "sf", "name": "Фантастика" }],
  "series": { "id": 10, "name": "Основание", "num": 1 },
  "collection": { "id": 1, "name": "lib.rus.ec" }
}
```

---

### GET /api/authors

**Используется для**: Вкладка «Авторы» (FR-003)

**Query parameters:**

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|-------------|----------|
| q | string | — | Поиск по имени (ILIKE) |
| page | int | 1 | Страница |
| limit | int | 20 | Записей на странице (max 100) |

**Response 200:**
```json
{
  "items": [
    { "id": 1, "name": "Азимов, Айзек", "books_count": 47 }
  ],
  "total": 5000,
  "page": 1,
  "limit": 20
}
```

---

### GET /api/genres

**Используется для**: Вкладка «Жанры» (FR-005)

**Response 200:**
```json
[
  {
    "id": 1,
    "code": "sf",
    "name": "Фантастика",
    "meta_group": "Художественная литература",
    "books_count": 15000,
    "children": [
      { "id": 2, "code": "sf_cyberpunk", "name": "Киберпанк", "books_count": 450, "children": [] }
    ]
  }
]
```

---

### GET /api/series

**Используется для**: Вкладка «Серии» (FR-004)

**Query parameters:**

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|-------------|----------|
| q | string | — | Поиск по названию (ILIKE) |
| page | int | 1 | Страница |
| limit | int | 20 | Записей на странице (max 100) |

**Response 200:**
```json
{
  "items": [
    { "id": 10, "name": "Основание", "books_count": 7 }
  ],
  "total": 2000,
  "page": 1,
  "limit": 20
}
```

---

### GET /api/stats

**Используется для**: Хедер — счётчик книг (FR-012)

**Response 200:**
```json
{
  "books_count": 600000,
  "authors_count": 50000,
  "genres_count": 150,
  "series_count": 20000,
  "languages": ["ru", "en", "de", "fr"],
  "formats": ["FB2", "EPUB", "PDF", "DJVU"]
}
```

---

### GET /api/me/progress

**Используется для**: Индикатор прогресса в таблице книг (опционально)

**Response 200:**
```json
{
  "165528": 59,
  "234567": 100
}
```

---

### GET /api/me/settings

**Используется для**: Загрузка настроек при старте (тема, размеры панелей)

**Response 200:**
```json
{
  "reader": { "theme": null, "fontSize": 18 },
  "catalog": { "theme": "dark", "panelSizes": { "leftWidth": 25, "tableHeight": 60 } }
}
```

---

## Изменяемый эндпоинт

### PUT /api/me/settings

**Изменение**: Добавить `"catalog"` в whitelist allowed keys.

**Текущий whitelist**: `["reader", "ui"]`
**Новый whitelist**: `["reader", "ui", "catalog"]`

**Request body (пример сохранения темы):**
```json
{
  "catalog": {
    "theme": "dark"
  }
}
```

**Request body (пример сохранения размеров панелей):**
```json
{
  "catalog": {
    "panelSizes": { "leftWidth": 30, "tableHeight": 55 }
  }
}
```

**Request body (пример изменения темы читалки):**
```json
{
  "reader": {
    "theme": "night"
  }
}
```

**Request body (пример сброса темы читалки к каталогу):**
```json
{
  "reader": {
    "theme": null
  }
}
```

**Response 200**: Полный объект settings после merge.

**Поведение**: JSONB merge (`settings || $patch`) — существующие ключи обновляются, остальные сохраняются.

---

## Маппинг FR → API

| FR | Описание | API эндпоинт | Статус |
|----|----------|-------------|--------|
| FR-001 | Трёхпанельный layout | — (фронтенд) | Нет API |
| FR-002 | Навигационные вкладки | — (фронтенд) | Нет API |
| FR-003 | Авторы с поиском | GET /api/authors | Существует |
| FR-004 | Серии с поиском | GET /api/series | Существует |
| FR-005 | Жанры (дерево) | GET /api/genres | Существует |
| FR-006 | Расширенный поиск | GET /api/books?q=...&... | Существует |
| FR-007 | Таблица книг | GET /api/books | Существует |
| FR-008 | Сортировка таблицы | GET /api/books?sort=...&order=... | Существует |
| FR-009 | Детали книги | GET /api/books/:id | Существует |
| FR-010 | Читать / Скачать | GET /api/books/:id/content, download | Существует |
| FR-011 | Resizable panels | — (фронтенд) | Нет API |
| FR-012 | Счётчик книг в хедере | GET /api/stats | Существует |
| FR-013 | Пользовательское меню | — (фронтенд) | Нет API |
| FR-014 | Статус-бар | — (фронтенд) | Нет API |
| FR-015 | 4 цветовые схемы | — (фронтенд) | Нет API |
| FR-016 | Фильтрация по выбору | GET /api/books?author_id=... | Существует |
| FR-017 | Ellipsis в таблице | — (CSS) | Нет API |
| FR-018 | Расширяемость тем | — (фронтенд) | Нет API |
| FR-019 | Быстрый переключатель | — (фронтенд) | Нет API |
| FR-020 | Диалог настроек | GET/PUT /me/settings | Существует |
| FR-021 | Наследование темы читалкой | GET /me/settings | Существует |
| FR-022 | Независимая тема читалки | PUT /me/settings | Существует |
| FR-023 | Сброс темы читалки | PUT /me/settings | Существует |
| FR-024 | Сохранение тем | PUT /me/settings | **Изменение whitelist** |

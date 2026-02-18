# API Contract: Reading Progress & Settings

**Feature**: 006-fb2-reader | **Date**: 2026-02-18

## Базовый путь

```
/api/me
```

Все эндпоинты требуют JWT-аутентификации. Данные привязаны к `user_id` из JWT claims.

---

## GET /api/me/books/:bookId/progress

Получить прогресс чтения конкретной книги.

### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| bookId | int64 | ID книги |

### Успешный ответ (200 OK)

```json
{
  "chapterId": "ch1-5",
  "chapterProgress": 42,
  "totalProgress": 15,
  "device": "desktop",
  "updatedAt": "2026-02-18T14:30:00Z"
}
```

### Поля ответа

| Поле | Тип | Описание |
|------|-----|----------|
| chapterId | string | ID последней прочитанной главы |
| chapterProgress | int | Прогресс внутри главы (0-100%) |
| totalProgress | int | Общий прогресс книги (0-100%) |
| device | string | Устройство последнего обновления |
| updatedAt | string (ISO 8601) | Время последнего обновления |

### Ответ, если прогресса нет (204 No Content)

Пустое тело — прогресс для этой книги ещё не сохранялся.

### Ошибки

| Код | Описание |
|-----|----------|
| 401 | Не авторизован |
| 404 | Книга не найдена |

---

## PUT /api/me/books/:bookId/progress

Сохранить/обновить прогресс чтения.

### Параметры пути

| Параметр | Тип | Описание |
|----------|-----|----------|
| bookId | int64 | ID книги |

### Тело запроса

```json
{
  "chapterId": "ch1-5",
  "chapterProgress": 42,
  "totalProgress": 15,
  "device": "desktop"
}
```

### Поля запроса

| Поле | Тип | Обязательное | Описание |
|------|-----|-------------|----------|
| chapterId | string | Да | ID текущей главы |
| chapterProgress | int | Да | Прогресс внутри главы (0-100) |
| totalProgress | int | Да | Общий прогресс книги (0-100) |
| device | string | Нет | Тип устройства (desktop/tablet/mobile) |

### Валидация

- `chapterProgress` — целое число от 0 до 100
- `totalProgress` — целое число от 0 до 100
- `chapterId` — не пустая строка

### Успешный ответ (200 OK)

```json
{
  "chapterId": "ch1-5",
  "chapterProgress": 42,
  "totalProgress": 15,
  "device": "desktop",
  "updatedAt": "2026-02-18T14:30:00Z"
}
```

### Поведение

- При первом сохранении создаётся запись (`INSERT`)
- При повторном — обновляется (`UPSERT` по `UNIQUE(user_id, book_id)`)
- Фронтенд вызывает с debounce (2 сек) после каждого перелистывания

### Ошибки

| Код | Описание |
|-----|----------|
| 400 | Невалидные данные (прогресс вне диапазона, пустой chapterId) |
| 401 | Не авторизован |
| 404 | Книга не найдена |

---

## GET /api/me/settings

Получить настройки пользователя (включая настройки читалки).

### Успешный ответ (200 OK)

Структура полностью соответствует архитектуре v8, раздел 8.5.

```json
{
  "reader": {
    "fontSize": 18,
    "fontFamily": "Georgia",
    "fontWeight": 400,
    "lineHeight": 1.6,
    "paragraphSpacing": 0.5,
    "letterSpacing": 0,
    "marginHorizontal": 5,
    "marginVertical": 3,
    "firstLineIndent": 1.5,
    "textAlign": "justify",
    "hyphenation": true,
    "theme": "light",
    "customColors": null,
    "viewMode": "paginated",
    "pageAnimation": "slide",
    "showProgress": true,
    "showClock": false,
    "tapZones": "lrc"
  }
}
```

### Поля reader

| Поле | Тип | Диапазон | По умолчанию | Описание |
|------|-----|----------|-------------|----------|
| fontSize | number | 12-36 | 18 | Размер шрифта (px) |
| fontFamily | string | см. ниже | "Georgia" | Семейство шрифта |
| fontWeight | number | 400 \| 500 | 400 | Насыщенность: нормальный / полужирный |
| lineHeight | number | 1.0-2.5 | 1.6 | Межстрочный интервал |
| paragraphSpacing | number | 0-2 | 0.5 | Отступ между абзацами (em) |
| letterSpacing | number | -0.05 — 0.1 | 0 | Межбуквенный интервал (em) |
| marginHorizontal | number | 0-20 | 5 | Горизонтальные поля (% ширины) |
| marginVertical | number | 0-10 | 3 | Вертикальные поля (% высоты) |
| firstLineIndent | number | 0-3 | 1.5 | Красная строка (em) |
| textAlign | string | "left" \| "justify" | "justify" | Выравнивание текста |
| hyphenation | boolean | — | true | Авто-переносы (CSS hyphens) |
| theme | string | см. ниже | "light" | Цветовая тема |
| customColors | object \| null | — | null | Пользовательские цвета (при theme="custom") |
| viewMode | string | "paginated" \| "scroll" | "paginated" | Режим отображения |
| pageAnimation | string | "slide" \| "fade" \| "none" | "slide" | Анимация перелистывания |
| showProgress | boolean | — | true | Показывать индикатор прогресса |
| showClock | boolean | — | false | Показывать время в углу |
| tapZones | string | "lr" \| "lrc" | "lrc" | Зоны тапа на мобильных |

**fontFamily допустимые значения**: 'Georgia', 'PT Serif', 'Literata', 'OpenDyslexic', 'System'

**theme допустимые значения**: 'light', 'sepia', 'dark', 'night', 'custom'

**customColors** (только при theme="custom"):
| Поле | Тип | Описание |
|------|-----|----------|
| background | string | Цвет фона (#hex) |
| text | string | Цвет текста (#hex) |
| link | string | Цвет ссылок (#hex) |
| selection | string | Цвет выделения (#hex) |

### Ответ, если настроек нет (200 OK)

```json
{}
```

Пустой объект — фронтенд применяет значения по умолчанию.

---

## PUT /api/me/settings

Обновить настройки пользователя.

### Тело запроса

```json
{
  "reader": {
    "fontSize": 20,
    "theme": "dark"
  }
}
```

### Поведение

- Partial update — передаются только изменённые поля
- Слияние с существующими настройками (JSON merge в PostgreSQL)
- Невалидные поля игнорируются

### Успешный ответ (200 OK)

Возвращает полные настройки после слияния:

```json
{
  "reader": {
    "fontSize": 20,
    "fontFamily": "Georgia",
    "fontWeight": 400,
    "lineHeight": 1.6,
    "paragraphSpacing": 0.5,
    "letterSpacing": 0,
    "marginHorizontal": 5,
    "marginVertical": 3,
    "firstLineIndent": 1.5,
    "textAlign": "justify",
    "hyphenation": true,
    "theme": "dark",
    "customColors": null,
    "viewMode": "paginated",
    "pageAnimation": "slide",
    "showProgress": true,
    "showClock": false,
    "tapZones": "lrc"
  }
}
```

### Ошибки

| Код | Описание |
|-----|----------|
| 400 | Невалидный JSON |
| 401 | Не авторизован |

---

## Сводная таблица эндпоинтов

| Метод | Путь | Описание | FR |
|-------|------|----------|----|
| GET | /api/books/:id/content | Метаданные + структура книги | FR-001, FR-002, FR-003 |
| GET | /api/books/:id/chapter/:chapterId | HTML-контент главы | FR-004, FR-014 |
| GET | /api/books/:id/image/:imageId | Изображение из книги | FR-005 |
| GET | /api/me/books/:bookId/progress | Прогресс чтения | FR-009 |
| PUT | /api/me/books/:bookId/progress | Сохранить прогресс | FR-008 |
| GET | /api/me/settings | Настройки пользователя | FR-012 |
| PUT | /api/me/settings | Обновить настройки | FR-012 |

## Маршрутизация (бэкенд)

```
# Reader content (авторизованные)
GET  /api/books/:id/content           → handler.GetBookContent
GET  /api/books/:id/chapter/:chapterId → handler.GetChapter
GET  /api/books/:id/image/:imageId    → handler.GetBookImage

# User progress & settings (авторизованные)
GET  /api/me/books/:bookId/progress   → handler.GetReadingProgress
PUT  /api/me/books/:bookId/progress   → handler.SaveReadingProgress
GET  /api/me/settings                 → handler.GetUserSettings
PUT  /api/me/settings                 → handler.UpdateUserSettings
```

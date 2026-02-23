# Data Model: Браузерная читалка FB2

**Branch**: `006-fb2-reader` | **Date**: 2026-02-18

## Новые сущности

### ReadingProgress (таблица `reading_progress`)

Прогресс чтения пользователя для конкретной книги.

| Поле | Тип | Описание |
|------|-----|----------|
| id | BIGSERIAL PK | Автоинкремент |
| user_id | UUID NOT NULL FK→users | Пользователь |
| book_id | BIGINT NOT NULL FK→books | Книга |
| chapter_id | TEXT NOT NULL | ID текущей главы |
| chapter_progress | SMALLINT DEFAULT 0 | Прогресс внутри главы (0-100%) |
| total_progress | SMALLINT DEFAULT 0 | Общий прогресс книги (0-100%) |
| device | TEXT | Тип устройства (desktop/tablet/mobile) |
| updated_at | TIMESTAMPTZ DEFAULT NOW() | Время последнего обновления |

**Ограничения**:
- `UNIQUE(user_id, book_id)` — один прогресс на пару пользователь-книга
- `ON DELETE CASCADE` для обоих FK

**Индексы**:
- `(user_id, book_id)` — уникальный, основной lookup
- `(user_id)` — для списка «мои книги» с прогрессом

---

### ReaderSettings (JSONB в `users.settings`)

Настройки читалки, хранящиеся в существующем поле `settings` таблицы `users`.
Структура полностью соответствует архитектуре v8, раздел 8.5.

#### Шрифт

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| fontSize | number | 18 | Размер шрифта (12-36 px) |
| fontFamily | string | "Georgia" | Семейство шрифта: 'Georgia', 'PT Serif', 'Literata', 'OpenDyslexic', 'System' |
| fontWeight | number | 400 | Насыщенность: 400 (нормальный) или 500 (полужирный) |

#### Интервалы

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| lineHeight | number | 1.6 | Межстрочный интервал (1.0-2.5) |
| paragraphSpacing | number | 0.5 | Отступ между абзацами (0-2 em) |
| letterSpacing | number | 0 | Межбуквенный интервал (-0.05 — 0.1 em) |

#### Отступы

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| marginHorizontal | number | 5 | Горизонтальные поля (0-20% ширины) |
| marginVertical | number | 3 | Вертикальные поля (0-10% высоты) |
| firstLineIndent | number | 1.5 | Красная строка (0-3 em) |

#### Текст

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| textAlign | string | "justify" | Выравнивание: "left" или "justify" |
| hyphenation | boolean | true | Авто-переносы (CSS `hyphens: auto`) |

#### Тема

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| theme | string | "light" | Тема: "light", "sepia", "dark", "night", "custom" |
| customColors | object \| null | null | Пользовательские цвета (только для theme="custom") |
| customColors.background | string | — | Цвет фона (#hex) |
| customColors.text | string | — | Цвет текста (#hex) |
| customColors.link | string | — | Цвет ссылок (#hex) |
| customColors.selection | string | — | Цвет выделения (#hex) |

#### Режим отображения

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| viewMode | string | "paginated" | Режим: "paginated" или "scroll" |
| pageAnimation | string | "slide" | Анимация перелистывания: "slide", "fade", "none" |

#### Дополнительно

| Поле | Тип | По умолчанию | Описание |
|------|-----|-------------|----------|
| showProgress | boolean | true | Показывать индикатор прогресса |
| showClock | boolean | false | Показывать время в углу |
| tapZones | string | "lrc" | Зоны тапа: "lr" (лево-право) или "lrc" (лево-центр-право) |

**Дополнительная миграция не нужна** — поле `settings JSONB DEFAULT '{}'` уже существует в `users`.

---

## Структуры данных бэкенда (Go)

### BookContent

Результат конвертации книги (метаданные + структура).

```
BookContent {
  Metadata: BookMetadata {
    Title    string
    Author   string
    Cover    string   (URL на обложку, если есть)
    Language string
    Format   string
  }
  TOC: []TOCEntry {
    ID    string
    Title string
    Level int      (0 = верхний уровень)
  }
  ChapterIDs:    []string
  TotalChapters: int
}
```

### ChapterContent

Содержимое одной главы.

```
ChapterContent {
  ID    string
  Title string
  HTML  string   (чистый HTML без стилей)
}
```

### ReadingProgress (Go model)

```
ReadingProgress {
  ID              int64
  UserID          uuid.UUID
  BookID          int64
  ChapterID       string
  ChapterProgress int      (0-100)
  TotalProgress   int      (0-100)
  Device          string
  UpdatedAt       time.Time
}
```

---

## Структуры данных фронтенда (TypeScript)

### BookContent

```
BookContent {
  metadata: {
    title: string
    author: string
    cover: string | null
    language: string
    format: string
  }
  toc: TOCEntry[]
  chapters: string[]
  totalChapters: number
}
```

### TOCEntry

```
TOCEntry {
  id: string
  title: string
  level: number
}
```

### ChapterContent

```
ChapterContent {
  id: string
  title: string
  html: string
}
```

### ReadingPosition

```
ReadingPosition {
  chapterId: string
  chapterProgress: number   (0-100)
  totalProgress: number     (0-100)
  device: string
}
```

### ReaderSettings

Соответствует архитектуре v8, раздел 8.5.

```
ReaderSettings {
  // === Шрифт ===
  fontSize: number              // 12-36 px
  fontFamily: string            // 'Georgia', 'PT Serif', 'Literata', 'OpenDyslexic', 'System'
  fontWeight: 400 | 500         // нормальный / полужирный

  // === Интервалы ===
  lineHeight: number            // 1.0 - 2.5
  paragraphSpacing: number      // 0 - 2 em
  letterSpacing: number         // -0.05 - 0.1 em

  // === Отступы ===
  marginHorizontal: number      // 0 - 20 % от ширины
  marginVertical: number        // 0 - 10 % от высоты
  firstLineIndent: number       // 0 - 3 em (красная строка)

  // === Текст ===
  textAlign: 'left' | 'justify'
  hyphenation: boolean          // авто-переносы

  // === Тема ===
  theme: 'light' | 'sepia' | 'dark' | 'night' | 'custom'
  customColors?: {
    background: string
    text: string
    link: string
    selection: string
  }

  // === Режим отображения ===
  viewMode: 'paginated' | 'scroll'
  pageAnimation: 'slide' | 'fade' | 'none'

  // === Дополнительно ===
  showProgress: boolean         // индикатор прогресса
  showClock: boolean            // время в углу
  tapZones: 'lr' | 'lrc'       // зоны тапа: лево-право или лево-центр-право
}
```

---

## Связи между сущностями

```
users ──1:N──▶ reading_progress ◀──N:1── books
users.settings ──contains──▶ ReaderSettings (JSONB)

books ──format──▶ BookConverter (fb2 → FB2Converter)
BookConverter ──produces──▶ BookContent + ChapterContent[]
```

---

## Миграция

### 003_reading_progress.up.sql

```sql
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
```

### 003_reading_progress.down.sql

```sql
DROP TABLE IF EXISTS reading_progress;
```

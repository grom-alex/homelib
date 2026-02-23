# Research: Браузерная читалка FB2

**Branch**: `006-fb2-reader` | **Date**: 2026-02-18

## R1: Парсинг FB2 в Go

### Decision: Стандартная библиотека `encoding/xml`

**Rationale**: FB2 — это XML-формат с чётко определённой схемой. Go `encoding/xml` достаточно для парсинга всех элементов FB2: section, paragraph, emphasis, poem, epigraph, image, footnote. Внешние зависимости не нужны.

**Alternatives considered**:
- `etree` (сторонняя XML-библиотека) — избыточна, добавляет зависимость
- `xmlquery` (XPath) — удобнее для навигации, но неоправданная зависимость

**Структура FB2**:
```xml
<FictionBook>
  <description>
    <title-info>
      <author>, <book-title>, <annotation>, <genre>, <lang>, <coverpage>
    </title-info>
  </description>
  <body>
    <section>            <!-- Глава -->
      <title>            <!-- Заголовок -->
      <epigraph>         <!-- Эпиграф -->
      <p>                <!-- Параграф -->
      <poem>             <!-- Стихотворение -->
        <stanza><v>      <!-- Строфа/стих -->
      <cite>             <!-- Цитата -->
      <image>            <!-- Изображение (ссылка на binary) -->
      <section>          <!-- Вложенная секция (рекурсивно) -->
    </section>
  </body>
  <binary id="img1" content-type="image/jpeg">base64data...</binary>
</FictionBook>
```

**Маппинг FB2 → HTML**:

| FB2 тег | HTML | CSS класс |
|---------|------|-----------|
| `<section>` | — (контейнер глав) | — |
| `<title>` | `<h2>`...`<h6>` (по уровню вложенности) | `chapter-title` |
| `<p>` | `<p>` | — |
| `<emphasis>` | `<em>` | — |
| `<strong>` | `<strong>` | — |
| `<strikethrough>` | `<del>` | — |
| `<code>` | `<code>` | — |
| `<sup>` | `<sup>` | — |
| `<sub>` | `<sub>` | — |
| `<epigraph>` | `<blockquote>` | `epigraph` |
| `<cite>` | `<blockquote>` | `cite` |
| `<poem>` | `<div>` | `poem` |
| `<stanza>` | `<div>` | `stanza` |
| `<v>` (verse) | `<p>` | `verse` |
| `<subtitle>` | `<p>` | `subtitle` |
| `<text-author>` | `<cite>` | `epigraph-author` / `poem-author` |
| `<image>` | `<img src="data:...;base64,...">` | — |
| `<a>` | `<a>` | — |
| `<empty-line>` | `<br>` | — |

---

## R2: Пагинация в браузере

### Decision: CSS multi-column layout

**Rationale**: CSS columns — нативный, производительный и кроссбраузерный способ разбиения текста на «страницы». Контент заливается в контейнер с `column-width: 100vw`, `column-gap`, `overflow: hidden`. Переключение «страниц» — это сдвиг `translateX` контейнера. Не требует JavaScript-вычислений для разбиения текста.

**Alternatives considered**:
- JavaScript-based splitting (разбиение DOM на страницы) — сложнее, медленнее, ломается при resize
- Scroll-based (просто скролл) — проще, но не даёт ощущения «страниц», сложнее отслеживать прогресс
- `column-count: 1` с overflow — выбранный вариант, лучший баланс

**Реализация**:
```css
.reader-columns {
  column-width: 100%;
  column-gap: 2rem;
  column-fill: auto;
  height: calc(100vh - header - footer);
  overflow: hidden;
}
```

Переключение страниц:
```typescript
const totalPages = Math.ceil(scrollWidth / columnWidth)
const currentPage = ref(0)

function nextPage() {
  if (currentPage.value < totalPages - 1) {
    currentPage.value++
    container.style.transform = `translateX(-${currentPage.value * columnWidth}px)`
  }
}
```

**Пересчёт при изменении настроек**: При изменении размера шрифта, интервалов или ширины окна — пересчитать `totalPages` и скорректировать `currentPage` пропорционально.

---

## R3: Кеширование конвертированного контента

### Decision: Файловый кеш в директории `/cache/books/`

**Rationale**: Конвертация FB2→HTML — CPU-bound операция (парсинг XML, рекурсивная обработка секций). Для книги ~500KB FB2 это ~50–100ms. Кеширование результата на диск экономит время при повторных открытиях. Файловый кеш прост, не требует Redis (§2.I запрещает дополнительные СУБД), легко регенерируется.

**Alternatives considered**:
- In-memory кеш (LRU) — быстрее, но расходует RAM при 600K книг
- Redis — запрещён конституцией (§2.I)
- Без кеша — приемлемо для единиц пользователей, но кеш прост и полезен

**Структура кеша**:
```
/cache/books/
├── {bookID}/
│   ├── content.json        # Метаданные + TOC + список глав
│   ├── ch_{chapterID}.html # HTML контент каждой главы
│   └── img_{id}.bin        # Бинарные изображения (декодированный base64)
```

**Инвалидация**: Кеш не инвалидируется (книги не меняются в архивах). TTL = бесконечный. Очистка — ручная (`make clean-cache`).

**Конфигурация**: Путь к кешу — параметр `reader.cache_path` в YAML-конфиге. По умолчанию — `/cache/books` в контейнере, `./cache/books` локально.

---

## R4: Хранение прогресса чтения

### Decision: Таблица `reading_progress` в PostgreSQL

**Rationale**: Конституция (§2.I, §2.V) требует хранить бизнес-данные в PostgreSQL. Прогресс — per-user, per-book данные. Архитектура v8 уже описывает таблицу `reading_progress`.

**Схема** (из архитектуры):
```sql
CREATE TABLE reading_progress (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  book_id BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  chapter_id TEXT NOT NULL,
  chapter_progress SMALLINT DEFAULT 0,  -- 0-100%
  total_progress SMALLINT DEFAULT 0,    -- 0-100%
  device TEXT,
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, book_id)
);
```

**API**: Debounce сохранения на фронтенде (2 секунды). PUT-запрос с upsert на бэкенде (`ON CONFLICT (user_id, book_id) DO UPDATE`).

---

## R5: Хранение настроек читалки

### Decision: JSONB поле `settings` в таблице `users`

**Rationale**: Архитектура v8 уже предусматривает поле `settings JSONB DEFAULT '{}'` в таблице `users`. Настройки читалки — часть пользовательских предпочтений. Дополнительная таблица не нужна.

**Структура JSON**:
```json
{
  "reader": {
    "fontSize": 18,
    "fontFamily": "Georgia",
    "lineHeight": 1.6,
    "marginHorizontal": 5,
    "marginVertical": 3,
    "firstLineIndent": 1.5,
    "textAlign": "justify",
    "theme": "light",
    "viewMode": "paginated"
  }
}
```

**Важно**: Миграция 001 уже содержит поле `settings JSONB DEFAULT '{}'` в таблице `users`. Дополнительной миграции для настроек не нужно.

---

## R6: Обработка изображений в FB2

### Decision: Data URI для inline-изображений

**Rationale**: FB2 хранит изображения как base64 в тегах `<binary>`. Два варианта отдачи: (1) встроить как data URI в HTML, (2) отдавать отдельным эндпоинтом. Для простоты и минимизации запросов — вариант 2: отдельный эндпоинт `/api/books/:id/image/:imageId`, который достаёт бинарные данные из FB2 и отдаёт с правильным Content-Type. Это позволяет кешировать изображения отдельно и не раздувать HTML главы.

**Alternatives considered**:
- Data URI в HTML — увеличивает размер HTML в ~1.3x, нет отдельного кеширования
- Предварительная распаковка в файлы — нарушает §1.IV (чтение из архивов без распаковки)

**В HTML конвертер подставляет**: `<img src="/api/books/{bookID}/image/{imageId}">`

---

## R7: Навигация и жесты

### Decision: Нативные события + простой gesture detection

**Rationale**: Для домашней библиотеки с единицами пользователей не нужна сложная gesture library. Достаточно нативного `touchstart`/`touchend` для определения свайпов и зон тапа.

**Зоны тапа** (3 зоны по умолчанию):
- Левые 25% — предыдущая страница
- Центральные 50% — показать/скрыть UI
- Правые 25% — следующая страница

**Клавиатура**: `→`, `Space`, `PageDown` — вперёд; `←`, `PageUp` — назад; `T` — оглавление; `Esc` — назад к книге; `+`/`-` — шрифт.

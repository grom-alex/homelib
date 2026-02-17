# HomeLib — Архитектура веб-приложения домашней библиотеки

## 1. Общая схема системы

```
                    Хоумлаб-сервер (без GPU)
┌───────────────────────────────────────────────────────────────┐
│                       Docker Compose                          │
│                                                               │
│  ┌──────────┐   ┌──────────────┐   ┌────────────────────────┐ │
│  │  Nginx   │──▶│  API Server  │──▶│  PostgreSQL            │ │
│  │ (reverse │   │   (Go/Gin)   │   │  + pgvector            │ │
│  │  proxy)  │   └──────┬───────┘   │  + pg_trgm             │ │
│  │          │          │           │  + tsvector            │ │
│  │          │   ┌──────▼───────┐   └────────────────────────┘ │
│  │          │──▶│  Frontend    │                              │
│  │          │   │  (Vue 3 SPA) │   ┌────────────────────────┐ │
│  └──────────┘   └──────────────┘   │  Worker (Go)           │ │
│                                    │                        │ │
│  ┌──────────────────────────┐      │  • Импорт .inpx        │ │
│  │  Ollama (CPU)            │◀─────│  • Обложки / мета      │ │
│  │  (опц. fallback)         │      │  • Summary Embedding   │ │
│  └──────────────────────────┘      │  • LLM Summarization   │ │
│                                    │     (Ollama Pool)      │ │
│                                    └────────┬──┬──┬─────────┘ │
└────────────────────────────────┬────────────┼──┼──┼───────────┘
        ▲                        │            │  │  │
        │ volume (ro)            │            │  │  │ HTTP (LAN)
  ┌─────┴──────┐                 │            │  │  │
  │ /library   │                 │       ┌────┘  │  └────┐
  │ .inpx      │                 │       │       │       │
  │ ZIP-архивы │                 │  ┌────┴──┐ ┌──┴───┐ ┌─┴─────┐
  └────────────┘                 │  │Ollama │ │Ollama│ │Ollama │
                                 │  │:11434 │ │:11434│ │:11434 │
                семантический    │  │ PC #1 │ │ PC#2 │ │ PC #3 │
                поиск — только ──┘  │5060 Ti│ │5060Ti│ │5060 Ti│
                на сервере          └───────┘ └──────┘ └───────┘
                (pgvector)
                                  На Windows — только стандартный
                                  Ollama, никакого доп. софта
```

**Ключевые принципы:**
- На Windows-машинах стоит **только Ollama** — стандартная установка, ноль кастомного кода
- Весь интеллект координации — на сервере
- Сервер обращается к Ollama-инстансам в сети для **эмбеддингов саммари** и **LLM-саммаризации**
- Семантический поиск выполняется локально на сервере через pgvector
- **Summary Embedding** вместо чанков всей книги: 500K векторов вместо 75M

---

## 2. Компоненты и их роли

### 2.1. API Server (Go)

**Фреймворк:** Gin или Echo.

**Эндпоинты:**

| Группа | Метод / Путь | Описание | Доступ |
|--------|-------------|----------|--------|
| **Аутентификация** | | | |
| Регистрация | `POST /api/auth/register` | Создание аккаунта (имя, email, пароль) | Публичный¹ |
| Вход | `POST /api/auth/login` | Логин → JWT access + refresh токены | Публичный |
| Обновление | `POST /api/auth/refresh` | Обновить access-токен по refresh-токену | Публичный |
| Выход | `POST /api/auth/logout` | Инвалидировать refresh-токен | Авториз. |
| **Каталог** (общие, read-only) | | | |
| Книги | `GET /api/books?q=&author=&genre=&lang=&format=&page=&limit=&sort=` | Список с фильтрацией, пагинацией, сортировкой | Авториз. |
| Книга | `GET /api/books/:id` | Метаданные, обложка, аннотация + статус текущего юзера | Авториз. |
| Скачивание | `GET /api/books/:id/download` | Файл из ZIP-архива на лету | Авториз. |
| Чтение | `GET /api/books/:id/read` | Конвертированный контент для браузерной читалки | Авториз. |
| Улучшить описание | `POST /api/books/:id/improve-summary` | Запрос на LLM-генерацию описания | Авториз. |
| Авторы | `GET /api/authors?q=&page=` | Список/поиск авторов | Авториз. |
| Автор | `GET /api/authors/:id` | Автор + его книги | Авториз. |
| Жанры | `GET /api/genres` | Дерево жанров | Авториз. |
| Серии | `GET /api/series?q=&page=` | Список/поиск серий | Авториз. |
| Поиск | `POST /api/search {query}` | Гибридный поиск (полнотекстовый + семантический) | Авториз. |
| **Пользовательские данные** (per-user) | | | |
| Прогресс | `GET /api/me/books/:id/progress` | Получить прогресс чтения книги | Авториз. |
| Прогресс | `PUT /api/me/books/:id/progress` | Сохранить прогресс чтения | Авториз. |
| Статус книги | `PUT /api/me/books/:id/status` | Установить статус: want / reading / read / dropped | Авториз. |
| Оценка | `PUT /api/me/books/:id/rating` | Оценить книгу (1–10) | Авториз. |
| Мои книги | `GET /api/me/books?status=&page=` | Список книг пользователя с фильтром по статусу | Авториз. |
| Полки | `GET /api/me/shelves` | Список полок пользователя | Авториз. |
| Полка | `POST /api/me/shelves` | Создать полку | Авториз. |
| Полка | `PUT /api/me/shelves/:id` | Переименовать полку | Авториз. |
| Полка | `DELETE /api/me/shelves/:id` | Удалить полку | Авториз. |
| Кн. на полке | `POST /api/me/shelves/:id/books` | Добавить книгу на полку | Авториз. |
| Кн. на полке | `DELETE /api/me/shelves/:id/books/:book_id` | Убрать книгу с полки | Авториз. |
| Профиль | `GET /api/me/profile` | Данные текущего пользователя | Авториз. |
| Профиль | `PUT /api/me/profile` | Обновить профиль (имя, аватар, настройки читалки) | Авториз. |
| Статистика | `GET /api/me/stats` | Личная статистика (прочитано, время и т.п.) | Авториз. |
| **Администрирование** | | | |
| Импорт | `POST /api/admin/import` | Запуск импорта .inpx | Админ |
| Summary stats | `GET /api/admin/summaries/stats` | Статистика саммаризации | Админ |
| Summary batch | `POST /api/admin/summaries/batch-generate` | Пакетная LLM-саммаризация | Админ |
| Summary single | `POST /api/admin/books/:id/generate-summary` | LLM-саммари для одной книги | Админ |
| Embedding | `GET /api/admin/embedding/stats` | Статус пула, очередь, прогресс | Админ |
| Embedding | `POST /api/admin/embedding/start` | Запустить/приостановить индексацию | Админ |
| Пользователи | `GET /api/admin/users` | Список пользователей | Админ |
| Пользователь | `PUT /api/admin/users/:id` | Изменить роль, заблокировать | Админ |
| Статистика | `GET /api/stats` | Общая статистика библиотеки | Авториз. |

> ¹ Регистрация может быть закрыта настройкой `auth.registration_enabled` или защищена инвайт-кодом — полезно для хоумлаба чтобы не давать доступ случайным людям.

**Ключевые решения:**

- Книги **не распаковываются** заранее — Go читает из ZIP на лету (`archive/zip` поддерживает random access по offset).
- Для чтения fb2/epub — конвертация при первом запросе с кешированием результата на диск.
- Семантический поиск — `POST /api/search` embed'ит запрос через любой доступный Ollama из пула, затем ищет в pgvector на сервере.
- Все пользовательские данные привязаны к `user_id` из JWT-токена — один эндпоинт обслуживает всех пользователей, каждый видит только свои полки, прогресс и оценки.

### 2.1a. Аутентификация

**Подход:** JWT (access + refresh токены). Простая, stateless аутентификация без внешних зависимостей.

**Почему JWT, а не сессии:**
- Stateless — не нужен Redis или таблица сессий для проверки каждого запроса
- Хорошо ложится на SPA (Vue хранит токен в памяти)
- Простая реализация на Go без фреймворков

**Схема токенов:**

| Токен | Время жизни | Хранение на клиенте | Назначение |
|-------|-------------|---------------------|------------|
| Access | 15 минут | Pinia store (память) | Авторизация запросов (заголовок `Authorization: Bearer ...`) |
| Refresh | 30 дней | httpOnly cookie | Обновление access-токена без повторного логина |

**Payload access-токена:**

```json
{
  "sub": "user_uuid",
  "role": "user",          // "user" | "admin"
  "name": "Иван",
  "exp": 1738500000
}
```

**Middleware (Gin):**

```go
func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }

        claims, err := jwt.ParseAndVerify(token, jwtSecret)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        c.Set("user_id", claims.Subject)
        c.Set("user_role", claims.Role)
        c.Next()
    }
}

func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.GetString("user_role") != "admin" {
            c.AbortWithStatusJSON(403, gin.H{"error": "admin only"})
            return
        }
        c.Next()
    }
}
```

**Роутинг:**

```go
r := gin.Default()

// Публичные
auth := r.Group("/api/auth")
auth.POST("/register", h.Register)
auth.POST("/login", h.Login)
auth.POST("/refresh", h.RefreshToken)

// Требуют авторизации
api := r.Group("/api", AuthMiddleware(cfg.JWTSecret))
api.GET("/books", h.ListBooks)
api.GET("/books/:id", h.GetBook)
// ...

// Пользовательские данные (всегда user_id из токена)
me := api.Group("/me")
me.GET("/books", h.MyBooks)
me.PUT("/books/:id/progress", h.SaveProgress)
me.PUT("/books/:id/status", h.SetBookStatus)
me.PUT("/books/:id/rating", h.SetBookRating)
me.GET("/shelves", h.MyShelves)
me.POST("/shelves", h.CreateShelf)
// ...

// Администрирование
admin := api.Group("/admin", AdminOnly())
admin.POST("/import", h.StartImport)
admin.GET("/embedding/stats", h.EmbeddingStats)
admin.GET("/users", h.ListUsers)
// ...
```

**Создание первого администратора:**

При первом запуске, если в БД нет пользователей, первый зарегистрированный автоматически получает роль `admin`. Альтернативно — CLI-команда:

```bash
docker compose exec api ./homelib-api --create-admin \
    --email admin@home.lab --password changeme
```

### 2.2. Worker (Go)

Отдельный процесс для фоновых задач. Общается с API через БД (таблицы-очереди).

**Задачи:**

| Задача | Описание | Приоритет |
|--------|----------|-----------|
| Импорт .inpx | Парсинг, заполнение БД | Разовая |
| Извлечение обложек | Из fb2/epub, сохранение на диск | При импорте |
| Извлечение аннотаций | Из fb2 `<annotation>`, epub metadata | При импорте |
| Конвертация fb2→HTML | Предварительный рендер для читалки | Фоновая / по требованию |
| Summary Extraction | Извлечение саммари из метаданных книг | Фоновая |
| Summary Embedding | Отправка саммари в Ollama, сохранение векторов | Фоновая |
| LLM Summarization | Генерация описаний через LLM для книг без аннотаций | По очереди |

### 2.3. PostgreSQL + pgvector

Единая БД для всего: каталог, полнотекстовый поиск (`tsvector`), нечёткий поиск (`pg_trgm`), векторный поиск (`pgvector`).

**Почему pgvector, а не pgvecto.rs:** pgvector проще в установке (готовые Docker-образы), хорошо документирован, для домашней библиотеки производительности хватит с запасом. При необходимости можно заменить на pgvecto.rs без изменения схемы.

### 2.4. Ollama Pool (распределённые GPU)

Центральная часть embedding- и LLM-пайплайна. На сервере нет GPU — вычисления отдаются Ollama-инстансам на Windows-машинах в локальной сети.

**Архитектура пула:**

```
┌──────────────────────────────────────────────────┐
│  Ollama Pool (внутри Worker на сервере)          │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Health Monitor                             │  │
│  │ • Пингует /api/tags каждые 30 сек          │  │
│  │ • Помечает инстансы online/offline         │  │
│  │ • Проверяет наличие нужных моделей         │  │
│  │ • Определяет capabilities (embed/llm)      │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Load Balancer                              │  │
│  │ • Least-connections: выбирает инстанс      │  │
│  │   с минимумом активных запросов            │  │
│  │ • Раздельный выбор для embed и llm         │  │
│  │ • Автоматический fallback на CPU           │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Summary Workers                            │  │
│  │ • Извлекают саммари из книг                │  │
│  │ • Отправляют в Ollama для эмбеддинга       │  │
│  │ • Сохраняют вектора в БД                   │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ LLM Workers                                │  │
│  │ • Берут задачи из llm_summary_tasks        │  │
│  │ • Генерируют описания через LLM            │  │
│  │ • Обновляют саммари книг                   │  │
│  └────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────┘
```

**Модели:**

| Задача | Модель | Размер | Примечание |
|--------|--------|--------|------------|
| Embeddings | `nomic-embed-text` | ~275 MB | Вектор 768d, ~100 embed/сек на GPU |
| Embeddings (альт.) | `mxbai-embed-large` | ~670 MB | Вектор 1024d, качественнее |
| LLM Summarization | `llama3` | ~4.7 GB | ~5 сек/книга на GPU |
| LLM Summarization (лёгкая) | `llama3.2:3b` | ~2 GB | ~2 сек/книга, качество ниже |

### 2.5. Ollama на Windows (настройка)

На каждой Windows-машине — одноразовая установка:

```powershell
# 1. Установить Ollama
winget install Ollama.Ollama

# 2. Скачать модель
ollama pull nomic-embed-text

# 3. Разрешить подключения из сети
#    (по умолчанию Ollama слушает только localhost)
[System.Environment]::SetEnvironmentVariable("OLLAMA_HOST", "0.0.0.0:11434", "User")

# 4. Опционально: увеличить параллелизм
[System.Environment]::SetEnvironmentVariable("OLLAMA_NUM_PARALLEL", "4", "User")

# 5. Перезапустить Ollama (или перелогиниться)
```

Всё. Никакого кастомного софта, никаких агентов. Ollama запускается при старте Windows автоматически.

### 2.6. Frontend (Vue 3)

**Стек:** Vue 3 + Composition API, Vue Router, Pinia, Vuetify 3 или Naive UI.

**Страницы:**

```
/login                 — Вход / Регистрация
/                      — Главная: статистика, последние, «продолжить чтение»
/books                 — Каталог с фильтрами (автор, жанр, язык, формат, серия)
/books/:id             — Карточка книги (мета, обложка, аннотация, статус, оценка)
/books/:id/read        — Читалка (прогресс сохраняется автоматически)
/authors               — Каталог авторов (алфавитный, поиск)
/authors/:id           — Страница автора со списком книг
/genres                — Дерево жанров
/series                — Серии книг
/search                — Семантический поиск (свободный запрос)
/my/books              — Мои книги (фильтр: хочу / читаю / прочитал / бросил)
/my/shelves            — Мои полки
/my/shelves/:id        — Содержимое полки
/my/stats              — Моя статистика чтения
/my/profile            — Профиль и настройки (имя, аватар, настройки читалки)
/admin/import          — Управление импортом .inpx (только admin)
/admin/embedding       — Мониторинг embedding-пула и очереди (только admin)
/admin/users           — Управление пользователями (только admin)
```

**Аутентификация на фронтенде (Vue):**

```
┌──────────────────────────────────────────────────────┐
│  Pinia store: useAuthStore                           │
│                                                      │
│  state:                                              │
│    user: { id, name, role } | null                   │
│    accessToken: string | null  (в памяти, не в LS)   │
│                                                      │
│  actions:                                            │
│    login(email, password)   → POST /api/auth/login   │
│    register(...)            → POST /api/auth/register│
│    refresh()                → POST /api/auth/refresh │
│    logout()                 → POST /api/auth/logout  │
│                                                      │
│  Axios interceptor:                                  │
│    request  → добавляет Authorization: Bearer ...    │
│    response → при 401 пытается refresh(),            │
│              если не удалось → redirect /login       │
└──────────────────────────────────────────────────────┘
```

Refresh-токен хранится в **httpOnly cookie** (задаётся сервером) — JavaScript не имеет к нему доступа, защита от XSS. Access-токен хранится **только в памяти** (Pinia store) — при перезагрузке страницы автоматически обновляется через refresh-эндпоинт.

**Читалка в браузере:**

| Формат | Решение |
|--------|---------|
| **epub** | [epub.js](https://github.com/futurepress/epub.js) — пагинация, закладки, настройки шрифта |
| **fb2** | Конвертация в HTML на бэкенде (Go: `encoding/xml` + шаблон), отображение в компоненте |
| **pdf** | [pdf.js](https://mozilla.github.io/pdf.js/) (Mozilla) |
| **djvu** | [djvu.js](https://github.com/nickel715/djvu.js) или конвертация в PDF через `ddjvu` на бэкенде |

### 2.7. Nginx

- Reverse proxy: `/api/*` → Go API, `/*` → Vue SPA
- Gzip/Brotli сжатие
- Кеширование статики и обложек
- Basic auth (опционально, для доступа извне)

---

## 3. Схема базы данных

```sql
-- === Расширения ===
CREATE EXTENSION IF NOT EXISTS pg_trgm;      -- нечёткий поиск
CREATE EXTENSION IF NOT EXISTS vector;        -- pgvector для embeddings

-- === Каталог ===

CREATE TABLE authors (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    name_sort   TEXT NOT NULL,              -- "Фамилия, Имя" для сортировки
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_authors_name_sort ON authors (name_sort);
CREATE INDEX idx_authors_name_trgm ON authors USING gin (name gin_trgm_ops);

CREATE TABLE genres (
    id          SERIAL PRIMARY KEY,
    code        TEXT UNIQUE NOT NULL,       -- sf_fantasy, detective, etc.
    name        TEXT NOT NULL,              -- человекочитаемое название
    parent_id   INTEGER REFERENCES genres(id),
    meta_group  TEXT                        -- группировка: Фантастика, Детектив...
);

-- Коллекции/библиотеки (из collection.info)
CREATE TABLE collections (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,              -- название коллекции
    code            TEXT UNIQUE NOT NULL,       -- идентификатор (имя файла без .inpx)
    collection_type INTEGER DEFAULT 0,          -- тип: 0=fb2, 1=non-fb2, др.=флаги
    description     TEXT,
    source_url      TEXT,
    version         TEXT,                       -- версия из version.info (YYYYMMDD)
    version_date    DATE,                       -- версия как дата
    books_count     INTEGER DEFAULT 0,
    last_import_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE series (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_series_name_trgm ON series USING gin (name gin_trgm_ops);

CREATE TABLE books (
    id              BIGSERIAL PRIMARY KEY,
    collection_id   INTEGER REFERENCES collections(id),
    title           TEXT NOT NULL,
    lang            TEXT NOT NULL DEFAULT 'ru',
    year            INTEGER,
    format          TEXT NOT NULL,               -- fb2, epub, pdf, djvu
    file_size       BIGINT,
    archive_name    TEXT NOT NULL,               -- fb2-012345-012456.zip
    file_in_archive TEXT NOT NULL,               -- 123456.fb2
    series_id       BIGINT REFERENCES series(id),
    series_num      INTEGER,
    series_type     CHAR(1),                     -- 'a'=авторская, 'p'=издательская, NULL=неизвестно
    lib_id          TEXT,                        -- ID из .inpx (уникален в пределах коллекции)
    lib_rate        SMALLINT,                    -- рейтинг из библиотеки (1-5)
    is_deleted      BOOLEAN DEFAULT FALSE,
    is_lost         BOOLEAN DEFAULT FALSE,       -- из _lost архива
    has_cover       BOOLEAN DEFAULT FALSE,
    description     TEXT,                        -- аннотация
    keywords        TEXT[],                      -- ключевые слова из INPX
    date_added      DATE,                        -- дата добавления в библиотеку (из DATE)
    search_vector   tsvector,                    -- полнотекстовый поиск
    added_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (collection_id, lib_id)               -- lib_id уникален в пределах коллекции
);
CREATE INDEX idx_books_title_trgm ON books USING gin (title gin_trgm_ops);
CREATE INDEX idx_books_lang       ON books (lang);
CREATE INDEX idx_books_format     ON books (format);
CREATE INDEX idx_books_archive    ON books (archive_name);
CREATE INDEX idx_books_search     ON books USING gin (search_vector);
CREATE INDEX idx_books_collection ON books (collection_id);
CREATE INDEX idx_books_lib_rate   ON books (lib_rate) WHERE lib_rate IS NOT NULL;
CREATE INDEX idx_books_keywords   ON books USING gin (keywords) WHERE keywords IS NOT NULL;

-- Автообновление tsvector при изменении книги
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

-- Связи M:N
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

-- === Саммари и семантический поиск ===

CREATE TYPE summary_status AS ENUM (
    'pending',          -- ожидает обработки
    'extracted',        -- саммари извлечено из метаданных
    'llm_generated',    -- саммари сгенерировано LLM
    'embedded',         -- эмбеддинг создан
    'failed',           -- ошибка
    'skipped'           -- пропущена (нет данных)
);

CREATE TYPE summary_source AS ENUM (
    'annotation',       -- аннотация из FB2/EPUB
    'toc',              -- оглавление (названия глав)
    'first_paragraphs', -- первые параграфы текста
    'metadata',         -- метаданные (автор, жанр, серия)
    'keywords',         -- ключевые слова из INPX
    'llm'               -- сгенерировано LLM
);

-- Саммари книги (для семантического поиска)
CREATE TABLE book_summaries (
    id              BIGSERIAL PRIMARY KEY,
    book_id         BIGINT UNIQUE NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    
    -- Составное саммари из разных источников
    summary_text    TEXT,                       -- итоговый текст для эмбеддинга (~500-2000 символов)
    sources         summary_source[] NOT NULL DEFAULT '{}', -- откуда собрали
    
    -- Отдельные компоненты (для возможности пересборки)
    annotation      TEXT,                       -- аннотация из книги
    toc_text        TEXT,                       -- названия глав через ";"
    first_paragraphs TEXT,                      -- первые 2-3 параграфа
    llm_summary     TEXT,                       -- сгенерированное LLM описание
    
    -- Эмбеддинг
    embedding       vector(768),
    
    -- Статус обработки
    status          summary_status NOT NULL DEFAULT 'pending',
    error_message   TEXT,
    
    -- Метаданные LLM-генерации
    llm_model       TEXT,                       -- какая модель генерировала
    llm_prompt_type TEXT,                       -- тип промпта (default, short, detailed)
    llm_generated_at TIMESTAMPTZ,
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Индекс для семантического поиска (500K записей вместо 75M!)
CREATE INDEX idx_book_summaries_embedding 
    ON book_summaries USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

CREATE INDEX idx_book_summaries_status ON book_summaries (status);
CREATE INDEX idx_book_summaries_sources ON book_summaries USING gin (sources);

-- Очередь задач на LLM-саммаризацию
CREATE TABLE llm_summary_tasks (
    id              BIGSERIAL PRIMARY KEY,
    book_id         BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    
    -- Причина запуска
    trigger_type    TEXT NOT NULL,              -- 'no_annotation', 'manual', 'popular', 'new_import', etc.
    triggered_by    UUID REFERENCES users(id),  -- кто запустил (NULL = система)
    priority        INTEGER NOT NULL DEFAULT 0, -- выше = важнее
    
    -- Параметры генерации
    prompt_type     TEXT DEFAULT 'default',     -- 'default', 'short', 'detailed', 'genre_specific'
    
    -- Статус
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, processing, done, failed
    error_message   TEXT,
    retry_count     INTEGER DEFAULT 0,
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX idx_llm_tasks_status ON llm_summary_tasks (status, priority DESC);
CREATE INDEX idx_llm_tasks_book ON llm_summary_tasks (book_id);

-- === Пользователи и аутентификация ===

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT UNIQUE NOT NULL,
    username        TEXT UNIQUE NOT NULL,
    display_name    TEXT NOT NULL,
    password_hash   TEXT NOT NULL,              -- bcrypt
    role            user_role NOT NULL DEFAULT 'user',
    avatar_url      TEXT,
    settings        JSONB DEFAULT '{}',         -- настройки читалки, тема и т.п.
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
    token_hash      TEXT NOT NULL,              -- SHA-256 хеш токена
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens (expires_at);

-- === Пользовательские данные (привязаны к user_id) ===

CREATE TYPE book_status AS ENUM (
    'want',         -- хочу прочитать
    'reading',      -- читаю
    'read',         -- прочитал
    'dropped',      -- бросил
    'favorite'      -- избранное
);

-- Статус книги у пользователя (хочу / читаю / прочитал / бросил)
CREATE TABLE user_books (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id     BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    status      book_status NOT NULL,
    rating      SMALLINT CHECK (rating >= 1 AND rating <= 10),
    notes       TEXT,                       -- личные заметки по книге
    started_at  TIMESTAMPTZ,                -- когда начал читать
    finished_at TIMESTAMPTZ,                -- когда закончил
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, book_id)
);
CREATE INDEX idx_user_books_user_status ON user_books (user_id, status);
CREATE INDEX idx_user_books_book ON user_books (book_id);

-- Прогресс чтения (per-user, per-book)
CREATE TABLE reading_progress (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    book_id     BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    position    JSONB NOT NULL,             -- {page, cfi, percent, chapter} зависит от формата
    device      TEXT,                       -- "desktop", "mobile" — для синхронизации
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, book_id)
);

-- Книжные полки (per-user)
CREATE TABLE shelves (
    id          SERIAL PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    is_public   BOOLEAN DEFAULT FALSE,      -- видна ли другим пользователям
    sort_order  INTEGER DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (user_id, name)                  -- у одного юзера не может быть двух полок с одним именем
);
CREATE INDEX idx_shelves_user ON shelves (user_id);

CREATE TABLE shelf_books (
    shelf_id    INTEGER NOT NULL REFERENCES shelves(id) ON DELETE CASCADE,
    book_id     BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    added_at    TIMESTAMPTZ DEFAULT NOW(),
    sort_order  INTEGER DEFAULT 0,
    PRIMARY KEY (shelf_id, book_id)
);

-- Агрегированный рейтинг книги (обновляется триггером)
-- Денормализация для быстрой сортировки каталога по рейтингу
ALTER TABLE books ADD COLUMN avg_rating    NUMERIC(3,1);
ALTER TABLE books ADD COLUMN rating_count  INTEGER DEFAULT 0;

CREATE OR REPLACE FUNCTION update_book_rating() RETURNS trigger AS $$
BEGIN
    UPDATE books SET
        avg_rating = (
            SELECT ROUND(AVG(rating)::numeric, 1)
            FROM user_books WHERE book_id = COALESCE(NEW.book_id, OLD.book_id)
            AND rating IS NOT NULL
        ),
        rating_count = (
            SELECT COUNT(*)
            FROM user_books WHERE book_id = COALESCE(NEW.book_id, OLD.book_id)
            AND rating IS NOT NULL
        )
    WHERE id = COALESCE(NEW.book_id, OLD.book_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_book_rating
    AFTER INSERT OR UPDATE OF rating OR DELETE ON user_books
    FOR EACH ROW EXECUTE FUNCTION update_book_rating();
```

---

## 4. Формат INPX и пайплайн импорта

### 4.1. Структура INPX-файла

INPX — это ZIP-архив с расширением `.inpx`, содержащий:

| Файл | Обязательность | Описание |
|------|----------------|----------|
| `collection.info` | Да | Метаданные коллекции |
| `version.info` | Да | Версия/дата коллекции |
| `structure.info` | Нет | Структура полей в .inp файлах |
| `archives.info` | Нет | Список ZIP-архивов с книгами |
| `*.inp` | Да | Описания книг (по одному на архив) |

#### collection.info

```
Lib.rus.ec Offline September 4, 2022    ← Название коллекции
librusec_all_local_2022-09-04           ← Идентификатор (имя без расширения)
65537                                    ← Тип: 0=fb2, 1=non-fb2, другие значения — флаги
Total: 636691 books                      ← Описание / статистика
http://lib.rus.ec/                       ← Источник (опционально)
```

**Применение в проекте:**
- Сохранять метаданные в таблице `collections` для поддержки нескольких библиотек
- Отображать информацию об источнике на странице статистики

#### version.info

```
20220904
```

Одна строка — версия коллекции в формате `YYYYMMDD`.

**Применение в проекте:**
- Хранить в `collections.version` для отслеживания обновлений
- При повторном импорте сравнивать версии — если новая версия больше, запускать инкрементальный импорт
- Показывать дату актуальности библиотеки в UI

#### structure.info

Определяет порядок полей в .inp файлах. Поля разделены `;`.

**Стандартная структура** (если файл отсутствует):
```
AUTHOR;GENRE;TITLE;SERIES;SERNO;FILE;SIZE;LIBID;DEL;EXT;DATE;
```

**Расширенная структура** (пример из librusec):
```
AUTHOR;GENRE;TITLE;SERIES;SERNO;FILE;SIZE;LIBID;DEL;EXT;DATE;INSNO;FOLDER;LANG;LIBRATE;KEYWORDS;
```

**Описание полей:**

| Поле | Тип | Формат | Описание |
|------|-----|--------|----------|
| `AUTHOR` | string | `Фамилия,Имя,Отчество:` | Авторы через `:`, части имени через `,` |
| `GENRE` | string | `genre_code:` | Жанры через `:` |
| `TITLE` | string | | Название книги |
| `SERIES` | string | | Серия (может содержать `[тип]` в конце: `[a]`=авторская, `[p]`=издательская) |
| `SERNO` | int | | Номер в серии |
| `FILE` | string | | Имя файла без расширения |
| `SIZE` | int | | Размер файла в байтах |
| `LIBID` | int | | ID книги в библиотеке (уникален в пределах коллекции) |
| `DEL` | int | | `1` = удалена, пусто или `0` = есть |
| `EXT` | string | | Расширение файла (`fb2`, `epub`, `pdf`, `djvu`) |
| `DATE` | string | `YYYY-MM-DD` | Дата добавления |
| `LANG` | string | | Язык (`ru`, `en`, `uk`, ...) |
| `INSNO` | int | | Номер вставки (порядок добавления в коллекцию) |
| `FOLDER` | string | | Имя архива (альтернатива определению по имени .inp) |
| `LIBRATE` | int | | Рейтинг библиотеки (1–5) |
| `KEYWORDS` | string | `keyword:` | Ключевые слова через `:` |

**Применение в проекте:**
- Парсер должен читать `structure.info` первым делом
- Динамически маппить поля по их именам, а не по позиции
- Неизвестные поля игнорировать (forward compatibility)

#### archives.info

Список всех ZIP-архивов коллекции:
```
fb2-000024-030559.zip
fb2-000065-572310_lost.zip
fb2-030560-060423.zip
...
usr-739000-740499.zip
```

**Применение в проекте:**
- **Валидация целостности:** при импорте проверять что все архивы из списка существуют на диске
- **Обработка `_lost` архивов:** суффикс `_lost` означает восстановленные/потерянные книги — их можно помечать флагом в БД
- **Инвентаризация без сканирования:** быстрый подсчёт архивов без обхода файловой системы
- **Опционально:** предупреждать пользователя о недостающих архивах

### 4.2. Формат записей .inp

Каждая строка — одна книга. Разделитель полей — `\x04` (EOT). Разделитель строк — `\r\n`.

**Пример записи:**
```
Булгаков,Михаил,Афанасьевич<0x04>dramaturgy<0x04>Мастер и Маргарита<0x04>
<0x04>0<0x04>94240<0x04>106260<0x04>94240<0x04><0x04>fb2<0x04>2007-06-20
<0x04>27<0x04>0<0x04>fb2-000024-030559.zip<0x04>ru<0x04>5<0x04>dramaturgy<0x04>
```

**Особенности парсинга:**
- Несколько авторов: `Бомонт,Френсис,:Флетчер,Джон,:`
- Несколько жанров: `tragedy:drama:` 
- Серия с типом: `Библиотека поэта[p]0` — `[p]` = издательская серия, число после — номер
- Пустые поля допустимы

### 4.3. Пайплайн импорта

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. Распаковка INPX                                              │
│                                                                 │
│    .inpx ──▶ collection.info                                    │
│              version.info                                       │
│              structure.info (опц.)                              │
│              archives.info (опц.)                               │
│              *.inp                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. Чтение метаданных коллекции                                  │
│                                                                 │
│    • Парсинг collection.info → upsert в таблицу collections     │
│    • Парсинг version.info → проверка: новая версия?             │
│    • Парсинг structure.info → определение маппинга полей        │
│    • Парсинг archives.info → валидация наличия архивов          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. Парсинг .inp файлов                                          │
│                                                                 │
│    Для каждого .inp:                                            │
│      • Определить архив: из FOLDER или из имени файла           │
│      • Для каждой строки:                                       │
│        - Разбить по \x04                                        │
│        - Маппить поля по structure.info                         │
│        - Парсить авторов (split по ":")                         │
│        - Парсить жанры (split по ":")                           │
│        - Парсить серию (извлечь [a]/[p] тип)                    │
│        - Добавить в batch                                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. Batch upsert в БД (по 1000-5000 за транзакцию)               │
│                                                                 │
│    1. Upsert авторов    → кеш map[string]int64                  │
│    2. Upsert жанров     → кеш map[string]int                    │
│    3. Upsert серий      → кеш map[string]int64                  │
│    4. Insert/update книг (ON CONFLICT по lib_id)                │
│    5. Insert связей M:N (book_authors, book_genres)             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. Post-processing (фоновые задачи)                             │
│                                                                 │
│    • Извлечение обложек из fb2/epub → сохранение на диск        │
│    • Извлечение аннотаций из fb2 <annotation>                   │
│    • Создание embed_tasks для новых книг                        │
│    • Пометка отсутствующих книг (DEL=1) как is_deleted          │
└─────────────────────────────────────────────────────────────────┘
```

### 4.5. Go-структуры для парсера

```go
// internal/inpx/parser.go

// Метаданные коллекции
type CollectionInfo struct {
    Name        string
    Code        string
    Type        int
    Description string
    SourceURL   string
}

// Структура полей (из structure.info или дефолтная)
type FieldMapping struct {
    Fields []string          // ["AUTHOR", "GENRE", "TITLE", ...]
    Index  map[string]int    // {"AUTHOR": 0, "GENRE": 1, ...}
}

var DefaultFieldMapping = FieldMapping{
    Fields: []string{"AUTHOR", "GENRE", "TITLE", "SERIES", "SERNO", 
                     "FILE", "SIZE", "LIBID", "DEL", "EXT", "DATE"},
}

// Запись о книге (универсальная, все возможные поля)
type BookRecord struct {
    Authors     []Author
    Genres      []string
    Title       string
    Series      string
    SeriesType  string    // "a" = авторская, "p" = издательская, "" = неизвестно
    SeriesNum   int
    FileName    string
    FileSize    int64
    LibID       string
    IsDeleted   bool
    Extension   string
    Date        string
    Language    string
    LibRate     int
    Keywords    []string
    ArchiveName string    // из FOLDER или из имени .inp файла
    InsNo       int
}

type Author struct {
    LastName   string
    FirstName  string
    MiddleName string
}

// Парсинг авторов: "Булгаков,Михаил,Афанасьевич:Петров,Иван,:"
func ParseAuthors(s string) []Author {
    var authors []Author
    for _, part := range strings.Split(s, ":") {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        names := strings.Split(part, ",")
        a := Author{}
        if len(names) > 0 { a.LastName = names[0] }
        if len(names) > 1 { a.FirstName = names[1] }
        if len(names) > 2 { a.MiddleName = names[2] }
        authors = append(authors, a)
    }
    return authors
}

// Парсинг серии: "Библиотека поэта[p]5" → ("Библиотека поэта", "p", 5)
func ParseSeries(s string) (name, seriesType string, num int) {
    // Извлечь тип серии [a] или [p]
    re := regexp.MustCompile(`^(.+?)\[([ap])\](\d*)$`)
    if m := re.FindStringSubmatch(s); m != nil {
        name = m[1]
        seriesType = m[2]
        if m[3] != "" {
            num, _ = strconv.Atoi(m[3])
        }
        return
    }
    // Без типа — просто имя
    name = s
    return
}
```

**Скорость импорта:** типичный INPX (600K+ книг) импортируется за 1–3 минуты.

---

## 5. Summary Embedding Pipeline

### 5.1. Концепция

Вместо эмбеддинга всего текста книги (75M чанков для 500K книг = 400+ GB) используем **эмбеддинг саммари** — краткого описания книги из разных источников.

```
┌────────────────────────────────────────────────────────────────────┐
│                            Книга                                   │
│                                                                    │
│  Источники саммари:                                                │
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐             │
│  │  Аннотация    │ │  Оглавление   │ │  Первые       │             │
│  │  (annotation) │ │  (TOC)        │ │  параграфы    │             │
│  └───────┬───────┘ └───────┬───────┘ └───────┬───────┘             │
│          │                 │                 │                     │
│          │     ┌───────────┴───────────┐     │                     │
│          │     │   Метаданные:         │     │                     │
│          │     │   автор, жанр, серия  │     │                     │
│          │     │   ключевые слова      │     │                     │
│          │     └───────────┬───────────┘     │                     │
│          │                 │                 │                     │
│          └─────────────────┼─────────────────┘                     │
│                            ▼                                       │
│               ┌────────────────────────┐                           │
│               │   Составное саммари    │                           │
│               │   (~500-2000 символов) │                           │
│               └────────────┬───────────┘                           │
│                            │                                       │
│            ┌───────────────┴───────────────┐                       │
│            ▼                               ▼                       │
│  ┌─────────────────────┐      ┌─────────────────────┐              │
│  │  Ollama Embedding   │      │  LLM Summarization  │              │
│  │  nomic-embed-text   │      │  (опционально)      │              │
│  │  → vector(768)      │      │  llama3 / mistral   │              │
│  └─────────────────────┘      └─────────────────────┘              │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

**Преимущества:**
- **500K векторов** вместо 75M (в 150 раз меньше)
- **~5 GB** вместо 400 GB (в 80 раз меньше)
- **~1.5 часа** обработки вместо 4-17 часов
- Поиск по **смыслу книги**, а не по случайным фрагментам

### 5.2. Конфигурация

```yaml
# config.yaml

summary:
  # Автоматическое извлечение саммари
  auto_extract: true
  max_annotation_length: 2000
  max_toc_entries: 20
  first_paragraphs_count: 3
  max_summary_length: 2000
  
  # Эмбеддинг
  embedding:
    enabled: true
    model: "nomic-embed-text"
    vector_dim: 768
    batch_size: 50
    concurrency: 9

  # LLM-саммаризация
  llm:
    enabled: true
    model: "llama3"                    # или mistral, gemma2
    max_input_tokens: 4000             # сколько текста отправлять в LLM
    max_output_tokens: 500             # максимум для саммари
    temperature: 0.3                   # низкая для консистентности
    
    # Автоматические триггеры
    auto_triggers:
      no_annotation: true              # книги без аннотации
      short_annotation: 100            # аннотация короче N символов
      popular_books: false             # популярные книги (много просмотров)
      high_rated: false                # высокий рейтинг
    
    # Лимиты
    daily_limit: 1000                  # максимум генераций в день
    queue_priority_boost_for_manual: 100  # приоритет ручных запросов

# Ollama-инстансы
ollama_pool:
  instances:
    - url: "http://192.168.1.50:11434"
      name: "pc-kitchen"
      capabilities: ["embedding", "llm"]
    - url: "http://192.168.1.51:11434"
      name: "pc-bedroom"  
      capabilities: ["embedding", "llm"]
    - url: "http://192.168.1.52:11434"
      name: "pc-office"
      capabilities: ["embedding"]      # только эмбеддинги
  
  fallback:
    enabled: true
    url: "http://ollama:11434"
    capabilities: ["embedding"]        # CPU — только эмбеддинги
    
  healthcheck_interval: "30s"
```

### 5.3. Сценарии запуска LLM-саммаризации

| Триггер | Описание | Приоритет |
|---------|----------|-----------|
| `manual_user` | Пользователь нажал "Улучшить описание" на странице книги | 100 |
| `manual_admin` | Админ запустил из панели управления | 90 |
| `batch_admin` | Админ запустил пакетную обработку (по жанру/автору) | 50 |
| `no_annotation` | Автоматически: книга без аннотации | 30 |
| `short_annotation` | Автоматически: аннотация < 100 символов | 25 |
| `poor_search_ctr` | Книга часто в результатах поиска, но не кликают | 20 |
| `high_rated` | Книга с высоким рейтингом (≥4.5) без LLM-саммари | 15 |
| `popular` | Книга с большим числом просмотров/скачиваний | 15 |
| `new_import` | Новая книга при импорте (если включено) | 10 |
| `series_complete` | Вся серия книг — для консистентности описаний | 10 |
| `regenerate` | Пересоздание по запросу (старое саммари плохое) | 80 |

### 5.4. Формирование саммари

```go
// backend/internal/service/summary.go

type SummaryBuilder struct {
    converters map[string]bookfile.BookConverter
    config     SummaryConfig
}

type BookSummary struct {
    Text       string          // итоговый текст для эмбеддинга
    Sources    []string        // откуда собрали
    Components SummaryComponents
}

type SummaryComponents struct {
    Annotation      string
    TOC             string
    FirstParagraphs string
    Metadata        string
    Keywords        string
    LLMSummary      string
}

func (b *SummaryBuilder) BuildSummary(ctx context.Context, book *models.Book, reader io.Reader) (*BookSummary, error) {
    var parts []string
    var sources []string
    comp := SummaryComponents{}
    
    // 1. Аннотация из книги (самый ценный источник)
    if annotation := b.extractAnnotation(book.Format, reader); annotation != "" {
        comp.Annotation = annotation
        parts = append(parts, annotation)
        sources = append(sources, "annotation")
    }
    
    // 2. Метаданные (всегда добавляем)
    meta := b.buildMetadataText(book)
    comp.Metadata = meta
    parts = append(parts, meta)
    sources = append(sources, "metadata")
    
    // 3. Оглавление (названия глав — отличный индикатор содержания)
    if toc := b.extractTOC(book.Format, reader); toc != "" {
        comp.TOC = toc
        parts = append(parts, toc)
        sources = append(sources, "toc")
    }
    
    // 4. Первые параграфы (если аннотации нет или она короткая)
    if len(comp.Annotation) < 200 {
        if intro := b.extractFirstParagraphs(book.Format, reader); intro != "" {
            comp.FirstParagraphs = intro
            parts = append(parts, "Начало книги: "+intro)
            sources = append(sources, "first_paragraphs")
        }
    }
    
    // 5. Ключевые слова из INPX
    if len(book.Keywords) > 0 {
        kw := "Ключевые слова: " + strings.Join(book.Keywords, ", ")
        comp.Keywords = kw
        parts = append(parts, kw)
        sources = append(sources, "keywords")
    }
    
    // Собираем итоговый текст
    summary := strings.Join(parts, "\n\n")
    if len(summary) > b.config.MaxSummaryLength {
        summary = summary[:b.config.MaxSummaryLength]
    }
    
    return &BookSummary{
        Text:       summary,
        Sources:    sources,
        Components: comp,
    }, nil
}

func (b *SummaryBuilder) buildMetadataText(book *models.Book) string {
    var parts []string
    
    parts = append(parts, fmt.Sprintf("Название: %s", book.Title))
    
    if book.AuthorName != "" {
        parts = append(parts, fmt.Sprintf("Автор: %s", book.AuthorName))
    }
    if book.GenreName != "" {
        parts = append(parts, fmt.Sprintf("Жанр: %s", book.GenreName))
    }
    if book.SeriesName != "" {
        s := fmt.Sprintf("Серия: %s", book.SeriesName)
        if book.SeriesNum > 0 {
            s += fmt.Sprintf(" (книга %d)", book.SeriesNum)
        }
        parts = append(parts, s)
    }
    if book.Year > 0 {
        parts = append(parts, fmt.Sprintf("Год: %d", book.Year))
    }
    
    return strings.Join(parts, ". ")
}

func (b *SummaryBuilder) extractAnnotation(format string, reader io.Reader) string {
    switch format {
    case "fb2":
        // <description><title-info><annotation>...</annotation>
        return b.converters["fb2"].ExtractAnnotation(reader)
    case "epub":
        // dc:description из OPF или calibre:annotation
        return b.converters["epub"].ExtractAnnotation(reader)
    default:
        return ""
    }
}

func (b *SummaryBuilder) extractTOC(format string, reader io.Reader) string {
    content, err := b.converters[format].Parse(reader)
    if err != nil || len(content.TOC) == 0 {
        return ""
    }
    
    var titles []string
    for i, entry := range content.TOC {
        if entry.Title != "" && i < b.config.MaxTOCEntries {
            titles = append(titles, entry.Title)
        }
    }
    
    if len(titles) == 0 {
        return ""
    }
    
    return "Содержание: " + strings.Join(titles, "; ")
}
```

### 5.5. LLM-саммаризация

```go
// backend/internal/service/llm_summary.go

type LLMSummarizer struct {
    pool   *OllamaPool
    config LLMConfig
}

type LLMPromptType string

const (
    PromptDefault     LLMPromptType = "default"
    PromptShort       LLMPromptType = "short"        // 1-2 предложения
    PromptDetailed    LLMPromptType = "detailed"     // развёрнутое описание
    PromptGenreAware  LLMPromptType = "genre_aware"  // с учётом жанра
)

var prompts = map[LLMPromptType]string{
    PromptDefault: `Напиши краткое описание книги (3-4 предложения) на основе предоставленной информации.
Описание должно передавать суть книги и заинтересовать читателя.
Не начинай с "Эта книга..." или "В этой книге...".
Пиши на русском языке.

Информация о книге:
%s

Краткое описание:`,

    PromptShort: `Опиши книгу в 1-2 предложениях, передав главную идею.
Информация: %s
Описание:`,

    PromptDetailed: `Напиши развёрнутое описание книги (5-7 предложений), включая:
- Основную тему и сюжет
- Главных героев (если применимо)
- Настроение и стиль
- Для кого подойдёт книга

Информация о книге:
%s

Описание:`,

    PromptGenreAware: `Ты эксперт по жанру "%s". Напиши описание книги, подчёркивая характерные для жанра элементы.
Информация: %s
Описание:`,
}

func (s *LLMSummarizer) GenerateSummary(ctx context.Context, book *models.Book, existingSummary *BookSummary, promptType LLMPromptType) (string, error) {
    // Собираем контекст для LLM
    var inputParts []string
    
    inputParts = append(inputParts, fmt.Sprintf("Название: %s", book.Title))
    inputParts = append(inputParts, fmt.Sprintf("Автор: %s", book.AuthorName))
    
    if book.GenreName != "" {
        inputParts = append(inputParts, fmt.Sprintf("Жанр: %s", book.GenreName))
    }
    if book.SeriesName != "" {
        inputParts = append(inputParts, fmt.Sprintf("Серия: %s", book.SeriesName))
    }
    
    // Добавляем существующие компоненты
    if existingSummary != nil {
        if existingSummary.Components.Annotation != "" {
            inputParts = append(inputParts, fmt.Sprintf("Аннотация: %s", existingSummary.Components.Annotation))
        }
        if existingSummary.Components.TOC != "" {
            inputParts = append(inputParts, existingSummary.Components.TOC)
        }
        if existingSummary.Components.FirstParagraphs != "" {
            inputParts = append(inputParts, fmt.Sprintf("Начало: %s", existingSummary.Components.FirstParagraphs))
        }
    }
    
    input := strings.Join(inputParts, "\n")
    
    // Обрезаем до лимита токенов
    if len(input) > s.config.MaxInputTokens*4 { // ~4 символа на токен
        input = input[:s.config.MaxInputTokens*4]
    }
    
    // Формируем промпт
    var prompt string
    if promptType == PromptGenreAware && book.GenreName != "" {
        prompt = fmt.Sprintf(prompts[promptType], book.GenreName, input)
    } else {
        prompt = fmt.Sprintf(prompts[promptType], input)
    }
    
    // Выбираем инстанс с LLM capability
    inst := s.pool.PickWithCapability("llm")
    if inst == nil {
        return "", fmt.Errorf("no LLM instances available")
    }
    
    // Генерируем
    response, err := s.pool.Generate(ctx, inst, s.config.Model, prompt, s.config.Temperature)
    if err != nil {
        return "", err
    }
    
    return strings.TrimSpace(response), nil
}
```

### 5.6. Воркер обработки саммари

```go
// backend/internal/worker/summary_worker.go

type SummaryWorker struct {
    repo           *repository.BookRepo
    summaryRepo    *repository.SummaryRepo
    archiveReader  *archive.Reader
    summaryBuilder *SummaryBuilder
    llmSummarizer  *LLMSummarizer
    embedder       *OllamaPool
    config         SummaryConfig
}

func (w *SummaryWorker) Run(ctx context.Context) {
    var wg sync.WaitGroup
    
    // 1. Воркеры извлечения саммари
    for i := 0; i < w.config.ExtractConcurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            w.extractLoop(ctx)
        }()
    }
    
    // 2. Воркеры эмбеддинга
    for i := 0; i < w.config.EmbedConcurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            w.embedLoop(ctx)
        }()
    }
    
    // 3. Воркеры LLM-саммаризации (меньше, т.к. дороже)
    for i := 0; i < w.config.LLMConcurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            w.llmLoop(ctx)
        }()
    }
    
    wg.Wait()
}

// Извлечение саммари из метаданных книги
func (w *SummaryWorker) extractLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Берём книги без саммари
        books, err := w.repo.GetBooksWithoutSummary(ctx, 50)
        if err != nil || len(books) == 0 {
            time.Sleep(5 * time.Second)
            continue
        }
        
        for _, book := range books {
            reader, err := w.archiveReader.OpenFile(book.ArchiveName, book.FileInArchive)
            if err != nil {
                w.summaryRepo.MarkFailed(ctx, book.ID, err.Error())
                continue
            }
            
            summary, err := w.summaryBuilder.BuildSummary(ctx, &book, reader)
            reader.Close()
            
            if err != nil {
                w.summaryRepo.MarkFailed(ctx, book.ID, err.Error())
                continue
            }
            
            // Сохраняем саммари (без эмбеддинга пока)
            w.summaryRepo.SaveSummary(ctx, book.ID, summary)
            
            // Проверяем, нужна ли LLM-саммаризация
            w.maybeQueueLLMTask(ctx, book.ID, summary)
        }
    }
}

// Создание эмбеддингов для саммари
func (w *SummaryWorker) embedLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Берём саммари без эмбеддинга
        summaries, err := w.summaryRepo.GetSummariesWithoutEmbedding(ctx, 50)
        if err != nil || len(summaries) == 0 {
            time.Sleep(5 * time.Second)
            continue
        }
        
        inst := w.embedder.Pick()
        if inst == nil {
            time.Sleep(30 * time.Second)
            continue
        }
        
        for _, sum := range summaries {
            embedding, err := w.embedder.Embed(ctx, inst, sum.SummaryText)
            if err != nil {
                slog.Error("embedding failed", "book_id", sum.BookID, "err", err)
                continue
            }
            
            w.summaryRepo.SaveEmbedding(ctx, sum.BookID, embedding)
        }
    }
}

// LLM-генерация для книг в очереди
func (w *SummaryWorker) llmLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }
        
        // Берём задачу из очереди (по приоритету)
        task, err := w.summaryRepo.PullLLMTask(ctx)
        if err != nil || task == nil {
            time.Sleep(10 * time.Second)
            continue
        }
        
        // Получаем книгу и существующее саммари
        book, _ := w.repo.GetByID(ctx, task.BookID)
        existingSummary, _ := w.summaryRepo.GetByBookID(ctx, task.BookID)
        
        // Генерируем LLM-саммари
        llmText, err := w.llmSummarizer.GenerateSummary(ctx, book, existingSummary, LLMPromptType(task.PromptType))
        if err != nil {
            w.summaryRepo.MarkLLMTaskFailed(ctx, task.ID, err.Error())
            continue
        }
        
        // Обновляем саммари с LLM-текстом
        w.summaryRepo.UpdateWithLLMSummary(ctx, task.BookID, llmText, w.config.LLM.Model, task.PromptType)
        w.summaryRepo.MarkLLMTaskDone(ctx, task.ID)
    }
}

// Проверка условий для автоматической LLM-саммаризации
func (w *SummaryWorker) maybeQueueLLMTask(ctx context.Context, bookID int64, summary *BookSummary) {
    cfg := w.config.LLM.AutoTriggers
    
    // Нет аннотации
    if cfg.NoAnnotation && summary.Components.Annotation == "" {
        w.summaryRepo.CreateLLMTask(ctx, bookID, "no_annotation", nil, 30, "default")
        return
    }
    
    // Короткая аннотация
    if cfg.ShortAnnotation > 0 && len(summary.Components.Annotation) < cfg.ShortAnnotation {
        w.summaryRepo.CreateLLMTask(ctx, bookID, "short_annotation", nil, 25, "default")
        return
    }
}
```

### 5.7. API для управления саммаризацией

```go
// backend/internal/handler/admin.go

// POST /api/admin/books/:id/generate-summary
// Ручной запуск LLM-саммаризации для одной книги
func (h *AdminHandler) GenerateSummary(c *gin.Context) {
    bookID := c.Param("id")
    userID := c.GetString("user_id")
    
    var req struct {
        PromptType string `json:"prompt_type"` // default, short, detailed, genre_aware
        Priority   int    `json:"priority"`    // опционально
    }
    c.BindJSON(&req)
    
    if req.PromptType == "" {
        req.PromptType = "default"
    }
    if req.Priority == 0 {
        req.Priority = 90 // высокий приоритет для ручных запросов
    }
    
    taskID, err := h.summaryRepo.CreateLLMTask(c, bookID, "manual_admin", userID, req.Priority, req.PromptType)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"task_id": taskID, "status": "queued"})
}

// POST /api/admin/summaries/batch-generate
// Пакетная LLM-саммаризация по критериям
func (h *AdminHandler) BatchGenerateSummaries(c *gin.Context) {
    userID := c.GetString("user_id")
    
    var req struct {
        Filter struct {
            GenreID      *int    `json:"genre_id"`
            AuthorID     *int64  `json:"author_id"`
            SeriesID     *int64  `json:"series_id"`
            NoAnnotation bool    `json:"no_annotation"`
            NoLLMSummary bool    `json:"no_llm_summary"`
            MinRating    float64 `json:"min_rating"`
        } `json:"filter"`
        PromptType string `json:"prompt_type"`
        Limit      int    `json:"limit"` // макс. книг
    }
    c.BindJSON(&req)
    
    if req.Limit == 0 || req.Limit > 1000 {
        req.Limit = 100
    }
    
    // Находим книги по критериям
    books, err := h.repo.FindBooksForSummarization(c, req.Filter, req.Limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // Создаём задачи
    created := 0
    for _, book := range books {
        _, err := h.summaryRepo.CreateLLMTask(c, book.ID, "batch_admin", userID, 50, req.PromptType)
        if err == nil {
            created++
        }
    }
    
    c.JSON(200, gin.H{
        "found":   len(books),
        "queued":  created,
        "message": fmt.Sprintf("Создано %d задач на LLM-саммаризацию", created),
    })
}

// GET /api/admin/summaries/stats
// Статистика саммаризации
func (h *AdminHandler) SummaryStats(c *gin.Context) {
    stats, _ := h.summaryRepo.GetStats(c)
    
    c.JSON(200, gin.H{
        "total_books":           stats.TotalBooks,
        "with_summary":          stats.WithSummary,
        "with_embedding":        stats.WithEmbedding,
        "with_llm_summary":      stats.WithLLMSummary,
        "without_annotation":    stats.WithoutAnnotation,
        "llm_queue_size":        stats.LLMQueueSize,
        "llm_processed_today":   stats.LLMProcessedToday,
        "llm_daily_limit":       h.config.LLM.DailyLimit,
    })
}

// POST /api/books/:id/improve-summary (для пользователей)
// Запрос на улучшение описания книги
func (h *BookHandler) RequestSummaryImprovement(c *gin.Context) {
    bookID := c.Param("id")
    userID := c.GetString("user_id")
    
    // Проверяем, не было ли недавно запроса от этого пользователя
    if h.summaryRepo.HasRecentRequest(c, bookID, userID, 24*time.Hour) {
        c.JSON(429, gin.H{"error": "Запрос уже отправлен, ожидайте обработки"})
        return
    }
    
    taskID, err := h.summaryRepo.CreateLLMTask(c, bookID, "manual_user", userID, 100, "default")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "task_id": taskID,
        "message": "Запрос на улучшение описания принят",
    })
}
```

### 5.8. Семантический поиск (упрощённый)

```go
// backend/internal/service/search.go

func (s *SearchService) SemanticSearch(ctx context.Context, query string, limit int) ([]BookResult, error) {
    // 1. Эмбеддинг запроса
    inst := s.pool.Pick()
    if inst == nil {
        return nil, fmt.Errorf("no ollama instances available")
    }
    
    queryVec, err := s.pool.Embed(ctx, inst, query)
    if err != nil {
        return nil, err
    }
    
    // 2. Поиск по саммари книг (не по чанкам!)
    rows, err := s.db.Query(ctx, `
        SELECT 
            b.id, b.title, b.lang, b.format,
            bs.summary_text,
            bs.embedding <=> $1::vector AS distance,
            a.name AS author_name,
            g.name AS genre_name
        FROM book_summaries bs
        JOIN books b ON b.id = bs.book_id
        LEFT JOIN book_authors ba ON ba.book_id = b.id
        LEFT JOIN authors a ON a.id = ba.author_id
        LEFT JOIN book_genres bg ON bg.book_id = b.id
        LEFT JOIN genres g ON g.id = bg.genre_id
        WHERE bs.embedding IS NOT NULL
          AND b.is_deleted = FALSE
        ORDER BY bs.embedding <=> $1::vector
        LIMIT $2
    `, pgvector.NewVector(queryVec), limit)
    
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var results []BookResult
    for rows.Next() {
        var r BookResult
        rows.Scan(&r.ID, &r.Title, &r.Lang, &r.Format, 
                  &r.Summary, &r.Distance, &r.Author, &r.Genre)
        results = append(results, r)
    }
    
    return results, nil
}

// Гибридный поиск: комбинация методов
func (s *SearchService) HybridSearch(ctx context.Context, query string, limit int) ([]BookResult, error) {
    var wg sync.WaitGroup
    var semantic, fulltext, fuzzy []BookResult
    var errSemantic, errFulltext, errFuzzy error
    
    // Параллельно запускаем три типа поиска
    wg.Add(3)
    
    go func() {
        defer wg.Done()
        semantic, errSemantic = s.SemanticSearch(ctx, query, limit*2)
    }()
    
    go func() {
        defer wg.Done()
        fulltext, errFulltext = s.FulltextSearch(ctx, query, limit*2)
    }()
    
    go func() {
        defer wg.Done()
        fuzzy, errFuzzy = s.FuzzySearch(ctx, query, limit)
    }()
    
    wg.Wait()
    
    // Объединяем результаты с весами
    return s.mergeResults(
        semantic, fulltext, fuzzy,
        limit,
        0.5,  // вес семантического поиска
        0.35, // вес полнотекстового
        0.15, // вес fuzzy (названия/авторы)
    ), nil
}
```

### 5.9. Оценка ресурсов

| Метрика | Chunk Embedding (старый) | Summary Embedding (новый) |
|---------|--------------------------|---------------------------|
| Записей в БД | 75 млн чанков | 500K саммари |
| Размер данных | 400-500 GB | ~5 GB |
| Запросов к Ollama | 75 млн | 500K (+ LLM по требованию) |
| Время обработки | 4-17 часов | ~1.5 часа |
| RAM для HNSW | 50-100 GB | ~2 GB |
| Качество поиска | Фрагменты текста | Смысл книги |

**Время LLM-саммаризации:**
- ~5 сек на книгу (llama3 на RTX 5060 Ti)
- 1000 книг/день = ~1.5 часа GPU-времени
- При 3 GPU можно обрабатывать ~3000 книг/день


### 5.10. Map-Reduce саммаризация (глубокий анализ)

Для особо важных книг или книг без метаданных доступна полная саммаризация через Map-Reduce — анализ всего текста книги с рекурсивным объединением.

**Запускается только вручную** администратором из-за высокой стоимости операции.

```
┌────────────────────────────────────────────────────────────────────────┐
│                       Map-Reduce Summarization                         │
│                                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Полный текст книги                           │   │
│  │                    (~300 KB, 500+ страниц)                      │   │
│  └───────────────────────────┬─────────────────────────────────────┘   │
│                              │                                         │
│                      ┌───────┴───────┐                                 │
│                      │   Chunking    │                                 │
│                      │ (~3K токенов) │                                 │
│                      │ overlap 5-10% │ ◀── Пересечение для контекста   │
│                      └───────┬───────┘                                 │
│                              │                                         │
│       ┌──────────────────────┼─────────────────────┐                   │
│       ▼                      ▼                     ▼                   │
│  ┌─────────┐            ┌─────────┐           ┌─────────┐              │
│  │ Chunk 1 │──overlap───│ Chunk 2 │──overlap──│ Chunk N │              │
│  └────┬────┘            └────┬────┘           └────┬────┘              │
│       │ MAP                  │ MAP                 │ MAP               │
│       ▼                      ▼                     ▼                   │
│  ┌───────────┐         ┌───────────┐         ┌───────────┐             │
│  │Summary 1  │         │Summary 2  │   ...   │Summary N  │             │
│  │(5-7 предл)│         │(5-7 предл)│         │(5-7 предл)│             │
│  └────┬──────┘         └─────┬─────┘         └─────┬─────┘             │
│       │                      │                     │                   │
│       └──────────────────────┼─────────────────────┘                   │
│                              │                                         │
│               ┌──────────────┴─────────────┐                           │
│               │      REDUCE Level 1        │                           │
│               │    (группы по 5 саммари)   │                           │
│               └──────────────┬─────────────┘                           │
│                              │                                         │
│               ┌──────────────┴─────────────┐                           │
│               │      REDUCE Level 2        │  ◀── Рекурсивно           │
│               │      (если нужно)          │      до max_level         │
│               └──────────────┬─────────────┘      или до 1 промпта     │
│                              │                                         │
│                              ▼                                         │
│                    ┌────────────────────┐                              │
│                    │  Финальное саммари │                              │
│                    │   (≤300 слов)      │                              │
│                    └────────────────────┘                              │
└────────────────────────────────────────────────────────────────────────┘
```

#### Оценка стоимости (показывается перед запуском)

| Размер книги | Чанков | MAP вызовов | REDUCE уровней | Всего LLM | Время (3 GPU) | Время (1 GPU) |
|--------------|--------|-------------|----------------|-----------|---------------|---------------|
| 50 KB | ~12 | 12 | 1 | ~15 | ~1 мин | ~2 мин |
| 100 KB | ~25 | 25 | 1 | ~30 | ~2 мин | ~4 мин |
| 300 KB | ~75 | 75 | 2 | ~90 | ~5 мин | ~12 мин |
| 500 KB | ~125 | 125 | 2 | ~145 | ~8 мин | ~20 мин |
| 1 MB | ~250 | 250 | 3 | ~290 | ~15 мин | ~40 мин |

#### Схема БД

```sql
-- Задачи Map-Reduce саммаризации
CREATE TABLE mapreduce_tasks (
    id              BIGSERIAL PRIMARY KEY,
    book_id         BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    triggered_by    UUID NOT NULL REFERENCES users(id),
    
    -- Параметры
    chunk_size      INTEGER NOT NULL DEFAULT 3000,      -- токенов
    chunk_overlap   INTEGER NOT NULL DEFAULT 300,       -- ~10% перехлёст
    max_reduce_level INTEGER DEFAULT NULL,              -- NULL = до 1 промпта
    
    -- Кастомные промпты (NULL = дефолтные)
    map_prompt      TEXT,
    reduce_prompt   TEXT,
    final_prompt    TEXT,
    
    -- Прогресс
    status          TEXT NOT NULL DEFAULT 'pending',
    -- pending → estimating → extracting → mapping → reducing → finalizing → done/failed
    
    total_chunks    INTEGER,
    mapped_chunks   INTEGER DEFAULT 0,
    reduce_levels   INTEGER DEFAULT 0,
    current_level   INTEGER DEFAULT 0,
    
    -- Оценка (заполняется на этапе estimating)
    estimated_llm_calls  INTEGER,
    estimated_time_sec   INTEGER,
    
    -- Результат
    final_summary   TEXT,
    
    -- Метаданные
    error_message   TEXT,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_mapreduce_tasks_status ON mapreduce_tasks (status);
CREATE INDEX idx_mapreduce_tasks_book ON mapreduce_tasks (book_id);

-- Промежуточные результаты (хранятся с возможностью очистки)
CREATE TABLE mapreduce_chunks (
    id              BIGSERIAL PRIMARY KEY,
    task_id         BIGINT NOT NULL REFERENCES mapreduce_tasks(id) ON DELETE CASCADE,
    
    level           INTEGER NOT NULL,       -- 0 = MAP, 1+ = REDUCE уровни
    chunk_index     INTEGER NOT NULL,
    
    input_text      TEXT NOT NULL,          -- входной текст
    output_summary  TEXT,                   -- результат (NULL пока не обработан)
    
    token_count     INTEGER,                -- кол-во токенов входа
    status          TEXT NOT NULL DEFAULT 'pending',
    
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);

CREATE INDEX idx_mapreduce_chunks_task ON mapreduce_chunks (task_id, level, chunk_index);
CREATE INDEX idx_mapreduce_chunks_pending ON mapreduce_chunks (task_id) 
    WHERE status = 'pending';
```

#### Конфигурация

```yaml
# config.yaml

mapreduce:
  enabled: true
  
  # Параметры чанкинга
  default_chunk_size: 3000        # токенов (~12000 символов)
  default_chunk_overlap: 300      # ~10% перехлёст
  min_chunk_size: 1000
  max_chunk_size: 6000
  
  # Reduce
  summaries_per_reduce: 5         # сколько саммари объединять за раз
  default_max_reduce_level: null  # null = до 1 промпта
  absolute_max_reduce_level: 5    # защита от бесконечности
  
  # Лимиты
  max_concurrent_tasks: 2         # одновременных задач
  max_parallel_maps: 10           # параллельных MAP на задачу
  max_book_size_mb: 5             # максимальный размер книги
  
  # Хранение
  keep_intermediate_days: 7       # сколько хранить промежуточные результаты
  
  # LLM
  model: "llama3"
  temperature: 0.3
  max_output_tokens: 500
```

#### Промпты

```go
// backend/internal/service/mapreduce_prompts.go

var DefaultMapPrompt = `Ты делаешь техническое саммари фрагмента книги.

Правила:
- Ровно 5-7 предложений
- Без интерпретаций и оценок
- Сохранить ключевые термины, имена, даты
- Сохранить причинно-следственные связи
- Не упоминать, что это саммари или фрагмент
- Писать в том же времени, что и оригинал

Текст фрагмента:
"""
{{chunk}}
"""

Саммари:`

var DefaultReducePrompt = `На входе набор саммари последовательных частей книги.

Задача:
- Удалить дословные повторы информации
- Объединить связанные идеи в единое повествование
- Сохранить хронологию и логическую структуру
- Сохранить все ключевые термины и имена
- Итог: не более 300 слов

Саммари частей:
"""
{{summaries}}
"""

Объединённое саммари:`

var DefaultFinalPrompt = `На основе подробного содержания книги напиши финальную аннотацию.

Требования:
- 4-6 предложений
- Передать основную тему и идею книги
- Заинтересовать потенциального читателя
- НЕ раскрывать ключевые повороты сюжета (спойлеры)
- Упомянуть жанр/тип книги, если очевидно
- НЕ начинать с "Эта книга..." или "В книге рассказывается..."

Подробное содержание:
"""
{{summary}}
"""

Аннотация:`
```

#### Реализация чанкинга с перехлёстом

```go
// backend/internal/service/mapreduce.go

// Разбиение текста на чанки с перехлёстом 5-10%
func (s *MapReduceService) chunkTextWithOverlap(text string, chunkTokens, overlapTokens int) []string {
    charsPerChunk := chunkTokens * 4    // ~4 символа на токен для русского
    overlapChars := overlapTokens * 4   // перехлёст в символах
    
    var chunks []string
    start := 0
    
    for start < len(text) {
        end := start + charsPerChunk
        if end > len(text) {
            end = len(text)
        }
        
        // Ищем конец предложения для чистого разреза
        if end < len(text) {
            for i := end; i > start+charsPerChunk*2/3; i-- {
                if text[i] == '.' || text[i] == '!' || text[i] == '?' || text[i] == '\n' {
                    if i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\n') {
                        end = i + 1
                        break
                    }
                }
            }
        }
        
        chunks = append(chunks, strings.TrimSpace(text[start:end]))
        
        // Следующий чанк начинается с перехлёстом (5-10% от размера чанка)
        start = end - overlapChars
        if start < 0 || start >= end {
            start = end
        }
    }
    
    return chunks
}
```

#### Основной алгоритм

```go
// Оценка стоимости перед запуском
func (s *MapReduceService) Estimate(ctx context.Context, bookID int64, params EstimateParams) (*CostEstimate, error) {
    book, _ := s.bookRepo.GetByID(ctx, bookID)
    textSize, _ := s.getTextSize(book)
    
    chunkSize := coalesce(params.ChunkSize, s.config.DefaultChunkSize)
    overlap := coalesce(params.ChunkOverlap, s.config.DefaultChunkOverlap)
    
    // Расчёт с учётом перехлёста
    charsPerChunk := chunkSize * 4
    overlapChars := overlap * 4
    effectiveChunk := charsPerChunk - overlapChars
    
    totalChunks := int(math.Ceil(float64(textSize) / float64(effectiveChunk)))
    
    // Расчёт REDUCE уровней
    reduceLevels := 0
    summariesCount := totalChunks
    for summariesCount > s.config.SummariesPerReduce {
        summariesCount = int(math.Ceil(float64(summariesCount) / float64(s.config.SummariesPerReduce)))
        reduceLevels++
    }
    
    totalLLMCalls := totalChunks + summariesCount + 1  // MAP + REDUCE + FINAL
    
    gpuCount := max(s.ollama.OnlineGPUCount(), 1)
    estimatedSeconds := (totalLLMCalls * 5) / gpuCount
    
    return &CostEstimate{
        BookID:           bookID,
        BookTitle:        book.Title,
        TextSizeKB:       textSize / 1024,
        TotalChunks:      totalChunks,
        ReduceLevels:     reduceLevels,
        TotalLLMCalls:    totalLLMCalls,
        EstimatedTimeSec: estimatedSeconds,
        EstimatedTimeStr: formatDuration(estimatedSeconds),
        GPUCount:         gpuCount,
    }, nil
}

// Основной процесс
func (s *MapReduceService) Process(ctx context.Context, taskID int64) error {
    task, _ := s.taskRepo.GetByID(ctx, taskID)
    
    // 1. EXTRACTING
    s.updateStatus(ctx, taskID, "extracting")
    book, _ := s.bookRepo.GetByID(ctx, task.BookID)
    fullText, _ := s.extractFullText(book)
    
    // 2. Чанкинг с перехлёстом
    chunks := s.chunkTextWithOverlap(fullText, task.ChunkSize, task.ChunkOverlap)
    for i, chunk := range chunks {
        s.taskRepo.CreateChunk(ctx, taskID, 0, i, chunk, s.estimateTokens(chunk))
    }
    
    // 3. MAPPING (параллельно)
    s.updateStatus(ctx, taskID, "mapping")
    err := s.processMapPhase(ctx, taskID, task.MapPrompt)
    if err != nil {
        return s.failTask(ctx, taskID, err)
    }
    
    // 4. REDUCING (рекурсивно)
    s.updateStatus(ctx, taskID, "reducing")
    finalSummary, err := s.processReducePhase(ctx, taskID, task.ReducePrompt, task.MaxReduceLevel)
    if err != nil {
        return s.failTask(ctx, taskID, err)
    }
    
    // 5. FINALIZING
    s.updateStatus(ctx, taskID, "finalizing")
    finalAnnotation, _ := s.generateFinal(ctx, finalSummary, task.FinalPrompt)
    
    // 6. Сохранение
    s.taskRepo.Complete(ctx, taskID, finalAnnotation)
    s.updateBookSummary(ctx, task.BookID, finalAnnotation, finalSummary)
    
    return nil
}

// REDUCE с настраиваемой глубиной
func (s *MapReduceService) processReducePhase(ctx context.Context, taskID int64, prompt string, maxLevel *int) (string, error) {
    if prompt == "" {
        prompt = DefaultReducePrompt
    }
    
    level := 1
    for {
        prevChunks, _ := s.taskRepo.GetChunksByLevel(ctx, taskID, level-1)
        var summaries []string
        for _, c := range prevChunks {
            summaries = append(summaries, c.OutputSummary)
        }
        
        combined := strings.Join(summaries, "\n\n---\n\n")
        
        // Условие остановки: влезает в контекст
        if s.estimateTokens(combined) <= s.config.DefaultChunkSize {
            return combined, nil
        }
        
        // ИЛИ достигли maxLevel — принудительно сжимаем в один вызов
        if maxLevel != nil && level > *maxLevel {
            return s.forceCombine(ctx, combined, prompt)
        }
        
        // Защита от бесконечности
        if level > s.config.AbsoluteMaxReduceLevel {
            return "", fmt.Errorf("too many reduce levels")
        }
        
        // Группируем и редуцируем
        s.taskRepo.UpdateReduceLevel(ctx, taskID, level)
        groups := s.groupSummaries(summaries, s.config.SummariesPerReduce)
        
        for i, group := range groups {
            groupText := strings.Join(group, "\n\n---\n\n")
            p := strings.Replace(prompt, "{{summaries}}", groupText, 1)
            
            chunkID, _ := s.taskRepo.CreateChunk(ctx, taskID, level, i, groupText, 0)
            result, _ := s.ollama.Generate(ctx, s.config.Model, p)
            s.taskRepo.UpdateChunkResult(ctx, chunkID, strings.TrimSpace(result))
        }
        
        level++
    }
}
```

#### API

```go
// GET /api/admin/books/:id/mapreduce-estimate
// Оценка стоимости перед запуском
func (h *AdminHandler) EstimateMapReduce(c *gin.Context) {
    bookID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    
    estimate, err := h.mapreduceService.Estimate(c, bookID, EstimateParams{
        ChunkSize:    c.QueryInt("chunk_size"),
        ChunkOverlap: c.QueryInt("chunk_overlap"),
    })
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "book_id":            estimate.BookID,
        "book_title":         estimate.BookTitle,
        "text_size_kb":       estimate.TextSizeKB,
        "total_chunks":       estimate.TotalChunks,
        "reduce_levels":      estimate.ReduceLevels,
        "total_llm_calls":    estimate.TotalLLMCalls,
        "estimated_time":     estimate.EstimatedTimeStr,
        "estimated_time_sec": estimate.EstimatedTimeSec,
        "gpu_count":          estimate.GPUCount,
    })
}

// POST /api/admin/books/:id/mapreduce-summary
// Запуск Map-Reduce саммаризации
func (h *AdminHandler) StartMapReduce(c *gin.Context) {
    bookID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    userID := c.GetString("user_id")
    
    var req struct {
        ChunkSize      int    `json:"chunk_size"`       // опционально
        ChunkOverlap   int    `json:"chunk_overlap"`    // опционально, default ~10%
        MaxReduceLevel *int   `json:"max_reduce_level"` // null = до 1 промпта
        MapPrompt      string `json:"map_prompt"`       // опционально
        ReducePrompt   string `json:"reduce_prompt"`    // опционально
        FinalPrompt    string `json:"final_prompt"`     // опционально
    }
    c.BindJSON(&req)
    
    taskID, err := h.mapreduceService.Start(c, StartRequest{
        BookID:         bookID,
        UserID:         userID,
        ChunkSize:      req.ChunkSize,
        ChunkOverlap:   req.ChunkOverlap,
        MaxReduceLevel: req.MaxReduceLevel,
        MapPrompt:      req.MapPrompt,
        ReducePrompt:   req.ReducePrompt,
        FinalPrompt:    req.FinalPrompt,
    })
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(202, gin.H{"task_id": taskID, "status": "pending"})
}

// GET /api/admin/mapreduce/:id — статус задачи
// GET /api/admin/mapreduce/:id/chunks?level=0 — промежуточные результаты
// DELETE /api/admin/mapreduce/:id — отмена
// DELETE /api/admin/mapreduce/:id/chunks — очистка промежуточных данных
// GET /api/admin/mapreduce/active — список активных задач
```

#### Автоматическая очистка

```go
// Запускается по cron раз в день
func (w *CleanupWorker) CleanupOldMapReduceChunks(ctx context.Context) {
    // Удаляем промежуточные результаты старше N дней для завершённых задач
    deleted, _ := w.repo.Exec(ctx, `
        DELETE FROM mapreduce_chunks
        WHERE task_id IN (
            SELECT id FROM mapreduce_tasks
            WHERE status IN ('done', 'failed', 'cancelled')
              AND completed_at < NOW() - INTERVAL '$1 days'
        )
    `, w.config.KeepIntermediateDays)
    
    slog.Info("cleaned up mapreduce chunks", "deleted", deleted)
}
```

---
## 6. Docker Compose

Пример конфигурации для продакшна (`docker/docker-compose.prod.yml`):

```yaml
version: "3.8"

services:
  # --- База данных ---
  postgres:
    image: pgvector/pgvector:pg16
    environment:
      POSTGRES_DB: homelib
      POSTGRES_USER: homelib
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "127.0.0.1:5432:5432"
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U homelib"]
      interval: 10s
      timeout: 5s
      retries: 5

  # --- API сервер ---
  api:
    build:
      context: ../backend
      dockerfile: ../docker/backend/Dockerfile.api
    environment:
      DB_URL: postgres://homelib:${DB_PASSWORD}@postgres:5432/homelib?sslmode=disable
      JWT_SECRET: ${JWT_SECRET}
      LIBRARY_PATH: /library
      CACHE_PATH: /cache
      CONFIG_PATH: /config/config.yaml
    volumes:
      - ${LIBRARY_PATH}:/library:ro
      - cache_data:/cache
      - ../config/config.prod.yaml:/config/config.yaml:ro
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  # --- Фоновый worker ---
  worker:
    build:
      context: ../backend
      dockerfile: ../docker/backend/Dockerfile.worker
    environment:
      DB_URL: postgres://homelib:${DB_PASSWORD}@postgres:5432/homelib?sslmode=disable
      LIBRARY_PATH: /library
      CACHE_PATH: /cache
      CONFIG_PATH: /config/config.yaml
    volumes:
      - ${LIBRARY_PATH}:/library:ro
      - cache_data:/cache
      - ../config/config.prod.yaml:/config/config.yaml:ro
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped

  # --- CPU Ollama fallback (опционально) ---
  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    ports:
      - "127.0.0.1:11434:11434"
    restart: unless-stopped
    profiles:
      - cpu-fallback      # запускается только с --profile cpu-fallback

  # --- Frontend + Reverse proxy ---
  nginx:
    build:
      context: ../frontend
      dockerfile: ../docker/frontend/Dockerfile
    volumes:
      - ./nginx/nginx.prod.conf:/etc/nginx/nginx.conf:ro
    ports:
      - "8080:80"
    depends_on:
      - api
    restart: unless-stopped

volumes:
  pgdata:
  ollama_data:
  cache_data:
```

**Запуск:**

```bash
# Из директории docker/
cd docker

# Разработка (hot-reload)
docker compose -f docker-compose.dev.yml up -d

# Staging
docker compose -f docker-compose.stage.yml up -d

# Продакшн
docker compose -f docker-compose.prod.yml up -d

# С CPU fallback (если нужен Ollama и на сервере)
docker compose -f docker-compose.prod.yml --profile cpu-fallback up -d
```

**`.env`:**

```env
DB_PASSWORD=changeme_strong_password
JWT_SECRET=changeme_generate_with_openssl_rand_hex_32
LIBRARY_PATH=/mnt/storage/library
```

---

## 7. Структура проекта

```
homelib/
├── backend/                            # Go-бэкенд
│   ├── cmd/                            # Точки входа
│   │   ├── api/
│   │   │   └── main.go                 # API-сервер
│   │   └── worker/
│   │       └── main.go                 # Фоновый воркер
│   │
│   ├── internal/                       # Внутренний код (не импортируется извне)
│   │   ├── api/                        # HTTP-слой (обёртка для Gin)
│   │   │   ├── handler/               # HTTP-хендлеры
│   │   │   │   ├── auth.go            # /api/auth/*
│   │   │   │   ├── books.go           # /api/books/*
│   │   │   │   ├── reader.go          # /api/books/:id/content, /chapter/:n
│   │   │   │   ├── authors.go
│   │   │   │   ├── genres.go
│   │   │   │   ├── series.go
│   │   │   │   ├── search.go
│   │   │   │   ├── me.go              # /api/me/*
│   │   │   │   ├── admin.go           # /api/admin/*
│   │   │   │   └── download.go        # /api/books/:id/download
│   │   │   ├── middleware/            # Middleware
│   │   │   │   └── auth.go            # JWT проверка, RequireAuth, RequireAdmin
│   │   │   ├── router.go              # Маршрутизация (SetupRouter)
│   │   │   └── server.go              # HTTP-сервер (graceful shutdown)
│   │   ├── config/
│   │   │   └── config.go               # Конфигурация (YAML + env)
│   │   ├── models/                     # Доменные модели
│   │   │   ├── book.go
│   │   │   ├── author.go
│   │   │   ├── genre.go
│   │   │   ├── collection.go           # Collection (метаданные библиотеки)
│   │   │   ├── series.go               # Series, SeriesListItem
│   │   │   ├── user.go                 # User, RefreshToken
│   │   │   ├── user_book.go            # UserBook (статус, оценка)
│   │   │   ├── reading_progress.go     # ReadingProgress (позиция чтения)
│   │   │   ├── shelf.go                # Shelf, ShelfBook
│   │   │   └── summary.go              # BookSummary, LLMSummaryTask
│   │   ├── repository/                 # Слой доступа к БД (pgx)
│   │   │   ├── db.go                  # NewPool, RunMigrations
│   │   │   ├── book.go
│   │   │   ├── author.go
│   │   │   ├── genre.go
│   │   │   ├── collection.go
│   │   │   ├── series.go              # SeriesRepo
│   │   │   ├── user.go                # CRUD пользователей
│   │   │   ├── refresh_token.go       # CRUD refresh-токенов
│   │   │   ├── user_book.go            # Статусы, оценки, прогресс, полки
│   │   │   ├── summary.go              # Саммари, LLM-задачи, эмбеддинги
│   │   │   └── search.go               # Полнотекстовый + векторный поиск
│   │   ├── service/                    # Бизнес-логика
│   │   │   ├── auth.go                 # Регистрация, логин, JWT, refresh
│   │   │   ├── catalog.go              # Каталогизация, фильтрация
│   │   │   ├── user_library.go         # Статусы книг, оценки, полки, прогресс
│   │   │   ├── import.go               # Импорт .inpx
│   │   │   ├── reader.go               # Конвертация книг для читалки, кеширование
│   │   │   ├── summary.go              # SummaryBuilder — извлечение саммари
│   │   │   ├── llm_summary.go          # LLMSummarizer — генерация через LLM
│   │   │   ├── download.go             # Извлечение из ZIP
│   │   │   └── search.go               # Гибридный поиск
│   │   ├── inpx/                       # Парсер .inpx / .inp
│   │   │   ├── parser.go
│   │   │   ├── records.go             # Парсинг записей .inp
│   │   │   └── types.go               # BookRecord, Author, CollectionInfo
│   │   ├── bookfile/                   # Конвертеры форматов книг
│   │   │   ├── converter.go            # Интерфейс BookConverter
│   │   │   ├── fb2.go                  # FB2 → HTML + извлечение аннотации
│   │   │   ├── epub.go                 # EPUB → HTML + извлечение аннотации
│   │   │   ├── pdf.go                  # PDF → HTML (pdftohtml)
│   │   │   └── djvu.go                 # DJVU → HTML (djvutxt)
│   │   ├── ollama/                     # Ollama Pool
│   │   │   ├── pool.go                 # OllamaPool: healthcheck, балансировка
│   │   │   ├── embed.go                # Embed() — получение эмбеддинга
│   │   │   └── generate.go             # Generate() — LLM-генерация текста
│   │   ├── worker/                     # Фоновые воркеры
│   │   │   ├── summary_worker.go       # Извлечение + эмбеддинг саммари
│   │   │   └── llm_worker.go           # LLM-саммаризация по очереди
│   │   └── archive/                    # Работа с ZIP-архивами
│   │       └── reader.go
│   │
│   ├── migrations/                     # SQL-миграции (golang-migrate)
│   │   ├── embed.go                   # go:embed *.sql
│   │   ├── 001_init.up.sql
│   │   ├── 001_init.down.sql
│   │   ├── 002_add_unique_constraints.up.sql
│   │   ├── 002_add_unique_constraints.down.sql
│   │   ├── 003_user_data.up.sql
│   │   ├── 003_user_data.down.sql
│   │   ├── 004_embedding.up.sql
│   │   └── 004_embedding.down.sql
│   │
│   ├── config.example.yaml            # Шаблон конфигурации
│   ├── go.mod
│   ├── go.sum
│   └── Makefile                        # Make-таргеты для бэкенда
│
├── frontend/                           # Vue 3 SPA
│   ├── src/
│   │   ├── components/
│   │   │   ├── AppHeader.vue          # Навигация, user menu (layout)
│   │   │   ├── common/                 # Общие компоненты
│   │   │   │   ├── BookCard.vue
│   │   │   │   ├── BookFilters.vue    # Фильтры каталога
│   │   │   │   ├── PaginationBar.vue  # Пагинация с выбором limit
│   │   │   │   ├── SearchBar.vue
│   │   │   │   ├── BookStatusButton.vue
│   │   │   │   ├── BookRating.vue
│   │   │   │   ├── GenreTree.vue
│   │   │   │   ├── ShelfList.vue
│   │   │   │   ├── UserMenu.vue
│   │   │   │   └── EmbeddingStatus.vue
│   │   │   └── reader/                 # Браузерная читалка
│   │   │       ├── BookReader.vue      # Главный контейнер
│   │   │       ├── ReaderContent.vue   # Область контента (пагинация/скролл)
│   │   │       ├── ReaderHeader.vue    # Верхняя панель
│   │   │       ├── ReaderFooter.vue    # Прогресс-бар, номер страницы
│   │   │       ├── ReaderSettings.vue  # Модальное окно настроек
│   │   │       ├── ReaderTOC.vue       # Оглавление
│   │   │       ├── ReaderBookmarks.vue # Закладки и заметки
│   │   │       ├── ReaderSearch.vue    # Поиск по книге
│   │   │       └── ReaderFontPicker.vue
│   │   ├── views/
│   │   │   ├── LoginView.vue
│   │   │   ├── HomeView.vue
│   │   │   ├── CatalogView.vue
│   │   │   ├── BookView.vue
│   │   │   ├── AuthorsView.vue        # Список авторов с поиском
│   │   │   ├── AuthorView.vue
│   │   │   ├── GenresView.vue         # Дерево жанров
│   │   │   ├── SeriesView.vue         # Список серий с поиском
│   │   │   ├── AdminImportView.vue    # Управление импортом INPX
│   │   │   ├── ReaderView.vue          # Страница читалки (обёртка над BookReader)
│   │   │   ├── SearchView.vue
│   │   │   ├── MyBooksView.vue
│   │   │   ├── MyShelvesView.vue
│   │   │   └── ProfileView.vue
│   │   ├── composables/                # Композиции (логика)
│   │   │   ├── useBookContent.ts       # Загрузка контента книги с API
│   │   │   ├── usePagination.ts        # Разбивка на страницы
│   │   │   ├── useReaderSettings.ts    # Управление настройками читалки
│   │   │   ├── useReaderGestures.ts    # Свайпы и тапы
│   │   │   ├── useReaderKeyboard.ts    # Горячие клавиши
│   │   │   ├── useTextSelection.ts     # Выделение → закладка/цитата
│   │   │   └── useReadingProgress.ts   # Сохранение/загрузка прогресса
│   │   ├── stores/                     # Pinia
│   │   │   ├── auth.ts
│   │   │   ├── catalog.ts
│   │   │   ├── reader.ts               # Состояние читалки
│   │   │   └── userLibrary.ts
│   │   ├── api/                        # HTTP-клиент
│   │   │   ├── client.ts              # Axios instance, interceptors
│   │   │   ├── auth.ts
│   │   │   ├── books.ts
│   │   │   ├── admin.ts               # Импорт INPX API
│   │   │   └── me.ts
│   │   ├── plugins/
│   │   │   └── vuetify.ts             # Конфигурация Vuetify 3
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── types/                      # TypeScript типы
│   │   │   ├── book.ts
│   │   │   ├── user.ts
│   │   │   └── reader.ts               # ReaderSettings, ReadingPosition
│   │   ├── assets/
│   │   │   └── styles/
│   │   │       ├── main.css
│   │   │       └── reader-themes.css   # Темы читалки (light/sepia/dark/night)
│   │   ├── App.vue                    # Корневой компонент
│   │   ├── main.ts                    # Точка входа
│   │   └── env.d.ts                   # TypeScript declarations
│   ├── public/
│   ├── index.html
│   ├── nginx.conf                     # SPA-конфиг nginx внутри контейнера
│   ├── vite.config.ts
│   ├── vitest.config.ts               # Конфигурация тестов Vitest
│   ├── tsconfig.json
│   ├── tsconfig.node.json
│   └── package.json
│
├── docker/                             # Docker-конфигурация
│   ├── backend/
│   │   ├── Dockerfile.api              # Образ API-сервера
│   │   └── Dockerfile.worker           # Образ воркера
│   ├── frontend/
│   │   └── Dockerfile                  # Сборка Vue + nginx
│   ├── nginx/
│   │   ├── nginx.dev.conf             # Конфиг для разработки
│   │   └── nginx.prod.conf            # Конфиг для production (security headers, rate limiting)
│   ├── docker-compose.dev.yml          # Разработка
│   ├── docker-compose.stage.yml        # Staging
│   └── docker-compose.prod.yml         # Продакшн
│
├── scripts/                            # Скрипты автоматизации
│   ├── build.sh                        # Сборка Go-бинарников
│   ├── deploy.sh                       # Деплой на сервер
│   ├── migrate.sh                      # Запуск миграций
│   ├── backup-db.sh                    # Бэкап PostgreSQL
│   ├── restore-db.sh                   # Восстановление из бэкапа
│   ├── import-inpx.sh                  # CLI для импорта .inpx
│   ├── setup-ollama-windows.ps1        # Установка Ollama на Windows
│   └── dev-setup.sh                    # Настройка окружения разработчика
│
├── config/                             # Конфигурационные файлы
│   ├── config.dev.yaml
│   ├── config.stage.yaml
│   ├── config.prod.yaml
│   └── genres.json                     # Справочник жанров
│
├── .github/
│   └── workflows/
│       └── ci.yml                     # CI pipeline (lint, test, build, docker)
│
├── .env.example
├── .gitignore
├── .golangci.yml                      # Конфигурация Go-линтера
├── Makefile                            # Корневой Makefile (вызывает backend/frontend)
└── README.md
```

> **Примечание:** Тест-файлы (`*_test.go`, `__tests__/*.test.ts`) расположены рядом с исходным кодом, но опущены в дереве для краткости.

### Описание ключевых директорий

| Директория | Назначение |
|------------|------------|
| `backend/` | Go-бэкенд: API-сервер и воркер |
| `backend/cmd/` | Точки входа (минимум кода, только инициализация) |
| `backend/internal/` | Основной код, не экспортируется как библиотека |
| `backend/internal/api/` | HTTP-слой: handler/, middleware/, router.go, server.go |
| `backend/internal/bookfile/` | Конвертеры форматов книг (FB2/EPUB/PDF/DJVU → HTML) |
| `backend/internal/ollama/` | Ollama Pool: балансировка, embed, generate |
| `backend/internal/worker/` | Фоновые воркеры: саммаризация, LLM |
| `backend/migrations/` | SQL-миграции, версионирование схемы БД |
| `frontend/` | Vue 3 SPA, отдельный npm-проект |
| `frontend/src/api/` | HTTP-клиент (axios), типы API |
| `frontend/src/views/` | Vue-страницы (*View.vue) |
| `frontend/src/components/common/` | Общие UI-компоненты |
| `frontend/src/components/reader/` | Компоненты браузерной читалки |
| `frontend/src/composables/` | Логика читалки (пагинация, жесты, настройки) |
| `docker/` | Dockerfile'ы и compose-файлы для разных окружений |
| `scripts/` | Bash/PowerShell скрипты для автоматизации |
| `config/` | YAML-конфиги для разных окружений |

### Docker Compose окружения

| Файл | Назначение | Особенности |
|------|------------|-------------|
| `docker/docker-compose.dev.yml` | Локальная разработка | Hot-reload, volume mounts для кода, debug-порты, Vite dev server |
| `docker/docker-compose.stage.yml` | Staging/тестирование | Собранные образы, тестовые данные, логирование |
| `docker/docker-compose.prod.yml` | Продакшн | Оптимизированные образы, ограничения ресурсов, healthchecks |

---

## 8. Браузерная читалка

Единый интерфейс для чтения всех форматов (FB2, EPUB, PDF, DJVU) — по образцу десктопных читалок AlReader, CoolReader, FBReader.

### 8.1. Архитектура

```
┌────────────────────────────────────────────────────────────────────┐
│                        Браузерная читалка                          │
│                                                                    │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    Единый UI читалки                        │   │
│  │  ┌─────────┐ ┌─────────────────────────────┐ ┌─────────┐    │   │
│  │  │ Боковая │ │                             │ │ Боковая │    │   │
│  │  │ панель  │ │      Область контента       │ │ панель  │    │   │
│  │  │         │ │                             │ │         │    │   │
│  │  │ • TOC   │ │   Пагинация / Скролл        │ │• Закл.  │    │   │
│  │  │ • Поиск │ │   Настраиваемые стили       │ │• Заметки│    │   │
│  │  │ • Инфо  │ │   Выделение текста          │ │• Цитаты │    │   │
│  │  └─────────┘ └─────────────────────────────┘ └─────────┘    │   │
│  │                                                             │   │
│  │  ┌────────────────────────────────────────────────────────┐ │   │
│  │  │ Toolbar: шрифт, размер, тема, яркость, поля, интервал  │ │   │
│  │  └────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                              ▲                                     │
│                              │ Унифицированный HTML                │
│  ┌───────────────────────────┴───────────────────────────────────┐ │
│  │                   Content Adapter Layer                       │ │
│  │                                                               │ │
│  │   • getMetadata() → {title, author, cover}                    │ │
│  │   • getTOC() → [{id, title, level}]                           │ │
│  │   • getChapter(id) → HTML                                     │ │
│  │   • search(query) → [{chapterId, snippet, position}]          │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                              ▲                                     │
│                              │ HTTP API                            │
└──────────────────────────────┼─────────────────────────────────────┘
                               │
┌──────────────────────────────┼─────────────────────────────────────┐
│                        Бэкенд (Go)                                 │
│                              │                                     │
│  ┌───────────────────────────┴───────────────────────────────────┐ │
│  │                   Book Content Service                        │ │
│  │                                                               │ │
│  │   GET /api/books/:id/content → метаданные + TOC + список глав │ │
│  │   GET /api/books/:id/chapter/:n → HTML контент главы          │ │
│  │   GET /api/books/:id/search?q= → результаты поиска            │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                              ▲                                     │
│         ┌────────────────────┼────────────────────┐                │
│         │                    │                    │                │
│  ┌──────┴──────┐    ┌────────┴────────┐    ┌──────┴──────┐         │
│  │FB2 Converter│    │ EPUB Converter  │    │PDF Converter│         │
│  │ Go XML→HTML │    │ Go unzip+XHTML  │    │ pdftohtml   │         │
│  └─────────────┘    └─────────────────┘    └─────────────┘         │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │                    Converted Content Cache                   │  │
│  │              (файловая система или Redis)                    │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
```

**Ключевой принцип:** все форматы конвертируются на бэкенде в унифицированный HTML. Фронтенд работает только с HTML и не знает об исходном формате книги.

### 8.2. API контента книги

```
GET /api/books/:id/content
```

Возвращает метаданные и структуру книги:

```json
{
  "metadata": {
    "title": "Мастер и Маргарита",
    "author": "Михаил Булгаков",
    "cover": "/api/books/123/cover",
    "language": "ru",
    "format": "fb2"
  },
  "toc": [
    {"id": "ch1", "title": "Часть первая", "level": 0},
    {"id": "ch1-1", "title": "Глава 1. Никогда не разговаривайте с неизвестными", "level": 1},
    {"id": "ch1-2", "title": "Глава 2. Понтий Пилат", "level": 1}
  ],
  "chapters": ["ch1", "ch1-1", "ch1-2", "..."],
  "totalChapters": 32
}
```

```
GET /api/books/:id/chapter/:chapterId
```

Возвращает HTML-контент главы (чистый, без стилей — стили применяет читалка):

```html
<h2 class="chapter-title">Глава 1. Никогда не разговаривайте с неизвестными</h2>
<p>Однажды весною, в час небывало жаркого заката, в Москве, на Патриарших прудах, появились два гражданина.</p>
<p>Первый из них, одетый в летнюю серенькую пару...</p>
```

```
GET /api/books/:id/search?q=Воланд
```

Поиск по тексту книги:

```json
{
  "results": [
    {"chapterId": "ch1-1", "snippet": "...представился: — Воланд. — Немец?..", "position": 1523},
    {"chapterId": "ch1-3", "snippet": "...Воланд усмехнулся...", "position": 4201}
  ],
  "total": 47
}
```

### 8.3. Конвертеры форматов (бэкенд)

```go
// backend/internal/bookfile/converter.go

// Унифицированный результат конвертации
type BookContent struct {
    Metadata   BookMetadata `json:"metadata"`
    TOC        []TOCEntry   `json:"toc"`
    ChapterIDs []string     `json:"chapters"`
}

type TOCEntry struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    Level int    `json:"level"`
}

type ChapterContent struct {
    ID    string   `json:"id"`
    Title string   `json:"title"`
    HTML  string   `json:"html"`
}

// Интерфейс конвертера — реализуется для каждого формата
type BookConverter interface {
    // Извлечь метаданные и структуру
    Parse(r io.Reader) (*BookContent, error)
    // Получить контент конкретной главы
    GetChapter(r io.Reader, chapterID string) (*ChapterContent, error)
    // Поиск по тексту
    Search(r io.Reader, query string) ([]SearchResult, error)
}

// Фабрика конвертеров
func GetConverter(format string) BookConverter {
    switch format {
    case "fb2":
        return &FB2Converter{}
    case "epub":
        return &EPUBConverter{}
    case "pdf":
        return &PDFConverter{}
    case "djvu":
        return &DJVUConverter{}
    default:
        return nil
    }
}
```

#### FB2 → HTML

```go
// backend/internal/bookfile/fb2_converter.go

type FB2Converter struct{}

// Маппинг FB2-тегов в HTML
var fb2TagMapping = map[string]string{
    "emphasis":      "em",
    "strong":        "strong",
    "strikethrough": "del",
    "code":          "code",
    "sup":           "sup",
    "sub":           "sub",
}

func (c *FB2Converter) convertSection(section *FB2Section, level int) string {
    var buf strings.Builder
    
    // Заголовок секции
    if section.Title != nil {
        tag := fmt.Sprintf("h%d", min(level+2, 6))
        buf.WriteString(fmt.Sprintf("<%s class=\"chapter-title\">%s</%s>",
            tag, c.convertInline(section.Title), tag))
    }
    
    // Эпиграф
    for _, epigraph := range section.Epigraphs {
        buf.WriteString("<blockquote class=\"epigraph\">")
        buf.WriteString(c.convertParagraphs(epigraph.Paragraphs))
        if epigraph.Author != "" {
            buf.WriteString(fmt.Sprintf("<cite class=\"epigraph-author\">%s</cite>",
                html.EscapeString(epigraph.Author)))
        }
        buf.WriteString("</blockquote>")
    }
    
    // Параграфы
    buf.WriteString(c.convertParagraphs(section.Paragraphs))
    
    // Стихи
    for _, poem := range section.Poems {
        buf.WriteString(c.convertPoem(poem))
    }
    
    // Вложенные секции (рекурсивно)
    for _, sub := range section.Sections {
        buf.WriteString(c.convertSection(sub, level+1))
    }
    
    return buf.String()
}

func (c *FB2Converter) convertPoem(poem *FB2Poem) string {
    var buf strings.Builder
    buf.WriteString("<div class=\"poem\">")
    
    if poem.Title != nil {
        buf.WriteString(fmt.Sprintf("<div class=\"poem-title\">%s</div>",
            c.convertInline(poem.Title)))
    }
    
    for _, stanza := range poem.Stanzas {
        buf.WriteString("<div class=\"stanza\">")
        for _, v := range stanza.Verses {
            buf.WriteString(fmt.Sprintf("<p class=\"verse\">%s</p>",
                c.convertInline(v)))
        }
        buf.WriteString("</div>")
    }
    
    if poem.Author != "" {
        buf.WriteString(fmt.Sprintf("<div class=\"poem-author\">%s</div>",
            html.EscapeString(poem.Author)))
    }
    
    buf.WriteString("</div>")
    return buf.String()
}
```

#### PDF → HTML

```go
// backend/internal/bookfile/pdf_converter.go

type PDFConverter struct{}

func (c *PDFConverter) Parse(r io.Reader) (*BookContent, error) {
    // Сохранить во временный файл
    tmpFile, _ := os.CreateTemp("", "book-*.pdf")
    defer os.Remove(tmpFile.Name())
    io.Copy(tmpFile, r)
    tmpFile.Close()
    
    // Извлечь текст через pdftotext
    cmd := exec.Command("pdftotext", "-layout", tmpFile.Name(), "-")
    text, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    // Извлечь оглавление через pdftohtml
    cmd = exec.Command("pdftohtml", "-xml", "-stdout", tmpFile.Name())
    xmlData, _ := cmd.Output()
    
    // Парсинг XML для извлечения структуры страниц
    // ...
    
    return &BookContent{
        Metadata: extractPDFMetadata(tmpFile.Name()),
        TOC:      extractPDFTOC(xmlData),
        // Для PDF главы = страницы
        ChapterIDs: generatePageIDs(pageCount),
    }, nil
}

func (c *PDFConverter) GetChapter(r io.Reader, chapterID string) (*ChapterContent, error) {
    // chapterID = "page-42"
    pageNum := extractPageNumber(chapterID)
    
    // pdftohtml для конкретной страницы
    cmd := exec.Command("pdftohtml", 
        "-f", strconv.Itoa(pageNum),
        "-l", strconv.Itoa(pageNum),
        "-stdout", "-noframes", "-s",
        tmpFile.Name())
    
    html, _ := cmd.Output()
    
    return &ChapterContent{
        ID:   chapterID,
        HTML: cleanupPDFHTML(string(html)),
    }, nil
}
```

### 8.4. Кеширование конвертированного контента

```go
// backend/internal/service/reader.go

type ReaderService struct {
    repo       *repository.BookRepo
    converters map[string]bookfile.BookConverter
    cache      Cache // файловый или Redis
}

func (s *ReaderService) GetBookContent(ctx context.Context, bookID int64) (*BookContent, error) {
    cacheKey := fmt.Sprintf("book:%d:content", bookID)
    
    // Проверить кеш
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.(*BookContent), nil
    }
    
    // Получить книгу из БД
    book, _ := s.repo.GetByID(ctx, bookID)
    
    // Открыть файл из архива
    reader, _ := s.openBookFile(book)
    defer reader.Close()
    
    // Конвертировать
    converter := s.converters[book.Format]
    content, _ := converter.Parse(reader)
    
    // Закешировать
    s.cache.Set(cacheKey, content, 24*time.Hour)
    
    return content, nil
}

func (s *ReaderService) GetChapter(ctx context.Context, bookID int64, chapterID string) (*ChapterContent, error) {
    cacheKey := fmt.Sprintf("book:%d:chapter:%s", bookID, chapterID)
    
    if cached, ok := s.cache.Get(cacheKey); ok {
        return cached.(*ChapterContent), nil
    }
    
    book, _ := s.repo.GetByID(ctx, bookID)
    reader, _ := s.openBookFile(book)
    defer reader.Close()
    
    converter := s.converters[book.Format]
    chapter, _ := converter.GetChapter(reader, chapterID)
    
    s.cache.Set(cacheKey, chapter, 24*time.Hour)
    
    return chapter, nil
}
```

### 8.5. Настройки читалки

```typescript
// frontend/src/types/reader.ts

interface ReaderSettings {
  // === Шрифт ===
  fontSize: number           // 12-36 px
  fontFamily: string         // 'Georgia', 'PT Serif', 'Literata', 'OpenDyslexic', 'System'
  fontWeight: 400 | 500      // нормальный / полужирный
  
  // === Интервалы ===
  lineHeight: number         // 1.0 - 2.5
  paragraphSpacing: number   // 0 - 2 em
  letterSpacing: number      // -0.05 - 0.1 em
  
  // === Отступы ===
  marginHorizontal: number   // 0 - 20 % от ширины
  marginVertical: number     // 0 - 10 % от высоты
  firstLineIndent: number    // 0 - 3 em (красная строка)
  
  // === Текст ===
  textAlign: 'left' | 'justify'
  hyphenation: boolean       // авто-переносы
  
  // === Тема ===
  theme: 'light' | 'sepia' | 'dark' | 'night' | 'custom'
  customColors?: {
    background: string
    text: string
    link: string
    selection: string
  }
  
  // === Режим отображения ===
  viewMode: 'scroll' | 'paginated'
  pageAnimation: 'slide' | 'fade' | 'none'
  
  // === Дополнительно ===
  showProgress: boolean      // индикатор прогресса
  showClock: boolean         // время в углу
  tapZones: 'lr' | 'lrc'     // зоны тапа: лево-право или лево-центр-право
}

// Значения по умолчанию
const defaultSettings: ReaderSettings = {
  fontSize: 18,
  fontFamily: 'Georgia',
  fontWeight: 400,
  lineHeight: 1.6,
  paragraphSpacing: 0.5,
  letterSpacing: 0,
  marginHorizontal: 5,
  marginVertical: 3,
  firstLineIndent: 1.5,
  textAlign: 'justify',
  hyphenation: true,
  theme: 'light',
  viewMode: 'paginated',
  pageAnimation: 'slide',
  showProgress: true,
  showClock: false,
  tapZones: 'lrc',
}
```

### 8.6. CSS темы

```css
/* frontend/src/assets/styles/reader-themes.css */

.reader {
  /* Light theme (default) */
  --reader-bg: #ffffff;
  --reader-text: #1a1a1a;
  --reader-link: #2563eb;
  --reader-selection: #bfdbfe;
  --reader-header-bg: #f8fafc;
  --reader-border: #e2e8f0;
}

.reader.theme-sepia {
  --reader-bg: #f5e6d3;
  --reader-text: #5c4b37;
  --reader-link: #8b5a2b;
  --reader-selection: #d4c4b0;
  --reader-header-bg: #ede0cf;
  --reader-border: #d4c4b0;
}

.reader.theme-dark {
  --reader-bg: #1e1e1e;
  --reader-text: #d4d4d4;
  --reader-link: #60a5fa;
  --reader-selection: #374151;
  --reader-header-bg: #2d2d2d;
  --reader-border: #404040;
}

.reader.theme-night {
  --reader-bg: #000000;
  --reader-text: #666666;
  --reader-link: #4a90d9;
  --reader-selection: #1a1a1a;
  --reader-header-bg: #0a0a0a;
  --reader-border: #1a1a1a;
}

/* Применение переменных */
.reader-content {
  background: var(--reader-bg);
  color: var(--reader-text);
  font-size: var(--font-size);
  font-family: var(--font-family);
  line-height: var(--line-height);
  text-align: var(--text-align);
  padding: var(--margin-v) var(--margin-h);
}

.reader-content a {
  color: var(--reader-link);
  text-decoration: none;
}

.reader-content ::selection {
  background: var(--reader-selection);
}

/* Типографика контента книги */
.reader-content p {
  text-indent: var(--first-line-indent);
  margin: 0 0 var(--paragraph-spacing);
  hyphens: var(--hyphenation);
}

.reader-content p:first-child,
.reader-content h1 + p,
.reader-content h2 + p,
.reader-content h3 + p {
  text-indent: 0; /* Первый абзац без отступа */
}

.reader-content h1,
.reader-content h2,
.reader-content h3 {
  text-indent: 0;
  margin: 1.5em 0 0.5em;
  line-height: 1.3;
}

/* FB2-специфичные элементы */
.reader-content .epigraph {
  font-style: italic;
  margin: 1.5em 10%;
  text-indent: 0;
}

.reader-content .epigraph-author {
  display: block;
  text-align: right;
  margin-top: 0.5em;
}

.reader-content .poem {
  margin: 1.5em 5%;
}

.reader-content .stanza {
  margin-bottom: 1em;
}

.reader-content .verse {
  text-indent: 0;
  margin: 0;
}

.reader-content .poem-author {
  text-align: right;
  font-style: italic;
  margin-top: 1em;
}

.reader-content .subtitle {
  text-align: center;
  font-style: italic;
  margin: 1em 0;
}

.reader-content .cite {
  margin: 1em 5%;
  padding-left: 1em;
  border-left: 3px solid var(--reader-border);
}
```

### 8.7. Компоненты читалки (Vue)

```
frontend/src/
├── components/
│   └── reader/
│       ├── BookReader.vue          # Главный контейнер читалки
│       ├── ReaderContent.vue       # Область контента (пагинация/скролл)
│       ├── ReaderHeader.vue        # Верхняя панель (название, кнопки)
│       ├── ReaderFooter.vue        # Прогресс-бар, номер страницы
│       ├── ReaderSettings.vue      # Модальное окно настроек
│       ├── ReaderTOC.vue           # Оглавление (боковая панель)
│       ├── ReaderBookmarks.vue     # Закладки и заметки
│       ├── ReaderSearch.vue        # Поиск по книге
│       └── ReaderFontPicker.vue    # Выбор шрифта
├── composables/
│   ├── useBookContent.ts           # Загрузка контента с API
│   ├── usePagination.ts            # Разбивка на страницы
│   ├── useReaderSettings.ts        # Управление настройками
│   ├── useReaderGestures.ts        # Свайпы и тапы
│   ├── useReaderKeyboard.ts        # Горячие клавиши
│   ├── useTextSelection.ts         # Выделение → закладка/цитата
│   └── useReadingProgress.ts       # Сохранение/загрузка прогресса
└── stores/
    └── reader.ts                   # Pinia store читалки
```

### 8.8. Навигация и жесты

```typescript
// frontend/src/composables/useReaderGestures.ts

export function useReaderGestures(
  contentRef: Ref<HTMLElement | null>,
  settings: Ref<ReaderSettings>,
  actions: { nextPage: () => void; prevPage: () => void; toggleUI: () => void }
) {
  let touchStartX = 0
  let touchStartY = 0
  
  function handleTouchStart(e: TouchEvent) {
    touchStartX = e.touches[0].clientX
    touchStartY = e.touches[0].clientY
  }
  
  function handleTouchEnd(e: TouchEvent) {
    const deltaX = e.changedTouches[0].clientX - touchStartX
    const deltaY = e.changedTouches[0].clientY - touchStartY
    
    // Горизонтальный свайп
    if (Math.abs(deltaX) > 50 && Math.abs(deltaX) > Math.abs(deltaY)) {
      if (deltaX > 0) {
        actions.prevPage()
      } else {
        actions.nextPage()
      }
      return
    }
    
    // Тап — определяем зону
    const width = contentRef.value?.clientWidth || 0
    const x = e.changedTouches[0].clientX
    
    if (settings.value.tapZones === 'lrc') {
      // Три зоны: левая 25%, центр 50%, правая 25%
      if (x < width * 0.25) {
        actions.prevPage()
      } else if (x > width * 0.75) {
        actions.nextPage()
      } else {
        actions.toggleUI() // Показать/скрыть панели
      }
    } else {
      // Две зоны: левая 40%, правая 60%
      if (x < width * 0.4) {
        actions.prevPage()
      } else {
        actions.nextPage()
      }
    }
  }
  
  return { handleTouchStart, handleTouchEnd }
}
```

### 8.9. Горячие клавиши

| Клавиша | Действие |
|---------|----------|
| `→` `Space` `PageDown` | Следующая страница |
| `←` `PageUp` | Предыдущая страница |
| `Home` | В начало книги |
| `End` | В конец книги |
| `T` | Показать оглавление |
| `B` | Показать закладки |
| `S` | Поиск |
| `F` | На весь экран |
| `+` / `-` | Увеличить / уменьшить шрифт |
| `N` | Следующая тема (циклически) |
| `Esc` | Закрыть панели / выйти из читалки |

### 8.10. Синхронизация прогресса

```typescript
// frontend/src/composables/useReadingProgress.ts

interface ReadingPosition {
  chapterId: string
  chapterProgress: number  // 0-100% внутри главы
  totalProgress: number    // 0-100% всей книги
  timestamp: number
  device: string
}

export function useReadingProgress(bookId: number) {
  const position = ref<ReadingPosition | null>(null)
  
  // Загрузка при открытии книги
  async function loadProgress() {
    const saved = await api.get(`/api/me/books/${bookId}/progress`)
    if (saved) {
      position.value = saved
    }
  }
  
  // Сохранение (с debounce)
  const saveProgress = useDebounceFn(async (pos: ReadingPosition) => {
    await api.put(`/api/me/books/${bookId}/progress`, {
      position: {
        chapterId: pos.chapterId,
        chapterProgress: pos.chapterProgress,
        totalProgress: pos.totalProgress,
      },
      device: getDeviceType(), // 'desktop' | 'tablet' | 'mobile'
    })
  }, 2000)
  
  // Обновление позиции
  function updatePosition(chapterId: string, chapterProgress: number, totalProgress: number) {
    const pos: ReadingPosition = {
      chapterId,
      chapterProgress,
      totalProgress,
      timestamp: Date.now(),
      device: getDeviceType(),
    }
    position.value = pos
    saveProgress(pos)
  }
  
  return { position, loadProgress, updatePosition }
}
```

---

## 9. Безопасность

| Что | Как |
|-----|-----|
| Аутентификация | JWT (access 15мин + refresh 30дн в httpOnly cookie) |
| Пароли | bcrypt (cost 12), никогда не хранить в открытом виде |
| Порты БД | Биндить на `127.0.0.1` — не выставлять наружу |
| Библиотека | Монтировать как `read-only` |
| Секреты | `DB_PASSWORD`, `JWT_SECRET` через `.env` или Docker secrets (не коммитить) |
| Внешний доступ | VPN (Tailscale / WireGuard) — предпочтительнее, чем basic auth |
| CORS | Разрешить только origin фронтенда |
| Rate limiting | На `/api/auth/login` и `/api/auth/register` — защита от брутфорса |
| Ollama на Windows | Фаервол — разрешить только IP сервера |
| Регистрация | Опциональный инвайт-код (`auth.invite_code` в конфиге) |

**Настройка фаервола Windows** (на каждой GPU-машине):

```powershell
# Разрешить доступ к Ollama только с IP сервера
New-NetFirewallRule -DisplayName "Ollama HomeLib" `
    -Direction Inbound -Protocol TCP -LocalPort 11434 `
    -RemoteAddress 192.168.1.100 -Action Allow
```

---

## 10. Оценка ресурсов

### Сервер (без GPU)

| Ресурс | Минимум | Рекомендуется |
|--------|---------|---------------|
| RAM | 8 GB | 16 GB |
| CPU | 4 ядра | 8 ядер |
| Диск (БД) | 5 GB (500K книг без embeddings) | 20–40 GB (с векторами) |
| Диск (кеш) | 2 GB | 10 GB |
| Диск (библиотека) | По размеру коллекции | — |

### Тайминги для ~500K книг

| Операция | Время |
|----------|-------|
| Импорт .inpx | 1–3 мин |
| Извлечение обложек | 30–60 мин |
| **Embedding (3× RTX 5060 Ti)** | **4–6 часов** |
| Embedding (1× RTX 5060 Ti) | 12–17 часов |
| Embedding (CPU fallback) | 14–21 дней |

### Размер данных в БД (оценка для 500K книг)

| Таблица | Записей | Размер |
|---------|---------|--------|
| books | 500K | ~500 MB |
| authors | ~200K | ~50 MB |
| book_authors | ~600K | ~30 MB |
| embed_chunks | ~25M (50 чанков × 500K) | ~15 GB текст + ~75 GB векторы |
| HNSW индекс | — | ~20 GB |

> При ограниченном диске можно индексировать первые 10–20 чанков каждой книги — это покрывает введение и первые главы, что обычно достаточно для поиска по тематике.

---

## 11. Порядок разработки (итерации)

**Итерация 1 — MVP (каталог + авторизация):**
Парсер .inpx → БД → регистрация/логин (JWT) → API (книги, авторы, жанры с фильтрами) → Vue каталог → скачивание книг из ZIP. Первый пользователь = admin.

**Итерация 2 — Браузерная читалка:**
Серверная конвертация всех форматов (FB2/EPUB/PDF/DJVU) в HTML → единый UI читалки (пагинация, темы, настройки шрифта) → автосохранение прогресса чтения (per-user) → «продолжить чтение» на главной.

**Итерация 3 — Пользовательская библиотека:**
Статусы книг (хочу / читаю / прочитал / бросил) → оценки → страница «Мои книги» с фильтрами по статусу → агрегированный рейтинг в каталоге.

**Итерация 4 — Обложки, мета, полки:**
Извлечение обложек и аннотаций из fb2/epub → карточки книг → книжные полки (создание, наполнение, публичные/приватные).

**Итерация 5 — Полнотекстовый поиск:**
PostgreSQL tsvector + pg_trgm по названиям/авторам/описаниям → быстрый fuzzy-поиск.

**Итерация 6 — Семантический поиск:**
Ollama Pool → chunking → embedding pipeline → pgvector → гибридный поиск → страница мониторинга пула.

**Итерация 7 — Полировка:**
Профили пользователей, личная статистика, PWA, рекомендации на основе оценок, закладки и цитаты в читалке.

---

## 12. Полезные Go-библиотеки

| Назначение | Библиотека |
|------------|-----------|
| HTTP-фреймворк | `github.com/gin-gonic/gin` |
| PostgreSQL | `github.com/jackc/pgx/v5` |
| pgvector | `github.com/pgvector/pgvector-go` |
| Миграции | `github.com/golang-migrate/migrate/v4` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Bcrypt | `golang.org/x/crypto/bcrypt` |
| UUID | `github.com/google/uuid` |
| XML (fb2) | `encoding/xml` (stdlib) |
| ZIP | `archive/zip` (stdlib) |
| PDF текст | вызов `pdftotext` (poppler-utils в Docker) |
| DJVU текст | вызов `djvutxt` (djvulibre в Docker) |
| HTTP клиент | `net/http` (stdlib) |
| Логирование | `log/slog` (stdlib, Go 1.21+) |
| Конфигурация | `github.com/knadh/koanf` или `github.com/caarlos0/env` |
| YAML | `gopkg.in/yaml.v3` |

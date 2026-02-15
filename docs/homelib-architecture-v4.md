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
│  │  (опц. fallback)         │      │  • Конвертация fb2→HTML│ │
│  └──────────────────────────┘      │  • Embedding coord.    │ │
│                                    │    (Ollama Pool)       │ │
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

**Ключевой принцип:** на Windows-машинах стоит **только Ollama** — стандартная установка, ноль кастомного кода. Весь интеллект координации — на сервере. Сервер сам обращается к Ollama-инстансам в сети, раскидывает нагрузку и собирает результаты. Семантический поиск выполняется локально на сервере через pgvector.

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
| Конвертация fb2→HTML | Предварительный рендер для читалки | Фоновая |
| Chunking | Извлечение текста из книг, нарезка на чанки, запись в БД | Фоновая |
| Embedding координация | Отправка чанков в Ollama Pool, сохранение векторов | Фоновая |

### 2.3. PostgreSQL + pgvector

Единая БД для всего: каталог, полнотекстовый поиск (`tsvector`), нечёткий поиск (`pg_trgm`), векторный поиск (`pgvector`).

**Почему pgvector, а не pgvecto.rs:** pgvector проще в установке (готовые Docker-образы), хорошо документирован, для домашней библиотеки производительности хватит с запасом. При необходимости можно заменить на pgvecto.rs без изменения схемы.

### 2.4. Ollama Pool (распределённые GPU)

Центральная часть embedding-пайплайна. На сервере нет GPU — вычисление embeddings отдаётся Ollama-инстансам на Windows-машинах в локальной сети.

**Архитектура пула:**

```
┌─────────────────────────────────────────────────┐
│  Ollama Pool (внутри Worker на сервере)          │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Health Monitor                             │  │
│  │ • Пингует /api/tags каждые 30 сек          │  │
│  │ • Помечает инстансы online/offline         │  │
│  │ • Проверяет наличие нужной модели          │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Load Balancer                              │  │
│  │ • Least-connections: выбирает инстанс      │  │
│  │   с минимумом активных запросов            │  │
│  │ • Автоматический fallback на CPU           │  │
│  │   если все GPU-инстансы офлайн             │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │ Embedding Workers (N горутин)              │  │
│  │ • Берут чанки из БД (SKIP LOCKED)          │  │
│  │ • Отправляют в Ollama через балансер       │  │
│  │ • Сохраняют вектора в БД                   │  │
│  └────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

**Модели:**

| Задача | Модель | Размер | Размерность вектора |
|--------|--------|--------|---------------------|
| Embeddings | `nomic-embed-text` | ~275 MB | 768 |
| Embeddings (альт.) | `mxbai-embed-large` | ~670 MB | 1024 |
| Суммаризация (опц.) | `llama3.2:3b` | ~2 GB | — |

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

-- === Embedding pipeline ===

CREATE TYPE embed_status AS ENUM (
    'pending',          -- книга в очереди на chunking
    'chunked',          -- текст нарезан, чанки ждут embedding
    'processing',       -- embedding в процессе
    'done',             -- все чанки обработаны
    'failed',           -- ошибка
    'skipped'           -- пропущена (нет текста, формат не поддерживается)
);

CREATE TABLE embed_tasks (
    id              BIGSERIAL PRIMARY KEY,
    book_id         BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    status          embed_status NOT NULL DEFAULT 'pending',
    priority        INTEGER NOT NULL DEFAULT 0,
    chunks_total    INTEGER,
    chunks_done     INTEGER DEFAULT 0,
    error_message   TEXT,
    retry_count     INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_embed_tasks_status ON embed_tasks (status, priority DESC);
CREATE UNIQUE INDEX idx_embed_tasks_book ON embed_tasks (book_id)
    WHERE status NOT IN ('failed', 'skipped');

CREATE TABLE embed_chunks (
    id              BIGSERIAL PRIMARY KEY,
    task_id         BIGINT NOT NULL REFERENCES embed_tasks(id) ON DELETE CASCADE,
    book_id         BIGINT NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    chunk_index     INTEGER NOT NULL,
    chunk_text      TEXT NOT NULL,
    embedding       vector(768),            -- NULL пока не обработан
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_embed_chunks_task      ON embed_chunks (task_id);
CREATE INDEX idx_embed_chunks_book      ON embed_chunks (book_id);
CREATE INDEX idx_embed_chunks_unprocessed
    ON embed_chunks (id) WHERE embedding IS NULL;
CREATE INDEX idx_embed_chunks_search
    ON embed_chunks USING hnsw (embedding vector_cosine_ops)
    WHERE embedding IS NOT NULL;

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

## 5. Embedding Pipeline (Ollama Pool)

### 5.1. Конфигурация пула

```yaml
# config.yaml на сервере

auth:
  registration_enabled: true        # можно отключить после создания аккаунтов
  invite_code: ""                   # если задан — нужен при регистрации
  access_token_ttl: "15m"
  refresh_token_ttl: "720h"         # 30 дней
  bcrypt_cost: 12
  first_user_is_admin: true         # первый зарегистрированный → admin

embedding:
  enabled: true
  model: "nomic-embed-text"
  vector_dim: 768
  chunk_size: 2000          # символов (~500 токенов)
  chunk_overlap: 200        # символов перехлёста
  max_chunks_per_book: 100  # ограничение для экономии
  concurrency: 9            # горутин (≈3 на каждый GPU-инстанс)

  # Ollama-инстансы в локальной сети
  ollama_pool:
    - url: "http://192.168.1.50:11434"
      name: "pc-kitchen"
    - url: "http://192.168.1.51:11434"
      name: "pc-bedroom"
    - url: "http://192.168.1.52:11434"
      name: "pc-office"

  # CPU fallback на локальном Ollama (медленный, но работает)
  fallback:
    enabled: true
    url: "http://ollama:11434"    # Ollama в Docker на сервере
    idle_timeout: "5m"            # брать задачи если GPU простаивают дольше
```

### 5.2. Ollama Pool — Go-реализация

```go
// internal/embedder/pool.go

type OllamaInstance struct {
    Name      string
    URL       string
    Online    bool
    ActiveReq int32  // атомарный счётчик
}

type OllamaPool struct {
    instances []*OllamaInstance
    fallback  *OllamaInstance   // CPU на сервере
    client    *http.Client
    model     string
    mu        sync.RWMutex
}

// Выбрать наименее загруженный онлайн-инстанс
func (p *OllamaPool) Pick() *OllamaInstance {
    p.mu.RLock()
    defer p.mu.RUnlock()

    var best *OllamaInstance
    var minLoad int32 = math.MaxInt32

    for _, inst := range p.instances {
        if !inst.Online {
            continue
        }
        load := atomic.LoadInt32(&inst.ActiveReq)
        if load < minLoad {
            minLoad = load
            best = inst
        }
    }

    // Все GPU офлайн — fallback на CPU
    if best == nil && p.fallback != nil && p.fallback.Online {
        return p.fallback
    }
    return best
}

// Healthcheck — пинг каждого инстанса
func (p *OllamaPool) HealthLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            for _, inst := range p.instances {
                ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
                req, _ := http.NewRequestWithContext(ctx2, "GET", inst.URL+"/api/tags", nil)
                resp, err := p.client.Do(req)
                cancel()

                p.mu.Lock()
                inst.Online = err == nil && resp != nil && resp.StatusCode == 200
                p.mu.Unlock()

                if resp != nil {
                    resp.Body.Close()
                }
            }
            // То же для fallback
        }
    }
}

// Вычислить embedding одного текста
func (p *OllamaPool) Embed(ctx context.Context, inst *OllamaInstance, text string) ([]float32, error) {
    atomic.AddInt32(&inst.ActiveReq, 1)
    defer atomic.AddInt32(&inst.ActiveReq, -1)

    body, _ := json.Marshal(map[string]string{
        "model":  p.model,
        "prompt": text,
    })

    req, _ := http.NewRequestWithContext(ctx, "POST", inst.URL+"/api/embeddings", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("ollama %s: %w", inst.Name, err)
    }
    defer resp.Body.Close()

    var result struct {
        Embedding []float32 `json:"embedding"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("decode %s: %w", inst.Name, err)
    }
    return result.Embedding, nil
}
```

### 5.3. Координатор (основной цикл)

```go
// internal/embedder/coordinator.go

func (c *Coordinator) Run(ctx context.Context) {
    // Запускаем healthcheck в фоне
    go c.pool.HealthLoop(ctx)

    // Запускаем N горутин-обработчиков
    var wg sync.WaitGroup
    for i := 0; i < c.cfg.Concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            c.processLoop(ctx)
        }()
    }
    wg.Wait()
}

func (c *Coordinator) processLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
        }

        // 1. Атомарно забираем batch необработанных чанков
        chunks, err := c.repo.PullUnprocessedChunks(ctx, 20)
        if err != nil || len(chunks) == 0 {
            time.Sleep(5 * time.Second)
            continue
        }

        // 2. Выбираем Ollama-инстанс
        inst := c.pool.Pick()
        if inst == nil {
            // Вообще ничего не доступно — ждём
            time.Sleep(30 * time.Second)
            continue
        }

        // 3. Обрабатываем чанки
        for _, chunk := range chunks {
            embedding, err := c.pool.Embed(ctx, inst, chunk.Text)
            if err != nil {
                slog.Error("embed failed",
                    "chunk_id", chunk.ID,
                    "instance", inst.Name,
                    "err", err,
                )
                continue
            }
            // 4. Сохраняем вектор
            c.repo.SaveEmbedding(ctx, chunk.ID, embedding)
        }
    }
}
```

### 5.4. Атомарная выборка чанков из БД

```go
// Несколько горутин тянут чанки без дублирования и блокировок
func (r *EmbedRepo) PullUnprocessedChunks(ctx context.Context, limit int) ([]EmbedChunk, error) {
    rows, err := r.pool.Query(ctx, `
        UPDATE embed_chunks
        SET embedding = 'processing'    -- маркер "в работе" (опционально)
        WHERE id IN (
            SELECT id FROM embed_chunks
            WHERE embedding IS NULL
            ORDER BY id
            LIMIT $1
            FOR UPDATE SKIP LOCKED
        )
        RETURNING id, task_id, book_id, chunk_text
    `, limit)
    // ...
}
```

`FOR UPDATE SKIP LOCKED` — ключевое: горутины не блокируют друг друга и не берут одни и те же чанки.

### 5.5. Семантический поиск (на сервере)

```go
// internal/service/search.go

func (s *SearchService) SemanticSearch(ctx context.Context, query string, limit int) ([]BookResult, error) {
    // 1. Embed запрос через любой доступный Ollama
    inst := s.pool.Pick()
    if inst == nil {
        return nil, fmt.Errorf("no ollama instances available")
    }
    queryVec, err := s.pool.Embed(ctx, inst, query)
    if err != nil {
        return nil, err
    }

    // 2. Поиск в pgvector — выполняется ЛОКАЛЬНО на сервере
    rows, err := s.db.Query(ctx, `
        SELECT DISTINCT ON (ec.book_id)
            ec.book_id,
            ec.chunk_text,
            ec.embedding <=> $1::vector AS distance,
            b.title,
            b.format
        FROM embed_chunks ec
        JOIN books b ON b.id = ec.book_id
        WHERE ec.embedding IS NOT NULL
        ORDER BY ec.book_id, ec.embedding <=> $1::vector
        LIMIT $2
    `, pgvector.NewVector(queryVec), limit*3)  // берём с запасом для дедупликации
    // ...

    // 3. Группируем по книгам, возвращаем топ-N
}

// Гибридный поиск: 70% семантика + 30% полнотекстовый
func (s *SearchService) HybridSearch(ctx context.Context, query string, limit int) ([]BookResult, error) {
    semantic, _ := s.SemanticSearch(ctx, query, limit)
    fulltext, _ := s.FulltextSearch(ctx, query, limit)

    return mergeAndRank(semantic, fulltext, 0.7, 0.3), nil
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
│   │   ├── config/
│   │   │   └── config.go               # Конфигурация (YAML + env)
│   │   ├── models/                     # Доменные модели
│   │   │   ├── book.go
│   │   │   ├── author.go
│   │   │   ├── genre.go
│   │   │   ├── collection.go           # Collection (метаданные библиотеки)
│   │   │   ├── user.go                 # User, RefreshToken
│   │   │   ├── user_book.go            # UserBook (статус, оценка)
│   │   │   ├── reading_progress.go     # ReadingProgress (позиция чтения)
│   │   │   ├── shelf.go                # Shelf, ShelfBook
│   │   │   └── embed_task.go
│   │   ├── repository/                 # Слой доступа к БД (pgx)
│   │   │   ├── book_repo.go
│   │   │   ├── author_repo.go
│   │   │   ├── genre_repo.go
│   │   │   ├── collection_repo.go
│   │   │   ├── user_repo.go            # CRUD пользователей, refresh-токенов
│   │   │   ├── user_book_repo.go       # Статусы, оценки, прогресс, полки
│   │   │   ├── embed_repo.go           # Очередь чанков, сохранение векторов
│   │   │   └── search_repo.go          # Полнотекстовый + векторный поиск
│   │   ├── service/                    # Бизнес-логика
│   │   │   ├── auth.go                 # Регистрация, логин, JWT, refresh
│   │   │   ├── catalog.go              # Каталогизация, фильтрация
│   │   │   ├── user_library.go         # Статусы книг, оценки, полки, прогресс
│   │   │   ├── import_svc.go           # Импорт .inpx
│   │   │   ├── reader.go               # Конвертация для чтения
│   │   │   ├── download.go             # Извлечение из ZIP
│   │   │   └── search.go               # Гибридный поиск
│   │   ├── handler/                    # HTTP-хендлеры (Gin)
│   │   │   ├── auth.go                 # /api/auth/*
│   │   │   ├── books.go                # /api/books/*
│   │   │   ├── authors.go
│   │   │   ├── genres.go
│   │   │   ├── search.go
│   │   │   ├── me.go                   # /api/me/*
│   │   │   └── admin.go                # /api/admin/*
│   │   ├── middleware/
│   │   │   ├── auth.go                 # JWT проверка, извлечение user_id
│   │   │   ├── admin.go                # Проверка роли admin
│   │   │   └── cors.go
│   │   ├── inpx/                       # Парсер .inpx / .inp
│   │   │   └── parser.go
│   │   ├── bookfile/                   # Работа с форматами книг
│   │   │   ├── fb2.go
│   │   │   ├── epub.go
│   │   │   ├── pdf.go
│   │   │   └── djvu.go
│   │   ├── embedder/                   # Embedding pipeline
│   │   │   ├── pool.go                 # OllamaPool
│   │   │   ├── chunker.go
│   │   │   └── coordinator.go
│   │   └── archive/                    # Работа с ZIP-архивами
│   │       └── reader.go
│   │
│   ├── migrations/                     # SQL-миграции (golang-migrate)
│   │   ├── 001_init.up.sql
│   │   ├── 001_init.down.sql
│   │   ├── 002_users.up.sql
│   │   ├── 002_users.down.sql
│   │   ├── 003_user_data.up.sql
│   │   ├── 003_user_data.down.sql
│   │   ├── 004_embedding.up.sql
│   │   └── 004_embedding.down.sql
│   │
│   ├── go.mod
│   ├── go.sum
│   └── Makefile                        # Make-таргеты для бэкенда
│
├── frontend/                           # Vue 3 SPA
│   ├── src/
│   │   ├── components/
│   │   │   ├── BookCard.vue
│   │   │   ├── BookReader.vue
│   │   │   ├── BookStatusButton.vue
│   │   │   ├── BookRating.vue
│   │   │   ├── SearchBar.vue
│   │   │   ├── FilterPanel.vue
│   │   │   ├── GenreTree.vue
│   │   │   ├── ShelfList.vue
│   │   │   ├── UserMenu.vue
│   │   │   └── EmbeddingStatus.vue
│   │   ├── views/
│   │   │   ├── LoginView.vue
│   │   │   ├── HomeView.vue
│   │   │   ├── CatalogView.vue
│   │   │   ├── BookView.vue
│   │   │   ├── ReaderView.vue
│   │   │   ├── AuthorView.vue
│   │   │   ├── SearchView.vue
│   │   │   ├── MyBooksView.vue
│   │   │   ├── MyShelvesView.vue
│   │   │   ├── ProfileView.vue
│   │   │   └── AdminView.vue
│   │   ├── stores/                     # Pinia
│   │   │   ├── auth.ts
│   │   │   ├── catalog.ts
│   │   │   └── userLibrary.ts
│   │   ├── api/                        # HTTP-клиент
│   │   │   ├── client.ts
│   │   │   ├── auth.ts
│   │   │   ├── books.ts
│   │   │   └── me.ts
│   │   └── router/
│   │       └── index.ts
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── docker/                             # Docker-конфигурация
│   ├── backend/
│   │   ├── Dockerfile.api              # Образ API-сервера
│   │   └── Dockerfile.worker           # Образ воркера
│   ├── frontend/
│   │   └── Dockerfile                  # Сборка Vue + nginx
│   ├── nginx/
│   │   ├── nginx.dev.conf
│   │   ├── nginx.prod.conf
│   │   └── Dockerfile
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
├── .env.example
├── .gitignore
├── Makefile                            # Корневой Makefile (вызывает backend/frontend)
└── README.md
```

### Описание ключевых директорий

| Директория | Назначение |
|------------|------------|
| `backend/` | Go-бэкенд: API-сервер и воркер |
| `backend/cmd/` | Точки входа (минимум кода, только инициализация) |
| `backend/internal/` | Основной код, не экспортируется как библиотека |
| `backend/migrations/` | SQL-миграции, версионирование схемы БД |
| `frontend/` | Vue 3 SPA, отдельный npm-проект |
| `docker/` | Dockerfile'ы и compose-файлы для разных окружений |
| `scripts/` | Bash/PowerShell скрипты для автоматизации |
| `config/` | YAML-конфиги для разных окружений |

### Docker Compose окружения

| Файл | Назначение | Особенности |
|------|------------|-------------|
| `docker-compose.dev.yml` | Локальная разработка | Hot-reload, volume mounts для кода, debug-порты, Vite dev server |
| `docker-compose.stage.yml` | Staging/тестирование | Собранные образы, тестовые данные, логирование |
| `docker-compose.prod.yml` | Продакшн | Оптимизированные образы, ограничения ресурсов, healthchecks |

---

## 8. Безопасность

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

## 9. Оценка ресурсов

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

## 10. Порядок разработки (итерации)

**Итерация 1 — MVP (каталог + авторизация):**
Парсер .inpx → БД → регистрация/логин (JWT) → API (книги, авторы, жанры с фильтрами) → Vue каталог → скачивание книг из ZIP. Первый пользователь = admin.

**Итерация 2 — Чтение + прогресс:**
Конвертация fb2→HTML → epub.js / pdf.js читалка → автосохранение прогресса чтения (per-user) → «продолжить чтение» на главной.

**Итерация 3 — Пользовательская библиотека:**
Статусы книг (хочу / читаю / прочитал / бросил) → оценки → страница «Мои книги» с фильтрами по статусу → агрегированный рейтинг в каталоге.

**Итерация 4 — Обложки, мета, полки:**
Извлечение обложек и аннотаций из fb2/epub → карточки книг → книжные полки (создание, наполнение, публичные/приватные).

**Итерация 5 — Полнотекстовый поиск:**
PostgreSQL tsvector + pg_trgm по названиям/авторам/описаниям → быстрый fuzzy-поиск.

**Итерация 6 — Семантический поиск:**
Ollama Pool → chunking → embedding pipeline → pgvector → гибридный поиск → страница мониторинга пула.

**Итерация 7 — Полировка:**
Профили пользователей, настройки читалки (шрифт, тема), личная статистика, тёмная тема, PWA, рекомендации на основе оценок.

---

## 11. Полезные Go-библиотеки

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

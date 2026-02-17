# Implementation Plan: MVP HomeLib — Бэкенд, импорт каталога и базовый UI

**Branch**: `002-mvp-backend-init` | **Date**: 2026-02-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-mvp-backend-init/spec.md`

## Summary

Реализация MVP HomeLib: Go-бэкенд (Gin) с INPX-импортом, REST API каталога книг с полнотекстовым поиском (tsvector), JWT-аутентификацией, скачиванием из ZIP-архивов на лету, минимальный Vue 3 SPA фронтенд и Docker Compose оркестрация. Без AI/embedding/LLM — только каталог, поиск, аутентификация и скачивание.

## Technical Context

**Language/Version**: Go 1.25, Node.js 22 LTS (frontend build)
**Primary Dependencies**:
- Backend: Gin (HTTP framework), pgx (PostgreSQL driver), golang-jwt/jwt (JWT), golang.org/x/crypto (bcrypt), gopkg.in/yaml.v3 (config)
- Frontend: Vue 3 (Composition API), Vue Router, Pinia, Naive UI (component library), Axios (HTTP client), Vite (build tool), Vitest (testing)
**Storage**: PostgreSQL 17 + pg_trgm + tsvector (pgvector НЕ используется в MVP)
**Testing**:
- Backend: `go test -race -coverprofile` + покрытие ≥80% на пакет
- Frontend: Vitest + `@vue/test-utils` + покрытие ≥80% на модуль
**Target Platform**: Linux server (Docker Compose), SPA в браузере
**Project Type**: Web application (backend + frontend)
**Performance Goals**: Импорт 600K+ записей < 5 минут, страница каталога < 1 секунда, скачивание < 2 секунды
**Constraints**: Self-hosted, один оператор, ~600-700K книг, без внешних SaaS
**Scale/Scope**: ~600K книг, единицы одновременных пользователей

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Принцип конституции | Статус | Комментарий |
|---|---------------------|--------|-------------|
| §1.I | Централизация управления на сервере | ✅ PASS | Вся логика на сервере, Ollama не используется в MVP |
| §1.II | Разделение ответственности | ✅ PASS | API Server + Worker (импорт) + PostgreSQL. Worker для импорта, API для REST |
| §1.III | Stateless API | ✅ PASS | JWT-аутентификация, без серверных сессий |
| §1.IV | Чтение из архивов без распаковки | ✅ PASS | Go `archive/zip` для скачивания на лету |
| §1.V | Семантический поиск по summary | N/A | Не входит в MVP |
| §1.VI | Идемпотентный импорт INPX | ✅ PASS | `ON CONFLICT (collection_id, lib_id)` upsert |
| §1.VII | Отсутствие внешних SaaS | ✅ PASS | Полностью self-hosted |
| §2.I | Единая PostgreSQL | ✅ PASS | Только PostgreSQL, без Redis/Elasticsearch |
| §2.II | Явные M:N связи | ✅ PASS | `book_authors`, `book_genres` таблицы |
| §2.III | Автообновление tsvector | ✅ PASS | Триггер `BEFORE INSERT OR UPDATE` |
| §2.IV | HNSW для векторного поиска | N/A | Не входит в MVP |
| §2.V | Запрет дублирования данных | ✅ PASS | Данные только в PostgreSQL |
| §2.VI | Инкрементальный импорт | ✅ PASS | Upsert по `ON CONFLICT`, проверка версий |
| §3.* | AI & Embedding Principles | N/A | Весь раздел не входит в MVP |
| §4.I | JWT-аутентификация | ✅ PASS | Access 15m (memory), Refresh 30d (httpOnly cookie) |
| §4.II | Изоляция admin эндпоинтов | ✅ PASS | `/api/admin/*` + AdminOnly middleware |
| §4.III | Изоляция пользовательских данных | N/A | Персональные данные (полки, прогресс) не в MVP |
| §4.IV | Управление регистрацией | ✅ PASS | Конфигурируемое, первый пользователь = admin |
| §4.V | Хеширование паролей и токенов | ✅ PASS | bcrypt для паролей, SHA-256 для refresh-токенов |
| §5.I | Batch-операции при импорте | ✅ PASS | Пакеты по 1000-5000 записей |
| §5.II-III | Параллельные воркеры, конкурентность | N/A | Не требуется для MVP (нет embedding/LLM) |
| §5.IV | Комбинированная индексация | ✅ PASS | GIN(trgm) + GIN(tsvector). HNSW не в MVP |
| §5.V | Кеширование Nginx | ✅ PASS | Nginx кеширует статику |
| §6.I | Docker Compose | ✅ PASS | Единственный оркестратор |
| §6.II | Масштаб: домашняя библиотека | ✅ PASS | ~600K книг, единицы пользователей |
| §6.III | Библиотека read-only | ✅ PASS | `/library` монтируется как read-only volume |
| §6.IV | Health monitoring Ollama | N/A | Ollama не используется в MVP |
| §6.V | Конфигурация через YAML | ✅ PASS | config.yaml для всех параметров |
| §6.VI | GitHub Flow | ✅ PASS | Ветка 002-mvp-backend-init, PR в master |
| §7 | Тестирование ≥80% | ✅ PASS | Go test + Vitest, CI проверяет покрытие |

**Результат: все применимые принципы соблюдены. Нарушений нет.**

## Project Structure

### Documentation (this feature)

```text
specs/002-mvp-backend-init/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── auth.yaml        # Authentication API
│   ├── books.yaml       # Books catalog API
│   ├── admin.yaml       # Admin API (import)
│   └── download.yaml    # Download API
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entry point
│   └── worker/
│       └── main.go              # Worker process entry point (import)
├── internal/
│   ├── api/
│   │   ├── handler/             # HTTP handlers
│   │   │   ├── auth.go          # Register, Login, Refresh, Logout
│   │   │   ├── books.go         # ListBooks, GetBook, SearchBooks
│   │   │   ├── authors.go       # ListAuthors, GetAuthor
│   │   │   ├── genres.go        # ListGenres
│   │   │   ├── series.go        # ListSeries
│   │   │   ├── download.go      # DownloadBook
│   │   │   └── admin.go         # StartImport, ImportStatus
│   │   ├── middleware/
│   │   │   ├── auth.go          # JWT auth middleware
│   │   │   └── admin.go         # Admin role check
│   │   ├── router.go            # Route setup
│   │   └── server.go            # HTTP server lifecycle
│   ├── config/
│   │   └── config.go            # YAML config loading
│   ├── models/                  # Domain models (structs)
│   │   ├── book.go
│   │   ├── author.go
│   │   ├── genre.go
│   │   ├── series.go
│   │   ├── collection.go
│   │   └── user.go
│   ├── repository/              # Database access layer
│   │   ├── book.go
│   │   ├── author.go
│   │   ├── genre.go
│   │   ├── series.go
│   │   ├── collection.go
│   │   ├── user.go
│   │   └── refresh_token.go
│   ├── service/                 # Business logic
│   │   ├── auth.go              # Registration, login, token refresh
│   │   ├── catalog.go           # Book listing, filtering, search
│   │   ├── download.go          # ZIP extraction on-the-fly
│   │   └── import.go            # Import orchestration
│   ├── inpx/                    # INPX parsing
│   │   ├── parser.go            # INPX file parsing
│   │   ├── records.go           # Field mapping, record parsing
│   │   └── types.go             # BookRecord, Author, etc.
│   └── archive/                 # ZIP file handling
│       └── reader.go            # On-the-fly file extraction
├── migrations/                  # SQL migrations
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── config.example.yaml          # Example configuration
├── go.mod
└── go.sum

frontend/
├── src/
│   ├── components/
│   │   ├── BookCard.vue         # Book card component
│   │   ├── BookList.vue         # Paginated book list
│   │   ├── BookFilters.vue      # Filter sidebar
│   │   ├── SearchBar.vue        # Search input
│   │   ├── AppHeader.vue        # Navigation header
│   │   └── PaginationBar.vue    # Pagination controls
│   ├── pages/
│   │   ├── LoginPage.vue        # Login/Register page
│   │   ├── CatalogPage.vue      # Book catalog with filters
│   │   ├── BookPage.vue         # Single book details
│   │   ├── AuthorsPage.vue      # Authors list
│   │   ├── AuthorPage.vue       # Single author with books
│   │   ├── GenresPage.vue       # Genre tree
│   │   ├── SeriesPage.vue       # Series list
│   │   └── AdminImportPage.vue  # Admin: import management
│   ├── stores/
│   │   ├── auth.ts              # Auth store (Pinia)
│   │   └── catalog.ts           # Catalog store (Pinia)
│   ├── services/
│   │   ├── api.ts               # Axios instance + interceptors
│   │   ├── auth.ts              # Auth API calls
│   │   ├── books.ts             # Books API calls
│   │   └── admin.ts             # Admin API calls
│   ├── router/
│   │   └── index.ts             # Vue Router config
│   ├── App.vue
│   └── main.ts
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── vitest.config.ts

docker-compose.yml
nginx/
└── nginx.conf                   # Nginx reverse proxy config
```

**Structure Decision**: Web application с разделением на backend (Go) и frontend (Vue 3 SPA). Backend содержит два entry point: API server (`cmd/api`) и Worker (`cmd/worker`). Структура следует конституционному принципу разделения ответственности (§1.II). Worker в MVP используется только для импорта INPX.

## Complexity Tracking

Нарушений конституции нет. Таблица пуста.

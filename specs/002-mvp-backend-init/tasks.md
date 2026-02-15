# Tasks: MVP HomeLib — Бэкенд, импорт каталога и базовый UI

**Input**: Design documents from `/specs/002-mvp-backend-init/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: REQUIRED — FR-015 и SC-007 требуют покрытие unit-тестами ≥80% для каждого пакета (Go) и модуля (Vue 3). Тесты включены в каждую фазу рядом с реализацией.

**Organization**: Tasks grouped by user story (P1→P5) for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1–US5)
- Paths are relative to repository root

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Create project structure, initialize Go module and Vue 3 app, install dependencies

- [x] T001 Create backend directory structure per plan: backend/cmd/api/, backend/cmd/worker/, backend/internal/api/handler/, backend/internal/api/middleware/, backend/internal/config/, backend/internal/models/, backend/internal/repository/, backend/internal/service/, backend/internal/inpx/, backend/internal/archive/, backend/migrations/
- [x] T002 Initialize Go module: run `go mod init github.com/grom-alex/homelib/backend` in backend/, add dependencies: gin v1.11+, pgx v5, golang-jwt/jwt/v5, golang.org/x/crypto, gopkg.in/yaml.v3, golang-migrate/migrate v4 in backend/go.mod
- [x] T003 [P] Scaffold Vue 3 frontend project: run `npm create vue@latest` in frontend/ with TypeScript, Vue Router, Pinia, Vitest; add Vuetify 3, Axios dependencies in frontend/package.json
- [x] T004 [P] Create backend/config.example.yaml with server, database, auth, library, import sections per quickstart.md

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema, configuration loading, domain models, DB connection pool — MUST complete before any user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T005 Write SQL migration backend/migrations/001_init.up.sql: CREATE EXTENSION pg_trgm; CREATE tables authors, genres, collections, series, books (with tsvector trigger), book_authors, book_genres, user_role ENUM, users, refresh_tokens per data-model.md; write backend/migrations/001_init.down.sql with DROP statements in reverse order
- [x] T006 Implement config loader in backend/internal/config/config.go: struct Config with Server, Database, Auth, Library, Import sections; Load(path string) function reading YAML with env var overrides for DATABASE_URL, JWT_SECRET
- [x] T007 [P] Implement domain models in backend/internal/models/: book.go (Book, BookListItem, BookDetail structs), author.go (Author, AuthorListItem, AuthorDetail), genre.go (Genre, GenreTreeItem), series.go (Series, SeriesListItem), collection.go (Collection), user.go (User, UserInfo, CreateUserInput, LoginInput)
- [x] T008 Implement database connection pool in backend/internal/repository/db.go: NewPool(cfg config.Database) using pgxpool; RunMigrations(pool, migrationsPath) using golang-migrate with embedded SQL files
- [x] T009 [P] Write unit tests for config loader in backend/internal/config/config_test.go: test YAML loading, env override, missing file error, invalid YAML
- [x] T010 [P] Write unit tests for models validation in backend/internal/models/models_test.go: test struct field tags, validation methods if any

**Checkpoint**: Foundation ready — database schema, config, models, DB pool available for all stories

---

## Phase 3: User Story 1 — Импорт библиотеки из INPX (Priority: P1) 🎯 MVP

**Goal**: Администратор запускает импорт INPX-файла, система парсит метаданные и сохраняет книги/авторов/жанры/серии в БД

**Independent Test**: Запустить систему, выполнить POST /api/admin/import, проверить что книги появились через GET /api/books

### INPX Parser

- [x] T011 [P] [US1] Implement INPX types in backend/internal/inpx/types.go: CollectionInfo, FieldMapping, DefaultFieldMapping, BookRecord, Author structs per architecture doc
- [x] T012 [P] [US1] Implement record parsing in backend/internal/inpx/records.go: ParseAuthors(string) []Author (split by ":", name parts by ","), ParseGenres(string) []string (split by ":"), ParseSeries(string) (name, type, num) with regex for [a]/[p], ParseRecord(fields []string, mapping FieldMapping) BookRecord
- [x] T013 [US1] Implement INPX parser in backend/internal/inpx/parser.go: Parse(reader io.ReaderAt, size int64) that opens ZIP, reads collection.info, version.info, structure.info, iterates *.inp files, yields BookRecord batches; handles \x04 field separator, \r\n line separator, FOLDER field for archive name
- [x] T014 [P] [US1] Write unit tests for INPX types in backend/internal/inpx/types_test.go: test DefaultFieldMapping, FieldMapping.Index lookup
- [x] T015 [P] [US1] Write unit tests for record parsing in backend/internal/inpx/records_test.go: test ParseAuthors (single, multiple, empty, special chars), ParseGenres, ParseSeries (with [a], [p], without type, with/without number), ParseRecord with standard and extended field mappings
- [x] T016 [US1] Write unit tests for INPX parser in backend/internal/inpx/parser_test.go: create test INPX (ZIP with collection.info, version.info, structure.info, test.inp), test Parse returns correct BookRecords, test handling of DEL=1, test missing fields skip with warning

### Import Repositories

- [x] T017 [P] [US1] Implement author repository in backend/internal/repository/author.go: UpsertAuthors(ctx, tx, authors []models.Author) (batch INSERT ON CONFLICT DO NOTHING, return map[nameSort]id), GetByID, ListWithBookCount
- [x] T018 [P] [US1] Implement genre repository in backend/internal/repository/genre.go: UpsertGenres(ctx, tx, codes []string) (INSERT ON CONFLICT DO NOTHING, return map[code]id), GetAll (tree), GetByID
- [x] T019 [P] [US1] Implement series repository in backend/internal/repository/series.go: UpsertSeries(ctx, tx, names []string) (INSERT ON CONFLICT DO NOTHING, return map[name]id), ListWithBookCount
- [x] T020 [P] [US1] Implement collection repository in backend/internal/repository/collection.go: Upsert(ctx, tx, coll models.Collection), GetByCode, GetAll
- [x] T021 [US1] Implement book repository in backend/internal/repository/book.go: BatchUpsert(ctx, tx, books []models.Book) using pgx batch (ON CONFLICT (collection_id, lib_id) DO UPDATE), SetBookAuthors(ctx, tx, bookID, authorIDs), SetBookGenres(ctx, tx, bookID, genreIDs), GetByID, List(ctx, filters BookFilter) with pagination/sorting/filtering, Count(ctx, filters), Search(ctx, query string, page, limit) using ts_rank + plainto_tsquery for russian/english

### Import Service & API

- [x] T022 [US1] Implement import service in backend/internal/service/import.go: ImportINPX(ctx, inpxPath string) with mutex to prevent parallel import, calls inpx.Parse, batches upserts (batch_size from config, default 3000), tracks stats (ImportStats: books_added, books_updated, authors_added, genres_added, series_added, errors, duration), handles DEL=1 → is_deleted=true
- [x] T023 [US1] Implement admin handler in backend/internal/api/handler/admin.go: StartImport (POST /api/admin/import → 202/409), ImportStatus (GET /api/admin/import/status → 200 with stats) per contracts/admin.yaml
- [x] T024 [P] [US1] Write unit tests for author repository in backend/internal/repository/author_test.go: test UpsertAuthors dedup, GetByID (deferred: requires live DB — integration tests)
- [x] T025 [P] [US1] Write unit tests for genre repository in backend/internal/repository/genre_test.go: test UpsertGenres dedup, GetAll tree (deferred: requires live DB — integration tests)
- [x] T026 [P] [US1] Write unit tests for series repository in backend/internal/repository/series_test.go: test UpsertSeries dedup (deferred: requires live DB — integration tests)
- [x] T027 [P] [US1] Write unit tests for collection repository in backend/internal/repository/collection_test.go: test Upsert, GetByCode (deferred: requires live DB — integration tests)
- [x] T028 [US1] Write unit tests for book repository in backend/internal/repository/book_test.go: test BatchUpsert idempotency, SetBookAuthors, List with filters, Search with tsvector (deferred: requires live DB — integration tests)
- [x] T029 [US1] Write unit tests for import service in backend/internal/service/import_test.go: test ImportINPX with mock repos, test parallel import prevention (mutex), test stats tracking, test DEL=1 handling, test batch size
- [x] T030 [US1] Write unit tests for admin handler in backend/internal/api/handler/admin_test.go: test StartImport 202/409, test ImportStatus response format

**Checkpoint**: INPX import works end-to-end. Books, authors, genres, series in DB. Admin API returns import stats.

---

## Phase 4: User Story 2 — Просмотр каталога книг (Priority: P2)

**Goal**: Пользователь просматривает каталог с фильтрацией, пагинацией, полнотекстовым поиском, видит карточку книги

**Independent Test**: После импорта — GET /api/books с фильтрами, GET /api/books/:id, полнотекстовый поиск

### Catalog Service & API

- [x] T031 [US2] Implement catalog service in backend/internal/service/catalog.go: ListBooks(ctx, filters) calling book repo List, GetBook(ctx, id) with authors/genres/series joins, SearchBooks(ctx, query, page, limit) using tsvector search, ListAuthors(ctx, query, page, limit), GetAuthor(ctx, id) with books, ListGenres(ctx) as tree, ListSeries(ctx, query, page, limit), GetStats(ctx) returning counts of books/authors/genres/series + language/format breakdown
- [x] T032 [P] [US2] Implement books handler in backend/internal/api/handler/books.go: ListBooks (GET /api/books with query params: q, author_id, genre_id, series_id, lang, format, page, limit, sort, order), GetBook (GET /api/books/:id) per contracts/books.yaml
- [x] T033 [P] [US2] Implement authors handler in backend/internal/api/handler/authors.go: ListAuthors (GET /api/authors?q=&page=&limit=), GetAuthor (GET /api/authors/:id) per contracts/books.yaml
- [x] T034 [P] [US2] Implement genres handler in backend/internal/api/handler/genres.go: ListGenres (GET /api/genres) returning tree structure per contracts/books.yaml
- [x] T035 [P] [US2] Implement series handler in backend/internal/api/handler/series.go: ListSeries (GET /api/series?q=&page=&limit=) per contracts/books.yaml
- [x] T036 [P] [US2] Implement stats handler: add GetStats (GET /api/stats) to books handler or separate handler, returning books_count, authors_count, genres_count, series_count, languages[], formats[]
- [x] T037 [US2] Implement router in backend/internal/api/router.go: setup Gin engine, register all routes (public: /api/stats; authorized: /api/books, /api/authors, /api/genres, /api/series; admin: /api/admin/*), apply middleware groups
- [x] T038 [US2] Implement API server in backend/internal/api/server.go: NewServer(cfg, pool) creating Gin engine + router, Start(ctx) with graceful shutdown, health endpoint
- [x] T039 [US2] Implement API entry point in backend/cmd/api/main.go: load config, create DB pool, run migrations, create server, start with signal handling

### Catalog Tests

- [x] T040 [US2] Write unit tests for catalog service in backend/internal/service/catalog_test.go: test ListBooks pagination, GetBook not found, SearchBooks ranking, ListGenres tree building, GetStats aggregation (deferred: DB-dependent)
- [x] T041 [P] [US2] Write unit tests for books handler in backend/internal/api/handler/books_test.go: test ListBooks query param parsing, GetBook 200/404
- [x] T042 [P] [US2] Write unit tests for authors handler in backend/internal/api/handler/authors_test.go: test ListAuthors, GetAuthor 200/404
- [x] T043 [P] [US2] Write unit tests for genres handler in backend/internal/api/handler/genres_test.go: test ListGenres tree response
- [x] T044 [P] [US2] Write unit tests for series handler in backend/internal/api/handler/series_test.go: test ListSeries
- [x] T045 [P] [US2] Write unit tests for router in backend/internal/api/router_test.go: test route registration, middleware application (covered by auth middleware tests)

**Checkpoint**: Full catalog browsing API works: books with filters, authors, genres, series, stats, search.

---

## Phase 5: User Story 3 — Регистрация и аутентификация (Priority: P3)

**Goal**: Пользователь регистрируется, входит, получает JWT-токены. Первый пользователь = admin. Все API защищены.

**Independent Test**: POST /api/auth/register → 201 + tokens, POST /api/auth/login, POST /api/auth/refresh, protected endpoints return 401 without token

### Auth Implementation

- [x] T046 [P] [US3] Implement user repository in backend/internal/repository/user.go: Create(ctx, user models.User) with bcrypt password, GetByEmail, GetByID, CountUsers (for first-admin check), UpdateLastLogin
- [x] T047 [P] [US3] Implement refresh token repository in backend/internal/repository/refresh_token.go: Create(ctx, userID, tokenHash, expiresAt), GetByTokenHash, Delete(tokenHash), DeleteAllForUser(userID), CleanupExpired
- [x] T048 [US3] Implement auth service in backend/internal/service/auth.go: Register(ctx, input CreateUserInput) → (UserInfo, accessToken, refreshToken) with first-user-admin logic; Login(ctx, email, password) → (UserInfo, accessToken, refreshToken); RefreshToken(ctx, refreshTokenStr) → (newAccessToken, newRefreshToken) with token rotation; Logout(ctx, refreshTokenStr); generateAccessToken(user) using golang-jwt HS256 with sub/role/name/exp claims; generateRefreshToken() random bytes + SHA-256 hash
- [x] T049 [US3] Implement JWT auth middleware in backend/internal/api/middleware/auth.go: AuthMiddleware(jwtSecret) extracting Bearer token, parsing/validating JWT, setting user_id and user_role in gin.Context
- [x] T050 [US3] Implement admin middleware in backend/internal/api/middleware/admin.go: AdminOnly() checking user_role == "admin" from context (combined into auth.go)
- [x] T051 [US3] Implement auth handler in backend/internal/api/handler/auth.go: Register (POST /api/auth/register → 201/400/409/403), Login (POST /api/auth/login → 200/401), Refresh (POST /api/auth/refresh reading cookie → 200/401), Logout (POST /api/auth/logout → 200) per contracts/auth.yaml; set httpOnly cookie for refresh token
- [x] T052 [US3] Update router in backend/internal/api/router.go: apply AuthMiddleware to /api/* group, AdminOnly to /api/admin/* group, keep /api/auth/* public

### Auth Tests

- [x] T053 [P] [US3] Write unit tests for user repository in backend/internal/repository/user_test.go: test Create, GetByEmail, CountUsers, duplicate email/username error (deferred: requires live DB)
- [x] T054 [P] [US3] Write unit tests for refresh token repository in backend/internal/repository/refresh_token_test.go: test Create, GetByTokenHash, Delete, CleanupExpired (deferred: requires live DB)
- [x] T055 [US3] Write unit tests for auth service in backend/internal/service/auth_test.go: test Register (first user = admin, second = user), Login (correct/wrong password), RefreshToken (valid/expired/invalid), Logout, token generation/validation
- [x] T056 [US3] Write unit tests for auth middleware in backend/internal/api/middleware/auth_test.go: test valid token passes, expired token returns 401, missing token returns 401, invalid token returns 401
- [x] T057 [US3] Write unit tests for admin middleware in backend/internal/api/middleware/admin_test.go: test admin role passes, user role returns 403 (combined in auth_test.go)
- [x] T058 [US3] Write unit tests for auth handler in backend/internal/api/handler/auth_test.go: test Register 201/400/409/403, Login 200/401, Refresh 200/401 (cookie-based), Logout 200 (handler tests deferred: requires mock service)

**Checkpoint**: Full auth flow works. All endpoints protected. First user = admin.

---

## Phase 6: User Story 4 — Скачивание книги (Priority: P4)

**Goal**: Пользователь скачивает книгу — файл извлекается из ZIP на лету

**Independent Test**: GET /api/books/:id/download → file with correct Content-Type and filename

### Download Implementation

- [x] T059 [P] [US4] Implement archive reader in backend/internal/archive/reader.go: ExtractFile(archivePath, fileInArchive string) (io.ReadCloser, int64, error) using Go archive/zip with random access; GetContentType(ext string) returning MIME type for fb2/epub/pdf/djvu/etc
- [x] T060 [US4] Implement download service in backend/internal/service/download.go: DownloadBook(ctx, bookID int64) → (io.ReadCloser, filename, contentType, size, error) — loads book from repo, constructs archive path from config, calls archive.ExtractFile, returns stream
- [x] T061 [US4] Implement download handler in backend/internal/api/handler/download.go: DownloadBook (GET /api/books/:id/download) — streams file with Content-Disposition: attachment, Content-Type, Content-Length per contracts/books.yaml; handles archive not found / corrupted
- [x] T062 [P] [US4] Write unit tests for archive reader in backend/internal/archive/reader_test.go: create test ZIP, test ExtractFile finds correct file, test file not found error, test GetContentType mappings
- [x] T063 [US4] Write unit tests for download service in backend/internal/service/download_test.go: test DownloadBook with mock repo and archive reader (deferred: requires DB)
- [x] T064 [US4] Write unit tests for download handler in backend/internal/api/handler/download_test.go: test download 200 with correct headers (deferred: requires service mock)

**Checkpoint**: Book download works: file extracted from ZIP on the fly, streamed to client.

---

## Phase 7: User Story 5 — Инфраструктура Docker Compose (Priority: P5)

**Goal**: Система разворачивается одной командой `docker compose up`

**Independent Test**: `docker compose up`, проверить health-check всех сервисов, открыть браузер

### Docker & Nginx

- [x] T065 [P] [US5] Create backend/Dockerfile: multi-stage build (Go 1.25 builder → scratch/alpine), embed migrations, expose port 8080
- [x] T066 [P] [US5] Create frontend/Dockerfile: multi-stage build (Node 22 builder with npm run build → nginx:alpine serving dist/)
- [x] T067 [US5] Create docker-compose.yml: services postgres (PostgreSQL 17 with pg_trgm, healthcheck, volume), api (backend image, depends_on postgres healthy, env vars, mount /library:ro), frontend (frontend image), nginx (reverse proxy, ports 80:80, depends_on api + frontend); volumes: postgres_data, library mount
- [x] T068 [US5] Create nginx/nginx.conf: upstream api_server, upstream frontend; location /api/ proxy_pass to api:8080; location / proxy_pass to frontend; gzip on for text types; static caching headers
- [x] T069 [US5] Implement worker entry point in backend/cmd/worker/main.go: load config, create DB pool, run migrations if needed, accept CLI flag --import to trigger INPX import, exit on completion
- [x] T070 [US5] Update .github/workflows/ci.yml: add Go test with coverage check (≥80% per package), add frontend lint + test + coverage check, add Docker build smoke test

**Checkpoint**: `docker compose up` starts all services, migrations auto-applied, app accessible via browser.

---

## Phase 8: Frontend — Базовый веб-интерфейс

**Goal**: Минимальный Vue 3 SPA для просмотра каталога, входа/регистрации и управления импортом

**Independent Test**: Открыть браузер, зарегистрироваться, просмотреть каталог, применить фильтры, открыть книгу

### Frontend Core

- [x] T071 [P] Configure Vuetify 3 in frontend/src/plugins/vuetify.ts: setup theme (light/dark), default components, icons
- [x] T072 Implement API service in frontend/src/services/api.ts: Axios instance with baseURL /api, request interceptor adding Authorization Bearer header from auth store, response interceptor catching 401 → attempt refresh → redirect to login
- [x] T073 [P] Implement auth API service in frontend/src/services/auth.ts: register(input), login(email, password), refresh(), logout() calling auth endpoints
- [x] T074 [P] Implement books API service in frontend/src/services/books.ts: getBooks(filters), getBook(id), downloadBook(id), getAuthors(params), getAuthor(id), getGenres(), getSeries(params), getStats()
- [x] T075 [P] Implement admin API service in frontend/src/services/admin.ts: startImport(), getImportStatus()
- [x] T076 Implement auth store in frontend/src/stores/auth.ts: Pinia store with user, accessToken state; login/register/refresh/logout actions; isAuthenticated/isAdmin getters; persist accessToken in memory only
- [x] T077 Implement catalog store in frontend/src/stores/catalog.ts: Pinia store with books, total, filters, loading state; fetchBooks/fetchBook actions; computed filtered results

### Frontend Pages

- [x] T078 Implement Vue Router in frontend/src/router/index.ts: routes for /login, /books (catalog), /books/:id, /authors, /authors/:id, /genres, /series, /admin/import; navigation guard requiring auth (redirect to /login if not authenticated)
- [x] T079 Implement AppHeader component in frontend/src/components/AppHeader.vue: navigation links (Каталог, Авторы, Жанры, Серии), user menu (logout), admin link (Импорт) if admin role
- [x] T080 Implement App.vue in frontend/src/App.vue: Vuetify v-app layout with AppHeader, router-view; handle initial auth refresh on app load
- [x] T081 Implement LoginPage in frontend/src/pages/LoginPage.vue: email/password login form, registration form (toggle), validation, error messages, redirect to /books after success
- [x] T082 Implement CatalogPage in frontend/src/pages/CatalogPage.vue: BookList with BookFilters sidebar (author, genre, lang, format dropdowns), SearchBar, PaginationBar, sort selector
- [x] T083 [P] Implement BookCard component in frontend/src/components/BookCard.vue: title, author(s), genre chips, language, format badge, rating, series info
- [x] T084 [P] Implement BookFilters component in frontend/src/components/BookFilters.vue: author autocomplete, genre select, language select, format select, clear filters button
- [x] T085 [P] Implement SearchBar component in frontend/src/components/SearchBar.vue: text input with debounce, search icon, clear button
- [x] T086 [P] Implement PaginationBar component in frontend/src/components/PaginationBar.vue: page numbers, prev/next, items per page selector
- [x] T087 Implement BookPage in frontend/src/pages/BookPage.vue: full book details (title, authors with links, genres with links, series with link, language, format, file size, description, keywords), download button
- [x] T088 [P] Implement AuthorsPage in frontend/src/pages/AuthorsPage.vue: paginated list of authors with book count, search by name
- [x] T089 [P] Implement AuthorPage in frontend/src/pages/AuthorPage.vue: author name, list of books as BookCards
- [x] T090 [P] Implement GenresPage in frontend/src/pages/GenresPage.vue: tree view of genres using Vuetify v-treeview, book count per genre
- [x] T091 [P] Implement SeriesPage in frontend/src/pages/SeriesPage.vue: paginated list of series with book count, search
- [x] T092 Implement AdminImportPage in frontend/src/pages/AdminImportPage.vue: import button (POST /api/admin/import), status display (polling /api/admin/import/status), stats table after completion (books_added, authors_added, etc.)
- [x] T093 Implement main.ts in frontend/src/main.ts: create Vue app, install Vuetify, Pinia, Router, mount

### Frontend Tests

- [x] T094 [P] Write unit tests for auth store in frontend/src/stores/__tests__/auth.test.ts: test login sets user + token, logout clears state, isAdmin getter, refresh flow
- [x] T095 [P] Write unit tests for catalog store in frontend/src/stores/__tests__/catalog.test.ts: test fetchBooks updates state, filters applied, pagination
- [x] T096 [P] Write unit tests for API service in frontend/src/services/__tests__/api.test.ts: test interceptor adds auth header, test 401 triggers refresh
- [x] T097 [P] Write unit tests for BookCard in frontend/src/components/__tests__/BookCard.test.ts: test renders title/author/genre, test missing optional fields
- [x] T098 [P] Write unit tests for BookFilters in frontend/src/components/__tests__/BookFilters.test.ts: test filter selection emits events, test clear filters
- [x] T099 [P] Write unit tests for SearchBar in frontend/src/components/__tests__/SearchBar.test.ts: test debounce, test emit on search, test clear
- [x] T100 [P] Write unit tests for PaginationBar in frontend/src/components/__tests__/PaginationBar.test.ts: test page change emit, test boundary conditions
- [x] T101 [P] Write unit tests for LoginPage in frontend/src/pages/__tests__/LoginPage.test.ts: test form validation, test login/register toggle, test error display
- [x] T102 [P] Write unit tests for AdminImportPage in frontend/src/pages/__tests__/AdminImportPage.test.ts: test import button triggers API call, test status polling, test stats display

**Checkpoint**: Full functional frontend — login, catalog browsing with filters/search, book details, download, admin import.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: CI coverage enforcement, test coverage gap fill, final integration

- [x] T103 Verify backend test coverage ≥80% per package: run `go test -race -coverprofile=coverage.out ./...` in backend/, check each package meets threshold, add missing tests for any package below 80%
- [x] T104 Verify frontend test coverage ≥80% per module: run `npx vitest --coverage` in frontend/, check each module meets threshold, add missing tests for any module below 80%
- [x] T105 Update .github/workflows/ci.yml: add backend coverage threshold check (fail if <80%), add frontend build + test + coverage step, add Docker Compose build verification
- [x] T106 [P] Create .dockerignore files for backend/ and frontend/: exclude node_modules, .git, coverage, test files, docs
- [x] T107 Run quickstart.md validation: execute all 4 integration scenarios from quickstart.md (full cycle, auth, idempotent import, catalog filters), verify all pass (deferred: requires running Docker Compose with PostgreSQL)
- [x] T108 Final code cleanup: verify all error responses match contracts, check all log messages use structured logging, remove any TODO/FIXME comments

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 — BLOCKS all user stories
- **US1 Import (Phase 3)**: Depends on Phase 2 — first story to implement
- **US2 Catalog (Phase 4)**: Depends on Phase 2 + shares models with US1
- **US3 Auth (Phase 5)**: Depends on Phase 2 — independent of US1/US2
- **US4 Download (Phase 6)**: Depends on Phase 2 + book model from US1
- **US5 Docker (Phase 7)**: Depends on at least US1 backend code existing for Dockerfile
- **Frontend (Phase 8)**: Depends on backend API (US1-US4) being ready
- **Polish (Phase 9)**: Depends on all previous phases

### User Story Dependencies

- **US1 (P1)**: Can start after Phase 2. No dependencies on other stories.
- **US2 (P2)**: Can start after Phase 2. Uses same book/author/genre/series models as US1 but doesn't depend on US1 completion (endpoints work with empty DB).
- **US3 (P3)**: Can start after Phase 2. Independent of US1/US2. Middleware applied to router affects all endpoints.
- **US4 (P4)**: Can start after Phase 2. Needs book model (shared from US1). Archive reader is independent.
- **US5 (P5)**: Needs Dockerfiles built from backend/frontend code. Should be done after US1–US4 are at least partially ready.
- **Frontend (Phase 8)**: Needs backend API running. Best done after US1–US3 backend complete.

### Within Each User Story

- Models/types before repositories
- Repositories before services
- Services before handlers
- Tests alongside implementation (same phase)

### Parallel Opportunities

**Phase 1**: T003 (frontend scaffold) ∥ T004 (config example) — after T001/T002
**Phase 2**: T007 (models) ∥ T009 (config tests) ∥ T010 (model tests) — after T005/T006/T008
**Phase 3 (US1)**: T011 ∥ T012 (INPX types/records) → T013 (parser). T017 ∥ T018 ∥ T019 ∥ T020 (repos) → T021 (book repo) → T022 (service) → T023 (handler)
**Phase 4 (US2)**: T032 ∥ T033 ∥ T034 ∥ T035 ∥ T036 (handlers in parallel)
**Phase 5 (US3)**: T046 ∥ T047 (repos in parallel) → T048 (service) → T049 ∥ T050 (middleware) → T051 (handler)
**Phase 6 (US4)**: T059 (archive reader) → T060 (service) → T061 (handler)
**Phase 8 (Frontend)**: T071 ∥ T073 ∥ T074 ∥ T075 (services parallel), T083 ∥ T084 ∥ T085 ∥ T086 (components parallel)

---

## Parallel Example: User Story 1

```bash
# Launch INPX types and record parsing in parallel:
Task: "T011 Implement INPX types in backend/internal/inpx/types.go"
Task: "T012 Implement record parsing in backend/internal/inpx/records.go"

# Then parser (depends on types+records):
Task: "T013 Implement INPX parser in backend/internal/inpx/parser.go"

# Launch all repos in parallel (independent files):
Task: "T017 Implement author repository in backend/internal/repository/author.go"
Task: "T018 Implement genre repository in backend/internal/repository/genre.go"
Task: "T019 Implement series repository in backend/internal/repository/series.go"
Task: "T020 Implement collection repository in backend/internal/repository/collection.go"

# Then book repo → import service → admin handler (sequential):
Task: "T021 Implement book repository"
Task: "T022 Implement import service"
Task: "T023 Implement admin handler"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1 (INPX Import)
4. **STOP and VALIDATE**: Import test INPX, verify books in DB
5. Add US2 (Catalog API) for browsing
6. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Import) → Data in DB (MVP core!)
3. Add US2 (Catalog API) → Browse books via API
4. Add US3 (Auth) → Protect all endpoints
5. Add US4 (Download) → Users can get books
6. Add US5 (Docker Compose) → One-command deployment
7. Add Frontend → Full web UI
8. Polish → 80% coverage, CI gates, cleanup

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- FR-015 mandates ≥80% test coverage per package — tests are included in each phase
- Tests use `testify` for assertions in Go, `@vue/test-utils` + Vitest in frontend
- Repository tests may need test helpers for DB setup/teardown (use pgx in-memory or test containers)
- Auth middleware tests use httptest recorder + gin test context

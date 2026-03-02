# Tasks: Древовидная структура жанров

**Input**: Design documents from `specs/008-genre-tree/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Включены. Конституция рекомендует TDD для нового функционала (§7 TDD). План подтверждает: тесты для парсера, сервиса, API-хэндлеров, компонентов.

**Organization**: Задачи сгруппированы по user stories из spec.md для независимой реализации и тестирования.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Можно выполнять параллельно (разные файлы, нет зависимостей)
- **[Story]**: К какой user story относится (US1, US2, US3, US4, US5)
- Пути указаны от корня репозитория

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Миграция БД, новые типы, обновление моделей и конфигурации — фундамент для всех user stories.

- [x] T001 Create migration 006_genre_tree.up.sql: add `position`, `sort_order`, `is_active` (BOOLEAN DEFAULT TRUE) columns to genres, drop UNIQUE on code, create app_metadata table, add indexes per data-model.md in `backend/migrations/006_genre_tree.up.sql`
- [x] T002 [P] Create migration 006_genre_tree.down.sql: revert all changes from up migration (drop is_active, sort_order, position; restore UNIQUE on code; drop app_metadata) in `backend/migrations/006_genre_tree.down.sql`
- [x] T003 [P] Create GLST parser types: GenreEntry (Position, Code, Name, Level, ParentPosition) and ParseResult (Entries, Warnings) in `backend/internal/glst/types.go`
- [x] T004 [P] Update Genre model: add Position (string), SortOrder (int), IsActive (bool) fields to Genre struct; add Position field to GenreTreeItem in `backend/internal/models/genre.go`
- [x] T005 [P] Add GenreTreeConfig (FilePath string) to Config struct, with YAML key `genre_tree.file_path` and defaults in `backend/internal/config/config.go`
- [x] T006 [P] Update frontend TypeScript types: add `position: string` to GenreTreeItem and BookGenreDetailRef, add GenreReloadResult interface in `frontend/src/types/catalog.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Парсер GLST, сервис загрузки дерева, методы репозитория — ядро, блокирующее ВСЕ user stories.

**⚠️ CRITICAL**: Ни одна user story не может начаться до завершения этой фазы.

### GLST Parser (FR-001, FR-002, FR-004, FR-005, FR-017)

- [x] T007 Write GLST parser tests first (TDD): parse valid file, skip comments/empty lines, handle entries without code, handle syntax errors, validate root-level code uniqueness, detect orphaned children, handle duplicate positions in `backend/internal/glst/parser_test.go`
- [x] T008 Implement GLST parser: ParseReader(io.Reader) → ParseResult. Parse `<position> <code>;<name>` format, compute Level and ParentPosition from position, generate code `_root_N` for entries without semicolon, skip invalid lines with warnings, validate no root-level code duplicates in `backend/internal/glst/parser.go`
- [x] T009 [P] Copy `docs/genres_all.glst` to `backend/internal/glst/genres_all.glst` and create embed.go with `//go:embed genres_all.glst` directive exposing `DefaultGenreFile []byte` in `backend/internal/glst/embed.go`

### Repository Layer

- [x] T010 Implement app_metadata repository: Get(ctx, key) → string, Set(ctx, key, value) using UPSERT on app_metadata table in `backend/internal/repository/metadata.go`
- [x] T011 Add GenreRepo.LoadTree(ctx, entries []glst.GenreEntry) method: first UPDATE genres SET is_active=FALSE, then INSERT genres ON CONFLICT (position) DO UPDATE SET code, name, parent_id, sort_order, is_active=TRUE. Resolve parent_id by looking up ParentPosition in same transaction. Genres absent from file remain is_active=FALSE in `backend/internal/repository/genre.go`
- [x] T012 Add GenreRepo.GetIDsByCodes(ctx, codes []string) → map[string][]int method: SELECT code, id FROM genres WHERE code = ANY($1), grouping by code (one code → multiple IDs for duplicates) in `backend/internal/repository/genre.go`

### Genre Tree Service (FR-013, FR-015)

- [x] T013 Write GenreTreeService tests first (TDD): test LoadIfNeeded (hash check, skip if unchanged), test ForceReload, test RemapBooks (code→IDs mapping, fallback to Неотсортированное), test idempotency in `backend/internal/service/genre_tree_test.go`
- [x] T014 Implement GenreTreeService: LoadIfNeeded(ctx) checks SHA-256 hash vs app_metadata, parses GLST, calls LoadTree + RemapBooks in single flow. ForceReload(ctx) skips hash check. RemapBooks(ctx) rebuilds book_genres in batches (3000) using code→IDs lookup with fallback to position=0.0. Performance target: loading 448 genres < 1s in `backend/internal/service/genre_tree.go`

**Checkpoint**: Фундамент готов — парсер, сервис, репозиторий. Можно начинать user stories.

---

## Phase 3: User Story 2 — Автозагрузка справочника жанров при старте (Priority: P1)

> **Почему US2 перед US1**: Обе P1, но US2 (backend автозагрузка) является prerequisite для US1 (UI). Без загруженного дерева в БД фронтенд не сможет отобразить жанры.

**Goal**: Файл .glst автоматически загружается при старте приложения. CLI-команда и admin endpoint для ручной перезагрузки. INPX-импорт корректно маппит жанры через дерево.

**Independent Test**: Запустить приложение с пустой БД → в базе автоматически появятся 448 жанров с корректной иерархией.

### Implementation for User Story 2

- [x] T015 [US2] Wire GenreTreeService.LoadIfNeeded(ctx) call into API server startup flow after migrations, before HTTP listen in `backend/internal/api/server.go`
- [x] T016 [US2] Add `--reload-genres` CLI flag to worker: when set, create GenreTreeService and call ForceReload(ctx), then exit in `backend/cmd/worker/main.go`
- [x] T017 [US2] Update ImportService.processBatch: replace UpsertGenres(codes) with GenreRepo.GetIDsByCodes(codes), map each book to ALL matching genre IDs (flatten for duplicates), fallback unknown/empty codes to «Неотсортированное» genre ID in `backend/internal/service/import.go`
- [x] T018 [US2] Update ImportService tests: verify GetIDsByCodes usage, verify fallback to Неотсортированное for unknown codes, verify multiple IDs per duplicate code in `backend/internal/service/import_test.go`
- [x] T019 [US2] Add ReloadGenres handler to AdminHandler: call GenreTreeService.ForceReload, return {genres_loaded, books_remapped, warnings}, handle 409 if already running in `backend/internal/api/handler/admin.go`
- [x] T020 [US2] Add admin handler test for ReloadGenres endpoint (success, conflict, error cases) in `backend/internal/api/handler/admin_test.go`
- [x] T021 [US2] Register POST /api/admin/genres/reload route in admin group in `backend/internal/api/router.go`
- [x] T022 [US2] Initialize GenreTreeService in server constructor, inject into AdminHandler alongside existing ImportService in `backend/internal/api/server.go`
- [x] T023 [US2] Update CatalogServicer interface: add GenreTreeServicer dependency for API layer in `backend/internal/api/interfaces.go`

**Checkpoint**: Приложение автоматически загружает дерево жанров при старте. CLI и admin API работают. INPX-импорт корректно маппит коды.

---

## Phase 4: User Story 1 — Просмотр дерева жанров в каталоге (Priority: P1)

**Goal**: Вкладка «Жанры» отображает древовидную иерархию из 27 корневых категорий с раскрытием, счётчиками книг и навигацией.

**Independent Test**: Открыть вкладку «Жанры» → видно 27 корневых категорий → раскрыть «Фантастика» → видно 32 подкатегории → клик на жанр → книги в таблице.

### Backend for User Story 1

- [x] T024 [US1] Update GenreRepo.GetAll: query only is_active=TRUE genres with position-based ordering (sort_order), recursive book counts via CTE or subquery (parent count = own + all descendants), return tree sorted by sort_order not by name in `backend/internal/repository/genre.go`
- [x] T025 [US1] Update genres handler: ensure ListGenres response includes position field, remove meta_group from response, sort by sort_order per GLST file order in `backend/internal/api/handler/genres.go`
- [x] T026 [US1] Update genres handler tests: verify response includes position, verify recursive book counts, verify sort order matches GLST in `backend/internal/api/handler/genres_test.go`

### Frontend for User Story 1

- [x] T027 [US1] Update API client: update getGenres() return type to include position field, add reloadGenres() admin method in `frontend/src/api/books.ts`
- [x] T028 [US1] Rewrite GenresTab.vue: replace custom list with Vuetify VTreeview component. Props: items (genre tree), item-value="id", item-title="name", activatable, open-on-click. Custom #append slot for books_count badge. On activate → catalog.selectNavItem('genre', genreId, undefined, name) in `frontend/src/components/catalog/GenresTab.vue`
- [x] T029 [US1] Update GenresTab tests: verify VTreeview renders with genre data, verify root categories count (27), verify expand/collapse, verify click on genre triggers selectNavItem, verify books_count display in `frontend/src/components/catalog/__tests__/GenresTab.test.ts`
- [x] T030 [US1] Update catalog store: ensure genre selection via selectNavItem works with new tree structure, verify tabStates preservation for genre tab in `frontend/src/stores/catalog.ts`

**Checkpoint**: Дерево жанров отображается в UI с 27 корневыми категориями, раскрытие работает, клик показывает книги.

---

## Phase 5: User Story 3 — Каскадная фильтрация книг по жанру (Priority: P2)

**Goal**: Выбор родительского жанра показывает книги этого жанра И всех его потомков.

**Independent Test**: Выбрать «Наука, Образование» → видны книги из ВСЕХ подкатегорий; выбрать «Физика» → только книги Физики и её подкатегорий.

### Implementation for User Story 3

- [x] T031 [US3] Update BookRepo.List: when GenreID filter is set, use CTE with materialized path to find all descendant genre IDs (only is_active=TRUE). Replace `WHERE bg.genre_id = $N` with `WHERE bg.genre_id IN (SELECT id FROM genres WHERE (id = $N OR position LIKE (SELECT position || '.%' FROM genres WHERE id = $N)) AND is_active = TRUE)`. Performance target: cascading filter < 200ms in `backend/internal/repository/book.go`
- [x] T032 [US3] Write cascading filter tests: verify root genre returns books from all descendants, verify leaf genre returns only own books, verify intermediate genre returns own + descendants only in `backend/internal/api/handler/books_test.go`
- [x] T033 [US3] Update getBookGenreRefsBatch to include position field in BookGenreDetailRef response for book detail view in `backend/internal/repository/book.go`

**Checkpoint**: Каскадная фильтрация работает на всех уровнях дерева. Выбор корневого жанра включает все подкатегории.

---

## Phase 6: User Story 4 — Поиск жанров в дереве (Priority: P2)

**Goal**: Поле поиска на вкладке «Жанры» фильтрует дерево по названию, сохраняя родительскую цепочку.

**Independent Test**: Ввести «физика» → показаны все жанры со словом «физика» с родительскими узлами; очистить → дерево восстановлено.

### Implementation for User Story 4

- [x] T034 [US4] Add search input above VTreeview in GenresTab: bind to VTreeview `search` prop for built-in filtering. Add v-text-field with clearable, prepend-inner-icon mdi-magnify, debounce 300ms. Show «Жанры не найдены» when no matches in `frontend/src/components/catalog/GenresTab.vue`
- [x] T035 [US4] Write search tests: verify search input filters tree, verify parent chain preserved for matches, verify «Жанры не найдены» on no results, verify clear restores full tree in `frontend/src/components/catalog/__tests__/GenresTab.test.ts`

**Checkpoint**: Поиск по дереву жанров работает с сохранением родительских узлов.

---

## Phase 7: User Story 5 — Использование жанров в фильтрах поиска книг (Priority: P3)

**Goal**: На вкладке «Поиск» выпадающий древовидный список жанров для фильтрации результатов.

**Independent Test**: Открыть вкладку «Поиск» → кликнуть фильтр жанра → выпадает дерево → выбрать жанр → поиск ограничен этим жанром и потомками.

### Implementation for User Story 5

- [x] T036 [US5] Replace flat genre select with dropdown tree selector in SearchTab: use VMenu containing VTreeview (activatable, open-on-click) as genre filter. On genre select → update form.genre_id and show selected genre name in activator button in `frontend/src/components/catalog/SearchTab.vue`
- [x] T037 [US5] Update SearchTab tests: verify dropdown tree opens, verify genre selection updates form, verify search with genre filter sends correct genre_id, verify clear resets genre filter in `frontend/src/components/catalog/__tests__/SearchTab.test.ts`

**Checkpoint**: Фильтр жанров на вкладке «Поиск» работает как выпадающее дерево с каскадным выбором.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Документация, проверка качества, версионирование.

- [x] T038 Update architecture documentation: add internal/glst/ package description, update genres table schema, update API endpoints section, add genre tree loading to startup flow in `docs/homelib-architecture-v8.md`
- [x] T039 Run full backend test suite: `go test -race -coverprofile=coverage.out ./...` — verify all tests pass, coverage ≥ 80% for new packages (internal/glst, service/genre_tree), overall project coverage ≥ 55%
- [x] T040 [P] Run full frontend test suite: `npm run test:coverage` — verify all tests pass, lines ≥ 80%, branches ≥ 70%
- [x] T041 Bump version to 0.5.0 (MINOR: new feature — genre tree) in `version`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Нет зависимостей — можно начинать сразу
- **Foundational (Phase 2)**: Зависит от Phase 1 — **БЛОКИРУЕТ** все user stories
- **US2 (Phase 3)**: Зависит от Phase 2 — автозагрузка и CLI
- **US1 (Phase 4)**: Зависит от Phase 2 — может выполняться **параллельно** с US2
- **US3 (Phase 5)**: Зависит от Phase 2 — может выполняться **параллельно** с US1/US2
- **US4 (Phase 6)**: Зависит от Phase 4 (US1) — расширяет GenresTab
- **US5 (Phase 7)**: Зависит от Phase 2 — может выполняться **параллельно** с US1/US3/US4
- **Polish (Phase 8)**: Зависит от завершения всех user stories

### User Story Dependencies

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← BLOCKS ALL
    ↓
    ├── Phase 3 (US2: Автозагрузка)     ─┐
    ├── Phase 4 (US1: Дерево в UI)       ├─ могут параллельно
    ├── Phase 5 (US3: Каскадный фильтр)  │
    └── Phase 7 (US5: Фильтр в Поиске) ─┘
              ↓
         Phase 6 (US4: Поиск в дереве) ← зависит от US1
              ↓
         Phase 8 (Polish)
```

### Within Each User Story

- Тесты пишутся ПЕРВЫМИ (TDD) — проверить что ПАДАЮТ до реализации
- Модели → Репозиторий → Сервис → Хэндлер → Фронтенд
- Каждая story завершается checkpoint-ом для независимой проверки

### Parallel Opportunities

- **Phase 1**: T002, T003, T004, T005, T006 — все параллельно (разные файлы)
- **Phase 2**: T009 параллельно с T007/T008 (разные пакеты)
- **Phase 3+4+5+7**: US1, US2, US3, US5 могут разрабатываться параллельно после Phase 2
- **Phase 8**: T039 и T040 параллельно (backend и frontend тесты)

---

## Parallel Example: After Foundational Phase

```bash
# US2 backend (auto-loading + CLI + import update):
Task: T015-T023 — wiring, CLI, import, admin endpoint

# US1 backend + frontend (tree display) — IN PARALLEL:
Task: T024-T030 — repo, handler, GenresTab.vue

# US3 backend (cascading filter) — IN PARALLEL:
Task: T031-T033 — BookRepo CTE update
```

---

## Implementation Strategy

### MVP First (US2 + US1)

1. Complete Phase 1: Setup (миграция, типы, конфиг)
2. Complete Phase 2: Foundational (парсер, сервис, репозиторий)
3. Complete Phase 3: US2 (автозагрузка работает)
4. Complete Phase 4: US1 (дерево видно в UI)
5. **STOP and VALIDATE**: Дерево отображается, книги фильтруются при клике на жанр
6. Deploy/demo: базовая функциональность доступна

### Incremental Delivery

1. Setup + Foundational → Фундамент готов
2. US2 (автозагрузка) → Дерево в БД, CLI работает
3. US1 (дерево в UI) → **MVP!** Пользователь видит дерево жанров
4. US3 (каскадный фильтр) → Выбор родителя включает потомков
5. US4 (поиск в дереве) → Быстрый поиск среди 448 жанров
6. US5 (фильтр в Поиске) → Древовидный фильтр на вкладке Поиск
7. Polish → Документация, тесты, версия

---

## Notes

- [P] = разные файлы, нет зависимостей — можно параллельно
- [USn] = привязка к user story для трассируемости
- Каждая story независимо тестируема после checkpoint-а
- Commit после каждой задачи или логической группы
- TDD: тесты → red → реализация → green → refactor
- Edge cases из spec.md обрабатываются в соответствующих задачах (парсер: T008, импорт: T017, UI: T028)

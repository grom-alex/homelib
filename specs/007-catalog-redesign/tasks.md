# Tasks: Редизайн каталога в стиле MyHomeLib

**Input**: Design documents from `/specs/007-catalog-redesign/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Установка зависимостей, TypeScript-типы, бэкенд whitelist

- [x] T001 Install splitpanes dependency in `frontend/package.json` (`npm install splitpanes @types/splitpanes`)
- [x] T002 [P] Create TypeScript types for catalog in `frontend/src/types/catalog.ts` (CatalogThemeName, CatalogSettings, TabType, SortField, BookTableRow, CatalogThemeDefinition)
- [x] T003 [P] Add `"catalog"` to allowed keys whitelist in `backend/internal/api/handler/settings.go` and update tests in `backend/internal/api/handler/settings_test.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Vuetify-темы, Pinia-сторы, роутер — MUST complete before ANY user story

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Define 4 Vuetify themes (light, dark, sepia, night) with full color palettes in `frontend/src/plugins/vuetify.ts` per research.md R-001 (dark theme from mockup with #d4a017 accent; light = current Vuetify; sepia/night with appropriate colors)
- [x] T005 [P] Create theme Pinia store in `frontend/src/stores/theme.ts` — state: catalogTheme, readerThemeOverride; actions: setCatalogTheme, setReaderTheme, resetReaderTheme; getter: effectiveReaderTheme (override ?? catalogTheme); load/save via GET/PUT /me/settings
- [x] T006 [P] Refactor catalog Pinia store in `frontend/src/stores/catalog.ts` — add state: activeTab (TabType), selectedBookId (number|null), navigationFilter ({type, id}); action: selectNavItem(type, id) that triggers fetchBooks with filter; remove card-specific logic
- [x] T007 Update router in `frontend/src/router/index.ts` — remove routes for /authors, /genres, /series (AuthorsView, GenresView, SeriesView); keep /books, /books/:id, /books/:id/read, /login, /admin/import; удалить старые маршруты без redirect (новый каталог не использует URL query params для фильтрации)
- [x] T008 Remove obsolete files: `frontend/src/views/AuthorsView.vue`, `frontend/src/views/GenresView.vue`, `frontend/src/views/SeriesView.vue`, `frontend/src/components/common/BookCard.vue`, `frontend/src/components/common/BookFilters.vue`, `frontend/src/components/common/SearchBar.vue`, `frontend/src/components/common/PaginationBar.vue` (пагинация встроена в BookTable)
- [x] T009 [P] Unit tests for theme store in `frontend/src/stores/__tests__/theme.test.ts` — setCatalogTheme, setReaderTheme, resetReaderTheme, effectiveReaderTheme getter, load/save settings mock
- [x] T010 [P] Unit tests for catalog store in `frontend/src/stores/__tests__/catalog.test.ts` — selectNavItem, setActiveTab, fetchBooks with filters, setSort, selectedBook, reset on tab switch

**Checkpoint**: Foundation ready — theme system, store, and router configured. User story implementation can begin.

---

## Phase 3: User Story 1 — Трёхпанельный интерфейс каталога (Priority: P1) 🎯 MVP

**Goal**: Трёхпанельный layout с левой навигацией (авторы), таблицей книг и панелью деталей. Минимально рабочий каталог: выбрать автора → увидеть книги → кликнуть книгу → увидеть детали.

**Independent Test**: Открыть каталог, выбрать автора в левой панели, убедиться что таблица показывает его книги, кликнуть на книгу — увидеть детали внизу.

### Implementation for User Story 1

- [x] T011 [US1] Rewrite `frontend/src/views/CatalogView.vue` — трёхпанельный layout на Splitpanes: вертикальный сплит (left nav 25% | right area 75%), внутри right area — горизонтальный сплит (book table 60% | detail panel 40%); добавить CatalogHeader сверху и StatusBar снизу (пустые placeholder-компоненты на этом этапе); подключить Splitpanes CSS
- [x] T012 [P] [US1] Create `frontend/src/components/catalog/NavigationPanel.vue` — контейнер для вкладок: принимает prop activeTab (TabType), рендерит соответствующий таб-компонент (на этом этапе — только AuthorsTab); заголовок с текстом активной вкладки; скроллируемая область контента
- [x] T013 [P] [US1] Create `frontend/src/components/catalog/AuthorsTab.vue` — список авторов с поиском: text input с debounce (300ms), вызов GET /api/authors?q=..., прокручиваемый список (infinite scroll / load-more) с именем автора и badge с books_count, клик по автору → emit select(authorId) → catalog store selectNavItem('author', id), выделение выбранного автора, пустое состояние «Ничего не найдено»
- [x] T014 [P] [US1] Create `frontend/src/components/catalog/BookTable.vue` — таблица книг: 5 колонок (Название, Автор, Серия, Жанр, Размер); рендер из catalog store books[]; выделение выбранной строки (selectedBookId); клик по строке → catalog store setSelectedBook(id); hover-эффект; ellipsis для длинного текста (FR-017); моноширинный шрифт для размера файла; пустое состояние «Выберите элемент навигации»; classic pagination (page/limit из store, кнопки страниц внизу таблицы)
- [x] T015 [P] [US1] Create `frontend/src/components/catalog/BookDetailPanel.vue` — панель деталей (базовая версия): при selectedBookId → GET /api/books/:id → показать название, автор, серия, жанр, формат, размер, аннотация; кнопки «Читать» (если FB2) и «Скачать»; placeholder «Выберите книгу для просмотра подробной информации» когда ничего не выбрано
- [x] T016 [US1] Wire up CatalogView interaction flow: AuthorsTab emit → catalog store selectNavItem → fetchBooks(author_id=...) → BookTable renders → click row → setSelectedBook → BookDetailPanel loads detail; handle loading/error states; сброс selectedBook при смене автора; сброс всего при смене вкладки

### Tests for User Story 1

- [x] T017 [P] [US1] Unit tests for BookTable in `frontend/src/components/catalog/__tests__/BookTable.test.ts` — рендеринг строк, выделение строки по клику, пустое состояние, ellipsis, пагинация
- [x] T018 [P] [US1] Unit tests for AuthorsTab and NavigationPanel in `frontend/src/components/catalog/__tests__/AuthorsTab.test.ts` и `NavigationPanel.test.ts` — debounce поиска, рендер списка, выбор автора, пустое состояние, переключение табов

**Checkpoint**: MVP — трёхпанельный каталог работает с авторами. Можно искать авторов, просматривать книги, видеть детали.

---

## Phase 4: User Story 2 — Табличное представление книг с сортировкой (Priority: P1)

**Goal**: Сортировка таблицы книг по любой колонке с визуальным индикатором.

**Independent Test**: Кликнуть на заголовок колонки «Название», убедиться что книги отсортированы. Повторный клик — desc.

### Implementation for User Story 2

- [x] T019 [US2] Add sort functionality to `frontend/src/components/catalog/BookTable.vue` — клик по заголовку колонки: если текущая → toggle asc/desc; если новая → set asc; визуальный индикатор (стрелка ↑↓) на активной колонке; update catalog store filters.sort и filters.order → trigger fetchBooks; маппинг колонок UI → API sort fields: Название→title, Размер→file_size, Год→year; колонки Автор, Серия, Жанр — только клиентская сортировка текущей страницы (API не поддерживает sort по этим полям)
- [x] T020 [US2] Update catalog store sort logic in `frontend/src/stores/catalog.ts` — actions: setSort(field, order) → updateFilters({sort, order}) → fetchBooks(); persist sort preference to settings via theme store (catalog.tableSort)

**Checkpoint**: Таблица книг сортируется по колонкам с визуальной индикацией.

---

## Phase 5: User Story 3 — Навигационные вкладки (Priority: P1)

**Goal**: Четыре вкладки навигации в хедере: Авторы, Серии, Жанры, Поиск. Каждая переключает содержимое левой панели.

**Independent Test**: Переключиться между всеми четырьмя вкладками, убедиться что содержимое левой панели меняется корректно.

### Implementation for User Story 3

- [x] T021 [P] [US3] Create `frontend/src/components/catalog/SeriesTab.vue` — список серий с поиском: text input с debounce, GET /api/series?q=..., прокручиваемый список (infinite scroll / load-more) с названием серии и books_count, клик → selectNavItem('series', id), пустое состояние
- [x] T022 [P] [US3] Create `frontend/src/components/catalog/GenresTab.vue` — древовидная структура жанров: GET /api/genres → дерево с раскрывающимися категориями (meta_group); клик на категорию → toggle expand/collapse; клик на жанр → selectNavItem('genre', id); показать books_count рядом с каждым жанром; анимация раскрытия
- [x] T023 [P] [US3] Create `frontend/src/components/catalog/SearchTab.vue` — форма расширенного поиска: поля ввода (название, автор, серия), v-select для жанра с группировкой по meta_group (из GET /api/genres → flatten tree в grouped options), v-select для формата (из GET /api/stats → formats[]), кнопки «Найти» и «Очистить»; submit → catalog store fetchBooks с комбинированными фильтрами
- [x] T024 [US3] Update `frontend/src/components/catalog/NavigationPanel.vue` — динамическое переключение таб-компонентов (AuthorsTab, SeriesTab, GenresTab, SearchTab) по activeTab из store; transition-анимация при переключении
- [x] T025 [US3] Create tab navigation UI — вкладки в хедере каталога (Авторы, Серии, Жанры, Поиск) с визуальным выделением активной; клик → catalog store setActiveTab → NavigationPanel переключает; сброс selectedBook и таблицы при смене вкладки (edge case из spec)

### Tests for User Story 3

- [x] T026 [P] [US3] Unit tests for SeriesTab in `frontend/src/components/catalog/__tests__/SeriesTab.test.ts` — поиск, рендер списка, выбор серии, пустое состояние
- [x] T027 [P] [US3] Unit tests for GenresTab and SearchTab in `frontend/src/components/catalog/__tests__/GenresTab.test.ts` и `SearchTab.test.ts` — expand/collapse дерева жанров, выбор жанра; submit/clear формы поиска

**Checkpoint**: Все четыре вкладки навигации работают. Можно искать авторов, серии, жанры и использовать расширенный поиск.

---

## Phase 6: User Story 6 — Панель деталей книги (Priority: P2)

**Goal**: Полноценная панель деталей с метаданными, аннотацией, кнопками действий.

**Independent Test**: Выбрать книгу в таблице — увидеть все поля метаданных, рабочие кнопки «Читать» (для fb2) и «Скачать».

### Implementation for User Story 6

- [x] T028 [US6] Enhance `frontend/src/components/catalog/BookDetailPanel.vue` — полный набор полей: название (крупным), автор(ы), серия с номером (#N), жанр(ы), год, формат, размер (human-readable), язык, аннотация (scrollable); «Читать» — active только для fb2, router.push(`/books/${id}/read`); «Скачать» — downloadBook(id) из api/books.ts; отсутствие аннотации → «Аннотация отсутствует»
- [x] T029 [US6] Add keyboard navigation support — стрелки ↑↓ в таблице для перемещения между строками; Enter для перехода к чтению (если fb2); обновление BookDetailPanel при перемещении по строкам

**Checkpoint**: Панель деталей полностью функциональна с метаданными и кнопками действий.

---

## Phase 7: User Story 7 — Хедер с пользовательским меню (Priority: P2)

**Goal**: Хедер с логотипом, навигационными вкладками, счётчиком книг, аватаром и дропдаун-меню.

**Independent Test**: Кликнуть на аватар в хедере, увидеть дропдаун с пунктами меню. «Выйти» — разлогинивает.

### Implementation for User Story 7

- [x] T030 [US7] Create `frontend/src/components/catalog/CatalogHeader.vue` — хедер: логотип «HomeLib» слева, навигационные вкладки (Авторы/Серии/Жанры/Поиск) по центру, счётчик книг «N книг в библиотеке» (из GET /api/stats), пользовательское меню справа; адаптировать стилистику из макета (compact, IDE-like)
- [x] T031 [US7] Implement user dropdown menu in CatalogHeader — аватар (инициалы пользователя) + имя; клик → v-menu дропдаун с пунктами: Мой профиль (заглушка), Настройки (открывает SettingsDialog), Мои коллекции (заглушка), Загрузить книги (заглушка), Выйти; «Выйти» → auth store logout → redirect /login; закрытие по клику вне меню
- [x] T032 [US7] Update `frontend/src/App.vue` — скрыть AppHeader на CatalogView (каталог использует CatalogHeader); оставить для BookView
- [x] T033 [US7] Wire CatalogHeader into CatalogView — заменить placeholder-хедер на CatalogHeader; вкладки из CatalogHeader → catalog store setActiveTab; счётчик книг загружается при mount

**Checkpoint**: Хедер с пользовательским меню, вкладками навигации и счётчиком книг полностью функционален.

---

## Phase 8: User Story 4 — Цветовые схемы каталога (Priority: P2)

**Goal**: 4 цветовые схемы (Light, Dark, Sepia, Night), быстрый переключатель, диалог настроек, наследование темы читалкой, серверная синхронизация.

**Independent Test**: Переключить тему между всеми 4 схемами. Открыть книгу в читалке — тема наследуется. Изменить тему читалки отдельно, закрыть и открыть — сохраняется.

### Implementation for User Story 4

- [x] T034 [US4] Create `frontend/src/components/catalog/ThemeSwitcher.vue` — быстрый переключатель тем: 4 кнопки-превью (цветные кружки/квадраты с названием), клик → theme store setCatalogTheme → Vuetify useTheme().global.name; показывается в дропдаун-меню пользователя (FR-019)
- [x] T035 [US4] Create `frontend/src/components/catalog/SettingsDialog.vue` — v-dialog с настройками: секция «Тема каталога» (4 варианта с preview), секция «Тема читалки» (4 варианта + «Использовать тему каталога» как default); сохранение через theme store; закрытие по ESC/кнопке
- [x] T036 [US4] Integrate ThemeSwitcher into user dropdown in CatalogHeader — добавить секцию быстрого переключения тем в дропдаун-меню (между пунктами меню и «Выйти»); пункт «Настройки» открывает SettingsDialog
- [x] T037 [US4] Implement theme persistence in `frontend/src/stores/theme.ts` — loadSettings при app mount (GET /me/settings → extract catalog.theme, reader.theme); apply Vuetify theme; saveSettings с debounce 1000ms (PUT /me/settings); localStorage fallback для мгновенного применения
- [x] T038 [US4] Update reader theme inheritance — theme store effectiveReaderTheme (readerThemeOverride ?? catalogTheme) обеспечивает наследование; SettingsDialog позволяет выбрать «Тема каталога» для сброса
- [x] T039 [US4] Apply theme CSS to all catalog components — все компоненты используют Vuetify theme colors через CSS variables (rgb(var(--v-theme-*))); dark theme золотистый акцент (#d4a017) определён в vuetify.ts

### Tests for User Story 4

- [x] T040 [P] [US4] Unit tests for ThemeSwitcher in `frontend/src/components/catalog/__tests__/ThemeSwitcher.test.ts` — переключение темы, рендер 4 опций

**Checkpoint**: 4 темы переключаются корректно. Читалка наследует тему каталога. Настройки сохраняются на сервере.

---

## Phase 9: User Story 5 — Изменяемые размеры панелей (Priority: P3)

**Goal**: Перетаскивание разделителей между панелями с min/max ограничениями и сохранением размеров.

**Independent Test**: Перетащить вертикальный разделитель — ширина левой панели меняется. Обновить страницу — размеры сохранились.

### Implementation for User Story 5

- [x] T041 [US5] Create composable `frontend/src/composables/usePanelResize.ts` — load panel sizes from localStorage (instant) + server (GET /me/settings → catalog.panelSizes); handle @resized from Splitpanes → update localStorage immediately + debounced save to server (PUT /me/settings); default sizes: leftWidth=25, tableHeight=60
- [x] T042 [US5] Configure Splitpanes constraints in `frontend/src/views/CatalogView.vue` — vertical split: left pane min-size=10 max-size=50; horizontal split: top pane min-size=20 max-size=80; connect usePanelResize composable for size persistence; custom gutter styling matching current theme
- [x] T043 [US5] Style splitpanes gutters — стиль разделителей: 4px ширина, cursor col-resize/row-resize, цвет из текущей темы, hover-подсветка; override default splitpanes CSS для соответствия дизайну

### Tests for User Story 5

- [x] T044 [US5] Unit tests for usePanelResize in `frontend/src/__tests__/composables/usePanelResize.test.ts` — загрузка размеров из localStorage, сохранение, debounce, значения по умолчанию, clamp в пределах min/max

**Checkpoint**: Панели ресайзятся с ограничениями, размеры персистятся между сессиями.

---

## Phase 10: User Story 8 — Статус-бар (Priority: P3)

**Goal**: Статус-бар внизу экрана с контекстной информацией.

**Independent Test**: Выбрать автора — статус-бар показывает «Автор: [имя]» и «Показано книг: [N]».

### Implementation for User Story 8

- [x] T045 [US8] Create `frontend/src/components/catalog/StatusBar.vue` — compact bar (24-28px height) в самом низу CatalogView; слева: текущий контекст из catalog store (buildStatusText: «Автор: Азимов, Айзек» / «Серия: Основание» / «Жанр: Фантастика» / «Поиск: [запрос]» / «Готов»); справа: «Показано книг: N из M» (total из API response)
- [x] T046 [US8] Wire StatusBar into CatalogView — заменить placeholder на StatusBar; reactive binding к catalog store (activeFilter, books.length, total)

**Checkpoint**: Статус-бар отображает контекст и количество книг.

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: Удаление старого кода, edge cases, тесты, визуальная сверка с макетом, обновление документации

- [x] T047 Remove obsolete test files for deleted components in `frontend/src/` (BookCard.test.ts, BookFilters.test.ts, SearchBar.test.ts, PaginationBar.test.ts if exist)
- [x] T048 [P] Handle edge cases across all components — пустая библиотека (нет книг/авторов/серий): показывать приветственное сообщение; длинные названия: ellipsis во всех панелях; отсутствие аннотации: placeholder текст; ошибки API: user-friendly сообщения
- [x] T049 [P] Visual QA — сверка с макетом docs/design/myhomelib-catalog.jsx при теме Dark: золотистые акценты (#d4a017), шрифт Source Sans 3, моноширинный JetBrains Mono для размеров, compact layout; проверить все 4 темы на всех компонентах
- [x] T050 [P] Verify existing functionality preserved (SC-005) — проверить что /books/:id (BookView), /books/:id/read (ReaderView), /login (LoginView), /admin/import (AdminImportView) работают корректно после редизайна; убедиться что reading progress по-прежнему загружается
- [x] T051 Update `docs/homelib-architecture-v8.md` section 7 — добавить `components/catalog/` (12 компонентов), `stores/theme.ts`, `composables/usePanelResize.ts`, `types/catalog.ts`; отметить удалённые файлы (BookCard, BookFilters, SearchBar, PaginationBar, AuthorsView, GenresView, SeriesView)
- [x] T052 Run backend tests `cd backend && go test -race ./...` and frontend tests `cd frontend && npx vitest run` and build `cd frontend && npm run build` to verify no regressions
- [x] T053 Build and deploy to staging: `./scripts/build-and-push.sh --bump minor && ./scripts/deploy-stage.sh --tag <TAG>`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (T001 for splitpanes, T002 for types) — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Phase 2 — core layout must exist first
- **US2 (Phase 4)**: Depends on US1 (T014 BookTable exists) — adds sorting
- **US3 (Phase 5)**: Depends on US1 (T012 NavigationPanel exists) — adds remaining tabs
- **US6 (Phase 6)**: Depends on US1 (T015 BookDetailPanel exists) — enhances detail panel
- **US7 (Phase 7)**: Depends on US1 (CatalogView exists) — adds header; can parallel with US2/US3/US6
- **US4 (Phase 8)**: Depends on Phase 2 (T004+T005 themes exist) + US7 (T031 dropdown exists for ThemeSwitcher integration) — best after US7
- **US5 (Phase 9)**: Depends on US1 (Splitpanes in CatalogView) — adds persistence
- **US8 (Phase 10)**: Depends on US1 (CatalogView exists) — can parallel with most phases
- **Polish (Phase 11)**: Depends on ALL user stories being complete

### User Story Dependencies

```
Phase 1 (Setup)
  └─> Phase 2 (Foundational)
        ├─> US1 (Phase 3) ─┬─> US2 (Phase 4) ───────────────────┐
        │                   ├─> US3 (Phase 5) ───────────────────┤
        │                   ├─> US6 (Phase 6) ───────────────────┤
        │                   ├─> US5 (Phase 9) ───────────────────┤
        │                   └─> US8 (Phase 10) ──────────────────┤
        └─> US7 (Phase 7) ──────────────────────────────────────┤
              └─> US4 (Phase 8) ────────────────────────────────┤
                                                                 └─> Phase 11 (Polish)
```

### Parallel Opportunities

**Within Phase 1**: T002, T003 can run in parallel (different repos)
**Within Phase 2**: T005, T006, T009, T010 can run in parallel (different files)
**Within Phase 3**: T012, T013, T014, T015 can run in parallel (different component files); T017, T018 parallel after implementation
**Within Phase 5**: T021, T022, T023 can run in parallel (different tab components); T026, T027 parallel after implementation
**Across Phases**: After US1 completes, US2+US3+US6+US7+US5+US8 can theoretically all start (but US4 best after US7)

---

## Parallel Example: User Story 1

```bash
# Launch all independent component tasks together:
Task: "Create NavigationPanel.vue" (T012)
Task: "Create AuthorsTab.vue" (T013)
Task: "Create BookTable.vue" (T014)
Task: "Create BookDetailPanel.vue" (T015)

# Then wire up integration (depends on above):
Task: "Wire up CatalogView interaction flow" (T016)

# Then tests (parallel):
Task: "Unit tests for BookTable" (T017)
Task: "Unit tests for AuthorsTab + NavigationPanel" (T018)
```

## Parallel Example: User Story 3

```bash
# Launch all tab components together:
Task: "Create SeriesTab.vue" (T021)
Task: "Create GenresTab.vue" (T022)
Task: "Create SearchTab.vue" (T023)

# Then update NavigationPanel (depends on above):
Task: "Update NavigationPanel.vue" (T024)

# Then tests (parallel):
Task: "Unit tests for SeriesTab" (T026)
Task: "Unit tests for GenresTab + SearchTab" (T027)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Открыть каталог → выбрать автора → увидеть книги → кликнуть → увидеть детали
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. US1 → Базовый трёхпанельный каталог с авторами (MVP!)
3. US2 → Сортировка таблицы
4. US3 → Все 4 вкладки навигации
5. US6 → Полная панель деталей
6. US7 → Хедер с меню
7. US4 → 4 цветовые схемы
8. US5 → Ресайз панелей
9. US8 → Статус-бар
10. Polish → Visual QA, edge cases, arch doc, deploy

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story independently testable after completion
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Макет-референс: `docs/design/myhomelib-catalog.jsx`
- Тесты включены в каждую фазу для соответствия конституции §7 (≥80% покрытие)
- API не поддерживает сортировку по автору/серии/жанру — для этих колонок используется клиентская сортировка текущей страницы

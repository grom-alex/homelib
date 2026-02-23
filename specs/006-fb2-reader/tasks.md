# Tasks: Браузерная читалка FB2

**Input**: Design documents from `/specs/006-fb2-reader/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/
**Architecture**: docs/homelib-architecture-v8.md §8.1-8.10 — строгая спецификация

**Tests**: Включены — SC-007 требует ≥80% покрытия, конституция §7 требует TDD.

**Organization**: Задачи сгруппированы по user story для независимой реализации и тестирования.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Можно выполнять параллельно (разные файлы, нет зависимостей)
- **[Story]**: К какой user story относится задача (US1, US2, US3)
- Пути файлов указаны точно
- Ссылки §N.N указывают на раздел архитектуры v8

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Инициализация структуры проекта и подготовка конфигурации

- [x] T001 Add reader config section (cache_path per §8.4, reader settings) to backend/internal/config/config.go and backend/config.example.yaml
- [x] T002 [P] Create test FB2 files (simple book, book with poems/epigraphs/images/footnotes, single-section book, malformed XML) in backend/internal/bookfile/fb2_testdata/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Базовая инфраструктура, необходимая для ВСЕХ user story

**⚠️ CRITICAL**: Ни одна user story не может начаться до завершения этой фазы

- [x] T003 Create BookConverter interface, types (BookContent, BookMetadata, TOCEntry, ChapterContent), and GetConverter factory per §8.3 in backend/internal/bookfile/converter.go
- [x] T004 [P] Create TypeScript types per §8.5: BookContent, TOCEntry, ChapterContent, ReadingPosition, ReaderSettings (all 18 fields: fontSize, fontFamily, fontWeight, lineHeight, paragraphSpacing, letterSpacing, marginHorizontal, marginVertical, firstLineIndent, textAlign, hyphenation, theme, customColors, viewMode, pageAnimation, showProgress, showClock, tapZones) with defaultSettings in frontend/src/types/reader.ts
- [x] T005 [P] Create reader API client (getBookContent, getChapter, getBookImage, getProgress, saveProgress, getSettings, updateSettings) in frontend/src/api/reader.ts
- [x] T006 [P] Add route `/books/:id/read` → ReaderView to frontend/src/router/index.ts

**Checkpoint**: Фундамент готов — реализация user stories может начинаться

---

## Phase 3: User Story 1 — Открытие и чтение книги FB2 (Priority: P1) 🎯 MVP

**Goal**: Пользователь открывает книгу FB2 из каталога и читает её в браузере с постраничной навигацией, оглавлением и корректным отображением FB2-элементов (эпиграфы, стихи, цитаты, изображения).

**Independent Test**: Открыть любую FB2-книгу из каталога → текст отображается → перелистывание работает → оглавление показывает структуру.

**FR**: FR-001, FR-002, FR-003, FR-004, FR-005, FR-006, FR-007, FR-013, FR-014, FR-015, FR-017

### Backend — FB2 конвертер (§8.3)

- [x] T007 [US1] Create FB2 XML struct definitions (FictionBook, Description, TitleInfo, Body, Section, Paragraph, Poem, Stanza, Epigraph, Cite, Binary) and Parse() method per §8.3 in backend/internal/bookfile/fb2.go
- [x] T008 [US1] Implement GetChapter(), convertSection() with recursive nesting, convertPoem() (title, stanzas, verses, author), convertEpigraph() (paragraphs, author), convertCite(), convertInline() (emphasis→em, strong→strong, strikethrough→del, code→code, sup→sup, sub→sub per §8.3 tag mapping), image URL substitution (/api/books/{bookID}/image/{imageId} per R6), subtitle, empty-line→br, footnotes: `<a type="note">` → `<a class="footnote-ref" data-note-id="{href}">`, `<body name="notes">` sections → `<div class="footnote-body" id="{id}">` appended to chapter HTML in backend/internal/bookfile/fb2.go
- [x] T009 [US1] Write FB2 converter tests: valid parsing with metadata extraction, chapter extraction with correct IDs, poem/epigraph/cite HTML rendering with correct CSS classes, image reference substitution, inline tag mapping (emphasis→em etc.), footnote-ref and footnote-body rendering, malformed XML error handling, single-section book (no TOC), empty sections in backend/internal/bookfile/fb2_test.go

### Backend — ReaderService и кеш (§8.4)

- [x] T010 [US1] Implement file-based cache per §8.4: Get/Set for content.json, ch_{chapterID}.html, img_{id}.bin; path structure {cache_dir}/books/{bookID}/; no TTL (books are immutable) in backend/internal/service/reader.go
- [x] T011 [US1] Implement ReaderService per §8.4: GetBookContent (check cache → extract from archive via archive.ExtractFile → convert → cache), GetChapter (same flow per chapter), GetBookImage (extract binary from FB2, decode base64, cache as bin) in backend/internal/service/reader.go
- [x] T012 [US1] Write ReaderService tests (cache hit returns cached, cache miss triggers conversion, archive.ExtractFile integration, error propagation for missing book/unsupported format) in backend/internal/service/reader_test.go

### Backend — Handlers и маршруты (§8.2)

- [x] T013 [US1] Implement reader handlers per §8.2 contracts: GetBookContent (200 with metadata+toc+chapters, 404, 415 unsupported format, 422 malformed FB2), GetChapter (200 with id+title+html, 404), GetBookImage (200 with correct Content-Type + Cache-Control: public max-age=86400, 404) in backend/internal/api/handler/reader.go
- [x] T014 [US1] Write reader handler tests (success JSON responses matching contract schema, format validation → 415, missing book → 404, missing chapter → 404, malformed FB2 → 422, image Content-Type detection) in backend/internal/api/handler/reader_test.go
- [x] T015 [US1] Add reader routes (GET /api/books/:id/content, GET /api/books/:id/chapter/:chapterId, GET /api/books/:id/image/:imageId) to auth-protected group in backend/internal/api/router.go

### Frontend — CSS темы и типографика (§8.6)

- [x] T016 [P] [US1] Create CSS themes and typography per §8.6 in frontend/src/assets/styles/reader-themes.css:
  - 4 themes with 6 CSS variables each: --reader-bg, --reader-text, --reader-link, --reader-selection, --reader-header-bg, --reader-border
  - Light: bg=#ffffff, text=#1a1a1a, link=#2563eb, selection=#bfdbfe, header-bg=#f8fafc, border=#e2e8f0
  - Sepia: bg=#f5e6d3, text=#5c4b37, link=#8b5a2b, selection=#d4c4b0, header-bg=#ede0cf, border=#d4c4b0
  - Dark: bg=#1e1e1e, text=#d4d4d4, link=#60a5fa, selection=#374151, header-bg=#2d2d2d, border=#404040
  - Night: bg=#000000, text=#666666, link=#4a90d9, selection=#1a1a1a, header-bg=#0a0a0a, border=#1a1a1a
  - .reader-content: applies CSS vars for bg, color, font-size, font-family, line-height, text-align, padding
  - .reader-content p: text-indent var(--first-line-indent), margin 0 0 var(--paragraph-spacing), hyphens var(--hyphenation)
  - .reader-content p:first-child, h1+p, h2+p, h3+p: text-indent 0 (first paragraph without indent)
  - .reader-content h1,h2,h3: text-indent 0, margin 1.5em 0 0.5em, line-height 1.3
  - .reader-content a: color var(--reader-link), no underline
  - .reader-content ::selection: background var(--reader-selection)
  - FB2-specific: .epigraph (italic, margin 1.5em 10%, text-indent 0), .epigraph-author (block, text-align right, margin-top 0.5em), .poem (margin 1.5em 5%), .stanza (margin-bottom 1em), .verse (text-indent 0, margin 0), .poem-author (text-align right, italic, margin-top 1em), .subtitle (text-align center, italic, margin 1em 0), .cite (margin 1em 5%, padding-left 1em, border-left 3px solid var(--reader-border))
  - Footnotes: .footnote-ref (color var(--reader-link), cursor pointer, vertical-align super, font-size 0.8em), .footnote-body (display none — hidden by default, shown via JS popup), .footnote-popup (position absolute, background var(--reader-bg), border 1px solid var(--reader-border), border-radius 8px, padding 12px 16px, max-width 300px, box-shadow, z-index 100, font-size 0.9em)

### Frontend — Базовые модули

- [x] T017 [P] [US1] Create Pinia reader store (bookContent, currentChapterId, currentChapterContent, currentPage, totalPages, loading, error, tocVisible, uiVisible) in frontend/src/stores/reader.ts
- [x] T018 [P] [US1] Implement useBookContent composable (loadBookContent, loadChapter, navigateToChapter, prefetch adjacent chapters, network error handling: show user-friendly message on fetch failure for chapter transitions per EC-5) in frontend/src/composables/useBookContent.ts
- [x] T019 [P] [US1] Implement usePagination composable per R2: CSS multi-column layout (column-width: 100%, column-gap, column-fill: auto, height: calc(100vh - header - footer), overflow: hidden), calculateTotalPages via scrollWidth/columnWidth, nextPage/prevPage via translateX, goToPage, proportional recalculation on resize/settings change in frontend/src/composables/usePagination.ts
- [x] T020 [P] [US1] Implement useReaderKeyboard composable per §8.9: →/Space/PageDown=next, ←/PageUp=prev, Home=start, End=end, T=TOC, F=fullscreen, +/-=font size, N=cycle theme, Esc=close panels or exit reader in frontend/src/composables/useReaderKeyboard.ts
- [x] T021 [P] [US1] Implement useReaderGestures composable per §8.8: touchstart/touchend events, horizontal swipe threshold 50px (deltaX > deltaY), swipe right=prev / swipe left=next; tap zone detection using settings.tapZones — 'lrc': left 25% prev, center 50% toggleUI, right 25% next; 'lr': left 40% prev, right 60% next in frontend/src/composables/useReaderGestures.ts

### Frontend — Компоненты читалки (§8.7)

- [x] T022 [US1] Create ReaderContent.vue component: content area with CSS multi-column pagination per R2, translateX page switching with pageAnimation setting (slide/fade/none), v-html chapter content, applies reader-content CSS class for typography per §8.6; footnote popup: click handler on .footnote-ref → find .footnote-body by data-note-id → show in positioned .footnote-popup tooltip, close on outside click/Esc in frontend/src/components/reader/ReaderContent.vue
- [x] T023 [P] [US1] Create ReaderHeader.vue component: book title, chapter title, TOC toggle button, settings button, back-to-catalog button; uses --reader-header-bg and --reader-border CSS variables per §8.6 in frontend/src/components/reader/ReaderHeader.vue
- [x] T024 [P] [US1] Create ReaderFooter.vue component: page X of Y, chapter progress bar; visibility controlled by showProgress setting; optional clock display controlled by showClock setting per §8.5 in frontend/src/components/reader/ReaderFooter.vue
- [x] T025 [US1] Create ReaderTOC.vue component per §8.7: sidebar drawer with hierarchical chapter list (indentation by TOCEntry.level), current chapter highlight, click-to-navigate (loads chapter + resets page), close on selection in frontend/src/components/reader/ReaderTOC.vue
- [x] T026 [US1] Create BookReader.vue main container per §8.7: assembles header, content, footer, TOC; wires useBookContent, usePagination, useReaderKeyboard, useReaderGestures; fullscreen layout; toggleUI on center tap; theme class on root element (.reader.theme-{name}) in frontend/src/components/reader/BookReader.vue
- [x] T027 [US1] Create ReaderView.vue page wrapper: route param parsing (:id), loadBookContent call, loading spinner, error states (404 book not found, 415 unsupported format, 422 malformed file with user-friendly messages per FR-017), BookReader mount in frontend/src/views/ReaderView.vue

### Frontend — Интеграция с каталогом

- [x] T028 [US1] Add «Читать» button (visible only for format=fb2 per FR-015) with router-link to /books/:id/read in frontend/src/components/common/BookCard.vue

**Checkpoint**: User Story 1 полностью функциональна — можно открыть и прочитать FB2-книгу от начала до конца

---

## Phase 4: User Story 2 — Сохранение и восстановление прогресса (Priority: P2)

**Goal**: Прогресс чтения автоматически сохраняется и восстанавливается при повторном открытии книги. На карточке книги в каталоге отображается процент прочитанного.

**Independent Test**: Открыть книгу → прочитать несколько страниц → закрыть вкладку → снова открыть → читалка на последнем месте. Карточка показывает процент.

**FR**: FR-008, FR-009, FR-016

### Backend — Миграция и модель

- [x] T029 [P] [US2] Create SQL migration files per data-model.md: CREATE TABLE reading_progress (id BIGSERIAL PK, user_id UUID NOT NULL FK→users ON DELETE CASCADE, book_id BIGINT NOT NULL FK→books ON DELETE CASCADE, chapter_id TEXT NOT NULL DEFAULT '', chapter_progress SMALLINT NOT NULL DEFAULT 0 CHECK 0-100, total_progress SMALLINT NOT NULL DEFAULT 0 CHECK 0-100, device TEXT DEFAULT '', updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), UNIQUE(user_id, book_id)); CREATE INDEX idx_reading_progress_user ON reading_progress(user_id) in backend/migrations/003_reading_progress.up.sql and DROP TABLE IF EXISTS reading_progress in backend/migrations/003_reading_progress.down.sql
- [x] T030 [P] [US2] Create ReadingProgress Go model struct with JSON tags per data-model.md in backend/internal/models/reading_progress.go

### Backend — Репозиторий

- [x] T031 [US2] Implement ReadingProgressRepo: Get(ctx, userID, bookID) → ReadingProgress/nil, Upsert(ctx, progress) with ON CONFLICT (user_id, book_id) DO UPDATE per R4, GetByUser(ctx, userID) → []ReadingProgress for catalog progress display in backend/internal/repository/reading_progress.go
- [x] T032 [US2] Write ReadingProgressRepo tests (insert new, upsert existing updates chapter_id/progress/updated_at, get returns record, get nonexistent returns nil, get by user returns list) in backend/internal/repository/reading_progress_test.go

### Backend — Handlers

- [x] T033 [US2] Implement progress handlers per contracts/progress-api.md: GetReadingProgress (200 with chapterId+chapterProgress+totalProgress+device+updatedAt, 204 No Content if no progress), SaveReadingProgress (validate chapterProgress 0-100, totalProgress 0-100, chapterId not empty → 200 with saved record, 400 on validation failure) in backend/internal/api/handler/reader.go
- [x] T034 [US2] Write progress handler tests (save new → 200, update existing → 200, get existing → 200, get nonexistent → 204, invalid progress range → 400, empty chapterId → 400) in backend/internal/api/handler/reader_test.go
- [x] T035 [US2] Add progress routes (GET/PUT /api/me/books/:bookId/progress) to auth-protected group in backend/internal/api/router.go

### Frontend — Прогресс (§8.10)

- [x] T036 [US2] Implement useReadingProgress composable per §8.10: loadProgress from API on book open, saveProgress with useDebounceFn 2000ms, calculateTotalProgress (chapterIndex/totalChapters * 100 + chapterProgress/totalChapters), getDeviceType() → 'desktop'|'tablet'|'mobile', save on beforeunload in frontend/src/composables/useReadingProgress.ts
- [x] T037 [US2] Integrate progress into BookReader.vue: call loadProgress on mount → navigate to saved chapter + restore page position, call updatePosition on every page turn and chapter change, wire saveProgress debounce, save on window beforeunload in frontend/src/components/reader/BookReader.vue
- [x] T038 [US2] Add reading progress indicator (% bar) to book card for books with saved progress in frontend/src/components/common/BookCard.vue

**Checkpoint**: User Story 1 + 2 функциональны — прогресс сохраняется и восстанавливается

---

## Phase 5: User Story 3 — Настройка внешнего вида (Priority: P3)

**Goal**: Пользователь настраивает шрифт, тему, интервалы и поля per §8.5. Настройки применяются мгновенно и сохраняются между сессиями.

**Independent Test**: Открыть книгу → изменить шрифт и тему → закрыть → открыть другую книгу → настройки сохранены.

**FR**: FR-010, FR-011, FR-012

### Backend — Настройки пользователя

- [x] T039 [P] [US3] Add GetSettings(ctx, userID) → JSONB and UpdateSettings(ctx, userID, settings) with JSONB merge (settings || $2) methods to user repository in backend/internal/repository/user.go
- [x] T040 [US3] Implement settings handlers per contracts/progress-api.md: GetUserSettings → 200 (full settings or {}), UpdateUserSettings → partial merge via JSONB → 200 (full merged settings) in backend/internal/api/handler/reader.go
- [x] T041 [US3] Write settings handler tests (get empty → 200 {}, get existing → 200 full, partial update merges correctly, full response after merge contains all 18 fields) in backend/internal/api/handler/reader_test.go
- [x] T042 [US3] Add settings routes (GET/PUT /api/me/settings) to auth-protected group in backend/internal/api/router.go

### Frontend — Настройки (§8.5)

- [x] T043 [US3] Implement useReaderSettings composable per §8.5: loadSettings from GET /api/me/settings, mergeWithDefaults (all 18 fields from defaultSettings), applySettings by setting CSS custom properties on reader element (--font-size, --font-family, --font-weight, --line-height, --paragraph-spacing, --letter-spacing, --margin-h, --margin-v, --first-line-indent, --text-align, --hyphenation), saveSettings with debounce via PUT /api/me/settings, theme class switching (.theme-sepia, .theme-dark, .theme-night, .theme-custom), custom colors application in frontend/src/composables/useReaderSettings.ts
- [x] T044 [US3] Create ReaderSettings.vue component per §8.5 with controls for all settings: font size ± buttons (12-36), fontWeight toggle (400/500), letterSpacing slider (-0.05 — 0.1), line height slider (1.0-2.5), paragraph spacing slider (0-2), margin sliders (H: 0-20%, V: 0-10%), first line indent slider (0-3), text align toggle (left/justify), hyphenation toggle, 5 theme buttons (light/sepia/dark/night/custom) with color previews, custom color pickers (when theme=custom), view mode toggle (paginated/scroll), page animation select (slide/fade/none), showProgress toggle, showClock toggle, tapZones select (lr/lrc) in frontend/src/components/reader/ReaderSettings.vue
- [x] T045 [P] [US3] Create ReaderFontPicker.vue subcomponent per §8.7: font family selector with preview (Georgia, PT Serif, Literata, OpenDyslexic, System), each option rendered in its own font in frontend/src/components/reader/ReaderFontPicker.vue
- [x] T046 [US3] Integrate settings into BookReader.vue: settings modal toggle from header button, apply all CSS variables from useReaderSettings to .reader-content element, recalculate pagination on any setting change via usePagination, preserve reading position proportionally on recalculation per R2 in frontend/src/components/reader/BookReader.vue

**Checkpoint**: Все три user story функциональны и независимо тестируемы

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Тесты, покрытие, верификация, документация

- [x] T047 Run all backend tests with coverage (`go test -race -coverprofile=coverage.out ./internal/bookfile/... ./internal/service/... ./internal/api/handler/... ./internal/repository/...`) and ensure ≥80% per package
- [x] T048 [P] Run all frontend tests with coverage (`vitest --coverage`) for reader components and composables, ensure ≥80%
- [x] T049 Run quickstart.md verification scenarios (all 4 scenarios + curl API checks) end-to-end; verify SC-002: measure time from «Читать» click to first page render — must be under 3 seconds for a typical book
- [x] T050 [P] Update docs/homelib-architecture-v8.md section 7 file tree to reflect new bookfile/, reader/ directories and files

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Нет зависимостей — можно начинать сразу
- **Foundational (Phase 2)**: Зависит от Phase 1 — БЛОКИРУЕТ все user stories
- **US1 (Phase 3)**: Зависит от Phase 2 — MVP, реализовать первой
- **US2 (Phase 4)**: Зависит от Phase 2, интегрируется с US1 (BookReader.vue)
- **US3 (Phase 5)**: Зависит от Phase 2, интегрируется с US1 (BookReader.vue, ReaderContent.vue)
- **Polish (Phase 6)**: Зависит от завершения всех желаемых user stories

### User Story Dependencies

- **US1 (P1)**: Может начаться после Phase 2. Независима от других stories. **MVP = Phase 1 + 2 + 3**
- **US2 (P2)**: Может начаться после Phase 2. Backend (T029-T035) параллелен с US1. Frontend (T036-T038) интегрируется в BookReader.vue из US1
- **US3 (P3)**: Может начаться после Phase 2. Backend (T039-T042) параллелен с US1/US2. Frontend (T043-T046) интегрируется в BookReader.vue из US1

### Within Each User Story

- Тесты конвертера пишутся после реализации (код + тесты в одной итерации)
- Модели → сервисы → хэндлеры → маршруты (последовательно)
- Composables → компоненты → интеграция (последовательно)
- Backend и frontend одной story можно делать параллельно

### Parallel Opportunities

**Backend параллельно с Frontend внутри каждой story:**
```
US1 Backend (T007-T015) ║ US1 Frontend (T016-T028)
US2 Backend (T029-T035) ║ US2 Frontend (T036-T038)
US3 Backend (T039-T042) ║ US3 Frontend (T043-T046)
```

**Backend US2/US3 параллельно с US1:**
```
US1 (T007-T028)
  ║ US2 Backend (T029-T035) — не зависит от US1 backend
  ║ US3 Backend (T039-T042) — не зависит от US1 backend
```

---

## Parallel Example: User Story 1

```bash
# Backend: FB2 конвертер (последовательно)
T007 → T008 → T009 (struct → methods → tests)

# Backend: Service + Handler (последовательно)
T010 → T011 → T012 (cache → service → tests)
T013 → T014 → T015 (handler → tests → routes)

# Frontend: CSS + composables параллельно
T016 ║ T017 ║ T018 ║ T019 ║ T020 ║ T021

# Frontend: компоненты (после composables)
T022 (зависит от T019, T016)
T023 ║ T024 (параллельно)
T025 (зависит от T017)
T026 → T027 → T028 (container → page → catalog integration)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003-T006)
3. Complete Phase 3: User Story 1 (T007-T028)
4. **STOP and VALIDATE**: Открыть FB2-книгу, прочитать несколько страниц, проверить оглавление
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → фундамент готов
2. Add US1 → тест → деплой (**MVP! Книги можно читать**)
3. Add US2 → тест → деплой (прогресс сохраняется)
4. Add US3 → тест → деплой (настройки работают)
5. Polish → финальное тестирование, покрытие ≥80%

---

## Architecture Compliance Notes

Все задачи строго соответствуют архитектуре v8:
- **§8.2**: API endpoints — contracts/reader-api.md и contracts/progress-api.md
- **§8.3**: BookConverter interface, FB2 tag mapping, convertSection/Poem/Epigraph
- **§8.4**: File-based cache structure {cache_dir}/books/{bookID}/
- **§8.5**: ReaderSettings — все 18 полей с defaultSettings
- **§8.6**: CSS темы — 4 темы × 6 CSS переменных, полная типографика, FB2-элементы
- **§8.7**: Компоненты — BookReader, ReaderContent, ReaderHeader, ReaderFooter, ReaderSettings, ReaderTOC, ReaderFontPicker
- **§8.8**: Жесты — touchstart/touchend, swipe 50px threshold, tapZones lr/lrc
- **§8.9**: Клавиатура — все 11 shortcuts (→/←/Space/PageUp/PageDown/Home/End/T/F/+/-/N/Esc)
- **§8.10**: Прогресс — debounce 2s, getDeviceType, loadProgress/saveProgress

**Deferred to future iterations** (в архитектуре §8.7, но вне scope spec.md):
- ReaderBookmarks.vue, ReaderSearch.vue — нет соответствующих user stories
- useTextSelection.ts — зависит от закладок
- Клавиши B (закладки), S (поиск) — зависят от соответствующих компонентов

---

## Notes

- [P] задачи = разные файлы, нет зависимостей
- [Story] привязывает задачу к конкретной user story
- Backend и frontend одной story можно делать параллельно
- Коммит после каждой логической группы задач
- Остановка на любом checkpoint для независимой верификации story
- FB2 тест-файлы (T002) нужны до начала T007 — обеспечить в Phase 1
- CSS значения (hex-цвета тем, отступы FB2-элементов) взяты из §8.6 дословно

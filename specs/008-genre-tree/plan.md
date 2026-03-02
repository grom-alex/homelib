# Implementation Plan: Древовидная структура жанров

**Branch**: `008-genre-tree` | **Date**: 2026-03-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/008-genre-tree/spec.md`

## Summary

Внедрение иерархической структуры жанров на основе файла `genres_all.glst` (448 жанров, 4 уровня вложенности). Ключевые изменения:

- **Backend**: новый парсер `.glst` (`internal/glst/`), миграция БД (позиция в дереве, отмена уникальности кода), автозагрузка дерева при старте, каскадная фильтрация книг по materialized path, CLI-команда перезагрузки
- **Frontend**: замена плоского списка жанров на `VTreeview` (Vuetify 3), поиск по дереву, выпадающий фильтр жанров на вкладке «Поиск»
- **Данные**: ремаппинг существующих book↔genre связей через INPX-коды

## Technical Context

**Language/Version**: Go 1.25.6 (backend), TypeScript 5.7 + Vue 3.5 (frontend)
**Primary Dependencies**: Gin 1.11, pgx/v5 5.8, Vuetify 3.8, Pinia 3.0, Vitest 3.1
**Storage**: PostgreSQL 17 + pg_trgm, tsvector; файл `.glst` встраивается через `//go:embed`
**Testing**: `go test -race` + testify/pgxmock (backend, порог 80% per new package, 55%+ overall), vitest + @vue/test-utils (frontend, порог 80% lines / 70% branches)
**Target Platform**: Linux server (Docker), SPA в браузере
**Project Type**: Web application (Go backend + Vue 3 SPA)
**Performance Goals**: Загрузка 448 жанров < 1 сек, рендеринг дерева без заметной задержки, каскадная фильтрация < 200ms
**Constraints**: Self-hosted, единственная БД — PostgreSQL, Docker Compose
**Scale/Scope**: ~600K книг, 448 жанров (27 корневых), 4 уровня глубины, единицы пользователей

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Принцип | Статус | Комментарий |
|---|---------|--------|-------------|
| §1.II | Разделение ответственности | OK | Парсер `.glst` — отдельный пакет `internal/glst`; загрузка — service; хранение — repository; UI — frontend |
| §1.VI | Идемпотентный импорт | OK | FR-013: повторная загрузка `.glst` не создаёт дубликатов (UPSERT по position) |
| §1.VII | Нет внешних SaaS | OK | Нет внешних зависимостей |
| §2.I | Единая PostgreSQL | OK | Все данные жанров в PostgreSQL |
| §2.II | Явные M:N связи | OK | `book_genres` сохраняется как связующая таблица |
| §2.V | Запрет дублирования вне БД | OK | Файл `.glst` — источник, но данные загружаются в БД; runtime работает только с БД |
| §2.VI | Инкрементальный импорт | OK | INPX-импорт продолжает использовать upsert; GLST — idempotent reload |
| §5.I | Batch-операции | OK | Загрузка GLST одной транзакцией; ремаппинг book_genres через batch |
| §6.I | Docker Compose | OK | Нет новых сервисов |
| §6.V | Конфигурация через YAML | OK | Путь к `.glst` добавляется в config.yaml |
| §6.VII | Архитектурная документация | OK | Обновление `homelib-architecture-v8.md` включено в задачи |
| §7 prod | Production-качество | OK | Полная обработка edge cases, тесты ≥ 80% |
| §7 test | TDD | OK | Тесты для парсера, сервиса, API-хэндлеров, компонентов |
| §7 compliance | Constitution check | OK | Данный раздел |

**Результат**: все gate-ы пройдены, нарушений нет.

## Project Structure

### Documentation (this feature)

```text
specs/008-genre-tree/
├── plan.md              # Этот файл
├── research.md          # Исследование (Phase 0)
├── data-model.md        # Модель данных (Phase 1)
├── quickstart.md        # Инструкция по запуску (Phase 1)
├── contracts/           # API-контракты (Phase 1)
│   └── genres-api.md
└── tasks.md             # Задачи (Phase 2 — /speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   ├── api/main.go                          # [MOD] автозагрузка дерева жанров при старте
│   └── worker/main.go                       # [MOD] флаг --reload-genres
├── internal/
│   ├── glst/                                # [NEW] парсер .glst файлов
│   │   ├── parser.go                        # ParseFile(), ParseReader()
│   │   ├── parser_test.go
│   │   ├── types.go                         # GenreEntry, ParseError
│   │   ├── embed.go                         # [NEW] //go:embed genres_all.glst
│   │   └── genres_all.glst                  # [EMBED] копия docs/genres_all.glst
│   ├── models/
│   │   └── genre.go                         # [MOD] + Position, SortOrder
│   ├── repository/
│   │   └── genre.go                         # [MOD] LoadTree, GetIDsByCodes, GetDescendantIDs
│   ├── service/
│   │   ├── genre_tree.go                    # [NEW] GenreTreeService: Load, Reload, Remap
│   │   ├── genre_tree_test.go               # [NEW]
│   │   ├── catalog.go                       # Без изменений — thin proxy, изменения в repo протекают автоматически
│   │   └── import.go                        # [MOD] маппинг через GetIDsByCodes + fallback
│   ├── api/
│   │   ├── handler/
│   │   │   ├── genres.go                    # [MOD] ListGenres возвращает позиционное дерево
│   │   │   └── admin.go                     # [MOD] ReloadGenres endpoint
│   │   ├── router.go                        # [MOD] POST /api/admin/genres/reload
│   │   └── server.go                        # [MOD] инициализация GenreTreeService
│   └── config/
│       └── config.go                        # [MOD] + GenreTreeConfig{FilePath}
├── migrations/
│   ├── 006_genre_tree.up.sql                # [NEW] ALTER genres + app_metadata
│   └── 006_genre_tree.down.sql              # [NEW]

frontend/
├── src/
│   ├── api/
│   │   └── books.ts                         # [MOD] GenreTreeItem + position, admin reload
│   ├── components/
│   │   └── catalog/
│   │       ├── GenresTab.vue                # [REWRITE] VTreeview + поиск + счётчики
│   │       └── SearchTab.vue                # [MOD] выпадающий VTreeview для фильтра жанра
│   ├── stores/
│   │   └── catalog.ts                       # [MOD] каскадная навигация по жанру
│   └── types/
│       └── catalog.ts                       # [MOD] GenreTreeItem type update
└── tests (в __tests__/ рядом с компонентами)
```

**Structure Decision**: Web application (Option 2). Все новые файлы соответствуют структуре из `docs/homelib-architecture-v8.md`, раздел 7. Новый пакет `internal/glst/` аналогичен существующему `internal/inpx/` — отдельный парсер формата.

## Complexity Tracking

> Нарушений конституции нет. Таблица пуста.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| §2.VI: Полный remap book_genres при загрузке GLST | Дубликаты кодов (13 шт.) делают инкрементальный remap ненадёжным: нельзя определить правильные genre_id без полного пересчёта code→[]id. Операция выполняется ТОЛЬКО при изменении хеша .glst файла (редкое событие). | Инкрементальный remap по diff — отвергнут: сложность O(n²) для сопоставления старых/новых кодов при дубликатах, с непропорциональным ростом кода без выигрыша (448 жанров, remap < 5 минут для 600K книг) |

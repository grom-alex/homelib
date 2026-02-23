# Implementation Plan: Редизайн каталога в стиле MyHomeLib

**Branch**: `007-catalog-redesign` | **Date**: 2026-02-23 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/007-catalog-redesign/spec.md`

## Summary

Полная переработка UI каталога из карточного представления в трёхпанельный интерфейс в стиле MyHomeLib: левая навигационная панель (авторы/серии/жанры/поиск), таблица книг с сортировкой, панель деталей. Система из 4 цветовых схем (Light, Dark, Sepia, Night) с наследованием читалкой. Преимущественно фронтенд-фича: существующие API-эндпоинты покрывают все потребности, изменения бэкенда минимальны (расширение whitelist настроек).

## Technical Context

**Language/Version**: Go 1.25 (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Gin (HTTP), pgx/v5 (DB), Vuetify 3 (UI), Pinia (state), Vue Router, axios
**Storage**: PostgreSQL 17 (JSONB `users.settings` для тем)
**Testing**: `go test -race ./...` (backend), `vitest` (frontend)
**Target Platform**: Web, desktop ≥1280px
**Project Type**: Web application (frontend + backend)
**Performance Goals**: Мгновенное переключение вкладок; таблица 100+ строк без задержек прокрутки
**Constraints**: Нет мобильной адаптации; CSS variables для тем (мгновенное переключение без перезагрузки)
**Scale/Scope**: ~600K книг, единицы пользователей, ~15 новых/изменённых Vue-компонентов

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Статус | Примечание |
|---------|--------|------------|
| §1.II Разделение ответственности | ✅ PASS | Фронтенд-фича, бэкенд минимально затрагивается |
| §1.III Stateless API | ✅ PASS | Тема хранится в DB через `PUT /me/settings`, не в сессии |
| §2.I PostgreSQL единый источник | ✅ PASS | Настройки темы в JSONB колонке `users.settings` |
| §2.V Запрет дублирования бизнес-данных | ✅ PASS | Тема кешируется в Pinia store (в памяти), source of truth — DB |
| §4.I JWT-аутентификация | ✅ PASS | Используется существующая auth-система |
| §4.III Изоляция пользовательских данных | ✅ PASS | Settings привязаны к user_id из JWT |
| §6.I Docker Compose | ✅ PASS | Без новых сервисов |
| §6.VII Архитектурная документация | ⚠️ NOTE | Новые компоненты должны соответствовать структуре из arch doc |
| §7 Тестирование ≥80% | ✅ PASS | Все новые компоненты/composables/stores покрываются тестами |
| §7 TDD | ✅ PASS | Рекомендуется для новых composables и store |
| §7 Стадия проекта (production) | ✅ PASS | Полная обработка edge cases, сохранение настроек между сессиями |

**Результат**: Нарушений конституции нет. Все принципы соблюдаются.

## Project Structure

### Documentation (this feature)

```text
specs/007-catalog-redesign/
├── plan.md              # This file
├── research.md          # Phase 0: исследование и решения
├── data-model.md        # Phase 1: модель данных
├── quickstart.md        # Phase 1: быстрый старт разработки
├── contracts/           # Phase 1: API-контракты
└── tasks.md             # Phase 2: задачи (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── api/
│   │   └── handler/
│   │       └── settings.go          # Расширение whitelist: добавить "catalog" ключ
│   └── models/                       # Без изменений
└── migrations/                       # Без новых миграций

frontend/
├── src/
│   ├── views/
│   │   └── CatalogView.vue          # Полная переработка: трёхпанельный layout
│   ├── components/
│   │   ├── catalog/                  # НОВАЯ директория
│   │   │   ├── CatalogHeader.vue    # Хедер с вкладками и меню
│   │   │   ├── NavigationPanel.vue  # Левая панель навигации
│   │   │   ├── AuthorsTab.vue       # Вкладка «Авторы»
│   │   │   ├── SeriesTab.vue        # Вкладка «Серии»
│   │   │   ├── GenresTab.vue        # Вкладка «Жанры»
│   │   │   ├── SearchTab.vue        # Вкладка «Поиск»
│   │   │   ├── BookTable.vue        # Таблица книг
│   │   │   ├── BookDetailPanel.vue  # Панель деталей
│   │   │   ├── PanelResizer.vue     # Разделитель панелей
│   │   │   ├── StatusBar.vue        # Статус-бар
│   │   │   ├── ThemeSwitcher.vue    # Быстрый переключатель тем
│   │   │   └── SettingsDialog.vue   # Диалог настроек (темы каталога и читалки)
│   │   ├── common/
│   │   │   └── AppHeader.vue        # Обновление: пользовательское дропдаун-меню
│   │   └── reader/                   # Минимальные изменения (наследование темы)
│   ├── stores/
│   │   ├── catalog.ts               # Обновление: навигация вкладками, выбор книги
│   │   └── theme.ts                 # НОВЫЙ: управление темами каталога и читалки
│   ├── composables/
│   │   ├── usePanelResize.ts        # НОВЫЙ: логика изменения размеров панелей
│   │   └── useReaderSettings.ts     # Обновление: интеграция с темой каталога
│   ├── assets/
│   │   └── styles/
│   │       ├── catalog-themes.css   # НОВЫЙ: CSS-переменные для 4 тем каталога
│   │       └── reader-themes.css    # Без изменений (уже есть 4 темы)
│   └── types/
│       └── catalog.ts               # НОВЫЙ: типы для каталога (Theme, TabType, etc.)
└── package.json
```

**Structure Decision**: Web application (Option 2). Фронтенд-центричная фича с минимальными изменениями бэкенда. Новые компоненты каталога выделяются в отдельную директорию `components/catalog/` по аналогии с `components/reader/`. Соответствует структуре из `docs/homelib-architecture-v8.md` раздел 7.

## Complexity Tracking

> Нарушений конституции нет — таблица не заполняется.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| §7 Заглушки в FR-013 | Пункты «Мой профиль», «Мои коллекции», «Загрузить книги» — функционал future features, не входит в scope 007-catalog-redesign | Реализовать все пункты в полном объёме — чрезмерное расширение scope; каждый пункт — отдельная фича |

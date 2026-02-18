# Implementation Plan: Браузерная читалка FB2

**Branch**: `006-fb2-reader` | **Date**: 2026-02-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/006-fb2-reader/spec.md`

## Summary

Реализация браузерной читалки для книг формата FB2. Бэкенд конвертирует FB2 (XML) в унифицированный HTML, разбивая на главы. Фронтенд отображает HTML с CSS-пагинацией (multi-column), поддерживает навигацию (клавиатура, свайпы, тапы), оглавление, 4 цветовые темы, настройки шрифта/интервалов. Прогресс чтения автоматически сохраняется с debounce и восстанавливается при повторном открытии.

## Technical Context

**Language/Version**: Go 1.25 (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Gin (HTTP), pgx/v5 (DB), encoding/xml (FB2 parsing), axios (HTTP client), Pinia (state), Vue Router, Vuetify 3
**Storage**: PostgreSQL 17 (reading_progress, user settings), файловый кеш (конвертированный HTML)
**Testing**: `go test -race` (backend), `vitest` (frontend)
**Target Platform**: Docker Compose, Linux server
**Project Type**: web (backend + frontend SPA)
**Performance Goals**: Открытие книги < 3 сек, мгновенное применение настроек
**Constraints**: Домашняя библиотека, единицы пользователей. Книги в ZIP-архивах (read-only). Только FB2 на первом этапе.
**Scale/Scope**: ~600K книг в каталоге, ~30% в формате FB2

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Статус | Комментарий |
|---------|--------|-------------|
| §1.I Централизация на сервере | ✅ PASS | Конвертация FB2→HTML на бэкенде |
| §1.II Разделение ответственности | ✅ PASS | API Server — конвертация и отдача контента; фронтенд — отображение |
| §1.III Stateless API | ✅ PASS | Кеш конвертации — файловый (допускается §2.V), не session state |
| §1.IV Чтение из архивов без распаковки | ✅ PASS | Используем существующий archive.ExtractFile() |
| §2.I Единая PostgreSQL | ✅ PASS | reading_progress и settings хранятся в PostgreSQL |
| §2.V Запрет дублирования бизнес-данных | ✅ PASS | Кеш конвертации — производительность чтения, регенерируемый |
| §4.I JWT-аутентификация | ✅ PASS | Все эндпоинты читалки — авторизованные |
| §4.III Изоляция пользовательских данных | ✅ PASS | reading_progress фильтруется по user_id из JWT |
| §6.V Конфигурация через YAML | ✅ PASS | Путь к кешу — в конфиге |
| §6.VII Архитектурная документация | ⚠️ NOTE | Потребуется обновление раздела 7 arch doc после реализации |
| §7 Тестирование ≥80% | ✅ PASS | Тесты для конвертера, хэндлеров, компонентов |
| §7 TDD | ✅ PASS | Рекомендуется для нового функционала |

**Нарушения**: Нет. Все изменения соответствуют конституции.

## Project Structure

### Documentation (this feature)

```text
specs/006-fb2-reader/
├── spec.md              # Спецификация
├── plan.md              # Этот файл
├── research.md          # Исследование технических решений
├── data-model.md        # Модель данных
├── quickstart.md        # Шаги верификации
├── contracts/           # API контракты
│   ├── reader-api.md    # Эндпоинты читалки
│   └── progress-api.md  # Эндпоинты прогресса
├── checklists/
│   └── requirements.md  # Чеклист качества спецификации
└── tasks.md             # Будет создан /speckit.tasks
```

### Source Code (новые и изменяемые файлы)

```text
# Backend — конвертер FB2 и API читалки
backend/
├── internal/
│   ├── bookfile/                     # НОВАЯ директория — конвертеры форматов
│   │   ├── converter.go              # Интерфейс BookConverter, фабрика, типы
│   │   ├── converter_test.go         # Тесты интерфейса
│   │   ├── fb2.go                    # FB2Converter: Parse, GetChapter, Search
│   │   ├── fb2_test.go              # Тесты FB2-конвертера (парсинг, edge cases)
│   │   └── fb2_testdata/            # Тестовые FB2-файлы
│   ├── service/
│   │   ├── reader.go                 # НОВЫЙ — ReaderService (GetContent, GetChapter, кеш)
│   │   └── reader_test.go           # Тесты ReaderService
│   ├── api/handler/
│   │   ├── reader.go                 # НОВЫЙ — GetBookContent, GetChapter, GetBookImage
│   │   └── reader_test.go           # Тесты хэндлеров читалки
│   ├── repository/
│   │   ├── reading_progress.go       # НОВЫЙ — ReadingProgressRepo (CRUD)
│   │   └── reading_progress_test.go  # Тесты репозитория
│   │   └── user.go                   # ИЗМЕНЕНИЕ — добавить GetSettings, UpdateSettings
│   ├── models/
│   │   └── reading_progress.go       # НОВЫЙ — ReadingProgress, ReaderSettings
│   └── api/
│       └── router.go                 # ИЗМЕНЕНИЕ — добавить маршруты читалки
├── migrations/
│   ├── 003_reading_progress.up.sql   # НОВЫЙ — таблица reading_progress
│   └── 003_reading_progress.down.sql # НОВЫЙ — откат

# Frontend — UI читалки
frontend/src/
├── api/
│   └── reader.ts                     # НОВЫЙ — API-клиент читалки
├── views/
│   └── ReaderView.vue                # НОВЫЙ — страница читалки (обёртка)
├── components/
│   ├── common/
│   │   └── BookCard.vue              # ИЗМЕНЕНИЕ — кнопка «Читать» для FB2
│   └── reader/                       # НОВАЯ директория — компоненты читалки (§8.7)
│       ├── BookReader.vue            # Главный контейнер
│       ├── ReaderContent.vue         # Область контента (пагинация/скролл)
│       ├── ReaderHeader.vue          # Верхняя панель (название, кнопки)
│       ├── ReaderFooter.vue          # Прогресс-бар, номер страницы
│       ├── ReaderSettings.vue        # Настройки (модальное окно)
│       ├── ReaderTOC.vue             # Оглавление (боковая панель)
│       └── ReaderFontPicker.vue      # Выбор шрифта (подкомпонент настроек)
├── composables/
│   ├── useBookContent.ts             # НОВЫЙ — загрузка контента с API
│   ├── usePagination.ts              # НОВЫЙ — CSS column-based пагинация
│   ├── useReaderSettings.ts          # НОВЫЙ — управление настройками (§8.5)
│   ├── useReaderGestures.ts          # НОВЫЙ — свайпы и тапы (§8.8)
│   ├── useReaderKeyboard.ts          # НОВЫЙ — горячие клавиши (§8.9)
│   └── useReadingProgress.ts         # НОВЫЙ — сохранение/загрузка прогресса (§8.10)
├── stores/
│   └── reader.ts                     # НОВЫЙ — Pinia store читалки
├── types/
│   └── reader.ts                     # НОВЫЙ — TypeScript типы
├── assets/styles/
│   └── reader-themes.css             # НОВЫЙ — CSS темы и типографика (§8.6)
└── router/
    └── index.ts                      # ИЗМЕНЕНИЕ — маршрут /books/:id/read
```

**Structure Decision**: Web application (backend + frontend). Новая директория `backend/internal/bookfile/` для конвертеров форматов (расширяемо для EPUB/PDF/DJVU). Новая директория `frontend/src/components/reader/` для компонентов читалки. Структура соответствует архитектуре v8, раздел 7.

## Complexity Tracking

> Нарушений конституции нет — секция не применяется.

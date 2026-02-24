# Quickstart: Редизайн каталога в стиле MyHomeLib

**Date**: 2026-02-23
**Feature**: [spec.md](spec.md)

## Предварительные условия

- Node.js 22 LTS
- Go 1.25
- Docker + Docker Compose
- Локальный деплой HomeLib (для тестирования)

## Быстрый старт

### 1. Переключение на ветку

```bash
git checkout 007-catalog-redesign
```

### 2. Установка новой зависимости (splitpanes)

```bash
cd frontend
npm install splitpanes
npm install -D @types/splitpanes  # если есть
```

### 3. Бэкенд — расширение whitelist настроек

Файл: `backend/internal/api/handler/settings.go`

Найти whitelist allowed keys и добавить `"catalog"`:
```go
// Было:
allowedKeys := map[string]bool{"reader": true, "ui": true}
// Стало:
allowedKeys := map[string]bool{"reader": true, "ui": true, "catalog": true}
```

Проверить:
```bash
cd backend && go test ./internal/api/handler/... -run TestSettings -v
```

### 4. Фронтенд — порядок разработки

Рекомендуемая последовательность (от ядра к деталям):

1. **Типы и темы** — `types/catalog.ts`, определения 4 Vuetify-тем в `plugins/vuetify.ts`
2. **Theme store** — `stores/theme.ts` (загрузка/сохранение темы, наследование читалкой)
3. **CatalogView layout** — трёхпанельный layout со splitpanes
4. **BookTable** — таблица с сортировкой
5. **NavigationPanel + вкладки** — AuthorsTab, SeriesTab, GenresTab, SearchTab
6. **BookDetailPanel** — панель деталей
7. **CatalogHeader** — хедер с вкладками и меню
8. **ThemeSwitcher + SettingsDialog** — управление темами
9. **StatusBar** — статус-бар
10. **Интеграция с читалкой** — наследование темы

### 5. Запуск тестов

```bash
# Бэкенд
cd backend && go test -race ./...

# Фронтенд
cd frontend && npx vitest run

# С покрытием
cd frontend && npx vitest run --coverage
```

### 6. Локальная сборка и проверка

```bash
# Сборка Docker-образов
./scripts/build-and-push.sh --bump minor

# Локальный деплой
./scripts/deploy-stage.sh --tag <TAG>
```

## Ключевые файлы

### Новые файлы (создать)

| Файл | Описание |
|------|----------|
| `frontend/src/types/catalog.ts` | TypeScript-типы каталога |
| `frontend/src/stores/theme.ts` | Pinia store для управления темами |
| `frontend/src/composables/usePanelResize.ts` | Composable для персистентности размеров |
| `frontend/src/assets/styles/catalog-themes.css` | CSS-переменные для тем каталога (если нужны сверх Vuetify) |
| `frontend/src/components/catalog/CatalogHeader.vue` | Хедер |
| `frontend/src/components/catalog/NavigationPanel.vue` | Левая панель |
| `frontend/src/components/catalog/AuthorsTab.vue` | Вкладка авторов |
| `frontend/src/components/catalog/SeriesTab.vue` | Вкладка серий |
| `frontend/src/components/catalog/GenresTab.vue` | Вкладка жанров |
| `frontend/src/components/catalog/SearchTab.vue` | Вкладка поиска |
| `frontend/src/components/catalog/BookTable.vue` | Таблица книг |
| `frontend/src/components/catalog/BookDetailPanel.vue` | Панель деталей |
| `frontend/src/components/catalog/PanelResizer.vue` | (Опционально, если splitpanes стилизация) |
| `frontend/src/components/catalog/StatusBar.vue` | Статус-бар |
| `frontend/src/components/catalog/ThemeSwitcher.vue` | Быстрый переключатель тем |
| `frontend/src/components/catalog/SettingsDialog.vue` | Диалог настроек |

### Изменяемые файлы

| Файл | Что меняется |
|------|-------------|
| `frontend/src/plugins/vuetify.ts` | Добавить 4 темы (light, dark, sepia, night) |
| `frontend/src/views/CatalogView.vue` | Полная переработка: трёхпанельный layout |
| `frontend/src/stores/catalog.ts` | Навигация вкладками, выбор книги для деталей |
| `frontend/src/components/common/AppHeader.vue` | Обновление дропдаун-меню |
| `frontend/src/composables/useReaderSettings.ts` | Интеграция с темой каталога |
| `frontend/src/router/index.ts` | Удаление маршрутов /authors, /genres, /series |
| `backend/internal/api/handler/settings.go` | Whitelist: добавить "catalog" |

### Удаляемые файлы

| Файл | Причина |
|------|---------|
| `frontend/src/views/AuthorsView.vue` | Заменён на AuthorsTab внутри каталога |
| `frontend/src/views/GenresView.vue` | Заменён на GenresTab внутри каталога |
| `frontend/src/views/SeriesView.vue` | Заменён на SeriesTab внутри каталога |
| `frontend/src/components/common/BookCard.vue` | Заменён на BookTable (табличный вид) |
| `frontend/src/components/common/BookFilters.vue` | Фильтрация через вкладки навигации |
| `frontend/src/components/common/PaginationBar.vue` | Пагинация встроена в таблицу |

## Архитектурные решения

- **Vuetify 3 native themes** — 4 темы определяются как объекты в `createVuetify()`, переключение через `useTheme().global.name`
- **Splitpanes** — библиотека для resizable panels (2KB, zero-dependency)
- **Кастомная таблица** — вместо v-data-table для точного соответствия макету
- **Серверная синхронизация настроек** — через существующий `PUT /me/settings` с debounce
- **Reader theme inheritance** — `reader.theme === null` → использует `catalog.theme`

# Research: Редизайн каталога в стиле MyHomeLib

**Date**: 2026-02-23
**Feature**: [spec.md](spec.md)

## R-001: Система тем каталога поверх Vuetify 3

### Decision

Использовать нативную систему тем Vuetify 3 с 4 кастомными темами.

### Rationale

Vuetify 3 полностью поддерживает произвольное количество именованных тем. Каждая тема — объект с `dark: boolean`, `colors: {}` и `variables: {}`. Переключение в runtime через `useTheme().global.name.value = 'themeName'`. Все цвета автоматически становятся CSS-переменными `--v-theme-{color}` (формат RGB-компонентов: `25,118,210`), что позволяет использовать `rgb(var(--v-theme-primary))` в кастомном CSS.

Ключевые возможности:
- `useTheme().global.name.value` — реактивное имя текущей темы
- Проп `theme="dark"` на компоненте — изолированная тема для поддерева
- Автоматическая генерация `on-*` контрастных цветов
- Пользовательские цвета в `colors` → утилитарные классы `bg-{name}`, `text-{name}`
- `variations.lighten/darken` — авто-генерация вариантов

### Alternatives considered

1. **Ручные CSS variables без Vuetify themes** — требует дублирования логики переключения и отсутствия интеграции с Vuetify-компонентами. Отклонено.
2. **CSS prefers-color-scheme** — только light/dark, не поддерживает sepia/night. Отклонено.

---

## R-002: Resizable split panels

### Decision

Использовать библиотеку **splitpanes** (v4.x).

### Rationale

Splitpanes — минимальная (2KB gzipped), zero-dependency библиотека с нативной поддержкой Vue 3 и вложенных панелей. Трёхпанельный layout реализуется декларативно:

```html
<Splitpanes>
  <Pane min-size="15" size="25">Навигация</Pane>
  <Pane>
    <Splitpanes horizontal>
      <Pane min-size="20" size="60">Таблица</Pane>
      <Pane min-size="15">Детали</Pane>
    </Splitpanes>
  </Pane>
</Splitpanes>
```

- 121K загрузок/неделю, зрелая библиотека
- `min-size` / `max-size` в процентах
- Событие `@resized` для персистентности
- Touch-поддержка из коробки
- Есть форк от автора Pinia (@posva/splitpanes) с актуальными исправлениями

Известное ограничение: баг с `v-tabs` внутри вложенных Splitpanes (#35). В нашем случае не актуально — табы в хедере, а не внутри панелей.

### Alternatives considered

1. **Reka UI Splitter** — встроенный autoSaveId и collapse, но добавляет целую UI-библиотеку (Reka UI) как зависимость только ради сплиттера. Избыточно.
2. **Split.js** — фреймворк-агностичный, нужна Vue-обёртка. Лишняя работа при наличии готового Vue 3-решения.
3. **Кастомная реализация** — ~200-300 строк composable. Работоспособно, но splitpanes уже решает все edge cases (touch, iframe, text selection).

---

## R-003: Интеграция тем каталога и читалки

### Decision

Две раздельные настройки тем (`catalogTheme` и `readerTheme`) с механизмом наследования. Читалка наследует тему каталога по умолчанию, но позволяет независимый override.

### Rationale

**Хранение в settings JSONB:**
```json
{
  "reader": { "theme": "dark", "fontSize": 18, ... },
  "catalog": { "theme": "dark" }
}
```

**Логика наследования:**
- `catalog.theme` — тема каталога, выбирается пользователем
- `reader.theme` — может быть `null` (наследует catalog.theme) или явное значение (override)
- «Сбросить к теме каталога» → устанавливает `reader.theme = null`

**Реализация через Vuetify:**
- Каталог: `theme.global.name.value = catalogTheme`
- Читалка: проп `theme` на корневом `<v-sheet>` читалки, подставляет `readerTheme ?? catalogTheme`

Существующие 4 темы читалки (light/sepia/dark/night) уже определены через CSS variables в `reader-themes.css`. Нужно мигрировать их на Vuetify theme objects для единообразия, сохранив обратную совместимость CSS-переменных через `variables` в теме.

### Alternatives considered

1. **Единая тема для всего** — не соответствует требованию: «читалку можно переключить отдельно». Отклонено.
2. **Полностью раздельные темы без наследования** — пользователю пришлось бы настраивать две темы вручную при каждом изменении. Менее удобно.

---

## R-004: Персистентность размеров панелей

### Decision

localStorage для мгновенной реакции + debounced sync на сервер через `PUT /me/settings`.

### Rationale

Размеры панелей хранятся в `catalog.panelSizes` в user settings:
```json
{
  "catalog": {
    "theme": "dark",
    "panelSizes": {
      "leftWidth": 25,
      "tableHeight": 60
    }
  }
}
```

- При `@resized` от splitpanes → обновляем Pinia store и localStorage (мгновенно)
- Debounced save (1000ms) на сервер через `PUT /me/settings`
- При загрузке: сначала localStorage (мгновенно), затем server fetch (для кросс-девайс синхронизации)

### Alternatives considered

1. **Только localStorage** — теряются настройки при смене браузера/устройства. Отклонено: противоречит конституции §2.V (source of truth в PostgreSQL).
2. **Только server** — задержка при начальной загрузке, видимый «прыжок» layout. Отклонено: плохой UX.

---

## R-005: Таблица книг

### Decision

Кастомная таблица на CSS Grid с ручной сортировкой. Без виртуализации.

### Rationale

- Vuetify `v-data-table` тяжеловесен и с трудом кастомизируется под «IDE-стилистику» из макета
- Максимум 100 записей на странице (ограничение API `limit=100`) — виртуализация не нужна
- Сортировка выполняется на сервере (параметры `sort` и `order` в API) — фронтенд только отображает
- Кастомная таблица позволяет точно воспроизвести стилистику макета: компактные строки, hover-эффекты, моноширинный шрифт для размеров

### Alternatives considered

1. **Vuetify v-data-table** — подходит для generic admin-панелей, но сложно кастомизировать под конкретный дизайн макета. Отклонено.
2. **AG Grid / TanStack Table** — enterprise-решения, избыточны для 100 строк. Отклонено.
3. **Виртуальный скроллинг (vue-virtual-scroller)** — нужен при 1000+ строк, для 100 строк overhead. Отклонено.

---

## R-006: Расширяемость тем

### Decision

Темы определяются как TypeScript-объекты с интерфейсом `CatalogThemeDefinition`. Регистрация в Vuetify — декларативная через массив тем.

### Rationale

```typescript
interface CatalogThemeDefinition {
  name: string        // 'light' | 'dark' | 'sepia' | 'night' | string
  label: string       // Человекочитаемое имя для UI
  dark: boolean       // Vuetify dark flag
  colors: Record<string, string>  // Vuetify color tokens
  variables?: Record<string, string | number>  // CSS variables
}
```

Добавление новой темы:
1. Создать объект `CatalogThemeDefinition`
2. Добавить в массив `catalogThemes`
3. Тема автоматически появляется в Vuetify, переключателе и диалоге настроек

Никаких изменений в компонентах не требуется — все завязаны на CSS variables, не на конкретные имена тем.

### Alternatives considered

1. **JSON-файл тем, загружаемый с сервера** — overengineering для 4 тем. Возможно добавить в будущем, когда появится UI для создания пользовательских тем. Отклонено на данном этапе.

---

## R-007: Бэкенд — расширение whitelist настроек

### Decision

Добавить `"catalog"` в whitelist allowed keys в `settings.go`. Без новых эндпоинтов, миграций или моделей.

### Rationale

Текущий handler `UpdateUserSettings` в `settings.go` имеет whitelist: `["reader", "ui"]`. Достаточно добавить `"catalog"` — все остальное (JSONB merge, validation, max size 64KB) уже работает.

Структура settings после изменения:
```json
{
  "reader": { "theme": null, "fontSize": 18, ... },
  "catalog": { "theme": "dark", "panelSizes": { "leftWidth": 25, "tableHeight": 60 } }
}
```

### Alternatives considered

1. **Новый эндпоинт `/me/catalog-settings`** — нарушает DRY, дублирует логику settings handler. Отклонено.
2. **Хранить тему каталога в `ui` ключе** — смешивает семантически разные данные. Лучше разделить `catalog` и `reader`. Отклонено.

---

## R-008: Удаление отдельных страниц авторов/серий/жанров

### Decision

Маршруты `/authors`, `/genres`, `/series` удаляются из router. Вкладки навигации внутри каталога заменяют их полностью.

### Rationale

Спецификация явно указывает (Assumptions): «Маршруты /authors, /genres, /series как отдельные страницы заменяются на вкладки в едином каталоге». Компоненты `AuthorsView.vue`, `GenresView.vue`, `SeriesView.vue` будут удалены. Маршруты `/authors/:id` и `/books/:id` сохраняются — детали автора и книги доступны через таблицу + панель деталей.

Фактически, `/authors/:id` тоже заменяется: выбор автора в левой панели фильтрует таблицу. Маршрут `/books/:id` (отдельная страница BookView) может остаться как fallback для прямых ссылок.

### Alternatives considered

1. **Сохранить отдельные страницы как альтернативный вид** — усложняет поддержку двух UI. Отклонено.

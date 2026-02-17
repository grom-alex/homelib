# Implementation Plan: Скрипты сборки и деплоя

**Branch**: `003-build-deploy-scripts` | **Date**: 2026-02-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-build-deploy-scripts/spec.md`

## Summary

Адаптация скриптов сборки и деплоя из проекта kids-accounting к архитектуре HomeLib. Скрипты уже скопированы в `scripts/`, но 5 из 9 содержат ссылки на kids-accounting (`IMAGE_PREFIX`, `REMOTE_APP_DIR`, container names). Необходимо: (1) заменить все ссылки на homelib, (2) добавить поддержку 3-х компонентов (api, worker, frontend) вместо 2-х, (3) обновить старые утилитарные скрипты для использования общих библиотек из `scripts/lib/`.

## Technical Context

**Language/Version**: Bash (POSIX-compatible shell scripts)
**Primary Dependencies**: Docker, Docker Compose v2, ssh, git, go 1.25, node 22
**Storage**: N/A (скрипты не добавляют хранилище)
**Testing**: ShellCheck (статический анализ bash), ручное тестирование
**Target Platform**: Linux server (homelab), macOS (разработка)
**Project Type**: Infrastructure/DevOps scripts для web-приложения
**Performance Goals**: Локальная сборка < 5 мин (с кэшем), деплой < 2 мин
**Constraints**: Простота обслуживания одним человеком, POSIX-совместимость
**Scale/Scope**: 13 скриптов + 3 библиотеки, 3 Docker-образа

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Pre-design | Post-design | Комментарий |
|---------|------------|-------------|-------------|
| §1.VII Нет внешних SaaS | PASS | PASS | Все скрипты работают локально, Docker Registry опционален |
| §6.I Docker Compose оркестратор | PASS | PASS | Скрипты используют Docker Compose для деплоя |
| §6.II Масштаб homelab | PASS | PASS | Один сервер, один администратор |
| §6.V Конфиг через YAML | PASS | PASS | Скрипты используют env vars + .env файлы, не конфликтуют с YAML-конфигом |
| §6.VI GitHub Flow | PASS | PASS | Разработка на ветке 003-build-deploy-scripts |
| §6.VII Архитектурная документация | PASS | PASS | Структура scripts/ уже отражена в architecture-v8 |
| §7 Тестирование 80% | N/A | N/A | Bash-скрипты не входят в scope покрытия unit-тестами |

**Результат**: Все gates пройдены, нарушений нет. Post-design проверка подтверждена.

## Project Structure

### Documentation (this feature)

```text
specs/003-build-deploy-scripts/
├── plan.md              # Этот файл
├── research.md          # Решения по адаптации
├── quickstart.md        # Инструкции по использованию
└── tasks.md             # Задачи (генерируется /speckit.tasks)
```

### Source Code (repository root)

```text
scripts/
├── lib/                     # Общие библиотеки
│   ├── logging.sh           # Логирование с уровнями и цветами
│   ├── prerequisites.sh     # Проверка зависимостей
│   └── docker-utils.sh      # Docker-операции (build, push, health)
├── build-local.sh           # Локальная сборка Docker-образов
├── build-and-push.sh        # Сборка + тесты + публикация в registry
├── deploy.sh                # Универсальный деплой-обёртка
├── deploy-local.sh          # Деплой в локальное окружение
├── deploy-stage.sh          # Деплой на staging через SSH
├── deploy-prod.sh           # Деплой на production через SSH
├── build.sh                 # Сборка Go-бинарей (без Docker)
├── backup-db.sh             # Бэкап PostgreSQL
├── restore-db.sh            # Восстановление PostgreSQL
├── migrate.sh               # Документация по миграциям
├── import-inpx.sh           # Запуск INPX-импорта
├── dev-setup.sh             # Настройка dev-окружения
└── setup-ollama-windows.ps1 # Настройка Ollama на Windows
```

**Structure Decision**: Все скрипты в `scripts/` с библиотеками в `scripts/lib/`. Старый `deploy.sh` → `deploy-old.sh`, новый `deploy.sh` — обёртка.

## Адаптация: kids-accounting → HomeLib

### Файлы с необходимыми исправлениями

| Файл | Проблема | Исправление |
|------|----------|-------------|
| `lib/docker-utils.sh` | `IMAGE_PREFIX=apps/kids-accounting` | → `apps/homelib` |
| `lib/docker-utils.sh` | Container names `kids_accounting_*` | → `homelib_*` |
| `build-and-push.sh` | `IMAGE_PREFIX=apps/kids-accounting` | → `apps/homelib` |
| `build-and-push.sh` | Только 2 образа (backend, frontend) | Добавить worker |
| `deploy-stage.sh` | `IMAGE_PREFIX=apps/kids-accounting` | → `apps/homelib` |
| `deploy-stage.sh` | `REMOTE_APP_DIR=/opt/kids-accounting` | → `/opt/homelib` |
| `deploy-prod.sh` | `IMAGE_PREFIX=apps/kids-accounting` | → `apps/homelib` |
| `deploy-prod.sh` | `REMOTE_APP_DIR=/opt/kids-accounting` | → `/opt/homelib` |
| `deploy.sh` | Захардкоженные порты, нет worker | Адаптировать к 3 компонентам |

### Файлы уже адаптированные (без изменений)

- `build-local.sh` — корректно работает с 3 компонентами HomeLib
- `deploy-local.sh` — корректные пути к docker-compose файлам
- `lib/logging.sh` — универсальная библиотека
- `lib/prerequisites.sh` — универсальная, минорная правка doc-ссылки

### Старые скрипты для обновления (использование lib/)

| Файл | Текущее состояние | Требуется |
|------|-------------------|-----------|
| `backup-db.sh` | Простой, без lib/ | Добавить sourcing lib/logging.sh |
| `restore-db.sh` | Простой, без lib/ | Добавить sourcing lib/logging.sh |
| `dev-setup.sh` | Свои проверки зависимостей | Перейти на lib/prerequisites.sh |
| `build.sh` | Минимальный | Добавить sourcing lib/logging.sh |
| `import-inpx.sh` | Простой wrapper | Добавить sourcing lib/logging.sh |
| `migrate.sh` | Документация | Добавить sourcing lib/logging.sh |

### Удаление

- `deploy-old.sh` — удалить после проверки нового `deploy.sh`

## Complexity Tracking

> Нарушений конституции нет. Таблица не заполняется.

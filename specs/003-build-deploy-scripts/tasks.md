# Tasks: Скрипты сборки и деплоя

**Input**: Design documents from `/specs/003-build-deploy-scripts/`
**Prerequisites**: plan.md (required), spec.md (required), research.md

**Tests**: Не запрашивались. Верификация — ручная проверка `--help`, `bash -n`, запуск скриптов.

**Organization**: Tasks grouped by user story. US4 (библиотеки) идёт как Foundational, т.к. от неё зависят все остальные скрипты.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1–US5)

---

## Phase 1: Setup

**Purpose**: Верификация базовой структуры и уже адаптированных файлов

- [x] T001 Проверить корректность `scripts/lib/logging.sh` — убедиться, что нет ссылок на kids-accounting
- [x] T002 [P] Проверить корректность `scripts/lib/prerequisites.sh` — исправить ссылку на docs/ если нужно

---

## Phase 2: Foundational — Библиотеки (US4, P2)

**Purpose**: Исправить общие библиотеки, от которых зависят все скрипты. БЛОКИРУЕТ все остальные фазы.

**⚠️ CRITICAL**: lib/docker-utils.sh используется всеми build/deploy скриптами. Исправления здесь — приоритет.

- [x] T003 [US4] Заменить `IMAGE_PREFIX=apps/kids-accounting` → `apps/homelib` в `scripts/lib/docker-utils.sh`
- [x] T004 [US4] Заменить container names `kids_accounting_*` → `homelib_*` в `scripts/lib/docker-utils.sh`
- [x] T005 [US4] Проверить `set -euo pipefail` и exit codes в `scripts/lib/docker-utils.sh`
- [x] T006 [US4] Убедиться, что `scripts/lib/logging.sh` поддерживает все уровни (DEBUG, INFO, SUCCESS, WARN, ERROR, FATAL) с `--help` документацией

**Checkpoint**: Библиотеки готовы, можно приступать к скриптам.

---

## Phase 3: User Story 1 — Локальная сборка Docker-образов (P1) 🎯 MVP

**Goal**: `./scripts/build-local.sh` собирает все 3 образа HomeLib (api, worker, frontend)

**Independent Test**: `./scripts/build-local.sh --help` выводит справку; `bash -n scripts/build-local.sh` проходит без ошибок

- [x] T007 [US1] Проверить `scripts/build-local.sh` — убедиться, что строит 3 компонента (api, worker, frontend) и использует `lib/` корректно
- [x] T008 [US1] Проверить, что `--component backend` строит оба образа (api и worker) в `scripts/build-local.sh`
- [x] T009 [US1] Проверить `--help` вывод и `set -euo pipefail` в `scripts/build-local.sh`

**Checkpoint**: Локальная сборка работает для всех компонентов.

---

## Phase 4: User Story 2 — Деплой в окружения (P1)

**Goal**: `deploy-local.sh`, `deploy-stage.sh`, `deploy.sh` работают с HomeLib

**Independent Test**: `./scripts/deploy-local.sh --help` выводит справку; скрипт проверяет `.env` и поднимает сервисы

### Implementation for User Story 2

- [x] T010 [US2] Адаптировать `scripts/deploy.sh` — заменить захардкоженные порты, добавить поддержку worker-компонента, проверить health check для 3 сервисов
- [x] T011 [US2] Заменить `IMAGE_PREFIX=apps/kids-accounting` → `apps/homelib` в `scripts/deploy-stage.sh`
- [x] T012 [US2] Заменить `REMOTE_APP_DIR=/opt/kids-accounting` → `/opt/homelib` в `scripts/deploy-stage.sh`
- [x] T013 [US2] Заменить container names `kids_accounting_*` → `homelib_*` в `scripts/deploy-stage.sh`
- [x] T014 [US2] Проверить `scripts/deploy-local.sh` — убедиться в корректных путях к `docker/docker-compose.*.yml` и health check
- [x] T015 [US2] Проверить `--help` и `set -euo pipefail` во всех deploy-скриптах: `scripts/deploy.sh`, `scripts/deploy-local.sh`, `scripts/deploy-stage.sh`

**Checkpoint**: Локальный и staging деплой работают.

---

## Phase 5: User Story 3 — Build & Push в Registry (P2)

**Goal**: `./scripts/build-and-push.sh` собирает 3 образа, тестирует и пушит в registry

**Independent Test**: `./scripts/build-and-push.sh --help` выводит справку; `bash -n scripts/build-and-push.sh` проходит

- [x] T016 [US3] Заменить `IMAGE_PREFIX=apps/kids-accounting` → `apps/homelib` в `scripts/build-and-push.sh`
- [x] T017 [US3] Добавить сборку worker-образа в `scripts/build-and-push.sh` (сейчас только backend + frontend, нужен ещё worker)
- [x] T018 [US3] Проверить логику тестирования перед сборкой — убедиться, что вызывается `go test` и `npm run test` из корректных директорий в `scripts/build-and-push.sh`
- [x] T019 [US3] Проверить стратегию тегирования: `sha-<hash>`, `<version>`, `latest` в `scripts/build-and-push.sh`

**Checkpoint**: Build & Push работает для всех 3 компонентов.

---

## Phase 6: User Story 4 — Обновление старых скриптов (P2)

**Goal**: Все старые утилитарные скрипты используют `lib/logging.sh` для единообразного вывода

**Independent Test**: Каждый скрипт выводит сообщения через `log_info`/`log_error` вместо plain `echo`

- [x] T020 [P] [US4] Обновить `scripts/backup-db.sh` — добавить sourcing `lib/logging.sh`, заменить `echo` на `log_info`/`log_error`
- [x] T021 [P] [US4] Обновить `scripts/restore-db.sh` — добавить sourcing `lib/logging.sh`, заменить `echo` на `log_info`/`log_error`
- [x] T022 [P] [US4] Обновить `scripts/build.sh` — добавить sourcing `lib/logging.sh`, заменить `echo` на `log_info`/`log_error`
- [x] T023 [P] [US4] Обновить `scripts/import-inpx.sh` — добавить sourcing `lib/logging.sh`, заменить `echo` на `log_info`/`log_error`
- [x] T024 [P] [US4] Обновить `scripts/migrate.sh` — добавить sourcing `lib/logging.sh`, заменить `echo` на `log_info`/`log_error`
- [x] T025 [US4] Обновить `scripts/dev-setup.sh` — заменить встроенные проверки зависимостей на sourcing `lib/prerequisites.sh`, заменить `echo` на `log_info`/`log_error`

**Checkpoint**: Все скрипты используют единые библиотеки.

---

## Phase 7: User Story 5 — Production деплой (P3)

**Goal**: `./scripts/deploy-prod.sh` выполняет безопасный SSH-деплой с бэкапом и dry-run

**Independent Test**: `./scripts/deploy-prod.sh --help` выводит справку; `--dry-run` показывает план без выполнения

- [x] T026 [US5] Заменить `IMAGE_PREFIX=apps/kids-accounting` → `apps/homelib` в `scripts/deploy-prod.sh`
- [x] T027 [US5] Заменить `REMOTE_APP_DIR=/opt/kids-accounting` → `/opt/homelib` в `scripts/deploy-prod.sh`
- [x] T028 [US5] Проверить логику бэкапа, health check и `--dry-run` в `scripts/deploy-prod.sh`
- [x] T029 [US5] Проверить `--help` и `set -euo pipefail` в `scripts/deploy-prod.sh`

**Checkpoint**: Production деплой адаптирован к HomeLib.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Финальная уборка и валидация

- [x] T030 Удалить `scripts/deploy-old.sh` — новый `deploy.sh` его заменяет
- [x] T031 Проверить `chmod +x` на всех скриптах в `scripts/` и `scripts/lib/`
- [x] T032 Выполнить `bash -n` проверку синтаксиса на всех .sh файлах в `scripts/` и `scripts/lib/`
- [x] T033 Проверить, что все скрипты поддерживают `--help` (FR-002)
- [x] T034 Выполнить сценарии из `specs/003-build-deploy-scripts/quickstart.md` для валидации

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: Нет зависимостей — начинать сразу
- **Phase 2 (Foundational/US4 lib)**: Зависит от Phase 1 — БЛОКИРУЕТ все остальные фазы
- **Phase 3 (US1)**: Зависит от Phase 2
- **Phase 4 (US2)**: Зависит от Phase 2, может идти параллельно с Phase 3
- **Phase 5 (US3)**: Зависит от Phase 2, может идти параллельно с Phase 3/4
- **Phase 6 (US4 старые скрипты)**: Зависит от Phase 2, может идти параллельно с Phase 3/4/5
- **Phase 7 (US5)**: Зависит от Phase 2, может идти параллельно с Phase 3–6
- **Phase 8 (Polish)**: Зависит от завершения всех фаз

### User Story Dependencies

- **US1 (build-local)**: Независима после Phase 2
- **US2 (deploy)**: Независима после Phase 2
- **US3 (build-and-push)**: Независима после Phase 2
- **US4 (lib + старые скрипты)**: Phase 2 = lib, Phase 6 = старые скрипты
- **US5 (deploy-prod)**: Независима после Phase 2

### Parallel Opportunities

- T001 и T002 (Phase 1) — параллельно
- T020–T024 (Phase 6) — все параллельно (разные файлы)
- Phase 3, 4, 5, 6, 7 — могут идти параллельно после Phase 2

---

## Parallel Example: Phase 6

```bash
# Все старые скрипты обновляются параллельно:
Task T020: "Обновить backup-db.sh"
Task T021: "Обновить restore-db.sh"
Task T022: "Обновить build.sh"
Task T023: "Обновить import-inpx.sh"
Task T024: "Обновить migrate.sh"
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Phase 1: Setup (T001–T002)
2. Phase 2: Foundational — исправить lib/ (T003–T006)
3. Phase 3: US1 — проверить build-local.sh (T007–T009)
4. **STOP and VALIDATE**: `./scripts/build-local.sh --help` и `bash -n`

### Incremental Delivery

1. Setup + Foundational → lib/ готовы
2. US1 (build-local) → MVP сборки
3. US2 (deploy) → деплой работает
4. US3 (build-and-push) → CI/CD pipeline
5. US4 (старые скрипты) → единообразие
6. US5 (deploy-prod) → production ready
7. Polish → финальная валидация

---

## Notes

- Большинство задач — замена строк (find & replace), не написание нового кода
- `build-local.sh` и `deploy-local.sh` уже адаптированы — задачи на верификацию
- 5 скриптов требуют замены kids-accounting → homelib
- 6 старых скриптов требуют добавления sourcing lib/
- 1 файл на удаление (deploy-old.sh)

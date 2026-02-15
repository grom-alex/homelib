# Tasks: Настройка GitHub CI/CD и GitHub Flow

**Input**: Design documents from `/specs/001-github-ci-setup/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: Тесты не запрошены в спецификации. Верификация выполняется вручную через создание PR и проверку в GitHub UI.

**Organization**: Задачи сгруппированы по user stories для независимой реализации и тестирования.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Может выполняться параллельно (разные файлы, нет зависимостей)
- **[Story]**: К какой user story относится задача (US1, US2, US3)
- Точные пути к файлам указаны в описании

---

## Phase 1: Setup

**Purpose**: Создание структуры каталогов для CI конфигурации

- [x] T001 Создать директорию `.github/workflows/` в корне репозитория

**Checkpoint**: Структура директорий готова.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Базовая конфигурация, необходимая для всех user stories

**⚠️ CRITICAL**: Работа над user stories невозможна без завершения этой фазы

- [x] T002 Создать конфигурацию golangci-lint v2 в `.golangci.yml` — version: "2", linters.default: standard

**Checkpoint**: Конфигурация линтера готова, можно создавать CI workflow.

---

## Phase 3: User Story 1 - Автоматическая проверка PR (Priority: P1) 🎯 MVP

**Goal**: При создании PR в master автоматически запускается CI pipeline: сборка, тесты, линтинг Go-кода. Результат виден как status check.

**Independent Test**: Создать PR из feature-ветки в master → CI должен запуститься → результат отображается в Checks.

### Implementation for User Story 1

- [x] T003 [US1] Создать CI workflow `.github/workflows/ci.yml` с trigger `pull_request` на ветку `master`, job name `CI`, runner `ubuntu-latest`
- [x] T004 [US1] Добавить шаг `Check for Go code` в `.github/workflows/ci.yml` — проверка наличия `go.mod`, вывод `has_go` в `GITHUB_OUTPUT`
- [x] T005 [US1] Добавить шаг `actions/checkout@v5` в `.github/workflows/ci.yml`
- [x] T006 [US1] Добавить шаг `actions/setup-go@v6` с `go-version: '1.25'` в `.github/workflows/ci.yml` — условно: `if: steps.check.outputs.has_go == 'true'`
- [x] T007 [US1] Добавить шаг `golangci/golangci-lint-action@v9` с `version: v2.9` в `.github/workflows/ci.yml` — условно: `if: steps.check.outputs.has_go == 'true'`
- [x] T008 [US1] Добавить шаг `go test -race ./...` в `.github/workflows/ci.yml` — условно: `if: steps.check.outputs.has_go == 'true'`
- [x] T009 [US1] Добавить шаг `go build ./cmd/...` в `.github/workflows/ci.yml` — условно: `if: steps.check.outputs.has_go == 'true'`
- [x] T010 [US1] Установить `timeout-minutes: 15` на уровне job в `.github/workflows/ci.yml`

**Checkpoint**: PR в master запускает CI с этапами lint → test → build. При отсутствии Go-кода — graceful success.

---

## Phase 4: User Story 2 - Проверка при пуше в feature-ветку (Priority: P2)

**Goal**: CI pipeline запускается при пуше в любую ветку кроме master, обеспечивая раннюю обратную связь.

**Independent Test**: Запушить коммит в feature-ветку → CI запускается → результат виден на вкладке Actions.

### Implementation for User Story 2

- [x] T011 [US2] Добавить trigger `push` с `branches-ignore: [master]` в `.github/workflows/ci.yml`

**Checkpoint**: Push в feature-ветку запускает CI pipeline. Результат виден на вкладке Actions.

---

## Phase 5: User Story 3 - Защита ветки master (Priority: P3)

**Goal**: Branch protection ruleset запрещает мерж в master без прохождения CI.

**Independent Test**: Проверить через `gh api repos/grom-alex/homelib/rulesets` наличие активного ruleset. Попытаться замержить PR с failed CI — кнопка Merge заблокирована.

### Implementation for User Story 3

- [x] T012 [US3] Создать branch protection ruleset для `master` через `gh api repos/grom-alex/homelib/rulesets` — required_status_checks (context: `CI`), non_fast_forward, deletion
- [x] T013 [US3] Верифицировать ruleset через `gh api repos/grom-alex/homelib/rulesets --jq '.[] | {id, name, enforcement}'`

**Checkpoint**: Ruleset "Protect master" активен. Мерж в master без прохождения CI заблокирован.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Верификация и завершение

- [ ] T014 Запушить ветку `001-github-ci-setup` в origin (`git push -u origin 001-github-ci-setup`)
- [ ] T015 Убедиться, что CI workflow запустился на пуше (проверить вкладку Actions)
- [ ] T016 Убедиться, что CI завершился со статусом success (нет Go-кода — graceful skip)
- [ ] T017 Выполнить валидацию по quickstart.md в `specs/001-github-ci-setup/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Нет зависимостей — начинается сразу
- **Foundational (Phase 2)**: Зависит от Phase 1 — БЛОКИРУЕТ все user stories
- **US1 (Phase 3)**: Зависит от Phase 2 — создаёт ci.yml
- **US2 (Phase 4)**: Зависит от US1 (Phase 3) — добавляет push trigger в тот же ci.yml
- **US3 (Phase 5)**: Зависит от US1 (Phase 3) — требует существования CI workflow для status check
- **Polish (Phase 6)**: Зависит от всех предыдущих фаз

### User Story Dependencies

- **US1 (P1)**: Может начинаться после Phase 2. Создаёт ci.yml.
- **US2 (P2)**: Зависит от US1 — модифицирует тот же ci.yml (добавляет trigger).
- **US3 (P3)**: Зависит от US1 — настраивает branch protection на status check `CI`, который должен существовать.

### Within Each User Story

- Шаги выполняются последовательно (модификация одного файла)
- T005 (checkout) должен быть ПЕРЕД T006 (setup-go)
- T006 (setup-go) должен быть ПЕРЕД T007 (lint) и T008 (test)

### Parallel Opportunities

- T002 (golangci config) может создаваться параллельно с T001 (directory)
- US3 (T012-T013) может выполняться параллельно с US2 (T011) после завершения US1
- T014-T017 (polish) выполняются последовательно

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002)
3. Complete Phase 3: User Story 1 (T003-T010)
4. **STOP and VALIDATE**: Создать PR, проверить CI

### Incremental Delivery

1. Setup + Foundational → конфигурация готова
2. Add US1 → CI на PR → Deploy/Verify (MVP!)
3. Add US2 → CI на push → Verify
4. Add US3 → Branch protection → Verify
5. Polish → End-to-end валидация

---

## Notes

- Все Go-шаги в ci.yml условны (`if: steps.check.outputs.has_go == 'true'`) — FR-007
- Один job `CI` с последовательными шагами (не отдельные job'ы) — решение из research.md
- Branch protection через Rulesets API (не legacy) — решение из research.md
- Code review requirement НЕ включён (один разработчик)

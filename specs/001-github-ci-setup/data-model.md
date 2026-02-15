# Data Model: Настройка GitHub CI/CD и GitHub Flow

**Feature Branch**: `001-github-ci-setup`
**Date**: 2026-02-15

## Обзор

Данная фича не создаёт данных в приложении — она конфигурирует инфраструктуру CI/CD.
Модель данных описывает конфигурационные артефакты.

## Сущности

### CI Workflow Configuration

**Файл**: `.github/workflows/ci.yml`

| Атрибут | Значение | Описание |
|---------|----------|----------|
| name | `CI` | Имя workflow, отображается в GitHub UI и status checks |
| triggers | `push` (все ветки кроме master), `pull_request` (в master) | Условия запуска |
| runner | `ubuntu-latest` | Среда выполнения |
| go-version | `1.25` | Версия Go |
| lint-version | `v2.9` | Версия golangci-lint |

**Состояния workflow**: `queued → in_progress → completed (success | failure | cancelled)`

### golangci-lint Configuration

**Файл**: `.golangci.yml`

| Атрибут | Значение | Описание |
|---------|----------|----------|
| version | `"2"` | Формат конфигурации v2 |
| linters.default | `standard` | Набор линтеров по умолчанию |

### Branch Protection Ruleset

**Конфигурируется через**: GitHub API (gh cli)

| Правило | Параметр | Описание |
|---------|----------|----------|
| required_status_checks | context: `CI` | Требует прохождения CI |
| non_fast_forward | — | Запрет force-push |
| deletion | — | Запрет удаления ветки |

## Связи

```
.github/workflows/ci.yml
    ├── triggers → push events, PR events
    ├── uses → actions/checkout@v5
    ├── uses → actions/setup-go@v6
    ├── uses → golangci/golangci-lint-action@v9
    └── reads → .golangci.yml (config)

Branch Protection Ruleset
    └── requires → CI job status check (name: "CI")
```

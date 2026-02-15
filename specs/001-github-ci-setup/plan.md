# Implementation Plan: Настройка GitHub CI/CD и GitHub Flow

**Branch**: `001-github-ci-setup` | **Date**: 2026-02-15 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-github-ci-setup/spec.md`

## Summary

Настройка GitHub Actions CI pipeline для Go-бэкенда проекта HomeLib: сборка, тестирование, линтинг. Настройка branch protection для `master` через GitHub Rulesets API. Пайплайн корректно обрабатывает отсутствие Go-кода (graceful skip). Конституция и CLAUDE.md уже обновлены (GitHub Flow).

## Technical Context

**Language/Version**: Go 1.25 (latest patch 1.25.7)
**Primary Dependencies**: GitHub Actions (`actions/checkout@v5`, `actions/setup-go@v6`, `golangci/golangci-lint-action@v9`)
**Storage**: N/A (конфигурационные файлы)
**Testing**: Ручная верификация через создание PR
**Target Platform**: GitHub Actions runners (ubuntu-latest)
**Project Type**: Infrastructure/CI configuration
**Performance Goals**: CI pipeline < 10 минут
**Constraints**: Один разработчик, homelab-проект
**Scale/Scope**: 1 workflow, 1 ruleset, 1 lint config

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Статус | Комментарий |
|---------|--------|-------------|
| §1.VII Отсутствие внешних SaaS | ✅ PASS | GitHub Actions — часть GitHub, который уже используется как хостинг репозитория, не является дополнительной SaaS-зависимостью |
| §6.I Docker Compose оркестратор | ✅ PASS | CI не затрагивает deployment, Docker Compose остаётся оркестратором |
| §6.VI GitHub Flow | ✅ PASS | Фича напрямую реализует этот принцип |
| §7 Change Governance | ✅ PASS | Конституция обновлена до v1.1.0 с добавлением §6.VI |

**Gate result**: PASS — все принципы соблюдены, нарушений нет.

## Project Structure

### Documentation (this feature)

```text
specs/001-github-ci-setup/
├── plan.md              # Этот файл
├── research.md          # Phase 0: исследование версий и best practices
├── data-model.md        # Phase 1: описание конфигурационных сущностей
├── quickstart.md        # Phase 1: инструкция по проверке
├── contracts/           # Phase 1: описание контрактов
└── checklists/
    └── requirements.md  # Чеклист качества спецификации
```

### Source Code (repository root)

```text
.github/
└── workflows/
    └── ci.yml           # GitHub Actions CI workflow

.golangci.yml            # golangci-lint v2 configuration
```

**Structure Decision**: Инфраструктурная фича — добавляются только конфигурационные файлы в корень репозитория. Файлы приложения не создаются и не изменяются.

## Complexity Tracking

Нарушений Constitution Check нет. Таблица не заполняется.

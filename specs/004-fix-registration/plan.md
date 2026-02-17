# Implementation Plan: Fix User Registration

**Branch**: `004-fix-registration` | **Date**: 2026-02-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-fix-registration/spec.md`

## Summary

Исправление критического бага: `SELECT COUNT(*) FROM users FOR UPDATE` несовместим с PostgreSQL (SQLSTATE 0A000 — FOR UPDATE запрещён с агрегатными функциями). Требуется заменить на корректный SQL с блокировкой таблицы и добавить серверное логирование ошибок регистрации.

## Technical Context

**Language/Version**: Go 1.25 (latest patch 1.25.7)
**Primary Dependencies**: Gin (HTTP framework), pgx/v5 (PostgreSQL driver), golang-jwt/jwt/v5, bcrypt
**Storage**: PostgreSQL 17 + pg_trgm + tsvector
**Testing**: `go test -race -coverprofile=coverage.out ./...` (backend), `vitest --coverage` (frontend)
**Target Platform**: Linux server (Docker containers)
**Project Type**: web (backend + frontend)
**Performance Goals**: N/A (bug fix, не влияет на производительность)
**Constraints**: Минимальное покрытие unit-тестами — 80% для каждого пакета
**Scale/Scope**: Bug fix — затрагивает 2-3 файла в backend

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Статус | Комментарий |
|---------|--------|-------------|
| §1.II Разделение ответственности | PASS | Изменения только в repository и handler layers |
| §2.I PostgreSQL как единственная БД | PASS | Исправляем SQL-запрос к PostgreSQL |
| §4.IV Управление регистрацией | PASS | Первый пользователь получает admin; блокировка предотвращает race condition |
| §4.V Хеширование паролей и токенов | PASS | Не затрагиваем — bcrypt и SHA-256 уже на месте |
| §6.VI GitHub Flow | PASS | Работаем в ветке 004-fix-registration |
| §7 Тестирование 80% | PASS | Добавляем unit-тесты для RegisterUser |

**Gate result**: PASS — все принципы соблюдены, нарушений нет.

## Project Structure

### Documentation (this feature)

```text
specs/004-fix-registration/
├── plan.md              # This file
├── research.md          # Phase 0: research on PostgreSQL locking
├── quickstart.md        # Phase 1: quick verification steps
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── api/
│   │   └── handler/
│   │       └── auth.go          # [MODIFY] Add error logging in Register handler
│   └── repository/
│       ├── user.go              # [MODIFY] Fix SQL query in RegisterUser
│       └── user_test.go         # [ADD] Unit tests for RegisterUser
```

**Structure Decision**: Bug fix — изменения только в существующих файлах backend. Новый файл — только `user_test.go` для unit-тестов. Полностью соответствует архитектуре из `docs/homelib-architecture-v8.md`.

## Complexity Tracking

Нарушений конституции нет — таблица не требуется.

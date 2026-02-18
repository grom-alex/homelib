# Implementation Plan: Исправление истечения сессии

**Branch**: `005-fix-session-expiry` | **Date**: 2026-02-18 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/005-fix-session-expiry/spec.md`

## Summary

Исправление ошибки 401 при истечении access token: корневая причина — `cookie_secure: true` в stage-конфиге при HTTP-соединении (refresh_token cookie отклоняется браузером). Дополнительно: синхронизация interceptor с Pinia store при неудачном refresh, redirect-back после повторного входа, увеличение TTL access token до 2 часов, исправление кнопки "Выход".

## Technical Context

**Language/Version**: Go 1.25 (backend), TypeScript + Vue 3 (frontend)
**Primary Dependencies**: Gin (HTTP), pgx/v5 (DB), axios (HTTP client), Pinia (state), Vue Router
**Storage**: PostgreSQL 17 (refresh_tokens table)
**Testing**: `go test -race` (backend), `vitest` (frontend)
**Target Platform**: Docker Compose, Linux server
**Project Type**: web (backend + frontend SPA)
**Performance Goals**: N/A (баг-фикс, не новый функционал)
**Constraints**: Домашняя библиотека, единицы пользователей
**Scale/Scope**: 5 файлов frontend + 3 файла config

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Принцип | Статус | Комментарий |
|---------|--------|-------------|
| §1.III Stateless API | ✅ PASS | JWT-аутентификация сохраняется, состояние в токенах |
| §4.I JWT-аутентификация | ✅ PASS | Access token в памяти, refresh в httpOnly cookie — без изменений |
| §4.IV Управление регистрацией | ✅ PASS | Не затрагивается |
| §4.V Хеширование паролей и токенов | ✅ PASS | SHA-256 для refresh tokens — без изменений |
| §6.V Конфигурация через YAML | ✅ PASS | TTL меняется в YAML-конфигах |
| §7 Требования к тестированию | ✅ PASS | Unit-тесты ≥80% для изменённых модулей |
| §7 TDD | ✅ PASS | Тесты для бага будут написаны первыми |

**Нарушения**: Нет. Все изменения соответствуют конституции.

## Project Structure

### Documentation (this feature)

```text
specs/005-fix-session-expiry/
├── spec.md              # Спецификация
├── plan.md              # Этот файл
├── research.md          # Исследование root cause
├── quickstart.md        # Шаги верификации
├── checklists/
│   └── requirements.md  # Чеклист качества спецификации
└── tasks.md             # Будет создан /speckit.tasks
```

### Source Code (изменяемые файлы)

```text
# Backend — конфигурация
config/
├── config.dev.yaml          # access_token_ttl: 15m → 2h
├── config.stage.yaml        # access_token_ttl: 15m → 2h, cookie_secure: true → false
└── config.prod.yaml         # access_token_ttl: 15m → 2h
backend/internal/config/
└── config.go                # Дефолт AccessTokenTTL: 15min → 2h

# Frontend — auth flow
frontend/src/
├── api/
│   ├── client.ts            # Синхронизация store при refresh failure, redirect с query
│   └── __tests__/
│       └── client.test.ts   # Новые тесты: store sync, redirect-back
├── stores/
│   ├── auth.ts              # Регистрация onAuthExpired callback
│   └── __tests__/
│       └── auth.test.ts     # Тесты callback регистрации
├── views/
│   └── LoginView.vue        # Обработка query.redirect после входа
├── components/
│   └── AppHeader.vue        # Обработка ошибок в handleLogout
└── router/
    └── index.ts             # (без изменений — guard уже корректный)
```

**Structure Decision**: Web application (backend + frontend). Изменения затрагивают только конфигурационные файлы на бэкенде и auth-flow на фронтенде. Новых файлов не создаётся, только модификация существующих.

## Complexity Tracking

> Нарушений конституции нет — секция не применяется.

# Research: PostgreSQL Table Locking for Atomic User Count

**Feature**: 004-fix-registration
**Date**: 2026-02-18

## Problem

`SELECT COUNT(*) FROM users FOR UPDATE` fails in PostgreSQL 17 with SQLSTATE 0A000: `FOR UPDATE is not allowed with aggregate functions`.

Goal: atomically count users and prevent concurrent INSERTs during first-user registration to guarantee exactly one admin.

## Decision

**`LOCK TABLE users IN EXCLUSIVE MODE`** followed by `SELECT COUNT(*) FROM users`.

## Rationale

1. **Полная корректность**: Блокирует конкурентные INSERT/UPDATE/DELETE на уровне таблицы. Работает на пустой таблице (в отличие от `FOR UPDATE` на строках).
2. **Простота**: Два стандартных SQL-запроса с очевидной семантикой.
3. **Совместимость с pgx/v5**: `tx.Exec()` + `tx.QueryRow().Scan()` — стандартный API.
4. **Производительность**: Блокировка держится миллисекунды. При масштабе 0-1000 пользователей влияние нулевое.
5. **Надёжность**: Enforcement на уровне СУБД — невозможно обойти случайно.

## SQL

```sql
-- Inside transaction:
LOCK TABLE users IN EXCLUSIVE MODE;
SELECT COUNT(*) FROM users;
-- if count == 0 → INSERT with role='admin', else role='user'
-- COMMIT releases the lock
```

## Alternatives Considered

| Approach | Verdict | Reason |
|----------|---------|--------|
| `SELECT 1 FROM users FOR UPDATE` + Go count | Rejected | Не блокирует пустую таблицу — race condition в основном сценарии |
| `SELECT COUNT(*) FROM (SELECT 1 FROM users FOR UPDATE) subq` | Rejected | Та же проблема — подзапрос не блокирует пустую таблицу |
| `pg_advisory_xact_lock()` | Rejected | Cooperative mechanism — хрупкий, требует дисциплины во всём коде |
| UPSERT / partial unique index | Rejected | Overengineering, требует изменения схемы для одноразовой операции |

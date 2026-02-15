# Research: Настройка GitHub CI/CD и GitHub Flow

**Feature Branch**: `001-github-ci-setup`
**Date**: 2026-02-15

## R1: Версия Go для CI

**Decision**: Go 1.25 (latest patch 1.25.7)

**Rationale**: Go 1.25 — текущая стабильная версия с LTS-поддержкой. Go 1.26 вышел 10 февраля 2026, но слишком нов для production CI. В `setup-go` указываем `'1.25'` — автоматически подтянет последний патч.

**Alternatives considered**:
- Go 1.26 (слишком новый, возможны нестабильности в тулчейне)
- `stable` (указывал бы на 1.26, нежелательно на раннем этапе)

## R2: Линтер и его версия

**Decision**: golangci-lint v2.9.0 через `golangci/golangci-lint-action@v9`

**Rationale**: golangci-lint v2 — текущая мажорная версия. Action v9 — последний релиз, поддерживает v2. Встроенное кеширование (отдельное от setup-go). Конфигурация v2 требует `version: "2"` в `.golangci.yml`.

**Alternatives considered**:
- golangci-lint v1.x (deprecated, не поддерживает Go 1.25+)
- Отдельные линтеры вручную (сложнее в поддержке, golangci-lint агрегирует 100+ линтеров)

## R3: Структура workflow

**Decision**: Один workflow файл `.github/workflows/ci.yml` с одним job `ci`, условные шаги через проверку `go.mod`

**Rationale**: Для homelab-проекта одного разработчика разделение на отдельные job'ы (lint, test, build) избыточно — создаёт overhead на запуск раннера. Один job с последовательными шагами быстрее и проще. Условная проверка `go.mod` обеспечивает graceful skip при отсутствии Go-кода (FR-007).

**Alternatives considered**:
- Отдельные job'ы lint/test/build (параллельное выполнение, но дольше из-за setup в каждом job, полезно для больших команд)
- Matrix build (разные версии Go — избыточно для одного разработчика)

## R4: Обработка отсутствия Go-кода

**Decision**: Проверка наличия `go.mod` через шаг `Check for Go code` с выводом в `GITHUB_OUTPUT`, условное выполнение всех Go-шагов через `if: steps.check.outputs.has_go == 'true'`

**Rationale**: Job-level `if: hashFiles('go.mod') != ''` может привести к тому, что весь job будет "skipped", и branch protection check не создаст запись. Лучше всегда запускать job, но условно пропускать шаги — так статус-чек всегда присутствует.

**Alternatives considered**:
- Job-level condition (check не появляется в PR при skipped job)
- Отдельный "gate" job (сложнее, избыточно)

## R5: Branch protection

**Decision**: GitHub Repository Rulesets API (`POST /repos/{owner}/{repo}/rulesets`) через `gh api`

**Rationale**: Rulesets — современный подход GitHub, заменяющий legacy branch protection. Более гибкий, поддерживает множественные правила. Для homelab-проекта: require status checks (CI) + запрет force-push + запрет удаления. Code review requirement НЕ включаем (один разработчик).

**Alternatives considered**:
- Legacy Branch Protection API (`PUT /repos/{owner}/{repo}/branches/{branch}/protection`) — устаревающий, менее гибкий
- Ручная настройка через GitHub UI — не автоматизируется, не воспроизводится

## R6: Кеширование

**Decision**: Встроенное кеширование `actions/setup-go@v6` (по `go.mod`) + встроенное кеширование `golangci-lint-action@v9`

**Rationale**: Оба action'а имеют встроенное кеширование, дополнительная настройка `actions/cache` не нужна. setup-go кеширует GOMODCACHE и GOCACHE, golangci-lint-action кеширует lint cache.

**Alternatives considered**:
- Ручной `actions/cache` (избыточно, встроенное кеширование достаточно)
- Отключение кеша (медленнее, нет причин)

## Версии компонентов

| Компонент | Версия | Примечание |
|-----------|--------|------------|
| Go | 1.25 | Последний патч 1.25.7 |
| golangci-lint | v2.9.0 | Требует config version: "2" |
| actions/checkout | v5 | Текущая стабильная |
| actions/setup-go | v6 | Встроенный кеш по go.mod |
| golangci/golangci-lint-action | v9 | Поддерживает golangci-lint v2 |

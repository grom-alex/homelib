# Research: Скрипты сборки и деплоя

## R-001: Адаптация kids-accounting скриптов к HomeLib

**Decision**: Прямая замена ссылок + расширение до 3 компонентов

**Rationale**: Скрипты из kids-accounting имеют зрелую архитектуру (общие библиотеки, единые exit codes, --help, health checks). Структура подходит для HomeLib с минимальными изменениями. Главное отличие — HomeLib имеет 3 Docker-образа (api, worker, frontend) вместо 2 (backend, frontend).

**Alternatives considered**:
- Написать скрипты с нуля → отвергнуто: дублирование усилий, kids-accounting скрипты уже проверены
- Использовать Makefile как единственный оркестратор → отвергнуто: Makefile уже есть, скрипты дополняют его для более сложных сценариев (SSH-деплой, registry push)

## R-002: Стратегия именования Docker-образов

**Decision**: `IMAGE_PREFIX=apps/homelib`, образы: `homelib-api`, `homelib-worker`, `homelib-frontend`

**Rationale**: Соответствует именованию в существующих Dockerfile (docker/backend/Dockerfile.api, Dockerfile.worker, docker/frontend/Dockerfile) и docker-compose файлах.

**Alternatives considered**:
- `homelib/api`, `homelib/worker`, `homelib/frontend` → отвергнуто: не совпадает с существующим именованием в CI

## R-003: Обновление старых скриптов

**Decision**: Минимальное обновление — добавить sourcing `lib/logging.sh` для единообразного вывода. Не переписывать полностью.

**Rationale**: Старые скрипты (backup-db, restore-db, dev-setup и др.) работают корректно. Полная переписка создаёт риск регрессий. Достаточно добавить общее логирование для консистентности.

**Alternatives considered**:
- Полностью переписать все старые скрипты → отвергнуто: избыточно, они просты и работают
- Оставить без изменений → отвергнуто: нарушает FR-003 (все скрипты используют общие библиотеки)

## R-004: Удаление deploy-old.sh

**Decision**: Удалить `deploy-old.sh` в рамках этого PR после проверки нового `deploy.sh`

**Rationale**: `deploy-old.sh` — это старый `deploy.sh` до замены. Новый `deploy.sh` покрывает все сценарии. Оставлять два файла создаёт путаницу.

**Alternatives considered**:
- Оставить для обратной совместимости → отвергнуто: проект в ранней стадии MVP, обратная совместимость не критична

## R-005: ShellCheck как линтер

**Decision**: Добавить ShellCheck в CI для проверки bash-скриптов не планируется в этом PR

**Rationale**: Текущий CI уже нагружен (Go lint, Go test, ESLint, Vitest, Docker build). Добавление ShellCheck — отдельная задача. В этом PR ограничимся ручной проверкой синтаксиса.

**Alternatives considered**:
- Добавить ShellCheck в CI сейчас → отложено на следующую итерацию

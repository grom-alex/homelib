# Quickstart: Скрипты сборки и деплоя HomeLib

## Предварительные требования

- Docker и Docker Compose v2
- Go 1.25+
- Node.js 22+
- SSH-доступ к серверу (для удалённого деплоя)

## Сценарий 1: Локальная сборка образов

```bash
# Собрать все образы (api, worker, frontend)
./scripts/build-local.sh

# Собрать только backend (api + worker)
./scripts/build-local.sh --component backend

# Собрать только frontend
./scripts/build-local.sh --component frontend

# Собрать с кастомным тегом
./scripts/build-local.sh --tag v1.0.0
```

**Ожидаемый результат**: Docker-образы `homelib-api`, `homelib-worker`, `homelib-frontend` видны в `docker images`.

## Сценарий 2: Локальный деплой (разработка)

```bash
# Поднять dev-окружение
./scripts/deploy-local.sh

# Остановить
docker compose -f docker/docker-compose.dev.yml down
```

**Ожидаемый результат**: Сервисы доступны, health check проходит.

## Сценарий 3: Сборка и публикация в registry

```bash
# Установить переменные registry
export DOCKER_REGISTRY=registry.example.com
export IMAGE_PREFIX=apps/homelib

# Собрать, протестировать, запушить
./scripts/build-and-push.sh v1.0.0
```

**Ожидаемый результат**: Образы запушены с тегами `v1.0.0`, `sha-<hash>`, `latest`.

## Сценарий 4: Деплой на staging

```bash
./scripts/deploy-stage.sh -t v1.0.0 -h staging.homelab.local
```

**Ожидаемый результат**: Контейнеры обновлены на сервере, health check проходит.

## Сценарий 5: Деплой на production

```bash
# Предварительный просмотр
./scripts/deploy-prod.sh -t v1.0.0 -h prod.homelab.local --dry-run

# Выполнить деплой
./scripts/deploy-prod.sh -t v1.0.0 -h prod.homelab.local
```

**Ожидаемый результат**: Бэкап текущего состояния, обновление контейнеров, подтверждение health check.

## Сценарий 6: Утилиты

```bash
# Бэкап базы данных
./scripts/backup-db.sh

# Восстановление из бэкапа
./scripts/restore-db.sh backups/homelib_20260217.sql.gz

# Запуск INPX-импорта
./scripts/import-inpx.sh

# Настройка dev-окружения с нуля
./scripts/dev-setup.sh
```

## Справка

Все скрипты поддерживают `--help`:

```bash
./scripts/build-local.sh --help
./scripts/deploy-prod.sh --help
```

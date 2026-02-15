# Quickstart: MVP HomeLib

**Branch**: `002-mvp-backend-init` | **Date**: 2026-02-15

## Prerequisites

- Docker & Docker Compose
- INPX-файл и ZIP-архивы с книгами (для импорта)

## Быстрый запуск

```bash
# 1. Клонировать и переключиться на ветку
git clone <repo> && cd homelib
git checkout 002-mvp-backend-init

# 2. Скопировать пример конфигурации
cp backend/config.example.yaml backend/config.yaml
# Отредактировать config.yaml: указать путь к библиотеке

# 3. Запустить всё через Docker Compose
docker compose up -d

# 4. Проверить что сервисы работают
docker compose ps
curl http://localhost/api/stats
```

## Сценарии интеграционного тестирования

### Сценарий 1: Полный цикл — от запуска до каталога

```bash
# 1. Запустить систему
docker compose up -d

# 2. Дождаться готовности
curl --retry 10 --retry-delay 2 http://localhost/api/stats

# 3. Зарегистрировать первого пользователя (станет admin)
curl -X POST http://localhost/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@home.lab","username":"admin","display_name":"Admin","password":"changeme123"}'
# → 201, получаем access_token

# 4. Запустить импорт
ACCESS_TOKEN="<token_from_step_3>"
curl -X POST http://localhost/api/admin/import \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → 202

# 5. Проверить статус импорта
curl http://localhost/api/admin/import/status \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → status: running → completed

# 6. Просмотреть каталог
curl "http://localhost/api/books?page=1&limit=20" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → 200, список книг с пагинацией

# 7. Полнотекстовый поиск
curl "http://localhost/api/books?q=Мастер+и+Маргарита" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → 200, результаты поиска

# 8. Скачать книгу
curl -O -J http://localhost/api/books/1/download \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → файл книги
```

### Сценарий 2: Аутентификация

```bash
# 1. Регистрация (первый = admin)
curl -X POST http://localhost/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@home.lab","username":"admin","display_name":"Admin","password":"changeme123"}' \
  -c cookies.txt
# → 201, access_token + refresh cookie

# 2. Вход
curl -X POST http://localhost/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@home.lab","password":"changeme123"}' \
  -c cookies.txt
# → 200, новый access_token

# 3. Обновление access-токена
curl -X POST http://localhost/api/auth/refresh \
  -b cookies.txt -c cookies.txt
# → 200, новый access_token

# 4. Запрос без токена → 401
curl http://localhost/api/books
# → 401 unauthorized

# 5. Выход
curl -X POST http://localhost/api/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -b cookies.txt
# → 200
```

### Сценарий 3: Идемпотентный импорт

```bash
# 1. Первый импорт
curl -X POST http://localhost/api/admin/import \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# Ждём завершения...

# 2. Запоминаем количество книг
curl http://localhost/api/stats \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → books_count: 636000

# 3. Повторный импорт того же файла
curl -X POST http://localhost/api/admin/import \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# Ждём завершения...

# 4. Количество книг НЕ изменилось (idempotent)
curl http://localhost/api/stats \
  -H "Authorization: Bearer $ACCESS_TOKEN"
# → books_count: 636000 (то же число)
```

### Сценарий 4: Фильтрация каталога

```bash
# По автору
curl "http://localhost/api/books?author_id=42" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# По жанру
curl "http://localhost/api/books?genre_id=5" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# По языку
curl "http://localhost/api/books?lang=en" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# По формату
curl "http://localhost/api/books?format=epub" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Комбинированный фильтр + сортировка
curl "http://localhost/api/books?lang=ru&format=fb2&sort=year&order=desc&page=1&limit=50" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Конфигурация (config.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  host: "postgres"
  port: 5432
  user: "homelib"
  password: "homelib"
  dbname: "homelib"
  sslmode: "disable"

auth:
  jwt_secret: "change-me-in-production"
  access_token_ttl: "15m"
  refresh_token_ttl: "720h"  # 30 days
  registration_enabled: true

library:
  inpx_path: "/library/librusec.inpx"
  archives_path: "/library"

import:
  batch_size: 3000
  log_every: 10000
```

## Docker Compose структура

```
services:
  postgres    — PostgreSQL 17 с pg_trgm
  api         — Go API server (порт 8080)
  nginx       — Reverse proxy (порт 80)
volumes:
  postgres_data — Персистентные данные БД
  /library      — Read-only mount с книгами
```

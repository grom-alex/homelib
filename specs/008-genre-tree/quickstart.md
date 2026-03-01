# Quickstart: Древовидная структура жанров

**Feature**: 008-genre-tree | **Date**: 2026-03-01

## Предварительные требования

- Docker + Docker Compose v2
- Go 1.25+
- Node.js 22 LTS + npm
- PostgreSQL 17 (через Docker или локально)

## Локальный запуск

### 1. Переключиться на ветку

```bash
git checkout 008-genre-tree
```

### 2. Запустить инфраструктуру

```bash
cd /home/test/projects/homelib
./scripts/deploy-local.sh local
```

Или вручную:

```bash
docker compose -f docker/docker-compose.dev.yml up -d postgres
```

### 3. Применить миграции

```bash
cd backend
go run cmd/worker/main.go --config ../config/config.dev.yaml
# Миграции применяются автоматически при старте worker/api
```

### 4. Запустить API-сервер

```bash
cd backend
go run cmd/api/main.go --config ../config/config.dev.yaml
# При старте автоматически загрузит дерево жанров из embedded .glst файла
```

### 5. Запустить фронтенд

```bash
cd frontend
npm install
npm run dev
```

Открыть http://localhost:5173 → вкладка «Жанры»

## Проверка работы дерева жанров

### API

```bash
# Получить дерево жанров
curl http://localhost:8080/api/genres | jq '.[0:3]'

# Проверить каскадную фильтрацию (книги по жанру и его потомкам)
# Сначала найти ID жанра «Фантастика»
GENRE_ID=$(curl -s http://localhost:8080/api/genres | jq '.[1].id')
curl "http://localhost:8080/api/books?genre_id=$GENRE_ID&limit=5" | jq '.total'

# Ручная перезагрузка дерева (admin)
curl -X POST http://localhost:8080/api/admin/genres/reload \
  -H "Authorization: Bearer $TOKEN"
```

### CLI

```bash
# Перезагрузка дерева жанров через worker
cd backend
go run cmd/worker/main.go --config ../config/config.dev.yaml --reload-genres
```

## Тестирование

### Backend

```bash
cd backend

# Все тесты
go test -race ./...

# Только парсер .glst
go test -v ./internal/glst/...

# Только сервис дерева жанров
go test -v ./internal/service/ -run TestGenreTree

# С покрытием
go test -race -coverprofile=coverage.out -coverpkg=./internal/... ./internal/...
go tool cover -func=coverage.out
```

### Frontend

```bash
cd frontend

# Все тесты
npm test

# Только компонент GenresTab
npx vitest run src/components/catalog/__tests__/GenresTab.test.ts

# С покрытием
npm run test:coverage
```

## Ключевые файлы для разработки

| Компонент | Путь |
|-----------|------|
| GLST парсер | `backend/internal/glst/parser.go` |
| GLST типы | `backend/internal/glst/types.go` |
| Миграция | `backend/migrations/006_genre_tree.up.sql` |
| Genre model | `backend/internal/models/genre.go` |
| Genre repo | `backend/internal/repository/genre.go` |
| Genre tree service | `backend/internal/service/genre_tree.go` |
| Import service | `backend/internal/service/import.go` |
| Genres handler | `backend/internal/api/handler/genres.go` |
| Admin handler | `backend/internal/api/handler/admin.go` |
| API router | `backend/internal/api/router.go` |
| Config | `backend/internal/config/config.go` |
| GenresTab | `frontend/src/components/catalog/GenresTab.vue` |
| SearchTab | `frontend/src/components/catalog/SearchTab.vue` |
| API client | `frontend/src/api/books.ts` |
| Catalog store | `frontend/src/stores/catalog.ts` |
| Types | `frontend/src/types/catalog.ts` |

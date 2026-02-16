# HomeLib — корневой Makefile
.PHONY: help dev stop logs build build-go build-frontend test test-backend test-frontend test-coverage lint lint-backend lint-frontend import backup restore deploy-stage deploy-prod clean docker-clean

# Переменные
DOCKER_DIR = docker
COMPOSE_DEV = $(DOCKER_DIR)/docker-compose.dev.yml
BACKEND_DIR = backend
FRONTEND_DIR = frontend

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# === Разработка ===

dev: ## Запустить dev-окружение
	docker compose -f $(COMPOSE_DEV) up -d
	@echo "HomeLib доступен на http://localhost"

stop: ## Остановить dev-окружение
	docker compose -f $(COMPOSE_DEV) down

logs: ## Показать логи
	docker compose -f $(COMPOSE_DEV) logs -f

# === Сборка ===

build: ## Собрать все Docker-образы
	docker compose -f $(COMPOSE_DEV) build

build-go: ## Собрать Go-бинарники
	./scripts/build.sh

build-frontend: ## Собрать фронтенд
	cd $(FRONTEND_DIR) && npm run build

# === Тестирование ===

test: test-backend test-frontend ## Запустить все тесты

test-backend: ## Тесты Go
	cd $(BACKEND_DIR) && go test -race ./...

test-frontend: ## Тесты Vue
	cd $(FRONTEND_DIR) && npx vitest run

test-coverage: ## Тесты с покрытием
	cd $(BACKEND_DIR) && go test -race -coverprofile=coverage.out ./...
	cd $(FRONTEND_DIR) && npm run test:coverage

# === Линтинг ===

lint: lint-backend lint-frontend ## Линтинг всего кода

lint-backend: ## Линтинг Go
	cd $(BACKEND_DIR) && golangci-lint run

lint-frontend: ## Линтинг Vue/TypeScript
	cd $(FRONTEND_DIR) && npm run lint

# === База данных ===

import: ## Запустить импорт INPX
	./scripts/import-inpx.sh

backup: ## Создать бэкап БД
	./scripts/backup-db.sh

restore: ## Восстановить БД из бэкапа (BACKUP=path)
	./scripts/restore-db.sh $(BACKUP)

# === Деплой ===

deploy-stage: ## Деплой на staging
	./scripts/deploy.sh stage

deploy-prod: ## Деплой на production
	./scripts/deploy.sh prod

# === Очистка ===

clean: ## Удалить артефакты сборки
	rm -rf build/
	rm -f $(BACKEND_DIR)/coverage.out
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/coverage

docker-clean: ## Удалить Docker-образы и volumes
	docker compose -f $(COMPOSE_DEV) down -v --rmi local

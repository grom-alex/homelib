# HomeLib

Веб-приложение домашней библиотеки для управления, поиска и чтения цифровых книжных коллекций. Развёртывается в хоумлабе, полностью self-hosted.

## Возможности (MVP)

- Импорт каталога из INPX-файлов (600K+ книг за 1-3 минуты)
- Каталог с фильтрацией по авторам, жанрам, сериям, языкам и форматам
- Скачивание книг из ZIP-архивов на лету (без предварительной распаковки)
- JWT-аутентификация с ролями (user/admin)
- Первый зарегистрированный пользователь становится администратором

## Технологии

| Компонент | Технология |
|-----------|-----------|
| Бэкенд | Go 1.25, Gin, pgx v5 |
| Фронтенд | Vue 3, Vuetify 3, Pinia, Vue Router |
| БД | PostgreSQL 17 + pg_trgm + tsvector |
| Контейнеризация | Docker Compose |
| Reverse Proxy | Nginx |

## Быстрый старт

### Требования

- Docker и Docker Compose
- Директория с библиотекой (INPX + ZIP-архивы)

### Установка

1. Клонировать репозиторий:
   ```bash
   git clone git@github.com:grom-alex/homelib.git
   cd homelib
   ```

2. Создать файл окружения:
   ```bash
   cp .env.example .env
   # Отредактировать .env: указать LIBRARY_PATH, сменить секреты
   ```

3. Запустить:
   ```bash
   make dev
   ```

4. Открыть http://localhost в браузере

5. Зарегистрировать первого пользователя (станет администратором)

6. Запустить импорт библиотеки:
   ```bash
   make import
   ```

## Структура проекта

```
homelib/
├── backend/          # Go API-сервер и воркер
├── frontend/         # Vue 3 SPA
├── docker/           # Dockerfile и compose-файлы
├── config/           # Конфигурации для разных окружений
├── scripts/          # Скрипты автоматизации
├── docs/             # Архитектурная документация
└── Makefile          # Основные команды
```

## Основные команды

```bash
make help             # Справка по всем командам
make dev              # Запустить dev-окружение
make stop             # Остановить
make logs             # Показать логи
make test             # Запустить все тесты
make lint             # Линтинг
make import           # Импорт INPX
make backup           # Бэкап БД
make deploy-prod      # Деплой в production
```

## Конфигурация

Приложение использует YAML-конфигурацию (`config/`) с переопределением через переменные окружения.

| Переменная | Описание | По умолчанию (dev) |
|------------|----------|-------------------|
| `DB_PASSWORD` | Пароль PostgreSQL | `homelib` |
| `JWT_SECRET` | Секрет JWT (мин. 32 символа) | dev-значение |
| `LIBRARY_PATH` | Путь к библиотеке | `/mnt/smb/media/...` |
| `NGINX_PORT` | Внешний порт | `80` |

## Окружения

| Окружение | Compose-файл | Описание |
|-----------|-------------|----------|
| dev | `docker/docker-compose.dev.yml` | Локальная разработка |
| stage | `docker/docker-compose.stage.yml` | Staging/тестирование |
| prod | `docker/docker-compose.prod.yml` | Production (resource limits, security headers) |

## Документация

Подробная архитектурная документация: [docs/homelib-architecture-v8.md](docs/homelib-architecture-v8.md)

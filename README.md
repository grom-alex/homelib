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
Полный пример с комментариями: [`backend/config.example.yaml`](backend/config.example.yaml).

### Переменные окружения

| Переменная | Описание | По умолчанию (dev) |
|------------|----------|-------------------|
| `DB_PASSWORD` | Пароль PostgreSQL | `homelib` |
| `DB_USER` | Пользователь PostgreSQL | `homelib` |
| `DB_NAME` | Имя базы данных | `homelib` |
| `DB_HOST` | Хост PostgreSQL | `postgres` |
| `JWT_SECRET` | Секрет JWT (мин. 32 символа) | dev-значение |
| `LIBRARY_PATH` | Путь к каталогу ZIP-архивов | `/mnt/smb/media/...` |
| `INPX_PATH` | Путь к INPX-файлу | см. config |
| `NGINX_PORT` | Внешний порт Nginx | `80` |

### Параметры YAML-конфигурации

| Секция | Параметр | Описание | По умолчанию |
|--------|----------|----------|-------------|
| `server` | `port` | Порт HTTP-сервера | `8080` |
| `server` | `host` | Адрес прослушивания | `0.0.0.0` |
| `database` | `host` | Хост PostgreSQL | — |
| `database` | `port` | Порт PostgreSQL | `5432` |
| `database` | `user` | Пользователь БД | — |
| `database` | `password` | Пароль БД | — |
| `database` | `dbname` | Имя базы данных | — |
| `database` | `sslmode` | Режим SSL (`disable` / `require` / `verify-full`) | `disable` |
| `auth` | `jwt_secret` | Секрет для подписи JWT (мин. 32 символа) | — |
| `auth` | `access_token_ttl` | Время жизни access token (JWT в памяти браузера) | `15m` |
| `auth` | `refresh_token_ttl` | Время жизни refresh token (httpOnly cookie). Определяет, как долго пользователь остаётся залогиненным без повторного ввода пароля | `720h` (30 дней) |
| `auth` | `registration_enabled` | Разрешена ли регистрация. Первый пользователь регистрируется всегда | `true` |
| `auth` | `cookie_secure` | Флаг Secure для cookie. `true` — только HTTPS, `false` — и HTTP. На HTTP-окружениях **обязательно** `false` | `false` |
| `library` | `inpx_path` | Путь к INPX-файлу внутри контейнера | — |
| `library` | `archives_path` | Путь к каталогу ZIP-архивов | — |
| `import` | `batch_size` | Размер пакета INSERT при импорте | `3000` |
| `import` | `log_every` | Логировать прогресс каждые N записей | `10000` |

## Окружения

| Окружение | Compose-файл | Описание |
|-----------|-------------|----------|
| dev | `docker/docker-compose.dev.yml` | Локальная разработка |
| stage | `docker/docker-compose.stage.yml` | Staging/тестирование |
| prod | `docker/docker-compose.prod.yml` | Production (resource limits, security headers) |

## Документация

Подробная архитектурная документация: [docs/homelib-architecture-v8.md](docs/homelib-architecture-v8.md)

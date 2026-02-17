# Research: MVP HomeLib

**Branch**: `002-mvp-backend-init` | **Date**: 2026-02-15

## R1. Go Web Framework: Gin

**Decision**: Gin v1.11.0+
**Rationale**: Gin — самый зрелый Go web-фреймворк с крупнейшей экосистемой (75K+ stars). Нативная поддержка route groups с middleware — идеально для JWT-аутентификации и изоляции admin-эндпоинтов. Отличная производительность для file streaming (скачивание из ZIP). Echo предлагает более идиоматичный API (error returns vs panic), но Gin выигрывает по экосистеме, документации и количеству готовых middleware.
**Alternatives considered**:
- Echo — чище API, хорошая документация, но меньшее комьюнити
- Chi — минималистичный, ближе к stdlib, но нет встроенных утилит (binding, validation)
- Fiber — основан на fasthttp (несовместим с net/http), ограничивает использование стандартных middleware

## R2. PostgreSQL Driver: pgx v5

**Decision**: jackc/pgx v5 (без database/sql обёртки)
**Rationale**: pgx — де-факто стандарт для Go + PostgreSQL в 2026. lib/pq в режиме maintenance. pgx на 50-100% быстрее, нативная поддержка COPY протокола (CopyFrom) для bulk-импорта 600K+ записей, встроенная поддержка PostgreSQL типов (JSONB, UUID, arrays, tsvector). Для batch upsert при импорте INPX — pgx.CopyFrom значительно превосходит multi-row INSERT.
**Alternatives considered**:
- lib/pq + database/sql — deprecated, нет COPY, медленнее
- sqlc — может использоваться поверх pgx для type-safe CRUD, но для MVP избыточен
- GORM — ORM добавляет абстракцию и снижает контроль над SQL; не нужен при наличии pgx

## R3. Database Migrations: golang-migrate v4

**Decision**: golang-migrate/migrate v4
**Rationale**: Наиболее распространённый инструмент миграций для Go. Простой, надёжный, встраивается в бинарник для containerized deployments. SQL-only миграции — без кодогенерации. Поддерживает PostgreSQL, Docker-friendly.
**Alternatives considered**:
- Goose — добавляет Go function migrations, избыточно для MVP
- Atlas — "Terraform for databases", мощный но сложный; оправдан только при автоматической генерации миграций
- Tern — минималистичный, но менее популярный

## R4. UI Component Library: Vuetify 3

**Decision**: Vuetify 3
**Rationale**: Более зрелая экосистема (39K+ stars vs 16K у Naive UI). Встроенные компоненты data table, tree view (для жанров), формы, пагинация — всё из коробки. Material Design обеспечивает консистентный UI без дизайнера. Нативная поддержка dark mode и responsive layout. Для каталожного приложения с таблицами и фильтрами Vuetify — оптимальный выбор.
**Alternatives considered**:
- Naive UI — легковеснее, хороший TypeScript, но слабее data table/tree view, меньше комьюнити
- Quasar — полнофункциональный фреймворк, избыточен для SPA
- PrimeVue — хорошие компоненты, но менее популярен в Vue 3 экосистеме

## R5. Frontend Toolchain: Vite + Vitest

**Decision**: Vite (build) + Vitest (test) + @vue/test-utils
**Rationale**: Стандартный toolchain для Vue 3 в 2026. create-vue генерирует Vite-based проект с Vitest из коробки. Vitest совместим с Jest API, быстрый (native ESM), поддерживает code coverage через v8/istanbul. @vue/test-utils — официальная библиотека для тестирования Vue-компонентов.
**Alternatives considered**:
- Vite+ — новый unified toolchain (анонсирован в начале 2026), ещё не стабилен; можно мигрировать позже

## R6. JWT Library: golang-jwt/jwt v5

**Decision**: golang-jwt/jwt/v5
**Rationale**: Стандартная библиотека для JWT в Go. Поддержка HS256/RS256, валидация claims, type-safe API. Форк оригинального dgrijalva/jwt-go с активной поддержкой.
**Alternatives considered**:
- lestrrat-go/jwx — более функциональный (JWE, JWK, JWS), но избыточен для простого JWT
- appleboy/gin-jwt — удобная обёртка, но привязывает к конкретному фреймворку; лучше использовать golang-jwt напрямую для гибкости

## R7. Конфигурация: YAML + env override

**Decision**: gopkg.in/yaml.v3 + os.Getenv для override
**Rationale**: Конституция требует YAML-конфигурацию (§6.V). gopkg.in/yaml.v3 — стандартный YAML парсер для Go. Environment variables для Docker-специфичных параметров (DATABASE_URL, JWT_SECRET). Не нужны сложные config-фреймворки типа Viper для MVP.
**Alternatives considered**:
- Viper — мощный, но тяжёлый; для MVP достаточно yaml.v3 + env
- koanf — лёгкая альтернатива Viper; рассмотреть если потребуется multi-source config

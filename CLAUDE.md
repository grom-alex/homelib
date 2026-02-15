# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Communication

Always respond in Russian (русский язык).

## Project Overview

HomeLib is a personal/home library web application for managing, searching, and reading digital book collections. It features AI-powered semantic search and automated metadata enhancement through distributed GPU processing.

**Current Status:** Architecture documentation only (in `/docs/`). Implementation not yet started.

## Development Workflow

Разработка ведётся по модели **GitHub Flow**:

- Основная ветка: `master` (всегда в deployable состоянии)
- Для каждой фичи/исправления создаётся отдельная ветка от `master`
- Именование веток: `NNN-short-name` (например, `001-github-ci-setup`)
- Изменения вливаются в `master` через Pull Request
- PR требует прохождения CI (сборка, тесты, линтеры) перед мержем
- Репозиторий: `git@github.com:grom-alex/homelib.git`

## Technology Stack

- **Backend:** Go with Gin/Echo framework
- **Frontend:** Vue 3 (Composition API), Vue Router, Pinia, Vuetify 3 or Naive UI
- **Database:** PostgreSQL with extensions: pgvector (semantic search), pg_trgm (fuzzy matching), tsvector (full-text search)
- **AI/ML:** Ollama instances running on Windows GPU machines in LAN (RTX 5060 Ti)
- **Containerization:** Docker Compose
- **Reverse Proxy:** Nginx

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│  HomeLib Server (Docker Compose)                            │
│  ┌─────────┐   ┌──────────────┐   ┌─────────────────────┐   │
│  │ Nginx   │──▶│ API Server   │──▶│ PostgreSQL          │   │
│  │         │   │ (Go/Gin)     │   │ + pgvector/pg_trgm  │   │
│  │         │──▶│ Vue 3 SPA    │   └─────────────────────┘   │
│  └─────────┘   └──────────────┘              ▲              │
│                      ▲                       │              │
│                      │            ┌──────────┴──────────┐   │
│                      │            │ Worker (Go)         │   │
│  ┌────────────────┐  │            │ • INPX import       │   │
│  │ /library (RO)  │  │            │ • Summary embedding │   │
│  │ .inpx + ZIPs   │  │            │ • LLM summarization │   │
│  └────────────────┘  │            └──────────┬──────────┘   │
└──────────────────────┼───────────────────────┼──────────────┘
                       │                       │ HTTP (LAN)
                       │              ┌────────┴────────┐
                       │              ▼                 ▼
                       │        ┌──────────┐     ┌──────────┐
                       │        │ Ollama   │     │ Ollama   │
                       │        │ PC #1    │     │ PC #2    │
                       │        │ (GPU)    │     │ (GPU)    │
                       │        └──────────┘     └──────────┘
```

## Key Components

### API Server (Go)
- RESTful API for catalog browsing, user data, authentication (JWT)
- On-the-fly file reading from ZIP archives (books not extracted)
- Format conversion: FB2→HTML, EPUB in browser, PDF.js, djvu.js
- Semantic search coordination via pgvector

### Worker (Go)
- INPX import and parsing (600K+ books in 1-3 minutes)
- Cover/metadata extraction from FB2/EPUB
- Summary extraction from multiple sources (annotation, TOC, first paragraphs, metadata)
- Embedding generation via distributed Ollama pool
- LLM summarization for books without annotations

### Ollama Pool
- Windows machines run standard Ollama only (no custom code)
- Health monitoring (30-sec heartbeat), least-connections load balancing
- Models: `nomic-embed-text` (768d vectors), `llama3`/`mistral` for LLM

### Database Schema
- Core tables: `books`, `authors`, `genres`, `series`, `collections`
- User tables: `users`, `user_books`, `reading_progress`, `shelves`
- Semantic search: `book_summaries` with `vector(768)` embedding

## Planned Directory Structure

```
homelib/
├── backend/
│   ├── cmd/
│   │   ├── api/        # API server entry point
│   │   └── worker/     # Worker process entry point
│   ├── internal/
│   │   ├── api/        # HTTP handlers
│   │   ├── models/     # Data models
│   │   ├── repository/ # Database access
│   │   ├── service/    # Business logic
│   │   ├── worker/     # Background tasks
│   │   ├── inpx/       # INPX parsing
│   │   └── archive/    # ZIP file handling
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── store/      # Pinia stores
│   │   └── services/   # API client
│   └── package.json
├── docker-compose.yml
└── docs/               # Architecture documentation (Russian)
```

## Book Formats

| Format | Reader Solution |
|--------|----------------|
| EPUB   | epub.js |
| FB2    | Convert to HTML on backend (Go `encoding/xml`) |
| PDF    | pdf.js (Mozilla) |
| DJVU   | djvu.js or backend conversion to PDF |

## INPX Format

INPX files are ZIP archives containing:
- `collection.info` - Collection metadata
- `version.info` - Version date (YYYYMMDD)
- `structure.info` - Field mapping for .inp files
- `*.inp` - Book records (fields separated by `\x04`, lines by `\r\n`)

Key parsing details:
- Authors: `Фамилия,Имя,Отчество:` (colon-separated, comma for name parts)
- Genres: `genre_code:` (colon-separated)
- Series with type: `Series Name[p]5` where `[a]`=author's, `[p]`=publisher's

## Authentication

- JWT-based: access token (15 min, in memory) + refresh token (30 days, httpOnly cookie)
- Roles: `user`, `admin`
- First registered user becomes admin (or via CLI command)

## Key Design Decisions

1. **Summary embedding instead of full-text:** 500K vectors vs 75M chunks (~80x storage reduction)
2. **Books stay in ZIPs:** Random access via Go `archive/zip`, no extraction needed
3. **No custom Windows code:** Standard Ollama with `OLLAMA_HOST=0.0.0.0:11434`
4. **Per-user data isolation:** All user data tied to `user_id` from JWT claims

## Documentation

Architecture documentation is in Russian in [docs/](docs/). The most current version is `homelib-architecture-v7.md`.

## Active Technologies
- Go 1.25 (latest patch 1.25.7) + GitHub Actions (`actions/checkout@v5`, `actions/setup-go@v6`, `golangci/golangci-lint-action@v9`) (001-github-ci-setup)
- N/A (конфигурационные файлы) (001-github-ci-setup)

## Recent Changes
- 001-github-ci-setup: Added Go 1.25 (latest patch 1.25.7) + GitHub Actions (`actions/checkout@v5`, `actions/setup-go@v6`, `golangci/golangci-lint-action@v9`)

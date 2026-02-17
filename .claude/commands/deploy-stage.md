---
description: Build Docker images and deploy to staging environment
---

## User Input

```text
$ARGUMENTS
```

## Goal

Build Docker images for backend and frontend, push them to the registry, and deploy to the staging environment.

## Execution Steps

### 1. Build and Push Images

Run the build script from the project root:

```bash
./scripts/build-and-push.sh
```

If user requests a version bump, use the appropriate flag:
- `./scripts/build-and-push.sh --bump patch` — bug fix
- `./scripts/build-and-push.sh --bump minor` — new feature
- `./scripts/build-and-push.sh --bump major` — breaking changes

Image tag format: `v<VERSION>-sha-<COMMIT>` (e.g., `v0.2.0-sha-abc1234`).
Version is read from `./version` file and optionally bumped with `--bump`.

This will:
- Run backend tests (`go test ./...`)
- Run frontend tests (`npm run test`)
- Build Docker images for api, worker, and frontend
- Push images to the configured registry
- Output the image tag

### 2. Deploy to Staging

After successful build, deploy to staging using the image tag from step 1:

```bash
./scripts/deploy-stage.sh --tag <IMAGE_TAG>
```

The script automatically reads `STAGE_HOST`, `NGINX_PORT`, `DB_PORT` from `.env` file.

### 3. Verify Deployment

After deployment completes:
- Report the staging URL (http://STAGE_HOST:NGINX_PORT)
- Confirm health check passed
- Suggest testing the deployed changes

## Configuration

The deploy script reads settings from `.env` file in project root:
- `STAGE_HOST` - Staging server hostname/IP
- `NGINX_PORT` - Nginx port on staging (default: 80)
- `DB_PORT` - PostgreSQL port on staging (default: 5432)
- `DOCKER_REGISTRY` - Docker registry URL
- `IMAGE_PREFIX` - Image path prefix in registry

## Error Handling

- If tests fail: Fix the failing tests before proceeding
- If build fails: Check Docker daemon and registry connectivity
- If deploy fails: Check SSH connectivity and disk space on staging server

## Arguments

If user provides arguments:
- `--skip-tests`: Pass `--skip-tests` to build script
- `--bump <part>`: Pass `--bump patch|minor|major` to build script
- `--tag <TAG>`: Skip build, deploy specific tag directly

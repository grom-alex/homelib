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

### 1. Pre-flight Checks

Run linting and tests before building:

```bash
# Backend
cd backend && gofmt -l . && go test ./...

# Frontend
cd frontend && npm run lint && npm run test:unit
```

If any check fails, stop and report the error. Do NOT proceed with build.

### 2. Build and Push Images

Run the build script:

```bash
./scripts/build-and-push.sh
```

This will:
- Run all tests
- Build Docker images for backend and frontend
- Push images to the configured registry
- Output the image tag (e.g., `sha-abc1234`)

### 3. Deploy to Staging

After successful build, deploy to staging using the image tag from step 2:

```bash
./scripts/deploy-stage.sh --tag <IMAGE_TAG>
```

The script automatically reads `STAGE_HOST` from `.env` file. No need to specify `--host`.

### 4. Verify Deployment

After deployment completes:
- Report the staging URL (from STAGE_HOST in .env)
- Confirm health check passed
- Suggest testing the deployed changes

## Configuration

The deploy script reads settings from `.env` file in project root:
- `STAGE_HOST` - Staging server hostname/IP
- `DOCKER_REGISTRY` - Docker registry URL
- `IMAGE_PREFIX` - Image path prefix in registry

## Error Handling

- If linting fails: Fix the issues first, then retry
- If tests fail: Fix the failing tests before proceeding
- If build fails: Check Docker daemon and registry connectivity
- If deploy fails: Check SSH connectivity and disk space on staging server

## Arguments

If user provides arguments:
- `--skip-tests`: Skip running tests (use with caution)
- `--tag <TAG>`: Use specific image tag instead of building new one

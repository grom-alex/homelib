# Quickstart: Fix User Registration

## Verification Steps

### 1. Run unit tests

```bash
cd backend && go test -race -v ./internal/repository/... ./internal/api/handler/...
```

### 2. Verify locally

```bash
# Rebuild and restart API
docker compose -f docker/docker-compose.dev.yml build api
docker compose -f docker/docker-compose.dev.yml up -d api

# Clean users table (if needed)
docker compose -f docker/docker-compose.dev.yml exec postgres psql -U homelib -d homelib -c "DELETE FROM refresh_tokens; DELETE FROM users;"

# Register first user
curl -s -X POST http://localhost:80/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@homelib.local","username":"admin","display_name":"Admin","password":"TestPass1234"}'

# Expected: HTTP 201 with {"user":{"id":"...","role":"admin",...},"access_token":"..."}
```

### 3. Verify race condition protection

```bash
# Two concurrent registrations — only one should get admin role
curl -s -X POST http://localhost:80/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user1@test.com","username":"user1","display_name":"User1","password":"TestPass1234"}' &
curl -s -X POST http://localhost:80/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user2@test.com","username":"user2","display_name":"User2","password":"TestPass1234"}' &
wait

# Check roles
docker compose -f docker/docker-compose.dev.yml exec postgres psql -U homelib -d homelib -c "SELECT email, role FROM users ORDER BY created_at;"
# Expected: exactly one admin, rest are users
```

## Files Modified

| File | Change |
|------|--------|
| `backend/internal/repository/user.go` | Replace `SELECT COUNT(*) FROM users FOR UPDATE` with `LOCK TABLE` + `SELECT COUNT(*)` |
| `backend/internal/api/handler/auth.go` | Add `log.Printf` for 500 errors in Register handler |
| `backend/internal/repository/user_test.go` | New file — unit tests for RegisterUser |

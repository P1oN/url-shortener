# URL Shortener (Go)

Production-ready URL shortener with PostgreSQL + Redis, structured JSON logs, API key auth, and a versioned HTTP API.

## Architecture Overview
Request flow:
`HTTP API (handlers)` → `service` → `repo (Postgres)` + `cache (Redis)`

## Features
- `/v1` API with API key auth.
- PostgreSQL persistence + Redis cache.
- Configurable timeouts, cache TTL, and DB pool sizing.
- Structured logs with request IDs.
- Explicit migration runner.

## Prerequisites
- Go 1.26+
- Docker + Docker Compose (for local DB/Redis)

## Configuration
Required env vars (see `.env.example`):
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB_NAME`
- `ADDRESS`, `BASE_URL`, `MIGRATIONS_PATH`, `API_KEY`

Optional tuning:
- `READ_TIMEOUT`, `WRITE_TIMEOUT`, `IDLE_TIMEOUT`, `GRACEFUL_SHUTDOWN_TIMEOUT`
- `REQUEST_TIMEOUT`, `CACHE_TTL`
- `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME`, `DB_CONN_MAX_IDLE_TIME`
- `ENABLE_SWAGGER` (enable `/swagger` docs UI and `/swagger/openapi.yaml`)

## Quick Start (Docker Compose)
1. Copy `.env.example` to `.env` and fill secrets.
2. Start services:
   ```bash
   docker compose up --build
   ```
3. Apply migrations:
   ```bash
   go run ./cmd/migrate --up
   ```
4. Run the API:
   ```bash
   go run ./cmd/server
   ```

## Local Commands
```bash
make run
make migrate-up
make migrate-down
make migrate-version VERSION=2
make test
make test-integration
make migrate-up-integration
make openapi-generate
make openapi-check
```

## API
All requests require:
```
Authorization: Bearer <API_KEY>
```

## Swagger / OpenAPI
- OpenAPI source: `api/openapi.yaml`
- Generated server contract: `internal/httpapi/openapi.gen.go` (via `go run ./cmd/openapi-gen`)
- Local docs UI (when enabled): `GET /swagger/`
- OpenAPI document endpoint: `GET /swagger/openapi.yaml`
- Set `ENABLE_SWAGGER=true` locally. Keep it `false` in production.

### Create short URL
`POST /v1/shorten`
```json
{
  "original_url": "https://example.com",
  "custom_code": "mycode",
  "expires_in_seconds": 3600
}
```

Response:
```json
{
  "short_url": "http://localhost:8080/abc123",
  "code": "abc123",
  "expires_at": "2026-02-15T10:00:00Z"
}
```

### Redirect
`GET /v1/{code}`

### Health
`GET /v1/health` (no auth)

### Error Format
```json
{
  "code": "invalid_url",
  "message": "invalid URL"
}
```

## Migrations
- Migrations are explicit (no auto-run on startup).
- `MIGRATIONS_PATH` must be a file URL, e.g. `file:///root/migrations`.
- Migration integration test:
  ```bash
  make migrate-up-integration
  ```

## Testing
```bash
go test ./...
INTEGRATION_TESTS=1 go test ./...
```

## Deployment Notes
- Use Docker Compose for local dev.
- Run migrations before deploying the server.

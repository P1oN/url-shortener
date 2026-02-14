# URL Shortener (Go)

Simple URL shortener service with PostgreSQL + Redis, structured logs, and a versioned HTTP API.

## Features
- PostgreSQL persistence + Redis cache.
- Configurable timeouts, cache TTL, and DB pool sizing.
- Explicit migration runner.

## Prerequisites
- Go 1.23+
- Docker + Docker Compose (for local DB/Redis)

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
```

## API
All requests require:
```
Authorization: Bearer <API_KEY>
```

### Create short URL
`POST /v1/shorten`
```json
{
  "original_url": "https://example.com",
  "custom_code": "mycode",
  "expires_in_seconds": 3600
}
```

### Redirect
`GET /v1/{code}`

### Health
`GET /v1/health` (no auth)

## Notes
- Migrations are explicit (no auto-run on startup).
- `MIGRATIONS_PATH` must be a file URL, e.g. `file:///root/migrations`.

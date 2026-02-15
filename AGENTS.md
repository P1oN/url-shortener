# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server`: main HTTP API entrypoint.
- `cmd/migrate`: migration runner for PostgreSQL.
- `internal/httpapi`: handlers and routing (`/v1/*`).
- `internal/httpapi/middleware`: auth + CORS middleware.
- `internal/service`: business logic and interfaces.
- `internal/repo`: database access (Postgres).
- `internal/cache`: Redis cache implementation.
- `internal/telemetry`: structured logging + request IDs.
- `internal/models`: domain models.
- `pkg/utils`: shared helpers.
- `config`: environment/config loading and validation.
- `migrations`: SQL migrations (`###_name.up.sql` / `###_name.down.sql`).
- `.env` / `.env.example`: runtime configuration (never commit real secrets).

## Build, Test, and Development Commands
- `go run ./cmd/server`: run the API locally (requires Postgres + Redis and a populated `.env`).
- `go run ./cmd/migrate --up`: apply all migrations.
- `go run ./cmd/migrate --down`: roll back all migrations.
- `go run ./cmd/migrate --version 2`: migrate to a specific version number.
- `go build ./cmd/server`: build the server binary.
- `go build ./cmd/migrate`: build the migration binary.
- `go test ./...`: run unit tests.
- `INTEGRATION_TESTS=1 go test ./...`: run integration tests (requires DB + Redis).
- `docker compose up --build`: start app + Postgres + Redis in containers.
- `make run`: run the API server.
- `make migrate-up`: apply all migrations.
- `make migrate-down`: roll back all migrations.
- `make migrate-version VERSION=2`: migrate to a specific version.
- `make test`: run unit tests.
- `make test-integration`: run integration tests.
- `make migrate-up-integration`: run the migration integration test.

## Coding Style & Naming Conventions
- Follow standard Go formatting; use `gofmt` before commit.
- Indentation is Go-standard tabs; keep line lengths reasonable.
- Naming: exported identifiers use `CamelCase`; unexported use `lowerCamel`.
- Tests live alongside packages and use `*_test.go` with `TestXxx` names.

## Testing Guidelines
- Use Go’s `testing` package; prefer table-driven tests.
- Integration tests are guarded by `INTEGRATION_TESTS=1` and require DB + Redis.
- Run `make migrate-up-integration` when migrations change.
- Run `go test ./...` locally and include any new test setup notes in your PR.

## Commit & Pull Request Guidelines
- Current history uses short, simple subject lines (e.g., “fix issues”); keep commits brief and imperative.
- PRs should include:
  - Summary of changes.
  - How you tested (commands + results).
  - Any config or migration impacts (e.g., new env vars, new migration files).

## Security & Configuration Tips
- Do not commit real secrets. Copy `.env.example` to `.env` and fill locally.
- `API_KEY` is required for all API calls (`Authorization: Bearer <API_KEY>`).
- `GET /v1/health` is unauthenticated for readiness checks.
- Ensure `MIGRATIONS_PATH` points to a file URL (e.g., `file:///root/migrations` in Docker).

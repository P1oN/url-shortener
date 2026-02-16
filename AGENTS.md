# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server`: main HTTP API entrypoint.
- `cmd/migrate`: migration runner for PostgreSQL.
- `cmd/openapi-gen`: local OpenAPI contract generator.
- `api/openapi.yaml`: source OpenAPI specification.
- `internal/httpapi`: handlers and routing (`/v1/*`).
- `internal/httpapi/openapi.gen.go`: generated OpenAPI server contract (do not hand-edit).
- `internal/httpapi/swagger`: embedded Swagger OpenAPI spec assets.
- `internal/httpapi/middleware`: auth + CORS middleware.
- `internal/service`: business logic and interfaces.
- `internal/repo`: database access (Postgres).
- `internal/cache`: Redis cache implementation.
- `internal/telemetry`: structured logging + request IDs.
- `internal/models`: domain models.
- `pkg/utils`: shared helpers.
- `config`: environment/config loading and validation.
- `migrations`: SQL migrations (`###_name.up.sql` / `###_name.down.sql`).
- `docker-compose.dev.yml`: Docker override for hot-reload development.
- `.air.toml`: hot-reload config for Go app container.
- `.env` / `.env.example`: runtime configuration (never commit real secrets).

## Build, Test, and Development Commands
- `go run ./cmd/server`: run the API locally (requires Postgres + Redis and a populated `.env`).
- `go run ./cmd/migrate --up`: apply all migrations.
- `go run ./cmd/migrate --down`: roll back all migrations.
- `go run ./cmd/migrate --version 2`: migrate to a specific version number.
- `go build ./cmd/server`: build the server binary.
- `go build ./cmd/migrate`: build the migration binary.
- `go run ./cmd/openapi-gen`: regenerate OpenAPI-derived artifacts.
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
- `make openapi-generate`: regenerate OpenAPI-derived artifacts.
- `make openapi-check`: verify generated OpenAPI artifacts are up to date.
- `make dev-up`: start Docker dev stack with hot reload.
- `make dev-migrate-up`: apply migrations in the dev app container.
- `make dev-logs`: tail dev app logs.
- `make dev-down`: stop Docker dev stack.

## Coding Style & Naming Conventions
- Follow standard Go formatting; use `gofmt` before commit.
- Indentation is Go-standard tabs; keep line lengths reasonable.
- Naming: exported identifiers use `CamelCase`; unexported use `lowerCamel`.
- Tests live alongside packages and use `*_test.go` with `TestXxx` names.

## Testing Guidelines
- Use Go’s `testing` package; prefer table-driven tests.
- Integration tests are guarded by `INTEGRATION_TESTS=1` and require DB + Redis.
- Run `make migrate-up-integration` when migrations change.
- Run `make openapi-generate` (or `make openapi-check`) when `api/openapi.yaml` changes.
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
- `ENABLE_SWAGGER` controls `/swagger` and `/swagger/openapi.yaml`; when enabled, these docs endpoints are served without API key auth. Keep it `false` in production.
- Ensure `MIGRATIONS_PATH` points to a file URL (e.g., `file:///root/migrations` in Docker).

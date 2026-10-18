# Repository Guidelines

## Project Structure & Module Organization

This repository is a Go/Fiber migration of the Smart Seaman mobile API.

- `cmd/api`: application entrypoint and graceful shutdown.
- `internal/app`, `internal/config`: application wiring and configuration.
- `internal/domain/<domain>`: domain models, services, repositories, and handlers.
- `internal/platform`: MySQL, object storage, SMTP, and external-service adapters.
- `internal/transport`: HTTP handlers, response helpers, and route registration.
- `internal/middleware`: authentication, API context, and request logging.
- `migrations`: SQL migrations; `README.md` and `MIGRATION.md` document operation and migration status.

Keep business rules in domain services, SQL in platform adapters, and HTTP concerns in transport handlers.

## Build, Test, and Development Commands

```bash
make run       # Run the API locally
make dev       # Run with Air hot reload
make test      # Run all Go tests
make vet       # Run go vet
make fmt       # Format all Go files
```

Copy `.env.example` to `.env` before local development. Use
`docker compose -f docker-compose.dev.yml up --build` for containerized hot reload.

## Coding Style & Naming Conventions

Use standard Go formatting with `gofmt`; tabs and idiomatic Go naming are expected.
Exported identifiers use PascalCase, unexported identifiers use camelCase, and
domain package names are lowercase. Prefer small, explicit functions and wrap
errors with useful context. Keep route paths and JSON field names compatible
with the existing Spring Boot API.

## Testing Guidelines

Tests use Go's standard `testing` package and are colocated with implementation
files using `_test.go` (for example, `internal/domain/auth/service_test.go`).
Add focused tests for validation, authorization, response envelopes, status
codes, and important side effects. Run `make test` before submitting changes;
there is no documented coverage threshold.

## Commit & Pull Request Guidelines

No Git history is available in this checkout to establish an existing commit
convention. Use short, imperative subjects such as `Add renewal status validation`
and keep unrelated changes separate.

Pull requests should explain the behavior change, identify affected routes or
domains, include test commands and results, and update `MIGRATION.md` or
`.env.example` when applicable. Do not include secrets, generated binaries, or
production credentials.

## Security & Configuration

Never commit JWT secrets, database credentials, storage keys, SMTP passwords, or
payment secrets. Inject them through environment variables or deployment secret
stores. Preserve request-log redaction and apply
`migrations/001_create_event_logs.sql` before deployments that require event logs.

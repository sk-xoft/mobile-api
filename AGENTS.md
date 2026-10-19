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

For authentication or database changes, also run `make vet`. Run targeted
`go test -race` checks when changing shared caches or concurrent behavior. Use
SQL mocks to verify ownership filters, transaction boundaries, rollback, query
counts, and error propagation; distinguish these checks from live MySQL tests.
For flows that perform multiple queries, include a one-connection pool regression
test when connection reuse or nested reads are affected.

## Authentication Requirements

- Activation and password-reset links use random, purpose-bound, expiring,
  single-use tokens stored as hashes. The legacy `uid` field carries the token,
  never a user UUID. Consume tokens and apply changes in one transaction.
- Activation requires a pending (`I`) account. Never restore email-only public
  activation or allow activation links to reactivate disabled accounts.
- Keep refresh credentials separate from JWT `jti`, which identifies the session.
  Store refresh tokens as hashes, enforce expiry, and rotate them atomically.
  Refresh responses include the new `refToken`; clients must persist it.
- Password changes/resets and account deactivation revoke sessions. Authentication
  and refresh must enforce active users, online sessions, and session expiry.
- Preserve redaction of passwords, JWTs, `refToken`, action-token `uid`, and file
  payloads. Do not expose credentials in logs, diagnostics, or benchmarks.
- Apply `migrations/002_secure_auth.sql` before deploying the hardened API, with
  old API instances stopped. It invalidates existing sessions; users must log in
  again. Coordinate client/link changes using `MIGRATION.md`.

## Performance Guidelines

- Reduce database round trips and unnecessary writes before changing security
  checks. Do not lower bcrypt cost or cache authentication decisions for speed.
- The auth adapter caches success descriptions in both languages for five minutes
  per process, coalescing concurrent misses. Failed reads and error descriptions
  are not cached. Account, password, session, and uploaded-file state stay fresh.
- When creating a login session, revoke only the user's online sessions while
  retaining the account lock and atomic session creation.
- Renewal validate-and-create loads files in batches by PROFILE/REQUEST scope.
  Avoid per-item queries; preserve ownership filters, ordering, completion/fix
  rules, and idempotency replay behavior.
- Close result sets before issuing another query. Do not use `db` for a second
  connection while holding a transaction; use `tx` for transactional reads, or
  perform independent reads before opening it. Keep sequence locks brief.
- Reuse metadata already read within the operation when safe. Treat lookup
  failures as errors; only explicit not-found results may trigger creation.
- Inspect execution plans, existing indexes, pool wait time, and end-to-end
  latency before adding indexes or increasing connection limits. Request event
  logging is synchronous and adds a database write after the handler.
- Report query-count reductions separately from measured latency. Local CPU
  benchmarks and SQL mocks do not establish production response times. Document
  relevant behavior, rollout requirements, and verification in `MIGRATION.md`.

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

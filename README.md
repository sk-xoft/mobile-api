# Smart Seaman Mobile API (Go)

Migration target for `../smart-seaman-mobile-api`, preserving the existing HTTP contract while replacing Spring Boot with Go and Fiber.

## Run

Requires Go 1.26 or newer. Both Docker build environments use Go 1.26.

1. Copy `.env.example` to `.env` and set the required values. The application loads `.env` automatically for local development.
2. Set the same database and JWT secret used by the Spring application.
3. Run `go mod download` and `make run`.

The migration is incremental. Routes that have not yet moved return HTTP 501 with code `MIGRATION_PENDING`; they must not receive production traffic until their domain is marked complete in `MIGRATION.md`.

## Database migrations

Database setup and upgrade scripts are collected under `migrations/`. Follow the
[migration guide](migrations/README.md) for legacy schema prerequisites, execution
order, and Go API migrations before deploying.

## Docker

Build the image:

```bash
docker build -t smart-seaman-mobile-api:latest .
```

Run it with environment variables supplied externally:

```bash
docker run --name smart-seaman-mobile-api --rm \
  --env-file .env \
  -p 8080:8080 \
  smart-seaman-mobile-api:latest
```

The container listens on port `8080`, runs as a non-root user, and exposes a Docker health check through `/health`. Do not copy secrets into the image; use deployment-secret or environment-variable injection.

## Hot reload

For local development, run directly; the Make target detects Air in `PATH`, detects the Go bin directory, or downloads it with `go run`:

```bash
make dev
```

To install Air explicitly:

```bash
make install-air
```

Or run the hot-reload development container:

```bash
docker compose -f docker-compose.dev.yml up --build
```

The production `Dockerfile` intentionally builds an immutable binary and does not include hot reload.



## API documentation
Endpoint specifications in Markdown, including the shared template, business
logic, request/response models and database query references, are indexed in
[docs/api/README.md](docs/api/README.md).

Set `API_DOCS_ENABLED=true` and open `/docs`. The spec is at `/docs/openapi.yaml` (JSON is valid YAML). Payment schemas are detailed; other endpoints currently provide a route inventory, not a complete integration contract.


```bash

codex resume 01a19023-de36-7a72-812b-cd5c6b834872

```

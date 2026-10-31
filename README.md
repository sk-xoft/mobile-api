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

## Deploy non-prod

Run the script on the Docker host from a checkout of this Go repository:

```bash
./deploy-non-prod.sh
```

Defaults: container `smart-seaman-mobile-api`, image
`smart-seaman-mobile-api:non-prod`, host port `30000` to container port `8080`,
and env file `/home/ssmuser/apps/config/mobile-api/non-prod/.env`. Override as needed:

```bash
ENV_FILE=/path/to/non-prod/.env HOST_PORT=30001 ./deploy-non-prod.sh
```

The env file must contain nonempty `DATABASE_DSN` (Go MySQL DSN format) and
`JWT_SECRET`, plus the Go integration settings in `.env.example`; legacy Spring
`DB_URL`, `DO_SPACES_*` and `MAIL_*` are not translated. The script injects the file
with Docker, sets `APP_PROFILE=non-prod`, and sets `APP_ADDRESS` to `CONTAINER_PORT`.

The multi-stage Dockerfile builds Go inside Docker; host Go/Maven and a Firebase
credential mount are not required. The script builds before replacing the existing
container, then waits up to `HEALTH_TIMEOUT=120` seconds for `/health`. `APP_NAME`,
`IMAGE_NAME` and `CONTAINER_PORT` can also be overridden. Health is a static process
probe, not proof that database/storage/payment integrations work. A failure after
replacement exits nonzero and does not automatically restore the previous container.

Logs go to stdout/stderr and Docker rotates JSON logs (10 MiB × 3). Read them with
`docker logs smart-seaman-mobile-api`; there is no `/logs` host mount. Apply required
SQL migrations manually using the [migration guide](migrations/README.md) before
deployment. The script does not execute SQL. `deploy-prod.sh` still contains the
legacy Java deployment flow and is not adapted by this change.

## Run dev (hot reload)

1. Install Go 1.26 or newer and run `go mod download` from the repository root.
2. Prepare `../config/mobile-api/non-prod/.env` using the Go configuration keys in
   `.env.example` (see the compatibility notes below).
3. Start the API with Air hot reload:

```bash
make dev
```

`make dev` loads `../config/mobile-api/non-prod/.env` by default (resolved to an
absolute path and passed to the application through `ENV_FILE`). In this checkout,
the default is `/Users/sk/works/products/smartseaman.com/source-code/config/mobile-api/non-prod/.env`.
It stops if the file is missing or unreadable. To select that file explicitly:

```bash
make dev ENV_FILE=/Users/sk/works/products/smartseaman.com/source-code/config/mobile-api/non-prod/.env
```

To use a repository-local `.env` or another configuration file:

```bash
make dev ENV_FILE=.env
make dev ENV_FILE=/Users/sk/works/products/smartseaman.com/source-code/config/mobile-api/non-prod/.env
```

The default listen address is `:8080`; change it with `APP_ADDRESS`. Air rebuilds
and restarts the API after Go source changes. Stop the process with `Ctrl+C`.

Existing process environment variables take precedence over values in the file.
The application reloads the selected file on each Air restart; changes to this
external file alone do not trigger hot reload, so restart `make dev` after editing
it. `make run` continues to use the
local `.env` unless `ENV_FILE` is explicitly set.

The selected file must use the Go variable names in `.env.example`. Legacy Spring
keys such as `DB_URL`/`DB_USERNAME`/`DB_PASSWORD`, `DO_SPACES_KEY`/`DO_SPACES_SECRET`
and `MAIL_USERNAME`/`MAIL_PASSWORD` are not mapped automatically. Supply
`DATABASE_DSN` in Go MySQL DSN format, `OBJECT_STORE_ACCESS_KEY`/
`OBJECT_STORE_SECRET_KEY` and `SMTP_USERNAME`/`SMTP_PASSWORD` respectively, plus
the endpoint/bucket/link settings required by your environment.

The API requires a nonempty `DATABASE_DSN` at startup. A file containing only the
legacy `DB_URL`/`DB_USERNAME`/`DB_PASSWORD` does not configure the Go database pool.
Use a DSN such as the following, replacing every placeholder with your settings:

```dotenv
DATABASE_DSN=app_user:replace-with-password@tcp(127.0.0.1:3306)/smart_seaman?parseTime=true&loc=Asia%2FBangkok
```

If startup reports `DATABASE_DSN is required`, set this key in the selected env
file and restart `make dev`. Previously, missing database configuration could
allow startup with a nil pool and cause a nil pointer panic on requests.

The Make target uses Air from `PATH` or the Go bin directory, falling back to
`go run` to download and run the pinned version. To install Air explicitly:

```bash
make install-air
```

Alternatively, run the development container, which uses the repository-local
`.env` through Docker Compose rather than the `make dev` default:

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

codex resume 01a19082-58e1-7250-aee8-ad246969e55f

```

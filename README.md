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


```bash

make dev ENV_FILE=/Users/sarunyook/workspaces/products/smartseaman.com/source-code/config/mobile-api/non-prod/.env
```
```


```

To use a repository-local `.env` or another configuration file:

```bash
make dev ENV_FILE=.env
jjjjjjjj
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
codex resume 01a19082-58e1-7250-aee8-ad246969e55f

```

# ดูว่าโปรเซสอะไรใช้ 8080 อยู่
lsof -nP -iTCP:8080

# หยุดทั้งหมดที่ใช้ 8080
kill $(lsof -ti tcp:8080)

# ถ้ายังไม่ยอมหยุด ให้บังคับปิด
kill -9 $(lsof -ti tcp:8080)


## Docker local development

### เตรียมก่อนรัน

1. เปิด Docker Desktop หรือ OrbStack และติดตั้ง Docker Compose
2. เข้าโฟลเดอร์ `mobile-api`
3. เตรียม `../config/mobile-api/non-prod/.env` โดยใช้ชื่อ config ของ Go ตาม `.env.example`
   เช่น `DATABASE_DSN` และ `JWT_SECRET`

### เริ่มและจัดการ Container

```bash
./run-docker-dev.sh          # Build image และเริ่มเบื้องหลัง
./run-docker-dev.sh logs     # ดู logs ต่อเนื่อง (Ctrl+C เพื่อออกจาก logs)
./run-docker-dev.sh status   # ดูสถานะ Container
./run-docker-dev.sh restart  # Build และสร้าง Container ใหม่ โหลด env ล่าสุด
./run-docker-dev.sh down     # หยุดและลบ Container โดยเก็บ cache volume
./run-docker-dev.sh --help   # ดูตัวเลือกทั้งหมด
```

API อยู่ที่ `http://localhost:8080` ตรวจ process health ได้ที่
`http://localhost:8080/health` (ไม่ได้ยืนยันการเชื่อมต่อฐานข้อมูลหรือบริการภายนอก)

ใช้ `Dockerfile.dev` และ Air hot reload เมื่อแก้ไฟล์ Go จะ build และ restart API
อัตโนมัติ โดย mount source ไปที่ `/app` ไม่ต้องติดตั้ง Go หรือ Air บนเครื่อง

### เลือก Environment file และ Port

ค่าเริ่มต้นใช้ `../config/mobile-api/non-prod/.env` และ host port `8080`
หากต้องการใช้ `.env` ภายในโปรเจกต์และพอร์ต `30000`:

```bash
ENV_FILE=.env HOST_PORT=30000 ./run-docker-dev.sh
ENV_FILE=.env HOST_PORT=30000 ./run-docker-dev.sh logs
ENV_FILE=.env HOST_PORT=30000 ./run-docker-dev.sh restart
ENV_FILE=.env HOST_PORT=30000 ./run-docker-dev.sh down
```

ใช้ overrides เดิมทุกครั้ง และรัน `restart` หลังแก้ environment file
ภายใน Container ใช้ `APP_ADDRESS=:8080` เสมอ ส่วน `HOST_PORT` กำหนดพอร์ตบนเครื่อง
เปิดรับเฉพาะ localhost โดย Compose inject config และไม่โหลด `.env` ซ้ำใน Container
ชื่อ Compose project เริ่มต้นคือ `smart-seaman-mobile-local`
เปลี่ยนได้ผ่าน `COMPOSE_PROJECT_NAME` โดยใช้ชื่อเดิมกับคำสั่งจัดการครั้งถัดไป

### ฐานข้อมูลและ Migration

สคริปต์ไม่สร้าง MySQL และไม่รัน SQL อัตโนมัติ ให้ apply สคริปต์ตาม
[migrations/README.md](migrations/README.md) แยกต่างหาก
หาก MySQL อยู่บนเครื่อง ให้ใช้ `host.docker.internal` แทน `localhost`
ใน `DATABASE_DSN` เพราะ localhost ภายใน Container หมายถึงตัว Container เอง

ไฟล์ config เริ่มต้นอาจชี้ไปบริการ non-prod หากต้องการแยกข้อมูล local ให้เลือก
ไฟล์ config ที่ชี้ฐานข้อมูลและบริการ local ของคุณ

### แก้ปัญหาเบื้องต้น

- `Docker is required`: ติดตั้ง Docker Desktop หรือ OrbStack และตรวจว่า `docker` อยู่ใน PATH
- `Start Docker Desktop or OrbStack first`: เปิด runtime และตรวจ `docker info` ให้สำเร็จ
- `Environment file is not readable`: ตรวจ path หรือระบุ `ENV_FILE` ให้ตรงไฟล์จริง
- พอร์ตถูกใช้งาน: เปลี่ยน `HOST_PORT` เช่น `HOST_PORT=30000 ./run-docker-dev.sh`
- API เริ่มไม่สำเร็จ: ดู `./run-docker-dev.sh logs` และตรวจ Go config/การเชื่อมต่อฐานข้อมูล


### Non-prod health check timeout

Deployment waits up to HEALTH_TIMEOUT (default 120 seconds), allowing transient
unhealthy/restarting states to recover. A running container alone does not confirm
API health. If it times out, inspect application and health probe output:

```bash
docker logs --tail 100 smart-seaman-mobile-api
docker inspect --format '{{json .State.Health}}' smart-seaman-mobile-api
docker exec smart-seaman-mobile-api wget -S -O- http://127.0.0.1:8080/health
```

Use the configured CONTAINER_PORT if it differs from 8080. Probe output identifies
HTTP failures versus connection failures; application logs identify startup errors.

### Health release metadata

`GET /health` returns `status`, `version`, `commit`, `buildDate`, `deployedAt`,
and `deployedBy`. It remains a process liveness check. Dates use UTC RFC3339.
Non-prod deployment records build/deploy timestamps automatically; version defaults
to the short Git commit, and deployer defaults to SUDO_USER or the current OS user.
Override the public release version and deployer label when deploying:

```bash
APP_VERSION=1.2.0 DEPLOYED_BY=release-team ./deploy-non-prod.sh
```

ตัวอย่างกำหนด version และผู้ deploy เอง:

```bash
APP_VERSION=1.2.0 DEPLOYED_BY=sarunyook ./deploy-non-prod.sh
```

หากไม่กำหนด `APP_VERSION` สคริปต์จะใช้ Git commit แบบสั้น เช่น `abc1234`
เป็น version อัตโนมัติ โดยไม่ได้เพิ่มเลข version เช่น `1.2.0` ให้อัตโนมัติ
หากไม่มี Git commit จะใช้ `unknown` ส่วนผู้ deploy ใช้ `SUDO_USER`
หรือชื่อ OS user ที่รันคำสั่ง:

```bash
./deploy-non-prod.sh
```

ตรวจข้อมูลหลัง deploy (ใช้ host port เริ่มต้น `30000`):

```bash
curl -fsS http://localhost:30000/health
```

ตัวอย่าง response เมื่อกำหนด version และผู้ deploy เอง
(commit และวันที่เป็นเพียงตัวอย่าง; วันที่เป็น UTC):

```json
{
  "status": "UP",
  "version": "1.2.0",
  "commit": "abc1234",
  "buildDate": "2026-11-14T08:00:00Z",
  "deployedAt": "2026-11-14T08:02:00Z",
  "deployedBy": "sarunyook"
}
```

These fields are public; use a team or CI label rather than private contact details.
For direct Docker builds pass APP_VERSION, APP_COMMIT and APP_BUILD_DATE as build
arguments; pass APP_DEPLOYED_AT and APP_DEPLOYED_BY at container creation.
Local runs default to version dev and unknown metadata. Recreating a container
without deployment metadata does not infer a deployment date from process startup.

### Persistent non-prod logs

`deploy-non-prod.sh` bind-mounts the host directory
`/home/ssmuser/apps/logs/mobile-api/non-prod` to `/app/logs`. Application JSON events
are written to both stdout (`docker logs`) and `api.log`. Files survive container
replacement. The logger rotates at approximately 10 MiB, retaining `api.log.1`
and `api.log.2` (about 30 MiB total; a single large event can exceed the limit).
Rotation limits retention even across deployments; this is not a permanent archive.

```bash
# Default log directory
APP_VERSION=1.2.0 DEPLOYED_BY=release-team ./deploy-non-prod.sh
# Override the host directory; use a dedicated directory per service instance
LOG_DIR=/home/ssmuser/apps/logs/mobile-api/custom ./deploy-non-prod.sh
# Follow the active file across rotation
sudo tail -F /home/ssmuser/apps/logs/mobile-api/non-prod/api.log
# Docker output remains available
docker logs -f --tail 100 smart-seaman-mobile-api
```

The image uses UID/GID `10001:10001`. Before stopping the old container, deployment
creates the dedicated directory, sets ownership and mode `750` through a temporary
root container, and verifies writes as the app user. Log files use mode `600`.
Set `LOG_DIR` in the application env file selected by `ENV_FILE`, without quotes:

```dotenv
LOG_DIR=/home/ssmuser/apps/logs/mobile-api/non-prod
```

The final host file is `LOG_DIR/api.log`. Precedence is a non-empty deployment
shell `LOG_DIR`, then the last `LOG_DIR=` entry in the env file, then the default
path above. Use a dedicated absolute directory; values are read literally with
no shell expansion. `LOG_FILE` remains the path inside the container and is set
automatically by the script to `/app/logs/api.log`.

Only one
application process may use a log file. For local runs, optional `LOG_FILE` enables
file logging; create its parent directory first. Leaving it empty uses stdout only.
Existing redaction applies equally to both destinations. Failures before logger
initialization and Go runtime stderr output remain available only in Docker logs.

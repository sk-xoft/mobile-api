# Migration status

The compatibility target is the Spring Boot service in `../smart-seaman-mobile-api`.

## Foundation

- [x] Fiber application bootstrap and graceful shutdown
- [x] MySQL connection pool configuration
- [x] CORS and security headers
- [x] Legacy response envelope
- [x] JWT authentication boundary
- [x] Route inventory
- [x] Message-code lookup and localized auth errors
- [ ] Static app-link files
- [x] Object storage client (S3-compatible Spaces)
- [x] SMTP email client
- [x] FCM and Omise clients
  - [x] FCM token update
- [ ] Scheduled notification refresh and user deletion jobs

## Domains

- [ ] Authentication and sessions
  - [x] Login, BCrypt verification, JWT issuance
  - [x] Refresh token and session token update
  - [x] Online-session enforcement for protected routes
  - [x] Activate user
  - [x] Change password
  - [x] Register with default certificates and confirmation email
  - [x] Reset-password email and forgot-password completion
  - [ ] Transaction log parity
- [x] Profile
  - [x] Read and update profile
  - [x] Profile image upload/download
  - [x] Deactivate/reactivate account
- [x] Master data and Thailand addresses
  - [x] Company and position master data
  - [x] Document and course dropdowns
  - [x] Province, district, and subdistrict lookup
- [x] Delivery addresses
  - [x] Default address read, create, and update
  - [x] Renewal request address snapshot
- [ ] Certificates and documents
  - [x] Certificate list, create, update, delete, edit, and file view
  - [x] Close-to-expiration list with display date and renewal-request flags
  - [ ] Request-item file upload and profile document endpoints
- [x] Document renewal workflow
  - [x] Active renewal statuses
  - [x] Renewal documents and active prices
  - [x] Request lifecycle, mobile update, delete/resubmit, timeline, stage, and payment transaction state
  - [x] Request validation
  - [x] Profile and renewal request-item upload/preview
- [x] Omise payments and webhook
  - [x] Payment transaction persistence and webhook state update
  - [x] Live Omise charge creation and signature verification
- [ ] Notifications and FCM
  - [x] FCM token update
  - [x] Notification list and read-state updates
- [x] Schools and courses
- [x] Forms
- [x] News, banners, and vouchers
  - [x] News list, detail, and image preview

An item is complete only after request validation, SQL behavior, response JSON, status codes, authorization, side effects, and focused compatibility tests match the Java implementation.

## Event logs

Before deploying the API, apply `migrations/001_create_event_logs.sql` to the application database. The API writes `http.request.completed` events with request/trace IDs, actor ID, route, status, and duration. Request/response bodies remain in structured application logs; Base64 files, multipart content, binary responses, and secrets are excluded from both logging paths.

## Request-item upload behavior

- Uploads validate the active master item before writing object storage. Profile
  uploads additionally require `storage_scope=PROFILE`.
- Writes lock the master row in a transaction to serialize uploads through this
  API. Renewal writes also lock and recheck the request/item status. Database
  tables must use a transactional engine with row-lock support, such as InnoDB.
- Switching document types marks previous types `file_uploaded=0` in the same
  transaction; their rows and stored files remain available as historical data.
  For ID cards, `MAIN` and the `FRONT`/`BACK` pair supersede each other.
- Preview and list return active files in upload-time/ID order. Completion requires
  an uploaded `MAIN` file for passport/general, or an uploaded ID-card `MAIN` or
  both `FRONT` and `BACK` of the matching type. Legacy mixed types use the latest
  uploaded type until a new upload resolves the active set.
- Profile insert affecting zero rows fails instead of reporting a generated ID
  as saved. Failed saves attempt to remove the newly uploaded object. An uncertain
  commit retains the object for reconciliation; existing historical objects are
  not removed automatically.
- Existing HTTP status codes and `MA...` response codes are retained for mobile
  compatibility. Size validation also runs in the service (maximum 10 MiB).
- Run `go test ./internal/domain/requestitem -cover` for the regression suite.

### Authentication security update

Apply `migrations/002_secure_auth.sql` before starting the updated API. Stop old
API instances first, apply the migration once, and deploy only the updated
instances. The migration invalidates existing sessions; users must log in again.
Do not roll back to the old API because it accepts UUID-based password resets.

- Registration links retain `?uid=`, but the value is now a random, single-use
  activation token (24 hours), not the user's UUID. Only pending (`I`) accounts
  can activate. Old activation/reset links are invalid.
- `POST /v1/profile/active` now requires `{"uid":"<activation token>"}`.
  Email-only reactivation is no longer supported. `GET /v1/activate-user?uid=...`
  consumes the same activation token. Suspended/deactivated accounts require a
  separate authorized recovery process.
- Reset links retain `?uid=` and the reset body retains `uid`, now containing a
  random, single-use reset token (30 minutes). Request a fresh email to reset.
  Both token purposes are stored as SHA-256 hashes and consumed transactionally.
- `GET /v1/refresh-token?refToken=...` retains its route. Its response now includes
  both `token` and a rotated `refToken`. Clients must save the new refresh token
  after every successful refresh and serialize concurrent refresh requests.
  Refresh credentials expire 30 days after login; refresh does not extend this.
  JWT `jti` contains only the session ID and cannot be used to refresh.
- Password change/reset and deactivation revoke sessions. Middleware and refresh
  enforce active accounts, online sessions and session expiry. Password changes
  also invalidate pending reset links. Login requires a fresh password afterward.
- Application logs redact `refToken` and email action `uid` fields. Reverse proxies
  and access logs must also omit query strings for refresh/activation links.
- Regression checks use local service tests and SQL mocks; no live MySQL, SMTP or
  deployed API was exercised. Test the email landing pages and refresh-token
  persistence in the mobile client during rollout.

### Login performance

- The MySQL auth adapter caches successful response descriptions in both
  languages for five minutes per API process. Concurrent cache misses share one
  query. Failed reads and error descriptions are not cached; database changes to
  success descriptions become visible after expiry or process restart. Passwords,
  account status, sessions, and authentication results are never cached.
- Login revokes only online sessions for the user. The account lock, atomic
  session creation, refresh expiry and session revocation guarantees remain.
  No additional migration or client change is needed for these optimizations.
- For production diagnosis, inspect existing indexes with `SHOW INDEX FROM
  m_mobile_users`, `SHOW INDEX FROM t_session`, and `SHOW INDEX FROM
  m_message_code`. Check execution plans for username lookup, mobile UUID lookup,
  session updates filtered by `(USER_ID, IS_ONLINE)`, and message-code lookup
  before adding indexes. Also inspect connection-pool wait time: the default
  maximum is five connections; increase only against measured concurrency and
  the database's connection budget.
- Request logging writes an event to MySQL synchronously after the handler;
  handler-only timing excludes that write. Measure end-to-end p50/p95 latency
  from the client as well as SQL latency before drawing conclusions.
- Local benchmark commands:
  `go test ./internal/platform/mysql/auth ./internal/domain/auth -run '^$' -bench 'BenchmarkSuccessMessage|BenchmarkLoginCost8' -benchtime=100x -benchmem`.
  On the development Apple M1 Pro, login CPU work with bcrypt cost 8 took about
  16.1 ms/op, excluding database and logging. The message benchmark uses SQL
  mocks and does not measure MySQL latency or end-to-end login improvement.
  Existing password hashes retain their own bcrypt cost; no cost was reduced.

### Validate-and-create performance

- `POST /v1/document-renewals/requests/validate-and-create` now loads uploaded
  files in batches: at most one PROFILE query and one REQUEST query instead of
  one per item. User and request ownership filters, upload-time/ID ordering,
  response fields and completion/fix rules are retained. Item/address result sets
  close before the next query, allowing this flow to use a single DB connection.
- For an existing request with N items across S scopes, response assembly uses
  `4 + S` queries instead of `4 + N` (S is at most two). Ten items across both
  scopes therefore use six response queries instead of fourteen. These counts
  exclude authentication, initial request lookup and middleware event logging.
- Creating a request reuses the document metadata and the newly inserted request
  identifiers/contact fields, avoiding two additional post-commit reads. The
  default address is read before opening the creation transaction, reducing the
  sequence-lock hold time and avoiding a second pooled connection inside the
  transaction. Request/items/address snapshot/history writes still commit
  together, and response file data is read fresh rather than cached.
- Lookup failures other than not-found now stop validation instead of falling
  through to draft creation. Address read failures likewise stop creation;
  absence of a default address remains supported.
- SQL-mock regression tests cover mixed-scope batching, query count independent
  of item count, file grouping/order/status, idempotency replay, database errors,
  new-draft response reuse and a one-connection pool. No live MySQL latency or
  concurrent-create load result is claimed. No migration/client change is needed.

### Payment attempts, expired QR codes and reconciliation

Apply `migrations/003_payment_attempts.sql` once with old payment writers stopped.
It imports CHARGE history into durable attempts, adds scoped idempotency/provider
reference constraints and flags ambiguous imported requests for manual review.
Do not discard history or force past duplicate-key errors. Verify against the
actual legacy schema and an Omise test account before release; see the detailed
[rollout checklist](docs/payments.md#rollout-and-release-checks).

- POST `/v1/document-renewals/:requestId/payments` replays an existing key, or
  reserves a new attempt before contacting Omise. A replacement needs a new key
  and a confirmed expired/failed previous attempt. An uncertain result blocks
  replacements and is reconciled without blindly creating another charge.
- GET on the same path resumes the current payment. Existing transaction GET
  remains available. Responses add `serverTime`, `qrExpired`, `nextAction` and
  `retryAfterSeconds` and retain saved QR/provider details where usable.
- Webhook, polling and background checks share charge verification and atomic
  ledger/request/timeline updates. Repeated/out-of-order results cannot downgrade
  success or duplicate fulfillment. Additional receipts require manual review.
- Cancellation is blocked during active/uncertain or successful payments; hard
  deletion is blocked for any payment history. All use the parent request lock.
- `PAYMENT_WORKER_ENABLED=true` runs a bounded, leased worker on each API instance.
  `OMISE_LIVE_MODE` must match the key. Shutdown stops workers before closing DB.
- Client action mapping and errors: [payments.md](docs/payments.md).
  Machine-readable contract: [payments.openapi.json](docs/payments.openapi.json).
- Local tests use real HTTP with a fake provider plus SQL mocks, including lost
  responses, parallel creation, expiry, replay, late/duplicate successes, rollback,
  one-connection pools and worker shutdown. They do not verify the live Omise
  account or apply the migration to a real MySQL instance.

## API documentation
Optional embedded Swagger UI: set `API_DOCS_ENABLED=true`. Disabled by default; when enabled, documentation is publicly readable. No database migration required. Non-payment request/response schemas remain incomplete.

## Go 1.26 toolchain
The module requires Go 1.26.0 or newer; production and development Docker builders use `golang:1.26-alpine`. Rebuild container images when deploying. No database migration or API contract changes are required.

# Renewal payments and QR replacement

Contract: [payments.openapi.json](payments.openapi.json). The API retains the
normal `code` / `description` / `data` envelope and existing authentication and
`Accept-Language`, `devicemodel`, `correlationid` headers.

## Mobile flow

1. On opening/resuming checkout, GET `/v1/document-renewals/{requestId}/payments`.
   A 404 means no attempt exists. POST to that same path to create one.
2. POST body contains `paymentMethod: "promptpay"` and a new UUID
   `idempotencyKey`. Persist that key before sending. Network retries use the
   **same** key. Store the returned `transactionId`.
3. Use `serverTime` and `expiresAt` (UTC) to display the countdown. Poll the GET
   collection or `/payments/{transactionId}` every five seconds while checkout
   is foregrounded; also check on resume and after returning from a banking app.
4. At countdown zero, hide the QR and check the backend. Do not infer failed
   payment or generate a new key solely from the phone's clock.
5. Only `nextAction=CREATE_NEW_QR` enables the manual “สร้าง QR ใหม่” button. A
   new click uses a **new** key. The old key always refers to the old attempt.

| `nextAction` | UI |
| --- | --- |
| `WAIT_FOR_PAYMENT` | Display the existing QR / bank authorization link and countdown. |
| `CHECK_PAYMENT` | Hide expired QR, show “กำลังตรวจสอบการชำระเงิน”, poll after `retryAfterSeconds`. Do not POST a replacement. |
| `CREATE_NEW_QR` | Stop polling; show expired/failed result and manual replacement button. |
| `COMPLETE` | Stop polling; display success and refresh the renewal request. |
| `CONTACT_SUPPORT` | Stop polling/creating; show a support message with transaction/request number. |

`qrExpired` is a display hint, not proof of nonpayment. `status` describes the
particular attempt, whereas `nextAction` also considers payment of the parent
request. A historical expired attempt can therefore have `nextAction=COMPLETE`
if another attempt paid the request. GET collection chooses the accepted paid
attempt first, then the active attempt, then the latest attempt.

`qrCodeDownloadUri` and `authorizeUri` are omitted once expired, failed, paid,
or requiring review. Successful replay restores QR data from the stored charge
rather than returning only the transaction summary.

## Errors and concurrency

- 400 `PAYMENT_INVALID_INPUT`: missing/oversized key or unsupported method shape.
- 400 `PAYMENT_DOCUMENTS_INCOMPLETE`: required request files are incomplete.
- 404 `PAYMENT_NOT_FOUND`: unknown payment/request or wrong owner.
- 409 `PAYMENT_IDEMPOTENCY_CONFLICT`: same key with a different payment payload.
- 409 `PAYMENT_IN_PROGRESS`: another key is already active/uncertain; `data`
  contains the payment to follow. Fetch it; do not rotate keys automatically.
- 409 `PAYMENT_ALREADY_PAID`: fetch current payment and refresh the request.
- 409 `PAYMENT_REQUEST_NOT_PAYABLE`: inactive/wrong-state/invalid-amount request.
- 409 `PAYMENT_RECONCILIATION_REQUIRED`: contact support; no new charge is allowed.
- 503 `PAYMENT_UNAVAILABLE`: provider credentials/service are not configured.

Creation can return HTTP 200 with `status=PROCESSING` after a provider timeout,
HTTP error, or dropped response. The durable attempt already exists. Retrying
its key or querying status reconciles it; it does **not** blindly POST again.
This conservative policy also applies to provider 409/429/5xx failures. It can
require operator investigation if no charge can be located. Never promise
exactly-once money movement: the guarantee is durable local deduplication,
provider reference reuse, and detection/reconciliation of unexpected receipts.

## Persistence and reconciliation

The new `mobile_payment_attempts` table records reservations before provider
HTTP calls. `(request_id,idempotency_key)`, provider reference, and non-null
charge ID are unique. `mobile_payment_request_state` stores the active attempt,
accepted payment, and review flag. The existing `m_payment_transaction` ledger
is inserted/updated after a verified charge is available.

All reservation/finalization/cancellation/deletion operations lock the parent
request first. Network calls occur outside DB transactions. A 60-second DB lease
prevents concurrent provider calls across API instances; recent checks are
coalesced for five seconds. Creation transitions RESERVED -> SENT before HTTP;
only RESERVED may create. SENT without a known charge ID searches the provider
list by server-generated reference/metadata, with a bounded creation-time window
and pagination (maximum 2,000 results per lookup). An inconclusive lookup stays
PROCESSING and raises an overdue log after 15 minutes.

Provider outcomes from POST, polling, webhook and worker share verification and
atomic finalization. Validate charge ID/reference, integer-satang amount, THB and
test/live mode. SUCCESS never regresses. Ledger, accepted-payment pointer,
renewal transition and timeline commit together; repeated success does not add
another timeline entry. Expiry/failure keeps the request PAYMENT_PENDING.

The worker starts with the API, scans up to 50 due attempts every 30 seconds,
uses two workers and a ten-second provider timeout. Failures back off from
30 seconds to five minutes plus jitter. Terminal expired/failed attempts are
checked every five minutes for 24 hours after attempt creation. Webhooks remain
usable afterward. Shutdown cancels workers and waits before closing the DB.

A second receipt or payment after cancellation requires manual review; preserve
both records and never auto-refund or fulfill the request twice. Manual review
flags prevent further creation. Operators must inspect provider/dashboard and
ledger evidence, resolve funds under the business's refund policy, and audit any
manual flag changes; there is intentionally no public override endpoint.

Provider references:
- [PromptPay](https://docs.omise.co/promptpay): QR expiration, completion flow,
  and the provider's void/refund limitations.
- [Charges API](https://docs.omise.co/charges-api): `ref_id`, charge retrieval,
  status/expiry fields and paginated charge lists.
- [Webhooks](https://docs.omise.co/api-webhooks): retrieve charges to verify events.

## Rollout and release checks

1. Back up the schema and payment tables. Inspect `SHOW CREATE TABLE` for the
   legacy ledger and requests and test `003_payment_attempts.sql` against a copy
   of the actual MySQL schema. Existing legacy schema is not defined in this repo. The migration enables
   strict SQL mode during import so oversized identifiers/keys fail rather than
   being silently truncated.
2. Audit duplicate CHARGE `(request_id,idempotency_key)` and nonempty
   `provider_charge_id` values, UUID-shaped IDs, key lengths <=100, currencies,
   modes and active payment counts. Do not delete duplicates to bypass a failure.
3. Stop old API/payment writers. Apply migration **once**, then deploy the updated
   API with `PAYMENT_WORKER_ENABLED=false` while inspecting imported records.
   Imported multi-active/multi-success or unsupported-provider records require
   review. Missing/invalid legacy provider data must be investigated.
4. Verify the configured Omise account/API version honors `ref_id` deduplication
   in test mode, including a duplicate reference, dropped response, delayed
   webhook and expiry. The implementation does not assume replaying a duplicate
   reference returns the original charge; it searches for the existing charge.
5. Match `OMISE_LIVE_MODE` to the secret-key mode, enable the worker, and release
   the updated mobile flow. Monitor `payment reconciliation overdue`, verification
   failures, worker failures, pending age and request review flags. Account for
   provider rate limits and the database's connection budget.
6. If rollback is needed, stop payment creation/writers and keep evidence for
   reconciliation. Do not restore the old payment writer while new attempts exist:
   it does not honor the reservation or review state.

Local checks: `make test`, `make vet`, and
`go test -race ./internal/domain/payment ./internal/platform/mysql/payment ./internal/platform/payment ./internal/domain/renewal`.
HTTP regression tests run the real provider client against a local HTTP server;
SQL mocks verify transaction order and rollback. These do not replace MySQL
migration/integration tests or the Omise account checks above. No live payment
credentials or real charges are needed for local tests.

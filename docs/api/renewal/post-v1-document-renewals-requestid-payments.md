# POST /v1/document-renewals/:requestId/payments

สร้างหรือ replay payment attempt

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.Payment`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestId` | string | ต้องมี; ID ของคำขอ (`m_document_request.id`) ไม่ใช่ requestNo |

Content-Type: `application/json`; model [payment.Input](../SCHEMAS.md#paymentinput)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `paymentMethod` | `string` | ใช่ |
| `idempotencyKey` | `string` | ใช่ |
| `returnUri` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `platformType` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "paymentMethod": "promptpay",
  "idempotencyKey": "example-request-key",
  "returnUri": "",
  "platformType": ""
}
```

## Business logic

1. paymentMethod trim/lowercase: promptpay หรือขึ้นต้น mobile_banking_; idempotencyKey 1–100 ตัว
2. provider ต้อง ready; ตรวจเจ้าของ request; key เดิมต้อง fingerprint input ตรงกัน
3. reconcile attempt เดิมก่อนตัดสินใจสร้างใหม่
4. Reserve ล็อก request/state; ต้อง active PAYMENT_PENDING, amount>0, ยังไม่ paid/review และ active attempt เดิมจบ EXPIRED/FAILED
5. ตรวจ required REQUEST items มีไฟล์ uploaded และไม่ถูก fix ก่อน reserve
6. สร้าง RESERVED attempt แล้ว lease/mark SENT ก่อนเรียก Omise; timeout ไม่ใช่หลักฐานว่า charge ไม่เกิด
7. ตรวจ amount/currency THB/livemode/provider reference; reconcile ledger/request/timeline แบบ atomic
8. คืน nextAction,qrExpired,serverTime,retryAfterSeconds; replacement ใช้ key ใหม่หลังยืนยัน FAILED/EXPIRED

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_master_request_item` | READ |
| `m_document_request` | READ, UPDATE |
| `m_document_request_item_files` | READ |
| `m_document_request_items` | READ |
| `m_document_status` | READ |
| `m_document_transaction` | INSERT |
| `m_payment_transaction` | READ, INSERT, UPDATE |
| `mobile_payment_attempts` | READ, INSERT, UPDATE |
| `mobile_payment_request_state` | READ, INSERT, UPDATE |

Query path: `Request` → `ByKey` → `Current` → `Get` → `Reserve` → `Acquire` → `MarkSent` → `Apply` → `Retry` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [payment.Request](../queries/payment.md#paymentrequest)
- [payment.readRequest](../queries/payment.md#paymentreadrequest)
- [payment.ByKey](../queries/payment.md#paymentbykey)
- [payment.readAttempt](../queries/payment.md#paymentreadattempt)
- [payment.Current](../queries/payment.md#paymentcurrent)
- [payment.Get](../queries/payment.md#paymentget)
- [payment.Reserve](../queries/payment.md#paymentreserve)
- [payment.Acquire](../queries/payment.md#paymentacquire)
- [payment.MarkSent](../queries/payment.md#paymentmarksent)
- [payment.Apply](../queries/payment.md#paymentapply)
- [payment.Retry](../queries/payment.md#paymentretry)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Provider authentication failures

Omise ตอบ 401/403 จะคืน HTTP 503 `PAYMENT_UNAVAILABLE` และ `Retry-After: 30`
ไม่มี success data snapshot; ไม่ใช้ HTTP 401 ของ JWT แทน error ของ provider
พัก provider calls 30 วินาทีต่อ process หลังพบปัญหา แล้วอนุญาตให้ probe ใหม่
ระหว่างติด lease จะไม่คืน success จาก PROCESSING ที่ cache ไว้ถ้า process ยังมี auth failure
provider HTTP สำเร็จจะ clear failure; restart ล้างเฉพาะสถานะพักนี้ ไม่ได้แก้ key
เก็บ attempt เดิมและความไม่แน่นอน ไม่สร้าง replacement หรือ ledger ปลอม
`m_payment_transaction` จะมีข้อมูลเมื่อได้ verified charge แล้วเท่านั้น
ดูรายละเอียดใน [payment guide](../../payments.md)

## Transaction และ side effects

- Transaction: reserve/mark sent/apply แยก transactions; ไม่เปิด transaction ค้างระหว่าง HTTP provider call
- Side effects: อาจสร้าง/อ่าน charge ที่ Omise; client ต้องทำตาม nextAction

## Response และ errors

HTTP 200; `data` เป็น [payment.Response](../SCHEMAS.md#paymentresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestId": "00000000-0000-4000-8000-000000000001",
    "requestNo": "261000001",
    "transactionId": "00000000-0000-4000-8000-000000000002",
    "transactionNo": "example",
    "status": "PROCESSING",
    "channel": "PROMPTPAY",
    "paymentMethod": "promptpay",
    "amount": 0,
    "currency": "THB",
    "provider": "OMISE",
    "serverTime": "2026-10-31T00:00:00Z",
    "qrExpired": false,
    "nextAction": "CHECK_PAYMENT",
    "retryAfterSeconds": 0
  }
}
```

Domain errors: 400 PAYMENT_INVALID_INPUT/PAYMENT_DOCUMENTS_INCOMPLETE; 404 PAYMENT_NOT_FOUND; 409 PAYMENT_IDEMPOTENCY_CONFLICT/PAYMENT_ALREADY_PAID/PAYMENT_IN_PROGRESS/PAYMENT_REQUEST_NOT_PAYABLE/PAYMENT_RECONCILIATION_REQUIRED; 503 PAYMENT_UNAVAILABLE

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [renewal/handler.go](../../../internal/domain/renewal/handler.go)
- [renewal/lifecycle_handler.go](../../../internal/domain/renewal/lifecycle_handler.go)
- [renewal/service.go](../../../internal/domain/renewal/service.go)
- [renewal/lifecycle_service.go](../../../internal/domain/renewal/lifecycle_service.go)
- [payment service](../../../internal/domain/payment/service.go), [payment client guide](../../payments.md), [payment OpenAPI](../../payments.openapi.json)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

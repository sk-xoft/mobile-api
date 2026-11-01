# GET /v1/document-renewals/:requestId/payments

อ่านและ reconcile payment

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.CurrentPayment`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestId` | string | ต้องมี; ID ของคำขอ (`m_document_request.id`) ไม่ใช่ requestNo |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. ใช้ requestId จาก column id ของคำขอ (ไม่มี uuid.Parse validation แยก) และตรวจ user ownership; PaymentStatus ตรวจ transactionId อยู่ใน request+user เดียวกัน
2. CurrentPayment เลือก paid attempt ก่อน active attempt ก่อนล่าสุด
3. GET สามารถ reconcile: lease, mark SENT และสร้าง charge หากยัง RESERVED หรือ retrieve/find charge ที่ Omise
4. ตรวจ provider result แล้ว UPDATE attempts/state/ledger; success อาจเปลี่ยน request เป็น PENDING_DOCUMENT_REVIEW และเพิ่ม PAYMENT_SUCCESS timeline
5. SUCCESS ไม่ถูก downgrade; success ซ้ำไม่ fulfill ซ้ำ; ambiguous receipt ส่ง CONTACT_SUPPORT
6. QR หมดเวลาอย่างเดียวยังไม่อนุญาต replacement จน provider ยืนยัน FAILED/EXPIRED

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ, UPDATE |
| `m_document_status` | READ |
| `m_document_transaction` | INSERT |
| `m_payment_transaction` | READ, INSERT, UPDATE |
| `mobile_payment_attempts` | READ, UPDATE |
| `mobile_payment_request_state` | READ, UPDATE |

Query path: `Request` → `Current` → `Get` → `Acquire` → `MarkSent` → `Apply` → `Retry` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [payment.Request](../queries/payment.md#paymentrequest)
- [payment.readRequest](../queries/payment.md#paymentreadrequest)
- [payment.Current](../queries/payment.md#paymentcurrent)
- [payment.readAttempt](../queries/payment.md#paymentreadattempt)
- [payment.Get](../queries/payment.md#paymentget)
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

- Transaction: reconciliation writes ใช้ request lock และ transaction; provider call อยู่นอก transaction
- Side effects: GET มี side effects และอาจเรียก Omise Create/Retrieve/Find

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

Domain errors: 404 PAYMENT_NOT_FOUND; 503 PAYMENT_UNAVAILABLE; 500 MA00013 database/provider processing error

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

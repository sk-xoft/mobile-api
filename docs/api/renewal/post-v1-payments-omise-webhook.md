# POST /v1/payments/omise/webhook

รับ Omise webhook

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.Webhook`
- Authentication: ไม่ใช้ Bearer JWT; ตรวจ signature ของ Omise
- Headers: ไม่ผ่าน APIContext; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| header | `Omise-Signature` | string | signature ของ raw payload |
| header | `Omise-Signature-Timestamp` | string | timestamp สำหรับ verify signature |

Content-Type: `application/json`; model [payment.WebhookInput](../SCHEMAS.md#paymentwebhookinput)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `key` | `string` | ประเภท event อื่น/ว่างตอบ 200 |
| `data` | `WebhookData` | ต้องมี data.id สำหรับ supported events; ไม่มี ID ตอบ 500 |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "key": "charge.complete",
  "data": {
    "id": "example"
  }
}
```

## Business logic

1. JSON ต้อง valid; ตรวจ Omise-Signature และ Omise-Signature-Timestamp ก่อนประมวลผล
2. สนใจ charge.complete/create/update/expire; event อื่นตอบ 200
3. ดึง charge จาก Omise ใหม่ด้วย data.id; ไม่เชื่อสถานะจาก payload โดยตรง
4. ค้น attempt ด้วย provider charge/ref; หาไม่พบตอบ 200
5. ตรวจ amount/currency/livemode/reference; mismatch บันทึก retry/review
6. Apply อัปเดต ledger/state/request/timeline แบบเดียวกับ polling; SUCCESS ไม่ downgrade และไม่ fulfill ซ้ำ

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

Query path: `Find` → `Retry` → `Apply` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [payment.Find](../queries/payment.md#paymentfind)
- [payment.readAttempt](../queries/payment.md#paymentreadattempt)
- [payment.Retry](../queries/payment.md#paymentretry)
- [payment.readRequest](../queries/payment.md#paymentreadrequest)
- [payment.Apply](../queries/payment.md#paymentapply)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: Apply หรือ Retry transaction; HTTP provider call อยู่นอก transaction
- Side effects: Omise retrieve; payment/request state อาจเปลี่ยน

## Response และ errors

ใช้ `SendStatus`: 200 ตอบข้อความสถานะ `OK`; 400/401/500 ตอบ status text ไม่ใช่ MA envelope

Domain errors: 400 malformed JSON; 401 signature ไม่ถูกต้อง; 500 processing error รวม event ที่ไม่มี data.id

Endpoint นี้ไม่มี domain error envelope เพิ่มเติมนอกเหนือจากที่ระบุ; framework errors อาจใช้ status/text ของ Fiber

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [renewal/handler.go](../../../internal/domain/renewal/handler.go)
- [renewal/lifecycle_handler.go](../../../internal/domain/renewal/lifecycle_handler.go)
- [renewal/service.go](../../../internal/domain/renewal/service.go)
- [renewal/lifecycle_service.go](../../../internal/domain/renewal/lifecycle_service.go)
- [payment service](../../../internal/domain/payment/service.go), [payment client guide](../../payments.md), [payment OpenAPI](../../payments.openapi.json)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

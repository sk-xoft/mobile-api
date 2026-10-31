# DELETE /v1/document-renewals/requests/:requestNo

ยกเลิกคำขอแบบ soft delete

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.DeleteRequest`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. อนุญาตเฉพาะ PAYMENT_PENDING ของเจ้าของ
2. ล็อก request และตรวจ status ไม่เปลี่ยนระหว่างทาง
3. ถ้ามี attempt ที่ยังไม่ EXPIRED/FAILED หรือ review_required ให้ปฏิเสธ PAYMENT_IN_PROGRESS
4. เปลี่ยนเป็น CANCELLED, is_active=NO และเขียน CANCEL timeline

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ, UPDATE |
| `m_document_status` | READ |
| `m_document_transaction` | INSERT |
| `mobile_payment_attempts` | READ |

Query path: `ownedRequest` → `transition` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.ownedRequest](../queries/renewal.md#renewalownedrequest)
- [renewal.transition](../queries/renewal.md#renewaltransition)
- [renewal.statusID](../queries/renewal.md#renewalstatusid)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: request lock, payment guard, status update และ timeline transaction เดียว
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.TransitionResponse](../SCHEMAS.md#renewaltransitionresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestNo": "261000001",
    "fromStatus": "example",
    "toStatus": "example",
    "action": "example"
  }
}
```

Domain errors: MA00005 documentStatus; MA00016 request; 409 PAYMENT_IN_PROGRESS

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [renewal/handler.go](../../../internal/domain/renewal/handler.go)
- [renewal/lifecycle_handler.go](../../../internal/domain/renewal/lifecycle_handler.go)
- [renewal/service.go](../../../internal/domain/renewal/service.go)
- [renewal/lifecycle_service.go](../../../internal/domain/renewal/lifecycle_service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

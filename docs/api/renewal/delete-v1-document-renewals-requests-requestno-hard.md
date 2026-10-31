# DELETE /v1/document-renewals/requests/:requestNo/hard

ลบคำขอและข้อมูลลูกถาวร

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.HardDelete`
- Authentication: Bearer JWT + active online session
- Authorization: ไม่ตรวจว่าคำขอเป็นของผู้ใช้ใน JWT; ผู้ใช้ที่ผ่าน authentication สามารถ hard delete คำขอของบัญชีอื่นได้
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. trim/uppercase requestNo; ต้องยาว 1–20 ตัว
2. ค้น requestNo ที่ is_active=YES โดยไม่กรอง mobile_user_uuid แล้วล็อกแถวใน transaction; อ่าน status master ผ่าน transaction เดียวกัน
3. ตรวจ information_schema แล้วลบ m_delivery/m_dept_submission ถ้ามี
4. ลบ payment state/attempts, request files/items/address snapshot, timeline, ledger แล้ว request
5. พฤติกรรมโค้ดปัจจุบันไม่มี payment-history guard หรือจำกัด document status ใน hardDelete แม้ docs/payments.md และ MIGRATION.md ระบุว่าบล็อกเมื่อมีประวัติ payment

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `information_schema.tables` | READ |
| `m_delivery` | DELETE |
| `m_dept_submission` | DELETE |
| `m_document_request` | READ, DELETE |
| `m_document_request_delivery_address` | READ, DELETE |
| `m_document_request_item_files` | DELETE |
| `m_document_request_items` | READ, DELETE |
| `m_document_status` | READ |
| `m_document_transaction` | READ, DELETE |
| `m_payment_transaction` | READ, UPDATE, DELETE |
| `mobile_payment_attempts` | READ, DELETE |
| `mobile_payment_request_state` | READ, DELETE |

Query path: `hardDelete` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.hardDelete](../queries/renewal.md#renewalharddelete)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ลบ parent/children transaction เดียว; ต้อง review ความคลาดเคลื่อนของ payment guard ก่อนใช้งาน
- Side effects: ลบข้อมูลถาวรใน DB; ไม่มีการลบ objects ใน storage

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

Domain errors: MA00005 requestNo; MA00016 request

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [renewal/handler.go](../../../internal/domain/renewal/handler.go)
- [renewal/lifecycle_handler.go](../../../internal/domain/renewal/lifecycle_handler.go)
- [renewal/service.go](../../../internal/domain/renewal/service.go)
- [renewal/lifecycle_service.go](../../../internal/domain/renewal/lifecycle_service.go)

ทดสอบ handler และ SQL mocks: ผู้ใช้คนละบัญชีลบได้, active filter, one-connection transaction และ rollback เมื่อผิดพลาด ยังไม่ได้ส่ง DELETE จริงหรือ execute SQL กับ live MySQL

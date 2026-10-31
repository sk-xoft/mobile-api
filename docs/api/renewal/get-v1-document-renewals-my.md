# GET /v1/document-renewals/my

รายการคำขอต่ออายุของฉัน

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.ListRequests`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| query | `offSet` | integer | ต้องมี; >=0; page size 10 |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. offSet ต้องเป็น integer >=0
2. กรอง user+is_active=YES; เรียง submitted_at DESC, LIMIT 10 OFFSET
3. นับ total แยก query; last=offset+จำนวนรายการ >= total

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ |
| `m_document_status` | READ |

Query path: `list` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.list](../queries/renewal.md#renewallist)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.Page](../SCHEMAS.md#renewalpage); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "itemTotal": 0,
    "last": false,
    "items": []
  }
}
```

Domain errors: MA00005 offSet

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

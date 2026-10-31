# GET /v1/documents/profile-items

รายการเอกสารโปรไฟล์และไฟล์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `requestitem.List`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. อ่าน active PROFILE master items เรียง sort_order/item code
2. โหลด files ของ user ทีละ item; เลือก type ที่ upload ล่าสุด
3. documentStatus เป็น MISSING หรือ COMPLETE ตามชุด slot; fileUploaded อาจเป็น 1 แต่ชุด FRONT/BACK ยังไม่ครบ

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_master_request_item` | READ |
| `m_document_profile_request_item` | READ |

Query path: `ProfileItems` → `Files` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [requestitem.ProfileItems](../queries/requestitem.md#requestitemprofileitems)
- [requestitem.Files](../queries/requestitem.md#requestitemfiles)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [requestitem.Item](../SCHEMAS.md#requestitemitem) array; envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": []
}
```

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [requestitem/handler.go](../../../internal/domain/requestitem/handler.go)
- [requestitem/service.go](../../../internal/domain/requestitem/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

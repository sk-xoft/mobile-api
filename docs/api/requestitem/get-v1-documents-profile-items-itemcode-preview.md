# GET /v1/documents/profile-items/:itemCode/preview

metadata ไฟล์เอกสารโปรไฟล์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `requestitem.Preview`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `itemCode` | string | ต้องมี; ดู validation ใน Business logic |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. ใช้ itemCode และ user จาก session อ่าน uploaded profile files
2. เลือกเฉพาะ type ที่ upload ล่าสุด; คืน metadata/filePath ไม่ใช่ binary download
3. ไม่มีไฟล์คืน files=[]

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_profile_request_item` | READ |

Query path: `Files` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [requestitem.Files](../queries/requestitem.md#requestitemfiles)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [requestitem.PreviewResponse](../SCHEMAS.md#requestitempreviewresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "documentMasterRequestItemCode": "MRI002",
    "files": []
  }
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

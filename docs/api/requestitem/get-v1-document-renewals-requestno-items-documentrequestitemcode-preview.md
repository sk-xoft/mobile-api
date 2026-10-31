# GET /v1/document-renewals/:requestNo/items/:documentRequestItemCode/preview

metadata ไฟล์ของคำขอต่ออายุ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `requestitem.RenewalPreview`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |
| path | `documentRequestItemCode` | string | ต้องมี; ดู validation ใน Business logic |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. ตรวจ requestNo+เจ้าของ+active+itemCode ก่อนอ่านไฟล์
2. อ่าน uploaded request files และเลือก type ล่าสุด; ส่ง metadata/filePath ไม่ใช่ bytes
3. alias /items/preview อ่าน request_no จาก query แต่ไม่ได้อ่าน itemCode จาก query; item ที่ส่งไป repository เป็นค่าว่าง จึงมักคืน MA00016

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ |
| `m_document_request_item_files` | READ |
| `m_document_request_items` | READ |
| `m_document_status` | READ |

Query path: `Renewal` → `Files` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [requestitem.Renewal](../queries/requestitem.md#requestitemrenewal)
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

Domain errors: MA00016 request/item ไม่พบ; alias มีข้อจำกัด itemCode ว่างใน handler ปัจจุบัน

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [requestitem/handler.go](../../../internal/domain/requestitem/handler.go)
- [requestitem/service.go](../../../internal/domain/requestitem/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# POST /v1/document-renewals/:requestNo/items/:documentRequestItemCode/file

อัปโหลดไฟล์ของคำขอต่ออายุ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `requestitem.RenewalUpload`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |
| path | `documentRequestItemCode` | string | ต้องมี; ดู validation ใน Business logic |

Content-Type: `multipart/form-data`

| Field | Type | กติกา |
|---|---|---|
| `file` | binary | ต้องมี; 1–10 MiB; PNG/JPEG/PDF |
| `documentType` | string | default GENERAL; MRI001 ต้องเลือก ID_CARD/PASSPORT |
| `slotCode` | string | default MAIN; ID_CARD รองรับ FRONT/BACK |

## Business logic

1. file ต้องมีขนาด 1–10 MiB และ magic bytes PNG/JPEG/PDF
2. documentType/slotCode default GENERAL/MAIN; trim และ uppercase
3. MRI001 รับ ID_CARD MAIN/FRONT/BACK หรือ PASSPORT MAIN; item อื่นรับ GENERAL MAIN
4. ตรวจ master item active ก่อนเขียน storage
5. ค้น requestNo+เจ้าของ+active+itemCode; อนุญาต PAYMENT_PENDING หรือ PENDING_APPLICANT_CORRECTION เฉพาะ item FIX
6. PUT object แล้วล็อก master/request/item และตรวจสถานะซ้ำก่อน upsert
7. บันทึก request-scoped file; supersede type/slot ที่ถูกแทนแล้วคืน active files/complete

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_master_request_item` | READ |
| `m_document_request` | READ |
| `m_document_request_item_files` | READ, INSERT, UPDATE |
| `m_document_request_items` | READ |
| `m_document_status` | READ |

Query path: `ValidateItem` → `Renewal` → `UpsertRenewal` → `Files` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [requestitem.ValidateItem](../queries/requestitem.md#requestitemvalidateitem)
- [requestitem.Renewal](../queries/requestitem.md#requestitemrenewal)
- [requestitem.UpsertRenewal](../queries/requestitem.md#requestitemupsertrenewal)
- [requestitem.saveLocked](../queries/requestitem.md#requestitemsavelocked)
- [requestitem.upsertRenewal](../queries/requestitem.md#requestitemupsertrenewal)
- [requestitem.Files](../queries/requestitem.md#requestitemfiles)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: master/request lock, upsert และ supersede อยู่ใน transaction เดียว
- Side effects: PUT object; cleanup object ใหม่เมื่อ save ล้มเหลวที่ไม่ใช่ uncertain commit; ไม่ลบ object เก่า

## Response และ errors

HTTP 200; `data` เป็น [requestitem.UploadResponse](../SCHEMAS.md#requestitemuploadresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "itemCode": "MRI002",
    "storageScope": "PROFILE",
    "documentType": "GENERAL",
    "complete": false,
    "files": []
  }
}
```

Domain errors: MA00005 file/type/slot/status; MA00016 request/item

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [requestitem/handler.go](../../../internal/domain/requestitem/handler.go)
- [requestitem/service.go](../../../internal/domain/requestitem/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

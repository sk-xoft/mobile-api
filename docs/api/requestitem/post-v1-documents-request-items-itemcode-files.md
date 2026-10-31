# POST /v1/documents/request-items/:itemCode/files

อัปโหลดไฟล์เอกสารโปรไฟล์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `requestitem.Upload`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `itemCode` | string | ต้องมี; ดู validation ใน Business logic |

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
4. ตรวจ master item active ก่อนเขียน storage และ scope ต้องเป็น PROFILE
5. PUT object ใหม่ แล้ว save ภายใต้ master row lock
6. upsert ตาม user+item+type+slot และ reset check_result/note เมื่อ update
7. mark ไฟล์ type เก่าไม่ active; ID_CARD MAIN กับ FRONT/BACK แทนกัน
8. คืนไฟล์ active type ล่าสุดและ complete: MAIN หรือ ID_CARD FRONT+BACK

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_master_request_item` | READ |
| `m_document_profile_request_item` | READ, INSERT, UPDATE |

Query path: `ValidateItem` → `UpsertProfile` → `Files` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [requestitem.ValidateItem](../queries/requestitem.md#requestitemvalidateitem)
- [requestitem.UpsertProfile](../queries/requestitem.md#requestitemupsertprofile)
- [requestitem.saveLocked](../queries/requestitem.md#requestitemsavelocked)
- [requestitem.upsertProfile](../queries/requestitem.md#requestitemupsertprofile)
- [requestitem.Files](../queries/requestitem.md#requestitemfiles)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: master lock และ upsert/supersede อยู่ใน transaction; storage อยู่นอก transaction
- Side effects: PUT object; save ล้มเหลวพยายาม Delete object ใหม่ ยกเว้น commit ไม่แน่นอน; ไม่ลบไฟล์เก่า

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

Domain errors: MA00005 file/type/slot; MA00016 master/scope

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [requestitem/handler.go](../../../internal/domain/requestitem/handler.go)
- [requestitem/service.go](../../../internal/domain/requestitem/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

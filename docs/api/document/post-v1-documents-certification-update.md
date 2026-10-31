# POST /v1/documents/certification/update

แก้ไข certificate

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `document.Update`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [document.UpdateRequest](../SCHEMAS.md#documentupdaterequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `isChangeFile` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `documentCode` | `string` | ใช่ |
| `certStartDate` | `string` | ใช่ |
| `certEndDateType` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `certEndDate` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `fileCert` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `fileCertName` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "isChangeFile": "N",
  "documentCode": "DOC001",
  "certStartDate": "2026-01-01",
  "certEndDateType": "A",
  "certEndDate": "2027-01-01",
  "fileCert": "example",
  "fileCertName": "example"
}
```

## Business logic

1. validate dates แบบเดียวกับ create; certificate ต้องเป็นของ user+document
2. isChangeFile=Y จึงตรวจ Base64/magic bytes/size และ PUT object ใหม่
3. UPDATE dates/status A; update file/name เฉพาะเมื่อเปลี่ยนไฟล์

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_certificates` | READ, UPDATE |
| `m_mobile_users` | READ |

Query path: `FindUser` → `Certificate` → `UpdateCertificate` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [document.FindUser](../queries/document.md#documentfinduser)
- [document.Certificate](../queries/document.md#documentcertificate)
- [document.UpdateCertificate](../queries/document.md#documentupdatecertificate)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบ storage และ UPDATE
- Side effects: อาจ PUT object; ไม่ลบไฟล์เก่าและไม่มี cleanup flow นี้

## Response และ errors

HTTP 200; `data` เป็น [document.Response](../SCHEMAS.md#documentresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "documentCode": "DOC001",
    "certStartDate": "2026-01-01",
    "certEndDateType": "A",
    "certEndDate": "2027-01-01",
    "fileCertName": "example"
  }
}
```

Domain errors: MA00005 dates/file; MA00016 certificate ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [document/handler.go](../../../internal/domain/document/handler.go)
- [document/service.go](../../../internal/domain/document/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

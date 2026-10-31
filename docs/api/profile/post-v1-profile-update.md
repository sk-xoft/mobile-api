# POST /v1/profile-update

แก้ไขโปรไฟล์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `profile.Update`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [profile.UpdateRequest](../SCHEMAS.md#profileupdaterequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `firstName` | `string` | ใช่ |
| `lastName` | `string` | ใช่ |
| `companyCode` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `dateOfBirth` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `positionCode` | `string` | ใช่ |
| `email` | `string` | ใช่ |
| `mobileNumber` | `string` | ใช่ |
| `isChangeFile` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `imageProfile` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "firstName": "Example",
  "lastName": "Seaman",
  "companyCode": "COMPANY_EXAMPLE",
  "dateOfBirth": "1990-01-01",
  "positionCode": "POSITION_EXAMPLE",
  "email": "seaman@example.com",
  "mobileNumber": "0800000000",
  "isChangeFile": "N",
  "imageProfile": "example"
}
```

## Business logic

1. ต้องมี firstName,lastName,positionCode,email,mobileNumber
2. เมื่อ isChangeFile=Y ถอด Base64 ตรวจ PNG/JPEG ขนาด 1–10 MiB แล้ว upload object
3. ล็อก user อ่านเบอร์เดิม; ถ้าเปลี่ยนให้เพิ่ม mobile number history
4. อัปเดตชื่อ บริษัท ตำแหน่ง เบอร์ วันเกิด และรูปเมื่อมีไฟล์ใหม่
5. email เป็น required input และถูกคืนใน response แต่ SQL ปัจจุบันไม่ได้ UPDATE EMAIL/USERNAME

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_mobile_number_history` | INSERT |
| `m_mobile_users` | READ, UPDATE |

Query path: `FindUser` → `Update` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [profile.FindUser](../queries/profile.md#profilefinduser)
- [profile.Update](../queries/profile.md#profileupdate)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: user update และ history อยู่ใน transaction เดียว; storage เขียนก่อน transaction
- Side effects: อาจ PUT profile image; object upload ไม่ได้อยู่ใน DB transaction และไม่มี cleanup ใน flow นี้

## Response และ errors

HTTP 200; `data` เป็น [profile.RegisterResponse](../SCHEMAS.md#profileregisterresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "username": "seaman@example.com",
    "email": "seaman@example.com"
  }
}
```

Domain errors: MA00004 required; MA00005 imageProfile ไม่ถูกต้อง

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [profile/handler.go](../../../internal/domain/profile/handler.go)
- [profile/service.go](../../../internal/domain/profile/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# POST /v1/register

สมัครสมาชิก

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.Register`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [auth.RegisterRequest](../SCHEMAS.md#authregisterrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `firstName` | `string` | ใช่ |
| `lastName` | `string` | ใช่ |
| `companyCode` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `dateOfBirth` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |
| `positionCode` | `string` | ใช่ |
| `email` | `string` | ใช่ |
| `password` | `string` | ใช่ |
| `mobileNumber` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "firstName": "Example",
  "lastName": "Seaman",
  "companyCode": "COMPANY_EXAMPLE",
  "dateOfBirth": "1990-01-01",
  "positionCode": "POSITION_EXAMPLE",
  "email": "seaman@example.com",
  "password": "<redacted-example>",
  "mobileNumber": "0800000000"
}
```

## Business logic

1. ต้องมี firstName,lastName,positionCode,email,password; โค้ดไม่ได้ตรวจรูปแบบ email หรือความยาวขั้นต่ำ password เพิ่ม
2. ตรวจ email ซ้ำ (MA00006), สร้าง SMART_SEAMAN_ID, UUID, bcrypt hash และ user สถานะ I
3. สร้าง certificate เริ่มต้นจาก m_documents ที่ DOCUMENT_DEFAULT_FLAG=Y
4. สร้าง activation token แบบสุ่ม เก็บ SHA-256 อายุ 24 ชั่วโมง แล้วส่งอีเมล

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_certificates` | INSERT |
| `m_documents` | READ |
| `m_message_code` | READ |
| `m_mobile_users` | READ, INSERT |
| `mobile_auth_tokens` | READ, INSERT, DELETE |

Query path: `Register` → `CreateActionToken` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.Register](../queries/auth.md#authregister)
- [auth.CreateActionToken](../queries/auth.md#authcreateactiontoken)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: สร้าง user/default certificates transaction หนึ่ง; สร้าง action token อีก transaction; SMTP หลัง commit จึงอาจล้มเหลวหลังบันทึกบัญชีแล้ว
- Side effects: ส่ง registration email; token ไม่ส่งกลับใน response

## Response และ errors

HTTP 200; `data` เป็น [auth.RegisterResponse](../SCHEMAS.md#authregisterresponse); envelope ตาม [COMMON.md](../COMMON.md)

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

Domain errors: MA00004 required; MA00006 email ซ้ำ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

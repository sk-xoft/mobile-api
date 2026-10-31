# POST /v1/reset-password

ขอลิงก์ตั้งรหัสผ่านใหม่

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.RequestPasswordReset`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [auth.ResetPasswordRequest](../SCHEMAS.md#authresetpasswordrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `email` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "email": "seaman@example.com"
}
```

## Business logic

1. ต้องมี email; user ไม่พบคืน MA00016
2. สร้าง random reset token อายุ 30 นาที เก็บเฉพาะ SHA-256
3. ล็อกบัญชี ลบ reset tokens เก่า แล้วบันทึก token ใหม่
4. ส่ง reset email หลัง commit

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_message_code` | READ |
| `m_mobile_users` | READ |
| `mobile_auth_tokens` | READ, INSERT, DELETE |

Query path: `FindUserByEmail` → `CreateActionToken` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.FindUserByEmail](../queries/auth.md#authfinduserbyemail)
- [auth.findUser](../queries/auth.md#authfinduser)
- [auth.CreateActionToken](../queries/auth.md#authcreateactiontoken)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: token issuance transaction แยกจาก SMTP
- Side effects: ส่ง password reset email; ถ้า SMTP ล้มเหลว token อาจถูกบันทึกแล้ว

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

Domain errors: MA00004 email ว่าง; MA00016 user ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

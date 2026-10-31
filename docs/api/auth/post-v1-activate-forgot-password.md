# POST /v1/activate-forgot-password

ตั้งรหัสผ่านจากลิงก์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.CompletePasswordReset`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [auth.ForgotPasswordRequest](../SCHEMAS.md#authforgotpasswordrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `password` | `string` | ใช่ |
| `confirmPassword` | `string` | ใช่ |
| `uid` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "password": "<redacted-example>",
  "confirmPassword": "<redacted-example>",
  "uid": "<redacted-example>"
}
```

## Business logic

1. ต้องมี password,confirmPassword,uid และรหัสผ่านต้องตรงกัน
2. ตรวจ random uid purpose=reset และ expiry; hash password ใหม่
3. ล็อก user/token; เปลี่ยน password, ปิด sessions และลบ reset tokens

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_message_code` | READ |
| `m_mobile_users` | READ, UPDATE |
| `mobile_auth_tokens` | READ, DELETE |
| `t_session` | READ, UPDATE |

Query path: `ConsumeActionToken` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.ConsumeActionToken](../queries/auth.md#authconsumeactiontoken)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: consume token/password/session revocation ใน transaction เดียว
- Side effects: sessions เดิมถูก revoke; ต้อง login ใหม่

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

Domain errors: MA00004 required; MA00020 password ไม่ตรง; MA00007 token ใช้ไม่ได้

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

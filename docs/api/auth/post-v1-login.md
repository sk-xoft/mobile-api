# POST /v1/login

เข้าสู่ระบบ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.Login`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [auth.LoginRequest](../SCHEMAS.md#authloginrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `username` | `string` | ใช่ |
| `password` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "username": "seaman@example.com",
  "password": "<redacted-example>"
}
```

## Business logic

1. ต้องมี username/password; ค้น username และตรวจ bcrypt
2. ปฏิเสธสถานะ D ด้วย MA00023 และสถานะอื่นที่ไม่ใช่ A ด้วย MA00002
3. ออก JWT HS256 อายุ 30 นาที; jti คือ session ID
4. ล็อกบัญชี ตรวจ active ซ้ำ ปิด online sessions เดิม แล้วสร้าง session ใหม่และ refresh hash อายุ 30 วัน

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_message_code` | READ |
| `m_mobile_users` | READ |
| `mobile_auth_sessions` | INSERT |
| `t_session` | READ, INSERT, UPDATE |

Query path: `FindUserByUsername` → `CreateSecureSession` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.FindUserByUsername](../queries/auth.md#authfinduserbyusername)
- [auth.findUser](../queries/auth.md#authfinduser)
- [auth.CreateSecureSession](../queries/auth.md#authcreatesecuresession)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: account lock, revoke และ insert sessions อยู่ใน transaction เดียว
- Side effects: คืน token/refToken ให้ client เก็บอย่างปลอดภัย

## Response และ errors

HTTP 200; `data` เป็น [auth.LoginResponse](../SCHEMAS.md#authloginresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "token": "<redacted-jwt>",
    "refToken": "<redacted-refresh-token>",
    "username": "seaman@example.com",
    "lastLoginDateTime": "example"
  }
}
```

Domain errors: MA00004 required; MA00027 username ไม่พบ; MA00003 password ผิด; MA00023/D; MA00002/inactive

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

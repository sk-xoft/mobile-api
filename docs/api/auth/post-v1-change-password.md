# POST /v1/change-password

เปลี่ยนรหัสผ่าน

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.ChangePassword`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [auth.ChangePasswordRequest](../SCHEMAS.md#authchangepasswordrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `oldPassword` | `string` | ใช่ |
| `newPassword` | `string` | ใช่ |
| `confirmPassword` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "oldPassword": "<redacted-example>",
  "newPassword": "<redacted-example>",
  "confirmPassword": "<redacted-example>"
}
```

## Business logic

1. ใช้ JWT subject ค้น user; ต้องมี oldPassword,newPassword,confirmPassword
2. ตรวจ password ใหม่ตรงกันและ bcrypt ของ password เก่า
3. อัปเดต hash, ปิดทุก session ของ user และลบ reset tokens

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_message_code` | READ |
| `m_mobile_users` | READ, UPDATE |
| `mobile_auth_tokens` | READ, DELETE |
| `t_session` | READ, UPDATE |

Query path: `FindUserByUsername` → `UpdatePassword` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.FindUserByUsername](../queries/auth.md#authfinduserbyusername)
- [auth.findUser](../queries/auth.md#authfinduser)
- [auth.UpdatePassword](../queries/auth.md#authupdatepassword)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: password/session revocation/token invalidation transaction เดียว
- Side effects: JWT/session ปัจจุบันใช้ต่อไม่ได้หลังสำเร็จ

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

Domain errors: MA00004 required; MA00020 confirm ไม่ตรง; MA00003 password เก่าไม่ถูก

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

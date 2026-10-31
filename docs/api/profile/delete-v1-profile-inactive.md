# DELETE /v1/profile/inactive

ปิดบัญชี

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `profile.Deactivate`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. อ่านบัญชีปัจจุบันแล้วเปลี่ยน USER_STATUS=D
2. ปิด sessions และลบ action tokens ทั้งหมดของ user

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_mobile_users` | READ, UPDATE |
| `mobile_auth_tokens` | READ, DELETE |
| `t_session` | READ, UPDATE |

Query path: `FindUser` → `UpdateStatus` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [profile.FindUser](../queries/profile.md#profilefinduser)
- [profile.UpdateStatus](../queries/profile.md#profileupdatestatus)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: status/session/token changes transaction เดียว
- Side effects: ผู้ใช้ต้องผ่าน recovery ที่ได้รับอนุญาตเพื่อกลับมาใช้งาน

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

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [profile/handler.go](../../../internal/domain/profile/handler.go)
- [profile/service.go](../../../internal/domain/profile/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# GET /v1/refresh-token

หมุน refresh credential

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `auth.Refresh`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| query | `refToken` | string | ต้องมี; random refresh credential ไม่ใช่ JWT/session ID |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. รับ refToken จาก query; ห้ามใช้ JWT jti แทน
2. ตรวจ hash, expiry, online session และ active user
3. ออก JWT ใหม่และ refresh token ใหม่; compare-and-swap hash แบบ atomic
4. client ต้องเก็บ refToken ใหม่และ serialize refresh requests; วันหมดอายุ session เดิมไม่ขยาย

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_message_code` | READ |
| `m_mobile_users` | READ |
| `mobile_auth_sessions` | READ, UPDATE |
| `t_session` | READ, UPDATE |

Query path: `FindRefreshSession` → `FindUserByUUID` → `RotateRefreshSession` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [auth.FindRefreshSession](../queries/auth.md#authfindrefreshsession)
- [auth.FindUserByUUID](../queries/auth.md#authfinduserbyuuid)
- [auth.findUser](../queries/auth.md#authfinduser)
- [auth.RotateRefreshSession](../queries/auth.md#authrotaterefreshsession)
- [auth.Message](../queries/auth.md#authmessage)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: UPDATE JOIN พร้อมเงื่อนไข old hash/expiry/active/online เป็น atomic statement
- Side effects: หมุน credentials; token เก่าจะใช้ refresh ไม่ได้

## Response และ errors

HTTP 200; `data` เป็น [auth.RefreshTokenResponse](../SCHEMAS.md#authrefreshtokenresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "refToken": "<redacted-refresh-token>",
    "token": "<redacted-jwt>"
  }
}
```

Domain errors: MA00018 ไม่มี refToken; MA00009 หา refresh ไม่พบ; MA00007 credential หมดอายุหรือถูกใช้แข่งกัน

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [auth/handler.go](../../../internal/domain/auth/handler.go)
- [auth/service.go](../../../internal/domain/auth/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

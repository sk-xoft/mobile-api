# POST /v1/fcm-update

บันทึก FCM token

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `fcm.Update`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [fcm.UpdateRequest](../SCHEMAS.md#fcmupdaterequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `tokenFcm` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "tokenFcm": "<redacted-example>"
}
```

## Business logic

1. tokenFcm ต้องไม่เป็น whitespace
2. ตรวจ user มีแถว m_fcm_notifications หรือไม่ แล้ว UPDATE หรือ INSERT
3. คืน username + " update fcm token is success."

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_fcm_notifications` | READ, INSERT, UPDATE |

Query path: `Exists` → `Insert` → `Update` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [fcm.Exists](../queries/fcm.md#fcmexists)
- [fcm.Insert](../queries/fcm.md#fcminsert)
- [fcm.Update](../queries/fcm.md#fcmupdate)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: existence check กับ write แยก statement
- Side effects: บันทึก token เท่านั้น ไม่มีการเรียก FCM ส่ง push

## Response และ errors

HTTP 200; `data` เป็น `string`; envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": "success"
}
```

Domain errors: MA00005 tokenFcm

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [fcm/handler.go](../../../internal/domain/fcm/handler.go)
- [fcm/service.go](../../../internal/domain/fcm/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# POST /v1/notifications/update

อ่าน notification ตาม ID

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `notification.MarkID`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [notification.MarkRequest](../SCHEMAS.md#notificationmarkrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `notiId` | `string` | ใช้เลือก notification; "0" อ่าน count อย่างเดียว |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "notiId": "0"
}
```

## Business logic

1. notiId="0" หมายถึงอ่าน unread count โดยไม่ UPDATE
2. ค่าอื่น UPDATE เฉพาะ id+เจ้าของ; affected rows=0 คืน MA00016
3. คืน countNotification เป็น string ของ unread count

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_send_notifications` | READ, UPDATE |

Query path: `MarkID` → `Count` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [notification.MarkID](../queries/notification.md#notificationmarkid)
- [notification.Count](../queries/notification.md#notificationcount)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [notification.CountResponse](../SCHEMAS.md#notificationcountresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "countNotification": "0"
  }
}
```

Domain errors: MA00016 notification ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [notification/handler.go](../../../internal/domain/notification/handler.go)
- [notification/service.go](../../../internal/domain/notification/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

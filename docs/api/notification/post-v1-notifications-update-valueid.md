# POST /v1/notifications/update/valueId

อ่าน notification ตาม value/type

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `notification.MarkValue`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [notification.MarkValueRequest](../SCHEMAS.md#notificationmarkvaluerequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `valueId` | `string` | ใช่ |
| `notiType` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "valueId": "example",
  "notiType": "example"
}
```

## Business logic

1. ต้องมี valueId และ notiType; ค่าว่างเป็น ordinary error จึงตอบ HTTP 500 MA00013 ตามโค้ดปัจจุบัน
2. UPDATE READ_STATUS=YES ด้วย user+VALUE_ID+NOTI_TYPE
3. คืนจำนวน unread เป็น string; ไม่พบแถวไม่ถือว่า error

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_send_notifications` | READ, UPDATE |

Query path: `MarkValue` → `Count` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [notification.MarkValue](../queries/notification.md#notificationmarkvalue)
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

Domain errors: 400 MA00005 เมื่อ parse body ไม่ได้; 500 MA00013 เมื่อ valueId/notiType ว่าง

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [notification/handler.go](../../../internal/domain/notification/handler.go)
- [notification/service.go](../../../internal/domain/notification/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

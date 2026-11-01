# GET /v1/delivery-addresses

ที่อยู่จัดส่งเริ่มต้น

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-11-01
- Handler: `delivery.Default`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. กรองเจ้าของ, is_default=1 และ is_active=YES
2. join user เพื่ออ่านเบอร์ล่าสุด; join address master เพื่อประกอบ description
3. ต้องมี default เพียงแถวเดียว; ไม่พบคืน MA00016 และซ้ำคืน HTTP 500 MA00013

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `districts` | READ |
| `m_delivery_address` | READ |
| `m_mobile_users` | READ |
| `provinces` | READ |
| `subdistricts` | READ |

Query path: `Default` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [delivery.Default](../queries/delivery.md#deliverydefault)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [delivery.Response](../SCHEMAS.md#deliveryresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "id": "example",
    "firstName": "Example",
    "lastName": "Seaman",
    "addressLine": "example",
    "province": "10",
    "district": "1001",
    "subDistrict": "100101",
    "postalCode": "10100",
    "mobileNumber": "0800000000",
    "description": "example",
    "isDefault": false
  }
}
```

Domain errors: MA00016 deliveryAddress ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [delivery/handler.go](../../../internal/domain/delivery/handler.go)
- [delivery/service.go](../../../internal/domain/delivery/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# GET /v1/profile

อ่านโปรไฟล์

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `profile.Get`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. อ่าน user ด้วย MOBILE_UUID จาก session
2. อ่านคำอธิบายบริษัท/ตำแหน่งตามภาษา; lookup description errors ถูกละไว้จึงอาจได้ค่าว่าง
3. คำนวณ age จากวันเกิด และ shortName จากอักษรแรกชื่อ/นามสกุล

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_mobile_users` | READ |

Query path: `FindUser` → `Descriptions` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [profile.FindUser](../queries/profile.md#profilefinduser)
- [profile.Descriptions](../queries/profile.md#profiledescriptions)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [profile.Response](../SCHEMAS.md#profileresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "firstName": "Example",
    "lastName": "Seaman",
    "dateOfBirth": "1990-01-01",
    "age": "example",
    "mobile": "example",
    "email": "seaman@example.com",
    "companyCode": "COMPANY_EXAMPLE",
    "companyDescription": "example",
    "positionCode": "POSITION_EXAMPLE",
    "positionDescription": "example",
    "smartSeamanId": "example",
    "shortName": "example"
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

# GET /v1/master

ข้อมูลบริษัทและตำแหน่ง

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `master.List`
- Authentication: Public
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. บริษัทต้องเป็น Shipping Company, MOBILE_FLAG=Y และ STATUS=A เรียง type/sequence
2. ตำแหน่งอ่านทั้งหมด ไม่มี status filter ใน SQL
3. คืนชื่อไทยหรืออังกฤษตาม Accept-Language

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_companys` | READ |
| `m_positions` | READ |

Query path: `Companies` → `Positions` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [master.Companies](../queries/master.md#mastercompanies)
- [master.Positions](../queries/master.md#masterpositions)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [master.MasterResponse](../SCHEMAS.md#mastermasterresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "company": [],
    "position": []
  }
}
```

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [master/handler.go](../../../internal/domain/master/handler.go)
- [master/service.go](../../../internal/domain/master/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

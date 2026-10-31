# GET /v1/document-renewals/prices

ราคาต่ออายุเอกสาร

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.Price`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| query | `documentCode` | string | ต้องมี |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. documentCode ต้องไม่ว่าง; trim และ uppercase
2. เลือก price active ที่ effective_from<=วันนี้ และ effective_to ว่างหรือยังไม่หมดอายุ; document ต้อง active
3. เลือก effective_from ล่าสุด; total=governmentFee+documentProcessingFee+shippingFee-shippingDiscount-serviceFeeDiscount

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_prices_setting` | READ |
| `m_documents` | READ |

Query path: `Price` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.Price](../queries/renewal.md#renewalprice)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.Price](../SCHEMAS.md#renewalprice); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "documentCode": "DOC001",
    "governmentFee": 0,
    "documentProcessingFee": 0,
    "shippingFee": 0,
    "shippingDiscount": 0,
    "serviceFeeDiscount": 0,
    "total": 0,
    "effectiveFrom": null,
    "effectiveTo": null
  }
}
```

Domain errors: MA00005 ไม่มี code; MA00016 ไม่พบราคาที่ใช้งานได้

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [renewal/handler.go](../../../internal/domain/renewal/handler.go)
- [renewal/lifecycle_handler.go](../../../internal/domain/renewal/lifecycle_handler.go)
- [renewal/service.go](../../../internal/domain/renewal/service.go)
- [renewal/lifecycle_service.go](../../../internal/domain/renewal/lifecycle_service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

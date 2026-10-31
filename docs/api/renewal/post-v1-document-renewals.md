# POST /v1/document-renewals

สร้างคำขอต่ออายุแบบเดิม

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.CreateRequest`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [renewal.CreateRequest](../SCHEMAS.md#renewalcreaterequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `documentCode` | `string` | ใช่ |
| `deliveryAddressId` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "documentCode": "DOC001",
  "deliveryAddressId": "example"
}
```

## Business logic

1. documentCode/deliveryAddressId ต้องไม่ว่าง; อ่านราคาปัจจุบัน
2. อ่าน mobile/email, PAYMENT_PENDING status และเพิ่ม sequence ตาม YYMM
3. INSERT คำขอพร้อม amount และ delivery_address_id; price_setting_id เป็น NULL
4. เส้นนี้ไม่ได้สร้าง request items, address snapshot หรือ CREATE timeline และไม่ได้ตรวจ address ownership ใน service

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_prices_setting` | READ |
| `m_document_request` | INSERT |
| `m_document_request_sequence` | READ, INSERT, UPDATE |
| `m_document_status` | READ |
| `m_documents` | READ |
| `m_mobile_users` | READ |

Query path: `Price` → `create` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.Price](../queries/renewal.md#renewalprice)
- [renewal.create](../queries/renewal.md#renewalcreate)
- [renewal.statusID](../queries/renewal.md#renewalstatusid)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: sequence increment/read และ INSERT ไม่มี transaction ครอบทั้ง flow; ไม่มี idempotency key
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.CreateResponse](../SCHEMAS.md#renewalcreateresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestId": "00000000-0000-4000-8000-000000000001",
    "requestNo": "261000001",
    "documentCode": "DOC001",
    "status": "PAYMENT_PENDING",
    "amount": 0,
    "deliveryAddressId": "example"
  }
}
```

Domain errors: MA00005 documentCode/deliveryAddressId; MA00016 price

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

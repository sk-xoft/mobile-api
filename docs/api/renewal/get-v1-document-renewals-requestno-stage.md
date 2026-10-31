# GET /v1/document-renewals/:requestNo/stage

ความคืบหน้า 5 ขั้นตอน

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.Stage`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. trim/uppercase requestNo; ยาว 1–20 ตัว รับเฉพาะ A–Z,0–9,_,-
2. อ่าน active request ของ user, current status และ mobile-visible active status master
3. สร้าง ladder ให้ครบ 5 ขั้น: review, department result, pickup, delivering, delivered
4. กำหนด DONE/CURRENT/PENDING ตาม current step; current step ไม่มี mapping เป็น null และทุก stage PENDING
5. status/ladder config ไม่ครบเป็น BusinessError MA00013 ซึ่ง handler ส่ง HTTP 400

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ |
| `m_document_status` | READ |

Query path: `ownedRequest` → `stageCurrentStatus` → `Statuses` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.ownedRequest](../queries/renewal.md#renewalownedrequest)
- [renewal.stageCurrentStatus](../queries/renewal.md#renewalstagecurrentstatus)
- [renewal.Statuses](../queries/renewal.md#renewalstatuses)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.StageResponse](../SCHEMAS.md#renewalstageresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestNo": "261000001",
    "currentStatus": {
      "id": "example",
      "documentStatusCode": "example",
      "nameTh": "example",
      "nameEn": "example",
      "cssColor": "example",
      "step": null,
      "mobileStatus": null
    },
    "stages": []
  }
}
```

Domain errors: MA00005 requestNo; MA00016 request; 400 MA00013 status/ladder config

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

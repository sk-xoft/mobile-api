# POST /v1/document-renewals/requests/validate-and-create

ตรวจและสร้าง unpaid draft

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.Validate`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

Content-Type: `application/json`; model [renewal.ValidateInput](../SCHEMAS.md#renewalvalidateinput)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `documentCode` | `string` | ใช่ |
| `idempotencyKey` | `string` | ไม่มี required check โดยตรง; ดูเงื่อนไขด้านล่าง |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "documentCode": "DOC001",
  "idempotencyKey": "example-request-key"
}
```

## Business logic

1. trim/uppercase documentCode (1–10 ตัว); trim idempotencyKey สูงสุด 100 ตัว
2. ถ้ามี key เดิมของ user ต้องเป็น documentCode เดิม แล้วคืนคำขอเดิม; key คนละ document คืน MA00005
3. ถ้าไม่มี key ที่พบ ใช้คำขอ active ล่าสุดของ user+document ที่สถานะไม่ใช่ DELIVERED
4. ถ้าไม่พบ: อ่านราคา เอกสาร active required item count และ default address ก่อนเปิด transaction
5. สร้าง sequence/request, contact+price/address snapshot, required items และ CREATE timeline
6. ประกอบ response: certificate expiry, address, items และ batch files แยก PROFILE/REQUEST
7. สถานะ item เป็น MISSING/NEED_FIX/INCOMPLETE/COMPLETE; ID_CARD ครบ MAIN หรือ FRONT+BACK, PASSPORT/GENERAL ครบ MAIN
8. lookup errors ที่ไม่ใช่ not found ต้องหยุด; ไม่มี default address สร้าง draft ได้

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `districts` | READ |
| `m_certificates` | READ |
| `m_delivery_address` | READ |
| `m_document_master_request_item` | READ |
| `m_document_prices_setting` | READ |
| `m_document_profile_request_item` | READ |
| `m_document_request` | READ, INSERT |
| `m_document_request_delivery_address` | READ, INSERT |
| `m_document_request_item_files` | READ |
| `m_document_request_items` | READ, INSERT |
| `m_document_request_sequence` | READ, INSERT, UPDATE |
| `m_document_setting_requires` | READ |
| `m_document_status` | READ |
| `m_document_transaction` | INSERT |
| `m_documents` | READ |
| `m_mobile_users` | READ |
| `provinces` | READ |
| `subdistricts` | READ |

Query path: `validateExisting` → `latestActive` → `validateResponse` → `Price` → `validateCreate` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.validateExisting](../queries/renewal.md#renewalvalidateexisting)
- [renewal.ownedRequest](../queries/renewal.md#renewalownedrequest)
- [renewal.latestActive](../queries/renewal.md#renewallatestactive)
- [renewal.validateResponse](../queries/renewal.md#renewalvalidateresponse)
- [renewal.documentInfo](../queries/renewal.md#renewaldocumentinfo)
- [renewal.requestItems](../queries/renewal.md#renewalrequestitems)
- [renewal.validateResponseWithDocument](../queries/renewal.md#renewalvalidateresponsewithdocument)
- [renewal.validateFilesBatch](../queries/renewal.md#renewalvalidatefilesbatch)
- [renewal.Price](../queries/renewal.md#renewalprice)
- [renewal.validateCreate](../queries/renewal.md#renewalvalidatecreate)
- [renewal.statusID](../queries/renewal.md#renewalstatusid)
- [renewal.requiredItemCount](../queries/renewal.md#renewalrequireditemcount)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: new draft writes อยู่ใน transaction เดียว; response reads หลัง commit; response read ล้มเหลวอาจเกิดหลังสร้าง draftแล้ว
- Side effects: สร้าง draft ได้แม้ไฟล์ยังไม่ครบ; ไม่ upload file หรือเรียก payment

## Response และ errors

HTTP 200; `data` เป็น [renewal.ValidateResponse](../SCHEMAS.md#renewalvalidateresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "documentCode": "DOC001",
    "idempotencyKey": null,
    "certEndDate": null,
    "address": [],
    "items": []
  }
}
```

Domain errors: MA00005 code/key; MA00016 price; MA00013 documentSetting/requestItems configuration

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

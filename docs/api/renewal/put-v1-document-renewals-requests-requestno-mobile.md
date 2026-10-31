# PUT /v1/document-renewals/requests/:requestNo/mobile

แก้เบอร์ snapshot ของคำขอ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `renewal.UpdateMobile`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |

Content-Type: `application/json`; model [renewal.MobileRequest](../SCHEMAS.md#renewalmobilerequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `mobileNumber` | `string` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "mobileNumber": "0800000000"
}
```

## Business logic

1. mobileNumber ยาว 10 ตัวและขึ้นต้น 0; โค้ดไม่ได้ตรวจว่าทุกตัวเป็นตัวเลข
2. ค้น active request ของ user แล้ว UPDATE mobile_number ในคำขอเท่านั้น
3. fromStatus/toStatus ใน response ของเส้นนี้หมายถึงเบอร์เดิม/ใหม่ ไม่ใช่ document status

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_document_request` | READ, UPDATE |
| `m_document_status` | READ |

Query path: `updateMobile` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [renewal.updateMobile](../queries/renewal.md#renewalupdatemobile)
- [renewal.ownedRequest](../queries/renewal.md#renewalownedrequest)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200; `data` เป็น [renewal.TransitionResponse](../SCHEMAS.md#renewaltransitionresponse); envelope ตาม [COMMON.md](../COMMON.md)

ตัวอย่างโครงสร้าง response (ค่า illustration; ไม่ใช่ fixture จากระบบจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestNo": "261000001",
    "fromStatus": "example",
    "toStatus": "example",
    "action": "example"
  }
}
```

Domain errors: MA00005 mobileNumber; MA00016 request

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

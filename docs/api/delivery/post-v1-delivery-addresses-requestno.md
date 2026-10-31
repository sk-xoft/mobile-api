# POST /v1/delivery-addresses/:requestNo

สร้าง snapshot ที่อยู่ของคำขอ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `delivery.CreateForRenewal`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `requestNo` | string | ต้องมี; ดู validation ใน Business logic |

Content-Type: `application/json`; model [delivery.Request](../SCHEMAS.md#deliveryrequest)

| Field | Type | Required ตาม implementation |
|---|---|---|
| `firstName` | `string` | ใช่ |
| `lastName` | `string` | ใช่ |
| `addressLine` | `string` | ใช่ |
| `province` | `string` | ใช่ |
| `district` | `string` | ใช่ |
| `subDistrict` | `string` | ใช่ |
| `postalCode` | `string` | ใช่ |
| `isDefault` | `*bool` | ใช่ |

ตัวอย่างรูปแบบ body (ต้องแทนค่าตัวอย่างด้วยค่าที่ผ่าน validation):

```json
{
  "firstName": "Example",
  "lastName": "Seaman",
  "addressLine": "example",
  "province": "10",
  "district": "1001",
  "subDistrict": "100101",
  "postalCode": "10100",
  "isDefault": true
}
```

## Business logic

1. ต้องมีชื่อ นามสกุล ที่อยู่ province,district,subDistrict และ isDefault; postalCode ยาว 5 ตัว
2. ตรวจ codes จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์ร่วมกันใน master
3. ค้น request_no ที่ active และเป็นของผู้ใช้
4. มี snapshot อยู่แล้วคืน MA00017; ถ้ายังไม่มี INSERT snapshot โดยอ่านเบอร์ user ปัจจุบัน
5. ไม่มีการสร้าง master address ใน flow นี้

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `districts` | READ |
| `m_document_request` | READ |
| `m_document_request_delivery_address` | READ, INSERT |
| `m_mobile_users` | READ |
| `provinces` | READ |
| `subdistricts` | READ |

Query path: `ValidAddress` → `RenewalRequest` → `SnapshotExists` → `UserMobile` → `InsertSnapshot` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [delivery.ValidAddress](../queries/delivery.md#deliveryvalidaddress)
- [delivery.RenewalRequest](../queries/delivery.md#deliveryrenewalrequest)
- [delivery.SnapshotExists](../queries/delivery.md#deliverysnapshotexists)
- [delivery.UserMobile](../queries/delivery.md#deliveryusermobile)
- [delivery.InsertSnapshot](../queries/delivery.md#deliveryinsertsnapshot)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: existence check และ INSERT แยก statement; unique request constraint เป็น guard
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

Domain errors: MA00005 address; MA00016 request; MA00017 snapshot ซ้ำ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [delivery/handler.go](../../../internal/domain/delivery/handler.go)
- [delivery/service.go](../../../internal/domain/delivery/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

# PUT /v1/delivery-addresses/:addressId

แก้ไขที่อยู่จัดส่ง

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-11-01
- Handler: `delivery.Update`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| path | `addressId` | string | ต้องมี; ดู validation ใน Business logic |

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

1. addressId ต้อง parse UUID ได้
2. ต้องมีชื่อ นามสกุล ที่อยู่ province,district,subDistrict และ isDefault; postalCode ต้องเป็นเลข 5 ตัว; ชื่อ/นามสกุลและ codes ยาวไม่เกิน 100 ตัว ที่อยู่ไม่เกิน 500 ตัว
3. ตรวจ codes จังหวัด/อำเภอ/ตำบล/รหัสไปรษณีย์ร่วมกันใน master
4. อ่านที่อยู่ด้วย id+เจ้าของ+active; ไม่อนุญาตเปลี่ยน default ปัจจุบันเป็น false
5. ถ้าตั้ง default ให้ clear ค่าเดิมแล้ว UPDATE ใน transaction เดียว; การส่งค่าเดิมซ้ำยอมรับได้
6. ถ้าไม่พบ master address ให้ค้น owned renewal snapshot แล้ว UPDATE โดยรักษา mobileNumber เดิม; ไม่ clear default ของ master

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `districts` | READ |
| `m_delivery_address` | READ, UPDATE |
| `m_mobile_users` | READ (lock) |
| `m_document_request_delivery_address` | READ, UPDATE |
| `provinces` | READ |
| `subdistricts` | READ |

Query path: `ValidAddress` → `LockUser` → `LockAddresses` → `Find` → `ClearDefault` → `Update`; fallback `FindSnapshot` → `UpdateSnapshot` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [delivery.ValidAddress](../queries/delivery.md#deliveryvalidaddress)
- [delivery.Find](../queries/delivery.md#deliveryfind)
- [delivery.ClearDefault](../queries/delivery.md#deliverycleardefault)
- [delivery.Update](../queries/delivery.md#deliveryupdate)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ล็อก user/active addresses และทำ lookup, clear default, UPDATE ใน transaction เดียว; snapshot ใช้ owner filter และ row lock
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

Domain errors: MA00008 addressId; MA00005 address/isDefault; MA00016 ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [delivery/handler.go](../../../internal/domain/delivery/handler.go)
- [delivery/service.go](../../../internal/domain/delivery/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

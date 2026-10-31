# GET /v1/documents/certification/view

ไฟล์ certificate

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `document.View`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| query | `certCode` | string | ใช้เป็น CERT_DOCUMENT_CODE ของ user ปัจจุบัน |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. certCode เป็น document code; ตรวจ certificate เป็นของ user
2. ไม่มีชื่อไฟล์คืน MA00016; อ่าน object และส่ง bytes/content type จาก storage

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_certificates` | READ |

Query path: `Certificate` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [document.Certificate](../queries/document.md#documentcertificate)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: GET object storage

## Response และ errors

สำเร็จ HTTP 200 เป็น bytes โดยตรง ไม่มี JSON envelope; Content-Type ตาม Business logic/handler
เมื่อผิดพลาดกลับเป็น JSON error envelope ตามกติกากลาง

Domain errors: MA00016 certificate/filename ไม่พบ

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [document/handler.go](../../../internal/domain/document/handler.go)
- [document/service.go](../../../internal/domain/document/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

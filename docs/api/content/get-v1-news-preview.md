# GET /v1/news/preview

รูปข่าว

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `content.NewsImage`
- Authentication: Bearer JWT + active online session
- Headers: ต้องมี `Accept-Language`, `devicemodel`, `correlationid`; ดู [กติกากลาง](../COMMON.md)

## Request

| Location | Name | Type | กติกา |
|---|---|---|---|
| query | `newsId` | string | ใช้ค้น NEWS_ID |

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. ใช้ newsId ค้นชื่อไฟล์; file ว่างคืน MA00016
2. อ่าน object แล้ว Base64 decode ส่ง image/png

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

| Table / relation | Access |
|---|---|
| `m_news` | READ |

Query path: `NewsByID` (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

SQL/source สำหรับตรวจ filter, JOIN และ statements:

- [content.NewsByID](../queries/content.md#contentnewsbyid)

Common access แยกจาก domain: `t_session`, `mobile_auth_sessions`, `m_mobile_users` (READ โดย Authenticate); `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี transaction ครอบทั้ง operation
- Side effects: GET object storage

## Response และ errors

สำเร็จ HTTP 200 เป็น bytes โดยตรง ไม่มี JSON envelope; Content-Type ตาม Business logic/handler
เมื่อผิดพลาดกลับเป็น JSON error envelope ตามกติกากลาง

Domain errors: MA00016 ไม่พบข่าว/ไฟล์; 500 MA00013 storage/decode error

Technical errors โดยทั่วไปเป็น HTTP 500 `MA00013`; business/not-found errors ของ legacy handlers โดยทั่วไปเป็น HTTP 400 (ไม่ใช่ 404) ยกเว้น payment/webhook ที่ระบุแยก
ก่อนถึง handler อาจได้ 400 `MA00004` เมื่อ header ไม่ครบ
Authentication อาจคืน 401 ตาม [COMMON.md](../COMMON.md#authentication)

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)
- [content/handler.go](../../../internal/domain/content/handler.go)
- [content/service.go](../../../internal/domain/content/service.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

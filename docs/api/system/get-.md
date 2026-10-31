# GET /

ข้อความต้อนรับ

## Overview

- สถานะ: อ้างอิง implementation ณ 2026-10-31
- Handler: `system.Root`
- Authentication: Public
- Headers: ไม่ผ่าน APIContext; ดู [กติกากลาง](../COMMON.md)

## Request

ไม่มี request body; ใช้ path/query ตามที่ระบุด้านบน

## Business logic

1. ส่งข้อความ Smart Seaman Mobile API

## Database queries

ตารางด้านล่างเป็น domain access ที่อาจเกิดตาม branch; ไม่ใช่ทุก request จะใช้ทุก query
`READ` รวม SELECT/JOIN และการอ่านเงื่อนไขของ write; ไม่ใช่จำนวน round trips

ไม่มี domain SQL

Query path: ไม่มี repository call (เลือก branch ตาม business logic ไม่ใช่ลำดับตายตัว)

Common access แยกจาก domain: `event_logs` (INSERT โดย RequestLogger หลัง handler ใน application wiring) ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

- Transaction: ไม่มี domain transaction
- Side effects: ไม่มี external side effect

## Response และ errors

HTTP 200 เป็นข้อความ `Smart Seaman Mobile API` โดยตรง

Endpoint นี้ไม่มี domain error envelope เพิ่มเติมนอกเหนือจากที่ระบุ; framework errors อาจใช้ status/text ของ Fiber

## Source และ verification

- [Route registration](../../../internal/transport/routes/routes.go)

ตรวจจาก static source เท่านั้น ยังไม่ได้เรียก endpoint, execute SQL หรือยืนยันผลกับ live MySQL/Omise/SMTP/storage

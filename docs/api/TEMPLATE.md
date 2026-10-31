# <METHOD> <PATH>

<ชื่อ endpoint และวัตถุประสงค์ที่ผู้เรียกใช้งานเข้าใจได้>

## Overview

- สถานะ/วันที่ตรวจ: <implemented / pending / proposed; YYYY-MM-DD>
- Domain / handler: <package.Method>
- Authentication / authorization: <public, JWT, signature, user ownership>
- Headers: <required/optional, ค่า default, link COMMON.md>

## Request

Content-Type: <application/json / multipart/form-data / ไม่มี body>

| Location | Field | Type | Required / default | Validation / ความหมาย |
|---|---|---|---|---|
| path/query/header/body/form | fieldName | string | yes/no/default | รูปแบบ, enum, ขนาด, null/empty semantics |

ระบุรูปแบบวันที่, pagination, identifier ว่าเป็น UUID/requestNo/documentCode
และตัวอย่าง request ที่ไม่ใส่ credentials หรือข้อมูลส่วนบุคคลจริง
หากไม่มี body ให้เขียนชัดเจน ไม่เพิ่ม required rules ที่ source ไม่ได้ตรวจ

## Business logic

1. <ตรวจ input และแปลงค่า>
2. <ตรวจสิทธิ์เจ้าของข้อมูลและสถานะที่อนุญาต>
3. <อ่าน/คำนวณ/เขียนตามลำดับ รวม alternative branches>
4. <การเปลี่ยนสถานะ, completion rules, idempotency, retry/replay>

ระบุกรณีไม่พบข้อมูล, ข้อมูลซ้ำ, race/concurrency ที่มี guard จริง
แยก expected behavior กับข้อจำกัดของ implementation เมื่อไม่ตรงกัน

## Database queries

| Table / relation | Access | Filter / JOIN / วัตถุประสงค์ |
|---|---|---|
| table_name | READ / INSERT / UPDATE / DELETE | ownership, active flag, ordering, columns |

ระบุ repository methods และลิงก์ SQL/source ที่ตรวจ WHERE/JOIN/bind parameters ได้
หาก SQL ใช้ dynamic table/column หรือ IN list ให้ระบุการประกอบ query
แยกตารางจาก domain ออกจาก middleware auth/session/event_logs
ไม่อ้าง query count หรือ latency ที่ไม่ได้วัด และไม่ใส่ SQL จาก domain อื่นที่เข้าไม่ถึง

## Transaction และ side effects

- Transaction boundary: <เริ่ม/commit/rollback, row lock, atomicity และส่วนที่อยู่นอก transaction>
- External services: <storage, SMTP, Omise, FCM; หรือไม่มี>
- Failure behavior: <partial write, cleanup, uncertain commit, retry>
- Idempotency: <key scope, replay rules; หรือไม่มี guarantee>

## Response และ errors

ระบุ HTTP status, Content-Type, envelope, data model และตัวอย่าง success/error
หากเป็น binary/status text ให้ระบุว่าไม่มี JSON envelope
แสดง field types และ optional/null semantics ผ่าน inline schema หรือลิงก์ SCHEMAS.md

| HTTP | Code | Trigger | Caller action |
|---|---|---|---|
| 200 | MA00000 / ไม่มี | สำเร็จ | ใช้ data ตาม model |
| 400/401/404/409/500/503 | code จริงจาก handler | เหตุที่เกิดจริง | แก้ input/login/retry ตาม guarantee |

ไม่ใช้ 404 แทน legacy not-found ที่จริงตอบ 400; ไม่เหมารวม business error ทุกตัวเหมือนกัน

## Source และ verification

- Route / handler / service / repository: <relative Markdown links>
- Contract / tests: <link ที่เกี่ยวข้อง>
- หลักฐานที่ตรวจ: <static review / unit / SQL mock / HTTP fake / live>
- ข้อจำกัด: <ส่วนที่ยังไม่ได้ verify; ความคลาดเคลื่อนที่ต้อง review>

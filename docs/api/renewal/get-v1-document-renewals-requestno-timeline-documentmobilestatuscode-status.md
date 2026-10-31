# GET /v1/document-renewals/:requestNo/timeline/:documentMobileStatusCode/status

ประวัติล่าสุดของ mobile status พร้อมรายการเอกสารและไฟล์ปัจจุบัน

## Overview

- Handler: `renewal.TimelineStatus`
- Authentication: Bearer JWT, active account และ online session ที่ยังไม่หมดอายุ
- Authorization: คำขอต้องเป็นของผู้ใช้ใน JWT และ `is_active='YES'`
- Headers: `Accept-Language`, `devicemodel`, `correlationid` ตาม [COMMON.md](../COMMON.md)

## Request

```http
GET /v1/document-renewals/261000023/timeline/DOCUMENT_REVIEW/status
```

ไม่มี request body

| Path parameter | Type | กติกา |
|---|---|---|
| `requestNo` | string | trim/uppercase; 1–20 ตัว; A–Z, 0–9, `_`, `-` |
| `documentMobileStatusCode` | string | trim/uppercase; รับ mobile status 5 ค่าด้านล่าง |

ค่าที่รับ: `DOCUMENT_REVIEW`, `MARINE_DEPARTMENT_RESULT`,
`DEPARTMENT_DOCUMENT_PICKUP`, `DELIVERING`, `DELIVERED`

## Business logic

1. ตรวจ parameters แล้วค้น active request ด้วย request number + JWT user UUID
2. อ่าน `m_document_transaction` โดยจับคู่ `to_status` กับ
   `m_document_status.document_status_code` แล้วกรอง `document_mobile_status_code`
3. เลือก transaction ล่าสุด เรียง `actioned_at DESC, id DESC LIMIT 1`
   หลาย internal statuses ที่ map เป็น mobile status เดียวกันรวมอยู่ในการค้นนี้
4. คืน `timeline.date_time` จาก `actioned_at` แบบ RFC3339 และ `note` จาก transaction
   ถ้ายังไม่มีประวัติที่ตรง คืนทั้งสอง field เป็น `""` แต่ยังคืน items
5. โหลด items ด้วยตัวประกอบข้อมูลร่วมกับ validate-and-create:
   เรียง master `sort_order`, request-item ID และอ่านไฟล์แบบ batch แยก PROFILE/REQUEST
6. คำนวณ `MISSING`, `NEED_FIX`, `INCOMPLETE`, `COMPLETE` ตาม completion rules เดิม
   ID_CARD ครบ MAIN หรือ FRONT+BACK; PASSPORT/GENERAL ครบ MAIN

`items` เป็นสถานะไฟล์ปัจจุบัน ไม่ใช่ snapshot ณ เวลาที่ transaction เกิดขึ้น
เส้นเดิม `/v1/document-renewals/:requestNo/timeline` ยังคงคืนประวัติทั้งหมดรูปแบบเดิม

## Database queries

| Table | Access | Purpose / filter |
|---|---|---|
| `m_document_request` | READ | requestNo + owner + active; ตรวจ owner ซ้ำใน batch REQUEST files |
| `m_document_status` | READ | สถานะคำขอและ mapping internal status ไป mobile status |
| `m_document_transaction` | READ | request_id + mobile status; เลือก actioned_at/id ล่าสุด |
| `m_document_request_items` | READ | items ของคำขอ เรียงตาม master sort order และ item ID |
| `m_document_master_request_item` | READ | ชื่อเอกสาร, storage scope, sort order |
| `m_document_profile_request_item` | READ | metadata และ uploaded PROFILE files ของเจ้าของคำขอ |
| `m_document_request_item_files` | READ | uploaded REQUEST files ของคำขอ |

Query path: `ownedRequest` → latest transaction → `requestItems` → `validateFilesBatch`
เมื่อสำเร็จใช้ domain queries 4–5 ครั้งตามจำนวน scopes; PROFILE และ REQUEST ใช้ไม่เกิน
scope ละหนึ่ง file query ไม่ขึ้นกับจำนวน items ตัวเลขนี้ไม่รวม Authenticate/event logging
และไม่ใช่ผลวัด latency

SQL: [timelineStatus](../queries/renewal.md#renewaltimelinestatus),
[ownedRequest](../queries/renewal.md#renewalownedrequest),
[requestItems](../queries/renewal.md#renewalrequestitems),
[validateFilesBatch](../queries/renewal.md#renewalvalidatefilesbatch)

Middleware อ่าน `t_session`, `mobile_auth_sessions`, `m_mobile_users`
และเขียน `event_logs` แยกจาก domain ตาม [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

Domain เป็น read-only ไม่มี external calls หรือ write และไม่ได้เปิด transaction
การอ่านหลาย query จึงไม่รับประกัน snapshot เดียวกันเมื่อมีการอัปโหลดพร้อมกัน
ปิด result sets ก่อน query ถัดไป รองรับ pool ที่มี connection เดียว

## Response และ errors

HTTP 200 ใช้ MA envelope เดิม; `data` เป็น
[TimelineStatusResponse](../SCHEMAS.md#renewaltimelinestatusresponse)
และ items/files ใช้ [ValidateItem](../SCHEMAS.md#renewalvalidateitem)
กับ [ValidateFile](../SCHEMAS.md#renewalvalidatefile)

ตัวอย่างย่อ (ค่า illustration ไม่ใช่ผลจากฐานข้อมูลจริง):

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {
    "requestNo": "261000023",
    "documentMobileStatusCode": "DOCUMENT_REVIEW",
    "timeline": {
      "date_time": "2026-10-31T10:30:00Z",
      "note": "Documents received"
    },
    "items": [
      {
        "id": "example-request-item",
        "documentCode": "DOC001",
        "documentMasterRequestItemCode": "MRI004",
        "storageScope": "REQUEST",
        "documentType": "GENERAL",
        "documentName": "Sea Service Record",
        "documentStatus": "COMPLETE",
        "sortOrder": 4,
        "fileUploaded": 1,
        "filePath": "documents/example/request-items/file",
        "fileUploadedAt": "31/10/2026 11:11",
        "files": [
          {
            "fileId": "example-file",
            "documentType": "GENERAL",
            "slotCode": "MAIN",
            "filePath": "documents/example/request-items/file",
            "originalFileName": "sample.pdf",
            "mimeType": "application/pdf",
            "fileSize": 18810,
            "fileUploadedAt": "31/10/2026 11:11"
          }
        ]
      }
    ]
  }
}
```

| HTTP | Code | Trigger |
|---|---|---|
| 400 | MA00004 | ขาด common header |
| 400 | MA00005 | requestNo/mobile status ไม่ถูกต้อง |
| 400 | MA00016 | ไม่พบคำขอ, ไม่ใช่เจ้าของ หรือคำขอ inactive |
| 400 | MA00013 | คำขอไม่มี request items ซึ่งเป็น configuration error |
| 401 | ตาม authentication mapping | JWT/session/account ใช้งานไม่ได้ |
| 500 | MA00013 | อ่าน history/items/files ล้มเหลว |

## Source และ verification

[routes](../../../internal/transport/routes/routes.go),
[handler](../../../internal/domain/renewal/lifecycle_handler.go),
[service](../../../internal/domain/renewal/lifecycle_service.go),
[timeline repository](../../../internal/domain/renewal/timeline_status.go),
[shared items](../../../internal/domain/renewal/request_items.go),
[SQL-mock/handler tests](../../../internal/domain/renewal/timeline_status_test.go)

ใช้ handler + SQL mocks ตรวจ response, filters, mixed-scope batching, validation,
not-found, history errors, empty history และ one-connection pool
ยังไม่ได้เรียก API ที่เปิดอยู่หรืออ่าน/เขียน live MySQL

# GET /docs

Swagger UI (optional)

## Overview

- Register เฉพาะเมื่อ `API_DOCS_ENABLED=true`
- Public; ไม่ผ่าน APIContext หรือ Authenticate จึงไม่บังคับ v1 headers

## Request

ไม่มี body, path parameter หรือ query parameter ที่ handler อ่าน

## Business logic

1. อ่าน embedded `index.html` ของ docs package
2. กำหนด Content-Type เป็น HTML แล้วส่งไฟล์ให้ browser
3. หน้า UI โหลด OpenAPI และ assets ผ่าน `/docs/*`

## Database queries

ไม่มี domain SQL; global RequestLogger พยายาม INSERT `event_logs` หลัง handler
ตาม [กติกากลาง](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

ไม่มี domain transaction หรือ external API call ใน handler
Browser อาจส่ง requests เพิ่มเพื่อโหลด assets/spec

## Response และ errors

สำเร็จ HTTP 200 HTML โดยตรง ไม่มี MA envelope
เมื่อปิด config ไม่ได้ register เส้นนี้ จึงใช้ unmatched-route behavior ของ Fiber

## Source และ verification

[docs/embed.go](../../../docs/embed.go), [app wiring](../../../internal/app/app.go)

ตรวจจาก source เท่านั้น ไม่ได้เปิด server หรือเรียก endpoint

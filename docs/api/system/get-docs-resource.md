# GET /docs/*

Embedded OpenAPI และ Swagger assets (optional)

## Overview

- Register เฉพาะเมื่อ `API_DOCS_ENABLED=true`
- Public; ไม่ผ่าน APIContext/Authenticate

## Request

Wildcard path ต้องตรง allowlist ต่อไปนี้ ไม่มี body หรือ query ที่ handler อ่าน:

| Path suffix | Content-Type |
|---|---|
| `openapi.yaml` | application/yaml |
| `assets/init.js` | JavaScript โดย Fiber c.Type("js") |
| `assets/swagger-ui-bundle.js` | JavaScript โดย Fiber c.Type("js") |
| `assets/swagger-ui.css` | CSS โดย Fiber c.Type("css") |

## Business logic

1. ตรวจ wildcard name อยู่ใน allowlist
2. อ่าน embedded file; path ไม่อนุญาตหรืออ่านไม่พบตอบ 404
3. ตั้ง Content-Type ตาม suffix แล้วส่ง bytes; OpenAPI มีเนื้อหา JSON ที่ใช้เป็น YAML ได้

## Database queries

ไม่มี domain SQL; RequestLogger พยายาม INSERT `event_logs` หลัง handler
ดู [COMMON.md](../COMMON.md#database-access-ร่วม)

## Transaction และ side effects

ไม่มี domain transaction หรือ external API call

## Response และ errors

HTTP 200 ส่ง file bytes โดยตรง ไม่มี MA envelope
HTTP 404 ใช้ SendStatus เป็นข้อความสถานะ; ปิด config แล้วเส้นนี้ไม่ถูก register

## Source และ verification

[docs/embed.go](../../../docs/embed.go), [app wiring](../../../internal/app/app.go)

ตรวจจาก source เท่านั้น ไม่ได้เรียก endpoint หรือ execute SQL

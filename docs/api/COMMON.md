# API conventions

อ้างอิง implementation ณ 2026-10-31 ใช้ร่วมกับไฟล์ endpoint ทุกเส้น
รายละเอียด input/business rules เฉพาะเส้นอยู่ใน spec และ SQL/source ที่ลิงก์

## Headers และ routing

เส้นที่ผ่าน `v1` group ต้องส่ง headers ต่อไปนี้ แม้เป็น public login/register/master:

| Header | Required | ความหมาย |
|---|---|---|
| `Accept-Language` | yes | ต้องไม่ว่าง; โดยทั่วไป en/EN เลือกอังกฤษ ค่าอื่นใช้ไทย ไม่ได้ validate enum กลาง |
| `devicemodel` | yes | ต้องไม่ว่าง |
| `correlationid` | yes | ต้องไม่ว่าง; login บันทึกลง session |
| `Authorization` | protected routes | `Bearer <JWT>`; prefix ต้องตรงตามนี้ |
| `X-Request-ID` | no | ใช้ค่าที่ส่งมา หรือสร้าง UUID แล้วส่งกลับใน response header |
| `Content-Type` | ตาม body | JSON หรือ multipart ตาม endpoint |

APIContext สร้าง trace ID และตรวจ headers ก่อน Authenticate/handler
header ว่างคืน HTTP 400 `MA00004` พร้อม description `[Header] Missing parameter ...`
`/`, `/health`, `/privacy_policy`, `/v1/payments/omise/webhook` register นอก v1 group
จึงไม่บังคับ headers กลุ่มนี้ Webhook ใช้ signature headers ของตัวเอง
ตัวอย่าง URL ใน spec อิง root ของ server ไม่มีการกำหนด production host

## Authentication

Protected routes ตรวจ HS256 JWT, issuer=`smart-seaman-mobile-api`, exp, subject และ jti
จากนั้นตรวจ session online พร้อม expiry และ active user โดยใช้ฐานข้อมูลทุก request
JWT อายุ 30 นาที; refresh credential แยกจาก jti และหมดอายุ 30 วันหลัง login
การ refresh หมุน hash atomically ไม่ขยาย session expiry; client ต้อง persist refToken ใหม่

| HTTP | Code | Trigger |
|---|---|---|
| 401 | `MA00004` | ไม่มี Authorization |
| 401 | `MA00011` | prefix ไม่ใช่ `Bearer ` |
| 401 | `MA00007` | JWT invalid/expired/issuer หรือ subject ไม่ถูกต้อง |
| 401 | `MA00018` | ไม่มี jti แบบ string |
| 401 | `MA00026` | ไม่พบ session, offline, หมดอายุ หรือ query session ล้มเหลว |
| 401 | `MA00016` | ไม่พบ active user, username ไม่ตรง subject หรือ query user ล้มเหลว |

Identity ที่ส่งต่อให้ handler คือ `user.mobile_uuid`, `user.username`, `jwt.subject`
Ownership ของ request/address/certificate ตรวจเพิ่มตาม SQL ของแต่ละเส้น
ไม่มี role-based admin authorization กลางใน Authenticate นี้

## Response envelope

JSON success ส่วนใหญ่ตอบ HTTP 200:

```json
{
  "code": "MA00000",
  "description": "Success",
  "data": {"example": "ดู model ของ endpoint"}
}
```

JSON error ส่วนใหญ่ไม่มี data:

```json
{
  "code": "MA00005",
  "description": "รายละเอียดขึ้นกับ handler และภาษา"
}
```

`data` ใช้ `omitempty`: nil/empty string อาจไม่ส่ง แต่ struct หรือ typed empty slice
มีพฤติกรรมตามชนิด Go ที่อยู่ใน interface; ดู source เมื่อพึ่ง null/empty behavior
`httpx.Success` ใช้ Success เมื่อภาษา en/EN นอกนั้นใช้ สำเร็จ
Auth handler อ่าน description จาก m_message_code; success description cache 5 นาที
ต่อ process พร้อม coalescing; error descriptions/failed reads ไม่ cache
เมื่อ lookup ล้มเหลวมี fallback description ไม่ได้ทำให้ operation ล้มเหลว

Legacy validation/not-found โดยทั่วไปตอบ 400; technical exception ตอบ 500 `MA00013`
Payment ใช้ PAYMENT_* กับ 400/404/409/503 ตาม mapping ในแต่ละ spec
บาง config BusinessError MA00013 ตอบ 400 ต่างจาก ordinary technical error
Binary downloads, root, health และ webhook มีรูปแบบต่างออกไปตาม spec
Fiber/framework errors เช่น body limit, CORS หรือ unmatched route อาจไม่ได้ใช้ MA envelope
Fiber instance ปัจจุบันใช้ default BodyLimit จึงอาจปฏิเสธ body ก่อนถึง validation
ขนาดไฟล์ 10 MiB ใน service ไม่ใช่ guarantee ว่า HTTP body ทุกขนาดถึง 10 MiB ผ่าน ingress ได้

## Database access ร่วม

| Component | Table | Access / behavior |
|---|---|---|
| Authenticate | `t_session` + `mobile_auth_sessions` | SELECT JOIN ด้วย CLIENT_SESSION_ID=jti, expires_at>NOW(); ตรวจ IS_ONLINE=YES |
| Authenticate | `m_mobile_users` | SELECT USERNAME ด้วย MOBILE_UUID ของ session, USER_STATUS=A; เทียบ JWT subject |
| Auth response descriptions | `m_message_code` | SELECT MESSAGE_CODE; auth success cache 5 นาที; error lookup ไม่ cache |
| RequestLogger | `event_logs` | INSERT event หลัง handler; synchronous, รวม root/health/webhook/docs ใน app wiring |

ตารางใน endpoint spec แยก domain reads/writes กับ common access นี้ เพื่อไม่ให้เข้าใจ
ว่า GET ที่ไม่มี domain SQL ไม่แตะ database เลย ตัว logger พยายามบันทึก success/failed event
และละ recorder error; เป็น write เพิ่มนอก domain transaction
Auth descriptions แสดงในตาราง domain ของ auth specs ด้วยเพื่อให้เห็น query path ครบ

SQL catalog อยู่ใน `queries/*.md` มี source links พร้อม line ที่อ้างอิง ณ วันที่จัดทำ
เป็น static literals/fragments; shared helpers มีหลาย branch, dynamic identifiers,
WHERE suffix และ placeholders ให้ดู arguments จาก source ไม่ใช่ SQL พร้อมรัน
`list_course_school_training` เป็น relation ที่โค้ด SELECT โดยตรง; checkout นี้ไม่ได้
ยืนยัน DDL หรือฐานตารางของ relation จึงไม่เดารายชื่อตารางด้านหลัง

## External services และ consistency

- Object storage เก็บ profile/certificate/request-item files และ Base64 legacy content
  โดยรูปแบบการ decode ต่างกันตาม service
- SMTP ส่ง activation/reset links หลังบันทึก token แล้ว; ไม่อยู่ใน DB transaction
- Omise ใช้ durable attempts; GET payment และ webhook อาจเขียน ledger/request state
- `/fcm-update` เก็บ token อย่างเดียว; `/send-noti` ปัจจุบันเป็น success placeholder
- โค้ด request-item มี cleanup object ใหม่เมื่อ save ล้มเหลว ยกเว้น uncertain commit
  ส่วน profile/certificate upload ไม่ได้มี cleanup guarantee เดียวกัน

## Optional documentation routes

เมื่อ `API_DOCS_ENABLED=true`:

| Method | Path | ผลลัพธ์ |
|---|---|---|
| GET | `/docs` | embedded Swagger HTML |
| GET | `/docs/openapi.yaml` | embedded OpenAPI; Content-Type application/yaml (เนื้อหา JSON) |
| GET | `/docs/assets/init.js` | JavaScript |
| GET | `/docs/assets/swagger-ui-bundle.js` | JavaScript |
| GET | `/docs/assets/swagger-ui.css` | CSS |

Public, ไม่ผ่าน APIContext/Authenticate, ไม่มี domain SQL; RequestLogger ยังทำงาน
`/docs/*` อนุญาตเฉพาะรายการด้านบน path อื่นตอบ 404 status text
เมื่อปิด config routes นี้ไม่ถูก register ดู [embed.go](../../docs/embed.go)

## Source และ verification

[routes](../../internal/transport/routes/routes.go),
[APIContext](../../internal/middleware/api.go),
[Authenticate](../../internal/middleware/auth.go),
[RequestLogger](../../internal/middleware/requestlog.go),
[response helper](../../internal/transport/httpx/response.go),
[app wiring](../../internal/app/app.go)

ชุดเอกสารนี้ตรวจจาก source ไม่ได้เรียก API, apply migration หรือรัน live integration
ไฟล์ [KNOWN-GAPS.md](KNOWN-GAPS.md) แยกข้อจำกัดที่มีผลต่อการนำ spec ไปใช้งาน

# Spring Boot / Go parity: remaining API groups

วันที่ตรวจ: 1 พฤศจิกายน 2026 — ต่อจาก [renewal/payment audit](migration-parity-audit.md)

## สถานะการแก้หลังตรวจ

แก้แล้ว: การอ่านไฟล์ legacy Base64/binary พร้อม MIME, delivery master JOIN,
transaction/locks, snapshot update, request UUID, สร้าง default address เมื่อยังไม่มี,
duplicate-default detection และ address length/postal validation
รายละเอียดข้างล่างเป็นข้อค้นพบก่อนแก้; รายการอื่นยังคงเปิดอยู่
ดูผลและข้อจำกัดใน [MIGRATION.md](../MIGRATION.md#priority-parity-fixes-stored-files-and-delivery-addresses)

## วิธีอ่านผล

ตรวจ controller, DTO, service และ SQL ของกลุ่มด้านล่างจาก checkout ทั้งสองฝั่ง
คำว่า **ใกล้เคียง** หมายถึง flow/contract หลักที่อ่านตรงกัน ไม่ใช่ผ่าน live compatibility test
ยังไม่ได้เรียกฐานข้อมูล, object storage, SMTP หรือ FCM และไม่ได้แก้ runtime code
ไม่ควรตีความรายการนี้ว่า validation, null handling และ concurrency ทุกกรณีตรงกันแล้ว

| กลุ่ม / routes | ผลตรวจ source |
| --- | --- |
| Auth: login/register/refresh/activation/reset/change password | มี security redesign ตาม AGENTS; พบ validation ต่างเพิ่มเติม |
| Profile: GET/update/image/inactive | พบ file-format incompatibility, mobile validation และ age calculation ต่าง |
| Delivery: default/create/update/create-for-renewal | พบ SQL schema mismatch และ flow/transaction ที่ย้ายไม่ครบ |
| Master: company/position/documents/courses/Thai addresses | query/response หลักใกล้เคียง; error behavior และรายละเอียดภาษาไม่เท่ากันทั้งหมด |
| Certificates: COT/DOC/expiration/create/update/delete/edit/view | พบ legacy file format, validation และ date/pagination ต่าง |
| Request items: profile list/upload/preview, renewal upload | upload rules หลักใกล้เคียง; preview ขาด signed URL และ retention ต่าง |
| Forms/news/banner | query และ successful response หลักใกล้เคียง; required-query/error behavior ต่าง |
| Vouchers | list/detail fields และ personal/global filtering ใกล้เคียง; list ordering ต่าง |
| School trainings | grouping, date range และ query หลักใกล้เคียง; null/error/tie ordering ยังไม่ยืนยันด้วย fixtures |
| FCM token update | required token, insert/update ตาม user และ success text ใกล้เคียง; ไม่ได้แปลว่าส่ง push ได้แล้ว |
| Notifications | list/count ใกล้เคียง; mark-by-value อัปเดตจำนวนรายการต่างกัน และ Go เพิ่ม ownership filter ใน mark-by-ID |
| Renewal my/documents/prices/statuses/stage | list และ price/document contract ต่าง; statuses/stage mapping หลักใกล้เคียง |
| Root/health/privacy/app-links/send-email | app-links/send-email ขาดตามรายงานแรก; privacy ทั้งคู่คืน success placeholder |

## P1 — รูปโปรไฟล์/ใบรับรองเดิมอ่านผิดรูปแบบ

Java `ProfileService.profileUpdate` และ `DocumentService.documentCreate/documentUpdate`
เขียน Base64 string ลง S3 แล้ว controller decode ตอนดาวน์โหลด
Go decode ก่อนเขียน binary และ `Image`/`View` ส่ง bytes จาก storage ตรง ๆ
adapter `storage.Client.Get` ไม่มี legacy decoding

ผลที่คาดจาก source: เมื่อชี้ Go ไปยัง bucket เดิม ไฟล์ที่ Java เคยเขียนจะถูกส่งเป็นข้อความ Base64
แทนรูป/PDF; ในทิศกลับ Java จะ decode ไฟล์ binary ที่ Go เขียนไม่ได้ตาม flow เดิม
ต้องกำหนดการอ่านได้ทั้งสองรูปแบบหรือวางแผนแปลงข้อมูลก่อนสลับระบบ
ข้อนี้ไม่รวม multipart request-item uploads ซึ่ง Java เขียน binary อยู่แล้ว

หลักฐาน: [Java profile](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/ProfileService.java),
[Java profile controller](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/ProfileController.java),
[Java certificates](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentService.java),
[Java document controller](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/DocumentController.java),
[Go profile](../internal/domain/profile/service.go), [Go certificates](../internal/domain/document/service.go),
[Go storage](../internal/platform/storage/storage.go)

## P1 — Delivery validation ใช้คอลัมน์ไม่ตรง legacy schema

Go `ValidAddress` join ด้วย `d.province_code` และ `s.district_code`
แต่ schema ที่เก็บใน repository และ Java ใช้ `d.province_id` / `s.district_id` join ไปที่ `id`
ดังนั้นบน schema ตามสคริปต์ต้นทาง create/update/create-for-renewal ของ Go จะเกิด SQL error
ก่อนบันทึกที่อยู่ ไม่ใช่แค่ validation ให้ผลต่างกัน
ยังไม่ได้ตรวจ schema ของฐานข้อมูลที่กำลังรันจริง

หลักฐาน: [Go ValidAddress](../internal/domain/delivery/repository.go),
[Java address validation](../../smart-seaman-mobile-api/src/main/java/com/seaman/repository/ThailandAddressRepository.java),
[legacy schema](../migrations/legacy-mvp1/04_import_thailand_address_master.sql)

## P1 — Delivery transaction และ snapshot update ย้ายไม่ครบ

- Java create/update ล็อก user/addresses และ clear/set default ใน transaction เดียว
  Go `ClearDefault` กับ insert/update เป็นคนละ autocommit operation
  หากคำสั่งหลังล้มเหลว default เดิมอาจถูกล้างไปแล้ว; concurrent creates อาจแข่งกัน
- Java `update` หา active address ไม่เจอแล้วลองอัปเดต owned renewal snapshot
  Go หาเฉพาะ `m_delivery_address` จึงแก้ snapshot ID ผ่านเส้นเดียวกันไม่ได้
- Java create-for-renewal รับ requestNo หรือ UUID, ล็อกคำขอ และสร้าง default address ให้ด้วยถ้ายังไม่มี
  Go lookup เฉพาะ requestNo และเพิ่ม snapshot อย่างเดียวโดยไม่ได้ล็อกตลอด check-and-insert
- Java GET default พบหลายแถวแล้วตอบ database error; Go `LIMIT 1` เลือกแถวล่าสุด ทำให้ซ่อนความผิดปกติ
- Go ยังขาด DTO max lengths ของ Java เช่นชื่อ 100 และ addressLine 500 ตัวอักษร

หลักฐาน: [Java delivery service](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DeliveryAddressService.java),
[Java DTO](../../smart-seaman-mobile-api/src/main/java/com/seaman/model/request/DeliveryAddressRequest.java),
[Go delivery service](../internal/domain/delivery/service.go), [Go SQL](../internal/domain/delivery/repository.go)

## P1 — Request-item preview ไม่มี URL สำหรับเปิดไฟล์

Java `/documents/profile-items/{itemCode}/preview` ตรวจ active item และคืน `fileUrl`
ที่เป็น presigned URL; Go `PreviewResponse` ใช้ file metadata ที่มี `filePath` แต่ไม่มี `fileUrl`
และเมธอด `Preview` ไม่ตรวจ active item ก่อนอ่าน files แบบ Java
client ที่เปิด URL เดิมจึงใช้ response Go แทนโดยตรงไม่ได้
สำหรับ renewal preview มี contract/query mismatch อีกชุดตามรายงานแรก

Upload ทั้งคู่รองรับ 10 MiB, PNG/JPEG/PDF และ ID_CARD/PASSPORT/GENERAL slot rules
แต่ Java ลบ record ต่างชนิดและ cleanup object ที่ถูกแทนหลัง commit
Go เก็บ row ต่างชนิดเป็น `file_uploaded=0` และไม่มี cleanup object เก่าเมื่อแทนที่สำเร็จ
นี่เป็นความต่างด้าน retention/storage ที่ต้องระบุให้ชัด ไม่ใช่ปัญหาการ decode แบบ certificates

หลักฐาน: [Java profile-item service](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRequestItemFileService.java),
[Java renewal-item service](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalRequestItemFileService.java),
[Go item service](../internal/domain/requestitem/service.go),
[Go models](../internal/domain/requestitem/model.go), [Go persistence](../internal/domain/requestitem/repository.go)

## P1 — Renewal list และ prices ยังไม่เท่ากัน

- `/document-renewals/my`: Java คืน documentName/documentNameEn, isResubmit,
  status object พร้อมชื่อ/สี/mobileStatus/step และเวลา Bangkok แบบแสดงผล
  Go คืน status string, statusName ภาษาอังกฤษ, RFC3339 และไม่มี field ที่เหลือ
- `/document-renewals/prices`: Java ปฏิเสธเมื่อมีราคาที่ active ซ้อนกันหรือ total ติดลบ,
  คืน priceSettingId และคำนวณด้วย BigDecimal
  Go เลือก `LIMIT 1`, ไม่ตรวจ total ติดลบ, ไม่คืน priceSettingId และใช้ float64
- `/document-renewals`: Java เติม documentMobileStatusCode และชื่อ renewal status
  แต่ Go model/query ไม่มีข้อมูลกลุ่มนี้

หลักฐาน: [Java my list](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalListService.java),
[Java renewal service](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalService.java),
[Java prices query](../../smart-seaman-mobile-api/src/main/java/com/seaman/repository/DocumentRenewalRepository.java),
[Go lifecycle models/SQL](../internal/domain/renewal/lifecycle.go),
[Go renewal models](../internal/domain/renewal/model.go), [Go renewal query](../internal/domain/renewal/repository.go)

## P2 — Validation และการคำนวณวันไม่ตรง

- `/profile-update`: Java `@Pattern` บังคับเบอร์ `^0[0-9]{9}$` และ service trim เบอร์
  Go ตรวจเพียงไม่ว่าง จึงยอมรับเบอร์ผิดรูปแบบได้; `@NotBlank` ของ Java ยังปฏิเสธ whitespace-only
  แต่ Go ใช้ `== ""` ใน profile/register/login หลาย field
- Certificates create/update: Java บังคับ certEndDateType/certEndDate และ update ต้องมี isChangeFile
  Go ไม่ตรวจ required fields กลุ่มนี้ครบ แม้มีการตรวจ parse วันที่
- GET profile: Go คำนวณ age ด้วย YearDay ซึ่งผิดได้เมื่อปีเกิดเป็น leap year เช่นเกิด 2000-11-01
  ณ 2026-11-01 ได้ 25 ทั้งที่ Java Period ให้ 26
- Certificate list/expiration: Java ใช้ `Period.between`; Go ลบ year/month/day แล้ว clamp วันติดลบเป็น 0
  ไม่ได้ยืมวันจริงจากเดือนก่อน เช่น Jan 31 → Mar 1 ในปีปกติ จึงแสดงระยะเหลือต่างกัน
  Go ยังใช้การเทียบเวลา ณ เที่ยงคืนสำหรับ expired ทำให้วันหมดอายุวันนี้อาจถือว่าหมดแล้ว
  แต่ Java เทียบ LocalDate และถือว่าวันนี้ยังไม่หมด
- Pagination `last` ใน certificate list/expiration คำนวณต่างกัน; Go ใช้ offset+จำนวนที่คืนเทียบ total
  ส่วน Java มีเงื่อนไขเทียบ page size กับ offset ที่อาจให้ผลผิด ไม่ควรคัดลอกกลับโดยไม่กำหนด expected behavior

หลักฐาน: [Java profile DTO](../../smart-seaman-mobile-api/src/main/java/com/seaman/model/request/ProfileRequest.java),
[Java create DTO](../../smart-seaman-mobile-api/src/main/java/com/seaman/model/request/DocumentCreateRequest.java),
[Java update DTO](../../smart-seaman-mobile-api/src/main/java/com/seaman/model/request/DocumentUpdateRequest.java),
[Java date utility](../../smart-seaman-mobile-api/src/main/java/com/seaman/utils/DateUtil.java),
[Go profile calculations/validation](../internal/domain/profile/service.go),
[Go certificate calculations/validation](../internal/domain/document/service.go),
[Go auth validation](../internal/domain/auth/service.go)

## P2 — Notifications และ voucher ordering

- `/notifications/update/valueId`: Java เลือก notification ล่าสุดที่ตรง user/valueId/notiType แล้ว mark แถวเดียว
  Go UPDATE ทุกแถวที่ตรง จึงเปลี่ยน unread count ต่างกันเมื่อมีรายการซ้ำ
- `/notifications/update`: Go เพิ่ม owner filter ใน UPDATE ซึ่ง Java ไม่มี
  ต้องรักษาการป้องกันแก้ notification ของผู้อื่นไว้ ไม่ย้อนเพื่อให้เหมือน Java
- `/vouchers`: Java เรียง `CREATE_DATE DESC` หลังรวม PERSONAL/GLOBAL; Go เรียง `VOUCHER_ID`
  field/filter หลักตรงกัน แต่ลำดับคูปองที่แสดงไม่ตรง

หลักฐาน: [Java notifications](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/SendNotificationService.java),
[Java notification SQL](../../smart-seaman-mobile-api/src/main/java/com/seaman/repository/SendNotificationRepository.java),
[Go notification SQL](../internal/domain/notification/repository.go),
[Java voucher SQL](../../smart-seaman-mobile-api/src/main/java/com/seaman/repository/VoucherRepository.java),
[Go catalog SQL](../internal/domain/catalog/repository.go)

## Error contract และส่วนที่ใกล้เคียง

Java หลายเส้นใช้ Spring required `@RequestParam`/`@Valid` ก่อนเข้า service
แต่ Go forms/news/catalog handlers อ่าน `c.Query` แล้วส่งค่าว่างต่อโดยไม่ตรวจ required query
เช่น formId/newsId/courseCode ผลจึงอาจกลายเป็น not-found, technical error หรือ empty success
แทน validation error เดิม และ Java BusinessException ใช้ HTTP 200 ขณะที่ Go หลาย handler ใช้ 400/500
จึงต้องเทียบทั้ง HTTP status กับ MA code; envelope ชื่อเดียวกันไม่พอ

Query หลักที่ตรวจแล้วใกล้เคียง: forms อ่าน `m_forms` และ decode stored Base64 PDF;
news แยก SHIP/GENERAL และเรียง UPDATE_DATE; banners เรียง CREATE_DATE;
master company/position/documents/courses และ Thai address lookups;
school grouping ตาม company+course/date range; FCM token insert/update
ยังต้องมี fixtures ครอบคลุม missing query, no rows, SQL NULL, language, MIME และเวลา
ก่อนรับรอง compatibility

หลักฐาน: [Java forms/news controllers](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/FormController.java),
[Java news controller](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/NewsController.java),
[Java exception handling](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/advice/ExceptionAdvice.java),
[Go content handlers](../internal/domain/content/handler.go), [Go catalog handlers](../internal/domain/catalog/handler.go),
[Go master](../internal/domain/master/repository.go), [Go FCM](../internal/domain/fcm/service.go)

## ขอบเขตที่ยังต้องทดสอบเพิ่ม

- Auth: แยก intentional token/session hardening ออกจาก contract regressions;
  ไม่ย้อน single-use activation/reset หรือ refresh rotation
- Validate-and-create, stage และ cancel: อ่าน flow หลักประกอบการ review แล้ว
  แต่ยังไม่ได้พิสูจน์ parity ทุก branch/error ด้วย fixtures; webhook ยังต้องตรวจเทียบเพิ่ม
  โดยเฉพาะ signature และ concurrent callbacks
- Profile email: Java SQL ก็ comment การแก้ USERNAME/EMAIL ไว้เหมือน Go
  จึงไม่จัดการไม่ persist email เป็น regression ของ Go โดยลำพัง
- Logging: Java เขียน transaction logs ราย service; Go ใช้ request event logs
  ไม่ควรถือเป็น audit trail เดียวกันโดยอัตโนมัติ

ลำดับแก้ที่แนะนำ: legacy file compatibility → delivery SQL/transactions/snapshots →
list/preview contracts → price invariants → validation/date/notification/order
รอบนี้เพิ่มเฉพาะเอกสาร ไม่ได้รัน test suites หรือบริการจริง

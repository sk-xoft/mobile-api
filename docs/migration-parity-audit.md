# Spring Boot / Go parity audit

วันที่ตรวจ: 1 พฤศจิกายน 2026

ดูผลตรวจกลุ่ม API อื่นเพิ่มเติมใน [remaining API groups](migration-parity-other-apis.md)

## ขอบเขตและข้อสรุป

เปรียบเทียบ source ใน `mobile-api` กับ `../smart-seaman-mobile-api` โดยเน้น
renewal/payment ที่พบปัญหา ตรวจ route inventory และตรวจ business logic ของเส้นที่ระบุด้านล่าง
การตรวจนี้เป็น static source review ไม่ได้พิสูจน์ว่า API ทุกเส้นทำงานเท่ากัน
ไม่ได้เรียก Omise, ส่ง notification, เขียนฐานข้อมูล หรือรัน SQL migration

Go ยังไม่ใช่การย้าย Spring Boot แบบ 1:1: มีทั้ง business logic ที่ย้ายไม่ครบ,
response contract ที่เปลี่ยน และ payment architecture ที่ออกแบบเพิ่ม
เครื่องหมาย `[x]` ใน MIGRATION.md จึงหมายถึงมี implementation แล้ว
ไม่ใช่หลักฐานว่าตรงกับ Java ทุกกรณี

## ช่องว่างที่ควรแก้ก่อน

P1 = กระทบความครบถ้วน/ความสอดคล้องของข้อมูลหรือ client contract หลัก;
P2 = ความต่างของ validation, compatibility หรือฟังก์ชันรอง

### P1 — POST /v1/document-renewals สร้างข้อมูลไม่ครบ

- Java ตรวจ UUID และเจ้าของที่อยู่ที่ยัง active, ตรวจเอกสาร PROFILE ที่จำเป็น,
  บันทึก price setting, สร้าง address snapshot, request items และ CREATE history
  ภายใน transaction เดียว และคืน snapshot ID
- Go ตรวจเพียงค่าบังคับและราคา แล้วเพิ่ม sequence และ request แยกคำสั่ง
  บันทึก `price_setting_id=NULL`; ไม่สร้าง snapshot/items/history และคืน address ID จาก input
- ผลกระทบ: API อาจตอบสำเร็จทั้งที่ข้อมูลสำหรับขั้นตอนถัดไปยังไม่ครบ
  และการอ่าน sequence หลัง increment โดยไม่มี transaction อาจแข่งกับ request อื่น
- หลักฐาน: [Java create](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalCreateService.java),
  [Go CreateRequest](../internal/domain/renewal/lifecycle_service.go),
  [Go repository create](../internal/domain/renewal/lifecycle.go) เมธอด `create`
- ข้อนี้เป็นเส้น create โดยตรง ไม่ใช่ `requests/validate-and-create` ซึ่งมี implementation อีกชุด

### P1 — GET /v1/document-renewals/{requestNo} คืน detail ไม่ตรง

- Java คืน `documentName`, `status` เป็น object, `items` และข้อมูล department/delivery ตามสถานะ
  ไฟล์มี presigned `fileUrl` อายุ 10 นาที; เวลาแสดงเป็น `dd/MM/yyyy HH:mm` ใน Asia/Bangkok
- Go `Detail` ไม่มีข้อมูลเหล่านี้ โดย `status` เป็น string และ `submittedAt` เป็น RFC3339
- ผลกระทบ: mobile client ที่อ่านโครงสร้างเดิมจะไม่พบข้อมูลหรือเจอชนิดข้อมูลผิด
- หลักฐาน: [Java detail](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalDetailService.java),
  [Go Detail และ repository detail](../internal/domain/renewal/lifecycle.go)

### P1 — PUT /v1/document-renewals/requests/{requestNo}/mobile อัปเดตไม่ครบ

- Java ล็อกคำขอและอัปเดตทั้งคำขอกับ delivery snapshot ใน transaction เดียว
  คืน `{requestNo,mobileNumber}`
- Go อัปเดตเฉพาะ `m_document_request` และคืน `TransitionResponse`
  โดยใส่เบอร์เก่า/ใหม่ลง `fromStatus`/`toStatus`
- ผลกระทบ: เบอร์ติดต่อกับเบอร์จัดส่งอาจไม่ตรง และ response ไม่ตรงกับ client เดิม
- หลักฐาน: [Java mobile](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalMobileService.java),
  [Go updateMobile](../internal/domain/renewal/lifecycle.go),
  [Go validation](../internal/domain/renewal/lifecycle_service.go)

### P1 — Resubmit ไม่ atomic ตลอด flow

- Java ล็อกคำขอก่อนตรวจ/reset FIX, เปลี่ยนสถานะ, เพิ่ม history และ publish event ใน transaction
- Go ตรวจและ `resetFixItems` ก่อนเข้า transaction ของ `transition`
  ถ้าขั้นตอนเปลี่ยนสถานะล้มเหลว การ reset ที่ทำไปแล้วจะไม่ rollback ตามกัน
  และไม่มีการ publish resubmission event แบบ Java
- หลักฐาน: [Java resubmit](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalResubmitService.java),
  [Go Resubmit](../internal/domain/renewal/lifecycle_service.go),
  [Go resetFixItems/transition](../internal/domain/renewal/lifecycle.go)

### P2 — Preview alias รับ query parameter ไม่ครบ และ response เปลี่ยน

- Java `/v1/document-renewals/items/preview` รับทั้ง `request_no` และ `request_master_items_code`
- Go อ่านเฉพาะ `request_no` ส่วน item code อ่านจาก path ซึ่งไม่มีใน alias นี้
  จึงส่ง item code ว่างไป lookup แม้ client ส่ง query ครบตาม Java
- Preview ของ Java คืน detail item พร้อม file URL แต่ Go คืน item code/type และ file metadata/path
- หลักฐาน: [Java preview controller](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/DocumentRenewalController.java),
  [Go RenewalPreview](../internal/domain/requestitem/handler.go),
  [Go PreviewResponse](../internal/domain/requestitem/model.go)

### P2 — Legacy timeline และ manual notification ยังไม่เท่ากัน

- Timeline เดิมของ Java คืนชื่อสถานะตามภาษา, `detail` จาก action และเวลา Bangkok แบบแสดงผล
  Go คืน status codes, `note` และ RFC3339 จึงไม่ใช่ contract เดิม
- Java `/v1/send-noti` เรียก service เพื่อสร้าง notification และ publish FCM event
  Go handler คืน success ทันที ไม่มี side effect นี้
- หลักฐาน: [Java timeline](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalTimelineService.java),
  [Go timeline](../internal/domain/renewal/lifecycle.go),
  [Java notification controller](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/NotiController.java),
  [Go Manual](../internal/domain/notification/handler.go)

## Payment: อะไรต่าง และอะไรเหมือน

| ประเด็น | Spring Boot | Go ปัจจุบัน |
| --- | --- | --- |
| ก่อนสร้าง charge | transaction ของ request/payment | บันทึก `mobile_payment_attempts` และ `mobile_payment_request_state`, lease และ worker |
| `m_payment_transaction` | insert หลัง Omise สร้าง charge สำเร็จ | insert/update เมื่อได้ charge ที่ตรวจสอบแล้ว |
| Replay/status GET | อ่าน payment ที่บันทึกไว้ | reconcile กับ provider; อาจเขียนข้อมูลและเริ่ม charge ของ RESERVED attempt |
| Provider HTTP error | BusinessException `MA00013` (constructor ใช้ HTTP 200) | 401/403 เป็น HTTP 503 `PAYMENT_UNAVAILABLE`; uncertain failures ยังอาจคืน PROCESSING เพื่อ reconcile |
| Input validation | จำกัด bank methods, key charset, returnUri max 500/จำเป็นสำหรับ mobile banking, platform IOS/ANDROID | ตรวจ key ว่าง/ความยาว และรับทุก method ที่ขึ้นต้น `mobile_banking_`; ข้อกำหนดอื่นยังไม่ครบ |
| Provider request | มี `Idempotency-Key` header และตรวจชนิด/trim secret key | ใช้ attempt reference/metadata และ local coordination; ไม่มี header ดังกล่าว และ Ready ไม่ตรวจรูปแบบ key |
| Response | มี `providerSourceId`, `paidAt` | ไม่มีสอง field นี้; เพิ่ม serverTime/qrExpired/nextAction/retryAfterSeconds |

หลักฐาน: [Java payment service](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/DocumentRenewalPaymentService.java),
[Java input](../../smart-seaman-mobile-api/src/main/java/com/seaman/model/request/DocumentRenewalPaymentRequest.java),
[Java provider](../../smart-seaman-mobile-api/src/main/java/com/seaman/service/OmisePaymentClient.java),
[Java BusinessException](../../smart-seaman-mobile-api/src/main/java/com/seaman/exception/BusinessException.java),
[Java exception response](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/advice/ExceptionAdvice.java),
[Go payment service](../internal/domain/payment/service.go), [Go response](../internal/domain/payment/model.go),
[Go store](../internal/platform/mysql/payment/store.go), [Go provider](../internal/platform/payment/omise.go)

ดังนั้นการไม่มีแถวใน `m_payment_transaction` เมื่อ authentication กับ Omise ล้มเหลว
ไม่ได้เกิดจาก Go ลืม insert เมื่อเทียบกับ Java ทั้งคู่ยังไม่ได้รับ charge ที่ยืนยันได้
ความต่างคือ Go มี attempt record และเคยกลืน provider error เป็น PROCESSING
การแก้ 401/403 ล่าสุดเปลี่ยนให้ตอบ error แล้ว ส่วน retry/reconciliation เป็นการออกแบบเพิ่ม
ควรกำหนด contract นี้ให้ชัดก่อนปรับ และคงการป้องกัน charge ซ้ำไว้
ไม่ได้ตรวจยืนยันความสามารถของ Omise ปัจจุบันจากภายนอกในการ review นี้

## Route inventory

นับ explicit method/path registrations จาก controllers เทียบ Go โดยแทนชื่อ path parameter
ด้วย placeholder เดียวกันและตัดรายการซ้ำ: Java 77, Go 76, ตรงกัน 74
ไม่รวม implicit HEAD/OPTIONS และ routes ที่ framework/config สร้าง เช่น actuator/Swagger

- Java มี แต่ Go ไม่มี: `GET /.well-known/assetlinks.json`,
  `GET /.well-known/apple-app-site-association`, `GET /v1/send-email`
- Go เพิ่ม: `GET /v1/document-renewals/{requestId}/payments` และ
  `GET /v1/document-renewals/{requestNo}/timeline/{documentMobileStatusCode}/status`
- Route ที่ตรงกันไม่ได้แปลว่า request/response, validation และ side effects ตรงกัน

หลักฐาน: [Java routes](../../smart-seaman-mobile-api/src/main/java/com/seaman/constant/Routes.java),
[Java app links](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/AppActivateController.java),
[Java email route](../../smart-seaman-mobile-api/src/main/java/com/seaman/controller/EmailController.java),
[Go routes](../internal/transport/routes/routes.go)

## ความต่างที่ต้องรักษาการตัดสินใจเดิม

- Auth hardening เช่น token hashes/rotation และ single-use action tokens มีข้อกำหนดใน
  [AGENTS.md](../AGENTS.md) และ [MIGRATION.md](../MIGRATION.md) ไม่ควรย้อนเพื่อเลียนแบบ Java
- การยกเลิก owner condition เฉพาะ hard-delete และ timeline/status endpoint ใหม่เป็นคำขอของผู้ใช้
- Config Go เปลี่ยนเป็น `DATABASE_DSN`, `OBJECT_STORE_*`, `SMTP_*`;
  ต้องแปลงทั้ง environment และค่าจาก Java properties ไม่ใช่คัดลอก `.env` เดิมอย่างเดียว
  การตรวจนี้ไม่ได้แก้ external config เพิ่ม

## ลำดับแก้และการพิสูจน์

1. เติม create transaction/snapshot/items/history และทำ resubmit atomic
2. คืน detail/mobile/preview contract ให้ตรง หรือประกาศ contract ใหม่พร้อมแก้ client
3. เติม payment validation/fields ที่ตกหล่น โดยคง lease/reconciliation และ auth-error handling
4. ตัดสินใจว่าจะรองรับ routes/app links/manual notification ที่เหลืออย่างไร
5. เพิ่ม compatibility fixtures จาก Java: เทียบ JSON types/fields, error code/status,
   database side effects, rollback, ownership และ concurrency ใช้ข้อมูลทดสอบเดียวกัน

รอบนี้ไม่ได้รัน test suite ใหม่เพราะไม่มีการแก้ runtime code
unit tests ของ Go ที่ผ่านก่อนหน้านี้พิสูจน์เฉพาะพฤติกรรมที่ทดสอบไว้
ยังใช้แทนการทดสอบเทียบ Java และ mobile client ไม่ได้

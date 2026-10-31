# Implementation notes และ contract gaps

รายการนี้บันทึกสิ่งที่เห็นจาก source ณ 2026-10-31 ไม่ใช่การแก้ implementation
หรือยืนยันผล live environment กฎที่ควรเป็นต้องแยกจากพฤติกรรมที่โค้ดทำอยู่

## Hard delete กับ payment history

`renewal.Repository.hardDelete` ล็อก request และลบ children รวม payment attempts/state
และ ledger โดยไม่มี query guard ปฏิเสธ payment history และไม่จำกัดสถานะคำขอ
ขัดกับคำอธิบายใน MIGRATION.md/docs/payments.md ที่ระบุว่าบล็อก hard delete เมื่อมีประวัติ
spec ของ endpoint ระบุพฤติกรรม source ปัจจุบันแล้ว ต้อง resolve ก่อนถือเป็น approved contract

## Renewal preview alias

`GET /v1/document-renewals/items/preview` อ่าน `request_no` จาก query แต่
`documentRequestItemCode` อ่านจาก path parameter ซึ่ง alias นี้ไม่มี จึงส่ง item code
ว่างไป query และโดยทั่วไปไม่พบข้อมูล ใช้เส้นที่มี requestNo/itemCode ใน path
หากต้องการ behavior ที่ source รองรับครบ

## Profile email update

`POST /v1/profile-update` บังคับส่ง email และคืน email นี้ใน response แต่ repository
ไม่ได้ UPDATE EMAIL/USERNAME จึงไม่ควรอธิบาย endpoint ว่าเปลี่ยนอีเมลได้

## Legacy create กับ validate-and-create

`POST /v1/document-renewals` เพิ่ม sequence และ request คนละ statement ไม่ได้อยู่ใน
transaction เดียว และไม่ได้สร้าง items/address snapshot/timeline แบบ validate-and-create
ทั้งสองเส้นจึงไม่มี atomicity หรือผลลัพธ์ข้างเคียงเหมือนกัน

## Transaction boundaries

- Delivery clear-default กับ insert/update ไม่ได้อยู่ใน transaction เดียวกัน
- Resubmit reset FIX items ก่อนเข้า transition transaction
- Register commit user/certificates และ action token ก่อนส่ง SMTP; SMTP fail ไม่ย้อนบัญชี
- Profile/certificate PUT object ก่อน SQL และไม่มี cleanup guarantee แบบ request-item

## Validation และ response ที่ควรทราบ

- Mobile snapshot update ตรวจเพียงยาว 10 ตัว/ขึ้นต้น 0 ไม่ตรวจทุกตัวเป็นตัวเลข
- Notification MarkValue ที่ valueId/notiType ว่างตอบ 500 เพราะคืน ordinary error
- Certificate validation ตรวจ parse dates แต่ไม่ได้บังคับ start<=end
- Voucher detail/image/QR กรอง active+ID แต่ไม่ตรวจ personal entitlement ซ้ำ
- News detail/image ไม่กรอง NEWS_STATUS ต่างจาก list
- `/privacy_policy` และ `/send-noti` ตอบ success placeholder ตาม source
- Service รับไฟล์สูงสุด 10 MiB แต่ Fiber default BodyLimit/ingress อาจปฏิเสธก่อนถึง service

## Schema compatibility

Delivery address validation ใช้ `districts.province_code` และ `subdistricts.district_code`
แต่ master dropdown queries ใช้ `province_id`/`district_id` ไม่ควรอนุมานว่าทั้งสองชุด
columns มีอยู่จากชื่อ relation เพียงอย่างเดียว ต้องตรวจ target DDL ก่อนนำขึ้น environment
SQL catalog คงข้อความจาก source เพื่อให้ตรวจความเข้ากันได้กับ migration ได้

## การดูแลเอกสาร

OpenAPI ที่มีอยู่ระบุ non-payment เป็น route inventory ส่วน Markdown ชุดนี้ลง business
rules และ query access ตาม Go source ทั้งสองส่วนยังต้องดูแลเมื่อ implementation เปลี่ยน
ตัวอย่าง JSON เป็นตัวอย่างโครงสร้าง ไม่ใช่ regression fixtures หรือ live responses

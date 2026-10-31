# Database migrations

## โครงสร้างและที่มา

- `legacy-mvp1/`: สำเนา SQL 25 ไฟล์และ README จาก
  `../smart-seaman-mobile-api/documents/mvp1/script` ณ วันที่ 2026-10-31
  โดยคงชื่อและเนื้อหาต้นฉบับ ดูรายละเอียดแต่ละไฟล์ใน
  [คู่มือต้นทาง](legacy-mvp1/README.md)
- `001_create_event_logs.sql`, `002_secure_auth.sql`,
  `003_payment_attempts.sql`: migrations เพิ่มเติมของ Go API
- `legacy-mvp1/SHA256SUMS`: checksum ของไฟล์ต้นทางที่นำเข้า ตรวจสำเนาด้วย
  `cd migrations/legacy-mvp1 && shasum -a 256 -c SHA256SUMS`

เลือกไฟล์ตามสถานะฐานข้อมูลและบันทึกไฟล์ที่ apply แล้ว ห้ามรัน SQL ทั้งโฟลเดอร์
ด้วย wildcard หรือ recursive runner เพราะมีทั้ง schema ทางเลือกและสคริปต์ลบข้อมูล
API ไม่ได้ apply migrations เหล่านี้ให้อัตโนมัติ

## สร้าง MVP1 schema ใหม่

ชุดนี้ไม่ใช่ bootstrap ฐานข้อมูลว่างทั้งหมด ต้องมี legacy core schema และ master
data ก่อน เช่น `m_mobile_users`, `t_session`, `m_documents`, `m_certificates`
และ `t_transaction_logs` ซึ่งไม่ได้ถูกสร้างในชุดนี้ ตรวจ foreign keys และ
document codes ให้ตรงกับฐานข้อมูลเป้าหมายก่อนเริ่ม

1. รัน `legacy-mvp1/01_create_mvp1_tables.sql`
2. รัน `legacy-mvp1/02_seed_mvp1_master_data.sql`
3. รัน `legacy-mvp1/03_create_core_indexes.sql` หลังตรวจ indexes ที่มีอยู่
4. รัน `legacy-mvp1/04_import_thailand_address_master.sql` เฉพาะเมื่อยังไม่มี
   ตาราง `provinces`, `districts` และ `subdistricts`
5. ตรวจและแก้ราคาในสำเนาสำหรับ deployment ก่อนรัน
   `legacy-mvp1/05_seed_document_price_settings.sql` เมื่อได้รับการยืนยันราคาแล้ว
6. รัน `legacy-mvp1/06_create_security_performance_indexes.sql`
   เฉพาะ indexes ที่ยังไม่มี
7. รัน Go migrations ตามลำดับด้านล่าง

ไม่ต้องรัน `90_*` หลังสร้างด้วย `01_*` เพราะ schema ใหม่รวมการเปลี่ยนแปลงเหล่านั้นแล้ว

**ราคา:** แม้ comment ต้นทางระบุว่าเป็น placeholders แต่ SQL `05_*` ใส่ค่าธรรมเนียม
0 บาทพร้อม `is_active='YES'` จริง และเมื่อพบ key เดิมจะอัปเดตเพียง `updated_at`
การรันซ้ำจึงไม่ปรับราคาเดิม ต้องตรวจทั้งราคาใหม่และราคาในฐานข้อมูลก่อนเปิดใช้งาน

## อัปเกรด MVP1 schema เดิม

1. ตรวจ schema และประวัติการ deploy แล้วเลือกเฉพาะ migrations ที่ยังไม่เคย apply
   ใน `legacy-mvp1/90_01_*` ถึง `90_17_*` เรียงตามหมายเลข
   ดูวัตถุประสงค์แต่ละไฟล์ในคู่มือต้นทาง
2. รัน `legacy-mvp1/02_seed_mvp1_master_data.sql` หลัง schema พร้อมแล้ว
3. ตรวจราคาก่อนพิจารณา `05_*`; ตรวจ existing indexes และ execution plans
   ก่อนพิจารณา `03_*` และ `06_*` อย่าเพิ่ม index ซ้ำโดยอัตโนมัติ
4. รัน Go migrations ที่ยังไม่เคย apply ตามลำดับด้านล่าง

`90_*` ส่วนใหญ่รันได้ครั้งเดียว การรันซ้ำอาจเกิด duplicate column/index/constraint
หากเคย apply บางส่วนให้ตรวจด้วย `SHOW CREATE TABLE` และ `SHOW INDEX` แล้วจัดทำ
SQL เฉพาะส่วนที่ขาด อย่าใช้ `--force` ข้ามข้อผิดพลาด
`90_06_*` จะตั้งไฟล์บัตรประชาชนเดิม (`MRI001`) เป็นยังไม่อัปโหลดเพราะแยกประเภทไม่ได้

## Go API migrations

| ลำดับ | ไฟล์ | ผลที่ต้องเตรียมก่อน deploy |
|---|---|---|
| 1 | [001_create_event_logs.sql](001_create_event_logs.sql) | สร้างตารางสำหรับ synchronous request event logs |
| 2 | [002_secure_auth.sql](002_secure_auth.sql) | หยุด API เดิมก่อนรัน; ยกเลิก sessions เดิมทั้งหมด ผู้ใช้ต้อง login ใหม่ |
| 3 | [003_payment_attempts.sql](003_payment_attempts.sql) | หยุด payment writers เดิม; ตรวจ duplicate keys/provider charges ก่อน import ประวัติ CHARGE |

ดูข้อกำหนด client/link และ rollout ใน [MIGRATION.md](../MIGRATION.md)
และ [payment rollout](../docs/payments.md#rollout-and-release-checks)
อย่ารัน `002_*` ซ้ำเป็นกิจวัตรเพราะจะ logout ผู้ใช้ และ `003_*` เป็น one-time migration

ตัวอย่างรันทีละไฟล์จาก repository root โดยให้ MySQL ถามรหัสผ่าน:

```bash
mysql --host=localhost --user=app_migrator --password smart_seaman < migrations/001_create_event_logs.sql
```

เปลี่ยน host, user และชื่อฐานข้อมูลให้ตรง environment สำรองข้อมูลและทดสอบใน
disposable/staging schema ก่อน deploy จริง MySQL DDL อาจ implicit commit;
อย่าคาดว่า rollback transaction จะย้อนทุกคำสั่งได้

## สคริปต์ลบข้อมูล

`legacy-mvp1/00_drop_mvp1_tables.sql` และ
`legacy-mvp1/99_delete_document_renewal_request_data.sql` เป็น utilities สำหรับ
reset local/dev/test เท่านั้น ไม่อยู่ในลำดับ migration ปกติ การใช้งานบน shared,
staging หรือ production ต้องได้รับอนุมัติชัดเจนและมี backup ตามคู่มือต้นทาง

สคริปต์ reset เดิมไม่จัดการ `event_logs`, `mobile_auth_tokens`,
`mobile_auth_sessions`, `mobile_payment_attempts` หรือ `mobile_payment_request_state`
จึงไม่ใช่ full reset ของ Go API และอาจทิ้งสถานะ auth/payment ที่ไม่สอดคล้องกัน

## การตรวจสอบสำเนา

ไฟล์ที่นำเข้าตรวจเทียบเนื้อหาและ SHA-256 กับต้นทางแล้ว ยังไม่ได้ execute SQL
กับ MySQL จริง จึงยังไม่ยืนยัน compatibility กับ schema/data ของ environment เป้าหมาย

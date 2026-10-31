# MVP1 Database Scripts

เอกสารนี้เป็นลำดับ execute หลักของ SQL scripts ใน folder นี้ แยกเป็น 2 กรณี: fresh install สำหรับฐานข้อมูลใหม่ และ existing DB migration สำหรับฐานข้อมูลที่มี schema เดิมอยู่แล้ว

## Fresh Install

ใช้ลำดับนี้เมื่อสร้าง MVP1 schema ใหม่ตั้งแต่ต้น

| Order | Script | Purpose |
|---:|---|---|
| 0 | `00_drop_mvp1_tables.sql` | Optional reset เฉพาะ disposable/local DB; ลบ MVP1 tables ตาม reverse dependency order |
| 1 | `01_create_mvp1_tables.sql` | สร้าง MVP1 tables, constraints, foreign keys และ indexes ที่จำเป็นต่อ schema |
| 2 | `02_seed_mvp1_master_data.sql` | seed status, document master request items และ required document setting แบบ rerunnable |
| 3 | `03_create_core_indexes.sql` | เพิ่ม core indexes ของ validation/session/mobile user paths |
| 4 | `04_import_thailand_address_master.sql` | import จังหวัด อำเภอ ตำบล และ postal code จาก external source |
| 5 | `05_seed_document_price_settings.sql` | seed/config ราคา renewal; เปิดใช้ production price หลัง Product/Admin ยืนยันราคา |
| 6 | `06_create_security_performance_indexes.sql` | เพิ่ม verified hot-path indexes จาก security/performance review |

หมายเหตุ:

- ห้ามรัน `00_drop_mvp1_tables.sql` บน shared/staging/production DB เว้นแต่ได้รับอนุมัติชัดเจน
- ไม่ต้องรัน scripts กลุ่ม `90_*` หลัง fresh install เพราะ changes เหล่านั้นถูก fold เข้า `01_create_mvp1_tables.sql` แล้ว
- `02_seed_mvp1_master_data.sql` ตั้ง `MRI004` เป็น `storage_scope = 'REQUEST'`

## Existing DB Migration

ใช้ลำดับนี้เมื่อ database มี MVP1 tables เดิมอยู่แล้ว และต้อง upgrade ให้ตรงกับ application ปัจจุบัน

| Order | Script | Purpose |
|---:|---|---|
| 1 | `90_01_migrate_document_status_code.sql` | เพิ่ม stable `document_status_code` ให้ status master |
| 2 | `90_02_migrate_request_scoped_document_items.sql` | เพิ่ม `storage_scope` และ request item file table สำหรับ request-scoped uploads |
| 3 | `90_03_migrate_document_renewal_price_effective_period.sql` | เพิ่ม effective period ให้ price setting |
| 4 | `90_04_migrate_renewal_request_draft.sql` | เพิ่ม unpaid draft foundation, sequence, price/address references |
| 5 | `90_05_migrate_delivery_address_default_guard.sql` | บังคับ active default delivery address ได้ไม่เกินหนึ่งรายการต่อ user |
| 6 | `90_06_migrate_identity_document_multi_file.sql` | ปรับ profile document upload ให้รองรับ multi-file slots |
| 7 | `90_07_migrate_request_no_on_request_items.sql` | เพิ่ม `request_no` ใน request items ให้ตรงกับ application insert/query |
| 8 | `90_08_migrate_mobile_number_history.sql` | เพิ่ม append-only mobile number history |
| 9 | `90_09_migrate_omise_payment_channels.sql` | ปรับ payment channel/action constraints สำหรับ Omise PromptPay/Mobile Banking |
| 10 | `90_10_migrate_document_request_soft_delete.sql` | เพิ่ม soft delete flag/index ให้ renewal request |
| 11 | `90_11_migrate_document_request_collation.sql` | ปรับ collation ของ renewal request logical keys ให้ join กับ profile/master tables ได้โดยไม่เกิด collation error |
| 12 | `90_12_migrate_document_request_user_contact_snapshot.sql` | เพิ่ม snapshot `mobile_number` และ `email` ของผู้ยื่นใน renewal request |
| 13 | `90_13_migrate_document_request_delivery_address_snapshot.sql` | เพิ่ม snapshot delivery address ต่อ renewal request และ backfill จาก `delivery_address_id` เท่าที่มีข้อมูล |
| 14 | `90_14_migrate_validate_create_performance_indexes.sql` | เพิ่ม indexes สำหรับ hot path ของ validate-and-create |
| 15 | `90_15_migrate_document_request_idempotency_key.sql` | เพิ่ม `idempotency_key` สำหรับกันการสร้าง renewal request draft ซ้ำ |
| 16 | `90_16_migrate_validate_create_covering_indexes.sql` | เพิ่ม covering indexes สำหรับ query validate-and-create ที่ aggregate file state |
| 17 | `90_17_migrate_document_mobile_status.sql` | เพิ่ม mobile-facing status code/name สำหรับ renewal progress display |

หลัง migration:

1. รัน `02_seed_mvp1_master_data.sql` เพื่อ upsert master data ล่าสุด
2. รัน `05_seed_document_price_settings.sql` เมื่อมีราคาที่ Product/Admin ยืนยันแล้ว
3. รัน `06_create_security_performance_indexes.sql` หาก indexes ยังไม่มีใน target DB

หมายเหตุ:

- Scripts กลุ่ม `90_*` เป็น one-time migration และส่วนใหญ่ไม่ idempotent; ตรวจ target schema ก่อนรันซ้ำ
- ถ้า target DB เคย apply บาง migration แล้ว ให้ข้าม script นั้นหรือทำ migration เฉพาะส่วนที่ยังขาด
- ควร backup และทดสอบใน staging/disposable schema ก่อนรันกับ production

## Utility Scripts

| Script | Purpose |
|---|---|
| `99_delete_document_renewal_request_data.sql` | ลบข้อมูล renewal request runtime data สำหรับ local/dev/test reset โดยลบ child tables ตาม foreign key order |

หมายเหตุ:

- Scripts กลุ่ม `99_*` เป็น destructive utility scripts ไม่ใช่ migration ปกติ
- ห้ามรันบน shared/staging/production DB เว้นแต่ได้รับอนุมัติชัดเจนและมี backup

# SQL catalog: catalog

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## catalog.Banner

Source: [internal/domain/catalog/repository.go:33](../../../internal/domain/catalog/repository.go#L33)

```sql
SELECT BANNER_ID,BANNER_FILE_NAME FROM m_banners WHERE BANNER_ID=?
```

## catalog.Banners

Source: [internal/domain/catalog/repository.go:15](../../../internal/domain/catalog/repository.go#L15)

```sql
SELECT BANNER_ID, BANNER_FILE_NAME FROM m_banners ORDER BY CREATE_DATE DESC
```

## catalog.Schools

Source: [internal/domain/catalog/repository.go:76](../../../internal/domain/catalog/repository.go#L76)

```sql
SELECT COMPANY_CODE, COMPANY_NAME_EN, COMPANY_NAME_TH, COMPANY_COLOUR, COURSE_COLOUR, COURSE_ID, COURSE_CODE, COURSE_NAME_EN, COURSE_NAME_TH, COURSE_TYPE, COURSE_ONLINE_DATE, COURSE_ONSITE_DATE, COURSE_TOTAL_DAYS, COURSE_PRICE, COMPANY_LINE, COMPANY_FACEBOOK, COMPANY_PHONE1, COURSE_START_DATE, COURSE_END_DATE, DATE_FOR_CHECK FROM list_course_school_training
```

## catalog.UserSmartID

Source: [internal/domain/catalog/repository.go:43](../../../internal/domain/catalog/repository.go#L43)

```sql
SELECT SMART_SEAMAN_ID FROM m_mobile_users WHERE MOBILE_UUID=? LIMIT 1
```

## catalog.Voucher

Source: [internal/domain/catalog/repository.go:64](../../../internal/domain/catalog/repository.go#L64)

```sql
SELECT VOUCHER_ID,VOUCHER_TITLE,VOUCHER_DETAILS,VOUCHER_START_DATE,VOUCHER_END_DATE,VOUCHER_PICTURE,VOUCHER_QRCODE FROM m_vouchers WHERE VOUCHER_STATUS='A' AND VOUCHER_ID=?
```

## catalog.Vouchers

Source: [internal/domain/catalog/repository.go:48](../../../internal/domain/catalog/repository.go#L48)

```sql
SELECT mv.VOUCHER_ID,mv.VOUCHER_TITLE,mv.VOUCHER_DETAILS,mv.VOUCHER_START_DATE,mv.VOUCHER_END_DATE FROM m_vouchers mv INNER JOIN m_voucher_details mvd ON mv.VOUCHER_ID=mvd.VOUCHER_ID WHERE mv.VOUCHER_STATUS='A' AND mv.VOUCHER_TYPE='PERSONAL' AND mvd.SMART_SEAMAN_ID=? UNION ALL SELECT VOUCHER_ID,VOUCHER_TITLE,VOUCHER_DETAILS,VOUCHER_START_DATE,VOUCHER_END_DATE FROM m_vouchers WHERE VOUCHER_STATUS='A' AND VOUCHER_TYPE='GLOBAL' ORDER BY VOUCHER_ID
```


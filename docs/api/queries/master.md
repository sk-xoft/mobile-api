# SQL catalog: master

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## master.Companies

Source: [internal/domain/master/repository.go:13](../../../internal/domain/master/repository.go#L13)

```sql
SELECT COMPANY_CODE, COMPANY_NAME_EN, COMPANY_NAME_TH
		FROM m_companys WHERE COMPANY_TYPE = 'Shipping Company' AND COMPANY_MOBILE_FLAG = 'Y'
		AND COMPANY_STATUS = 'A' ORDER BY COMPANY_TYPE, COMPANY_SEQ
```

## master.Courses

Source: [internal/domain/master/repository.go:67](../../../internal/domain/master/repository.go#L67)

```sql
SELECT COURSE_CODE, COURSE_NAME_EN, COURSE_NAME_TH
		FROM m_course_name WHERE COURSE_STATUS = 'A' ORDER BY COURSE_SEQ
```

## master.Districts

Source: [internal/domain/master/repository.go:89](../../../internal/domain/master/repository.go#L89)

```sql
SELECT d.code, d.name_in_thai, d.name_in_english, NULL
		FROM districts d INNER JOIN provinces p ON p.id = d.province_id
		WHERE p.code = ? ORDER BY d.name_in_thai
```

## master.Documents

Source: [internal/domain/master/repository.go:49](../../../internal/domain/master/repository.go#L49)

```sql
SELECT DOCUMENT_CODE, DOCUMENT_NAME_EN, DOCUMENT_NAME_TH
		FROM m_documents WHERE DOCUMENT_TYPE = ? ORDER BY DOCUMENT_SEQ
```

## master.Positions

Source: [internal/domain/master/repository.go:32](../../../internal/domain/master/repository.go#L32)

```sql
SELECT POSITION_CODE, POSITION_NAME_EN, POSITION_NAME_TH FROM m_positions
```

## master.Provinces

Source: [internal/domain/master/repository.go:85](../../../internal/domain/master/repository.go#L85)

```sql
SELECT code, name_in_thai, name_in_english, NULL FROM provinces ORDER BY name_in_thai
```

## master.Subdistricts

Source: [internal/domain/master/repository.go:95](../../../internal/domain/master/repository.go#L95)

```sql
SELECT s.code, s.name_in_thai, s.name_in_english, s.zip_code
		FROM subdistricts s INNER JOIN districts d ON d.id = s.district_id
		WHERE d.code = ? ORDER BY s.name_in_thai
```

## master.addresses

Source: [internal/domain/master/repository.go:101](../../../internal/domain/master/repository.go#L101)


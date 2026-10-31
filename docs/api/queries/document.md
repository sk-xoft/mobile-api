# SQL catalog: document

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## document.Certificate

Source: [internal/domain/document/repository.go:23](../../../internal/domain/document/repository.go#L23)

```sql
SELECT CERT_ID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE,
		CERT_FILE, ORIGINAL_FILE_NAME, CERT_STATUS FROM m_certificates
		WHERE CERT_MOBILE_UUID = ? AND CERT_DOCUMENT_CODE = ? LIMIT 1
```

## document.CloseToExpiration

Source: [internal/domain/document/repository.go:100](../../../internal/domain/document/repository.go#L100)

```sql
SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,
		md.DOCUMENT_TYPE, md.DOCUMENT_DEFAULT_FLAG, md.DOCUMENT_RENEWAL_FLAG, md.DOCUMENT_SEQ, md.DOCUMENT_MOBILE_FLAG, md.DOCUMENT_COURSE_CODE,
		mc.CERT_START_DATE, mc.CERT_END_DATE, mc.ORIGINAL_FILE_NAME,
		CASE WHEN EXISTS (SELECT 1 FROM m_document_request dr INNER JOIN m_document_status ds ON ds.id=dr.document_status_id
			WHERE dr.mobile_user_uuid=? AND dr.document_code=md.DOCUMENT_CODE AND dr.is_active='YES'
			AND ds.document_status_code NOT IN ('DELIVERED','CANCELLED')) THEN 'Y' ELSE 'N' END
		FROM m_documents md INNER JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND mc.CERT_END_DATE <= NOW() + INTERVAL 18 MONTH
		ORDER BY mc.CERT_END_DATE, md.DOCUMENT_SEQ LIMIT 10 OFFSET ?
```

```sql
SELECT COUNT(*) FROM m_documents md INNER JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=? WHERE md.DOCUMENT_STATUS='A' AND mc.CERT_END_DATE <= NOW() + INTERVAL 18 MONTH
```

## document.CreateCertificate

Source: [internal/domain/document/repository.go:35](../../../internal/domain/document/repository.go#L35)

```sql
INSERT INTO m_certificates
		(CERT_MOBILE_UUID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE, CERT_FILE, ORIGINAL_FILE_NAME,
		 CERT_STATUS, CREATE_DATE, CREATE_BY, UPDATE_DATE, UPDATE_BY)
		VALUES (?, ?, ?, ?, ?, ?, 'A', NOW(), ?, NOW(), ?)
```

## document.DeleteCertificate

Source: [internal/domain/document/repository.go:55](../../../internal/domain/document/repository.go#L55)

```sql
DELETE FROM m_certificates WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?
```

## document.FindUser

Source: [internal/domain/document/repository.go:17](../../../internal/domain/document/repository.go#L17)

```sql
SELECT MOBILE_UUID, USERNAME, EMAIL FROM m_mobile_users WHERE MOBILE_UUID = ? LIMIT 1
```

## document.Page

Source: [internal/domain/document/repository.go:70](../../../internal/domain/document/repository.go#L70)

```sql
SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,
		md.DOCUMENT_TYPE, md.DOCUMENT_DEFAULT_FLAG, md.DOCUMENT_RENEWAL_FLAG, md.DOCUMENT_SEQ, md.DOCUMENT_MOBILE_FLAG, md.DOCUMENT_COURSE_CODE,
		mc.CERT_START_DATE, mc.CERT_END_DATE, mc.ORIGINAL_FILE_NAME
		FROM m_documents md LEFT JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND md.DOCUMENT_TYPE=? AND mc.CERT_MOBILE_UUID IS NOT NULL ORDER BY md.DOCUMENT_SEQ LIMIT 10 OFFSET ?
```

```sql
SELECT COUNT(*) FROM m_documents md
		LEFT JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND md.DOCUMENT_TYPE=? AND mc.CERT_MOBILE_UUID IS NOT NULL
```

## document.UpdateCertificate

Source: [internal/domain/document/repository.go:44](../../../internal/domain/document/repository.go#L44)

```sql
UPDATE m_certificates SET CERT_START_DATE=?, CERT_END_DATE=?, CERT_STATUS='A', UPDATE_DATE=NOW(), UPDATE_BY=?
			WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?
```

```sql
UPDATE m_certificates SET CERT_START_DATE=?, CERT_END_DATE=?, CERT_FILE=?, ORIGINAL_FILE_NAME=?, CERT_STATUS='A', UPDATE_DATE=NOW(), UPDATE_BY=?
		WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?
```


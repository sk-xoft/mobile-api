# SQL catalog: delivery

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## delivery.ClearDefault

Source: [internal/domain/delivery/repository.go:65](../../../internal/domain/delivery/repository.go#L65)

```sql
UPDATE m_delivery_address SET is_default=0, updated_at=NOW() WHERE mobile_user_uuid=? AND is_active='YES' AND is_default=1
```

## delivery.Count

Source: [internal/domain/delivery/repository.go:59](../../../internal/domain/delivery/repository.go#L59)

```sql
SELECT COUNT(*) FROM m_delivery_address WHERE mobile_user_uuid=? AND is_active='YES'
```

## delivery.Create

Source: [internal/domain/delivery/repository.go:70](../../../internal/domain/delivery/repository.go#L70)

```sql
INSERT INTO m_delivery_address
		(id, mobile_user_uuid, first_name, last_name, address_line, province, district, sub_district, postal_code, is_default, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'YES')
```

## delivery.Default

Source: [internal/domain/delivery/repository.go:31](../../../internal/domain/delivery/repository.go#L31)

```sql
SELECT da.id, da.mobile_user_uuid, da.first_name, da.last_name, da.address_line,
		da.province, da.district, da.sub_district, da.postal_code, mu.MOBILE_NUMBER,
		CONCAT_WS(' ', da.address_line, CONCAT('ตำบล', sd.name_in_thai), CONCAT('อำเภอ', d.name_in_thai), CONCAT('จังหวัด', p.name_in_thai), da.postal_code), da.is_default
		FROM m_delivery_address da INNER JOIN m_mobile_users mu ON mu.MOBILE_UUID=da.mobile_user_uuid
		LEFT JOIN provinces p ON p.code=da.province LEFT JOIN districts d ON d.code=da.district LEFT JOIN subdistricts sd ON sd.code=da.sub_district
		WHERE da.mobile_user_uuid=? AND da.is_default=1 AND da.is_active='YES' ORDER BY da.updated_at DESC, da.id LIMIT 1
```

## delivery.Find

Source: [internal/domain/delivery/repository.go:47](../../../internal/domain/delivery/repository.go#L47)

```sql
SELECT id, mobile_user_uuid, first_name, last_name, address_line, province, district,
		sub_district, postal_code, is_default FROM m_delivery_address WHERE id=? AND mobile_user_uuid=? AND is_active='YES'
```

## delivery.InsertSnapshot

Source: [internal/domain/delivery/repository.go:111](../../../internal/domain/delivery/repository.go#L111)

```sql
INSERT INTO m_document_request_delivery_address
		(id, request_id, source_delivery_address_id, mobile_user_uuid, first_name, last_name, address_line,
		province, district, sub_district, postal_code, mobile_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
```

## delivery.RenewalRequest

Source: [internal/domain/delivery/repository.go:96](../../../internal/domain/delivery/repository.go#L96)

```sql
SELECT id FROM m_document_request WHERE request_no=? AND mobile_user_uuid=? AND is_active='YES' LIMIT 1
```

## delivery.SnapshotExists

Source: [internal/domain/delivery/repository.go:105](../../../internal/domain/delivery/repository.go#L105)

```sql
SELECT COUNT(*) FROM m_document_request_delivery_address WHERE request_id=? AND mobile_user_uuid=?
```

## delivery.Update

Source: [internal/domain/delivery/repository.go:78](../../../internal/domain/delivery/repository.go#L78)

```sql
UPDATE m_delivery_address SET first_name=?, last_name=?, address_line=?, province=?, district=?,
		sub_district=?, postal_code=?, is_default=?, updated_at=NOW() WHERE id=? AND mobile_user_uuid=? AND is_active='YES'
```

## delivery.UserMobile

Source: [internal/domain/delivery/repository.go:16](../../../internal/domain/delivery/repository.go#L16)

```sql
SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=? LIMIT 1
```

## delivery.ValidAddress

Source: [internal/domain/delivery/repository.go:22](../../../internal/domain/delivery/repository.go#L22)

```sql
SELECT COUNT(*) FROM provinces p
		INNER JOIN districts d ON d.province_code=p.code
		INNER JOIN subdistricts s ON s.district_code=d.code
		WHERE p.code=? AND d.code=? AND s.code=? AND s.zip_code=?
```


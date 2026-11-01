# SQL catalog: delivery

Static SQL literals/fragments จาก source ณ 2026-11-01 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## delivery.ClearDefault

Source: [internal/domain/delivery/repository.go:94](../../../internal/domain/delivery/repository.go#L94)

```sql
UPDATE m_delivery_address SET is_default=0, updated_at=NOW() WHERE mobile_user_uuid=? AND is_active='YES' AND is_default=1
```

## delivery.Count

Source: [internal/domain/delivery/repository.go:88](../../../internal/domain/delivery/repository.go#L88)

```sql
SELECT COUNT(*) FROM m_delivery_address WHERE mobile_user_uuid=? AND is_active='YES'
```

## delivery.Create

Source: [internal/domain/delivery/repository.go:99](../../../internal/domain/delivery/repository.go#L99)

```sql
INSERT INTO m_delivery_address
		(id, mobile_user_uuid, first_name, last_name, address_line, province, district, sub_district, postal_code, is_default, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'YES')
```

## delivery.Default

Source: [internal/domain/delivery/repository.go:44](../../../internal/domain/delivery/repository.go#L44)

```sql
SELECT da.id, da.mobile_user_uuid, da.first_name, da.last_name, da.address_line,
		da.province, da.district, da.sub_district, da.postal_code, mu.MOBILE_NUMBER,
		CONCAT_WS(' ', da.address_line, CONCAT('ตำบล', sd.name_in_thai), CONCAT('อำเภอ', d.name_in_thai), CONCAT('จังหวัด', p.name_in_thai), da.postal_code), da.is_default
		FROM m_delivery_address da INNER JOIN m_mobile_users mu ON mu.MOBILE_UUID=da.mobile_user_uuid
		LEFT JOIN provinces p ON p.code=da.province LEFT JOIN districts d ON d.code=da.district LEFT JOIN subdistricts sd ON sd.code=da.sub_district
		WHERE da.mobile_user_uuid=? AND da.is_default=1 AND da.is_active='YES' ORDER BY da.updated_at DESC, da.id
```

## delivery.Find

Source: [internal/domain/delivery/repository.go:76](../../../internal/domain/delivery/repository.go#L76)

```sql
SELECT id, mobile_user_uuid, first_name, last_name, address_line, province, district,
		sub_district, postal_code, is_default FROM m_delivery_address WHERE id=? AND mobile_user_uuid=? AND is_active='YES'
```

## delivery.InsertSnapshot

Source: [internal/domain/delivery/repository.go:139](../../../internal/domain/delivery/repository.go#L139)

```sql
INSERT INTO m_document_request_delivery_address
		(id, request_id, source_delivery_address_id, mobile_user_uuid, first_name, last_name, address_line,
		province, district, sub_district, postal_code, mobile_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
```

## delivery.RenewalRequest

Source: [internal/domain/delivery/repository.go:119](../../../internal/domain/delivery/repository.go#L119)

```sql
SELECT id FROM m_document_request WHERE request_no=? AND mobile_user_uuid=? AND is_active='YES' FOR UPDATE
-- ใช้ id=? แทน request_no=? เมื่อ input เป็น UUID; column มาจาก allowlist ในโค้ด
```

## delivery.SnapshotExists

Source: [internal/domain/delivery/repository.go:133](../../../internal/domain/delivery/repository.go#L133)

```sql
SELECT COUNT(*) FROM m_document_request_delivery_address WHERE request_id=? AND mobile_user_uuid=?
```

## delivery.Update

Source: [internal/domain/delivery/repository.go:107](../../../internal/domain/delivery/repository.go#L107)

```sql
UPDATE m_delivery_address SET first_name=?, last_name=?, address_line=?, province=?, district=?,
		sub_district=?, postal_code=?, is_default=?, updated_at=NOW() WHERE id=? AND mobile_user_uuid=? AND is_active='YES'
```

## delivery.UserMobile

ชื่ออ้างอิงเดิม; implementation ปัจจุบันคือ `LockUser` และใช้ใน transaction เท่านั้น

Source: [internal/domain/delivery/repository.go:26](../../../internal/domain/delivery/repository.go#L26)

```sql
SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE
```

## delivery.ValidAddress

Source: [internal/domain/delivery/repository.go:35](../../../internal/domain/delivery/repository.go#L35)

```sql
SELECT COUNT(*) FROM provinces p
		INNER JOIN districts d ON d.province_id=p.id
		INNER JOIN subdistricts s ON s.district_id=d.id
		WHERE p.code=? AND d.code=? AND s.code=? AND CAST(s.zip_code AS CHAR)=?
```


## delivery.LockUser

Source: [delivery repository](../../../internal/domain/delivery/repository.go)

```sql
SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE
```

## delivery.LockAddresses

Source: [delivery repository](../../../internal/domain/delivery/repository.go)

```sql
SELECT id FROM m_delivery_address WHERE mobile_user_uuid=? AND is_active='YES' FOR UPDATE
```

## delivery.FindSnapshot

Source: [delivery repository](../../../internal/domain/delivery/repository.go)

```sql
SELECT id,mobile_user_uuid,first_name,last_name,address_line,province,district,sub_district,postal_code,mobile_number
FROM m_document_request_delivery_address WHERE id=? AND mobile_user_uuid=? FOR UPDATE
```

## delivery.UpdateSnapshot

Source: [delivery repository](../../../internal/domain/delivery/repository.go)

```sql
UPDATE m_document_request_delivery_address SET first_name=?,last_name=?,address_line=?,province=?,district=?,sub_district=?,postal_code=? WHERE id=? AND mobile_user_uuid=?
```

ทุก mutation ใช้ repository ที่ผูกกับ transaction เดียวกัน รวมการล็อก ตรวจซ้ำ และเขียนข้อมูล
`source_delivery_address_id` bind เป็น SQL NULL เมื่อไม่มี master address ต้นทาง

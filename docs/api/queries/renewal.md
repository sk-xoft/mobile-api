# SQL catalog: renewal

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## renewal.Documents

Source: [internal/domain/renewal/repository.go:55](../../../internal/domain/renewal/repository.go#L55)

```sql
SELECT DOCUMENT_CODE, DOCUMENT_NAME_TH, DOCUMENT_NAME_EN FROM m_documents WHERE DOCUMENT_STATUS='A' AND DOCUMENT_RENEWAL_FLAG IN ('Y','YES') ORDER BY DOCUMENT_SEQ
```

## renewal.Price

Source: [internal/domain/renewal/repository.go:72](../../../internal/domain/renewal/repository.go#L72)

```sql
SELECT p.document_code, p.government_fee, p.document_processing_fee, p.shipping_fee,
		p.shipping_discount, p.service_fee_discount, p.effective_from, p.effective_to
		FROM m_document_prices_setting p INNER JOIN m_documents d ON d.DOCUMENT_CODE=p.document_code
		WHERE p.document_code=? AND p.is_active='YES' AND p.effective_from<=CURRENT_DATE
		AND (p.effective_to IS NULL OR p.effective_to>=CURRENT_DATE) AND d.DOCUMENT_STATUS='A'
		ORDER BY p.effective_from DESC LIMIT 1
```

## renewal.Statuses

Source: [internal/domain/renewal/repository.go:24](../../../internal/domain/renewal/repository.go#L24)

```sql
SELECT id, document_status_code, document_mobile_status_code, document_mobile_status_name_th,
		document_mobile_status_name_en, name_th, name_en, css_color FROM m_document_status
		WHERE is_active='YES' AND is_mobile_visible='YES'
		ORDER BY CASE document_status_code WHEN 'PENDING_DOCUMENT_REVIEW' THEN 1 WHEN 'PENDING_APPLICANT_CORRECTION' THEN 2
		WHEN 'PENDING_MARINE_DEPARTMENT_RESULT' THEN 3 WHEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP' THEN 4
		WHEN 'DELIVERING' THEN 5 WHEN 'DELIVERED' THEN 6 WHEN 'CANCELLED' THEN 7 ELSE 99 END, name_th
```

## renewal.countFixItems

Source: [internal/domain/renewal/lifecycle.go:593](../../../internal/domain/renewal/lifecycle.go#L593)

```sql
SELECT COUNT(*) FROM m_document_request_items WHERE request_id=? AND approve_status='FIX'
```

## renewal.countIncompleteFixItems

Source: [internal/domain/renewal/lifecycle.go:599](../../../internal/domain/renewal/lifecycle.go#L599)

```sql
SELECT COUNT(*)
		FROM m_document_request_items i
		WHERE i.request_id=? AND i.approve_status='FIX' AND NOT EXISTS (
			SELECT 1 FROM m_document_request_item_files f
			WHERE f.request_item_id=i.id AND f.file_uploaded=1 AND f.is_updated=1
			AND (f.check_result IS NULL OR f.check_result<>'fix')
		)
```

## renewal.create

Source: [internal/domain/renewal/lifecycle.go:202](../../../internal/domain/renewal/lifecycle.go#L202)

```sql
SELECT MOBILE_NUMBER,EMAIL FROM m_mobile_users WHERE MOBILE_UUID=?
```

```sql
INSERT INTO m_document_request_sequence(period,last_number) VALUES(?,1) ON DUPLICATE KEY UPDATE last_number=last_number+1
```

```sql
SELECT last_number FROM m_document_request_sequence WHERE period=?
```

```sql
INSERT INTO m_document_request(id,request_no,mobile_user_uuid,mobile_number,email,document_code,document_status_id,price_setting_id,delivery_address_id,is_active,amount) VALUES(?,?,?,?,?,?,?,NULL,?,'YES',?)
```

## renewal.detail

Source: [internal/domain/renewal/lifecycle.go:623](../../../internal/domain/renewal/lifecycle.go#L623)

## renewal.documentInfo

Source: [internal/domain/renewal/lifecycle.go:232](../../../internal/domain/renewal/lifecycle.go#L232)

```sql
SELECT COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_EN,'') FROM m_documents WHERE DOCUMENT_CODE=? AND DOCUMENT_STATUS='A'
```

## renewal.hardDelete

Source: [internal/domain/renewal/lifecycle.go:688](../../../internal/domain/renewal/lifecycle.go#L688)

```sql
SELECT id FROM m_document_request WHERE id=? AND mobile_user_uuid=? FOR UPDATE
```

```sql
SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?
```

```sql
DELETE FROM
```

```sql
DELETE FROM mobile_payment_request_state WHERE request_id=?
```

```sql
DELETE FROM mobile_payment_attempts WHERE request_id=?
```

```sql
DELETE FROM m_document_request_items WHERE request_id=?
```

```sql
DELETE FROM m_document_request_delivery_address WHERE request_id=?
```

```sql
UPDATE m_payment_transaction SET parent_transaction_id=NULL WHERE request_id=?
```

```sql
DELETE FROM m_document_transaction WHERE request_id=?
```

```sql
DELETE FROM m_payment_transaction WHERE request_id=?
```

```sql
DELETE FROM m_document_request WHERE id=? AND mobile_user_uuid=?
```

## renewal.latestActive

Source: [internal/domain/renewal/lifecycle.go:188](../../../internal/domain/renewal/lifecycle.go#L188)

```sql
SELECT r.id,r.request_no,r.mobile_user_uuid,r.mobile_number,r.email,r.document_code,r.document_status_id,s.document_status_code,s.name_en,COALESCE(r.idempotency_key,''),r.amount,r.is_resubmit,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.mobile_user_uuid=? AND r.document_code=? AND r.is_active='YES' AND s.document_status_code<>'DELIVERED' ORDER BY r.submitted_at DESC,r.id DESC LIMIT 1
```

## renewal.list

Source: [internal/domain/renewal/lifecycle.go:507](../../../internal/domain/renewal/lifecycle.go#L507)

```sql
SELECT r.id,r.request_no,r.document_code,s.document_status_code,s.name_en,r.amount,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.mobile_user_uuid=? AND r.is_active='YES' ORDER BY r.submitted_at DESC LIMIT 10 OFFSET ?
```

```sql
SELECT COUNT(*) FROM m_document_request WHERE mobile_user_uuid=? AND is_active='YES'
```

## renewal.ownedRequest

Source: [internal/domain/renewal/lifecycle.go:180](../../../internal/domain/renewal/lifecycle.go#L180)

```sql
SELECT r.id,r.request_no,r.mobile_user_uuid,r.mobile_number,r.email,r.document_code,r.document_status_id,s.document_status_code,s.name_en,COALESCE(r.idempotency_key,''),r.amount,r.is_resubmit,r.submitted_at FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.
```

## renewal.requiredItemCount

Source: [internal/domain/renewal/lifecycle.go:238](../../../internal/domain/renewal/lifecycle.go#L238)

```sql
SELECT COUNT(*) FROM m_document_setting_requires WHERE document_code=? AND is_required=1 AND is_active='YES'
```

## renewal.resetFixItems

Source: [internal/domain/renewal/lifecycle.go:611](../../../internal/domain/renewal/lifecycle.go#L611)

```sql
UPDATE m_document_request_items SET approve_status='PENDING',updated_at=NOW() WHERE request_id=? AND approve_status='FIX'
```

## renewal.stageCurrentStatus

Source: [internal/domain/renewal/lifecycle.go:165](../../../internal/domain/renewal/lifecycle.go#L165)

```sql
SELECT s.id,s.document_status_code,COALESCE(s.name_th,''),COALESCE(s.name_en,''),COALESCE(s.css_color,''),COALESCE(s.document_mobile_status_code,''),COALESCE(s.document_mobile_status_name_th,''),COALESCE(s.document_mobile_status_name_en,'') FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.id=?
```

## renewal.statusID

Source: [internal/domain/renewal/lifecycle.go:196](../../../internal/domain/renewal/lifecycle.go#L196)

```sql
SELECT id FROM m_document_status WHERE document_status_code=? AND is_active='YES' LIMIT 1
```

## renewal.timeline

Source: [internal/domain/renewal/lifecycle.go:630](../../../internal/domain/renewal/lifecycle.go#L630)

```sql
SELECT action,from_status,to_status,note,actioned_at FROM m_document_transaction WHERE request_id=? ORDER BY actioned_at
```

## renewal.transition

Source: [internal/domain/renewal/lifecycle.go:527](../../../internal/domain/renewal/lifecycle.go#L527)

```sql
SELECT document_status_id FROM m_document_request WHERE id=? AND mobile_user_uuid=? AND is_active='YES' FOR UPDATE
```

```sql
SELECT COUNT(*) FROM mobile_payment_attempts WHERE request_id=? AND (status NOT IN ('EXPIRED','FAILED') OR review_required=1)
```

```sql
UPDATE m_document_request SET document_status_id=?,is_resubmit=CASE WHEN ?='RESUBMIT' THEN 1 ELSE is_resubmit END,updated_at=NOW()
```

```sql
INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,?,?,?,?,?)
```

## renewal.updateMobile

Source: [internal/domain/renewal/lifecycle.go:615](../../../internal/domain/renewal/lifecycle.go#L615)

```sql
UPDATE m_document_request SET mobile_number=?,updated_at=NOW() WHERE id=?
```

## renewal.validateCreate

Source: [internal/domain/renewal/lifecycle.go:430](../../../internal/domain/renewal/lifecycle.go#L430)

```sql
SELECT MOBILE_NUMBER,EMAIL FROM m_mobile_users WHERE MOBILE_UUID=?
```

```sql
SELECT id,first_name,last_name,address_line,province,district,sub_district,postal_code FROM m_delivery_address WHERE mobile_user_uuid=? AND is_default=1 AND is_active='YES' ORDER BY updated_at DESC,id LIMIT 1
```

```sql
INSERT INTO m_document_request_sequence(period,last_number) VALUES(?,1) ON DUPLICATE KEY UPDATE last_number=last_number+1
```

```sql
SELECT last_number FROM m_document_request_sequence WHERE period=? FOR UPDATE
```

```sql
INSERT INTO m_document_request(id,request_no,mobile_user_uuid,mobile_number,email,document_code,document_status_id,price_setting_id,delivery_address_id,is_active,amount,idempotency_key) VALUES(?,?,?,?,?,?,?,(SELECT id FROM m_document_prices_setting WHERE document_code=? AND is_active='YES' AND effective_from<=CURRENT_DATE AND (effective_to IS NULL OR effective_to>=CURRENT_DATE) ORDER BY effective_from DESC LIMIT 1),?,'YES',?,?)
```

```sql
INSERT INTO m_document_request_delivery_address(id,request_id,source_delivery_address_id,mobile_user_uuid,first_name,last_name,address_line,province,district,sub_district,postal_code,mobile_number) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)
```

```sql
INSERT INTO m_document_request_items(id,request_id,request_no,document_master_request_item_code,approve_status) SELECT UUID(),?,?,document_master_request_item_code,'PENDING' FROM m_document_setting_requires WHERE document_code=? AND is_required=1 AND is_active='YES'
```

```sql
INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,?,?,?,?,?)
```

## renewal.validateExisting

Source: [internal/domain/renewal/lifecycle.go:244](../../../internal/domain/renewal/lifecycle.go#L244)

## renewal.validateFilesBatch

Source: [internal/domain/renewal/lifecycle.go:380](../../../internal/domain/renewal/lifecycle.go#L380)

```sql
SELECT document_master_request_item_code,id,document_type,slot_code,COALESCE(file_path,''),COALESCE(original_file_name,''),COALESCE(mime_type,''),COALESCE(file_size,0),COALESCE(DATE_FORMAT(file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(check_result,''),COALESCE(check_note,''),COALESCE(is_updated,0) FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code IN (
```

```sql
SELECT i.document_master_request_item_code,f.id,f.document_type,f.slot_code,COALESCE(f.file_path,''),COALESCE(f.original_file_name,''),COALESCE(f.mime_type,''),COALESCE(f.file_size,0),COALESCE(DATE_FORMAT(f.file_uploaded_at,'%d/%m/%Y %H:%i'),''),COALESCE(f.check_result,''),COALESCE(f.check_note,''),COALESCE(f.is_updated,0) FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id INNER JOIN m_document_request r ON r.id=i.request_id WHERE i.request_id=? AND r.mobile_user_uuid=? AND i.document_master_request_item_code IN (
```

## renewal.validateResponse

Source: [internal/domain/renewal/lifecycle.go:248](../../../internal/domain/renewal/lifecycle.go#L248)

## renewal.validateResponseWithDocument

Source: [internal/domain/renewal/lifecycle.go:256](../../../internal/domain/renewal/lifecycle.go#L256)

```sql
SELECT CERT_END_DATE FROM m_certificates WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=? ORDER BY CERT_END_DATE DESC LIMIT 1
```

```sql
SELECT a.id,a.first_name,a.last_name,a.address_line,a.province,a.district,a.sub_district,a.postal_code,a.mobile_number,CONCAT_WS(' ',a.address_line,CONCAT('ตำบล',sd.name_in_thai),CONCAT('อำเภอ',d.name_in_thai),CONCAT('จังหวัด',p.name_in_thai),a.postal_code) FROM m_document_request_delivery_address a LEFT JOIN provinces p ON p.code=a.province LEFT JOIN districts d ON d.code=a.district LEFT JOIN subdistricts sd ON sd.code=a.sub_district WHERE a.request_id=? AND a.mobile_user_uuid=? ORDER BY a.created_at DESC,a.id DESC
```

```sql
SELECT i.id,COALESCE(p.id,''),m.document_master_items_code,m.storage_scope,COALESCE(p.document_type,''),COALESCE(m.document_master_items_name,''),m.sort_order,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 1 ELSE 0 END,CASE WHEN COALESCE(p.file_uploaded,0)=1 THEN 'COMPLETE' WHEN p.id IS NULL AND m.storage_scope='PROFILE' THEN 'MISSING' ELSE 'INCOMPLETE' END,COALESCE(p.file_path,''),COALESCE(DATE_FORMAT(p.file_uploaded_at,'%Y-%m-%dT%H:%i:%sZ'),''),COALESCE(p.check_result,''),COALESCE(p.check_note,'') FROM m_document_request_items i INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code LEFT JOIN m_document_profile_request_item p ON p.id=(SELECT MIN(p2.id) FROM m_document_profile_request_item p2 WHERE p2.mobile_user_uuid=? AND p2.document_master_request_item_code=i.document_master_request_item_code) WHERE i.request_id=? ORDER BY m.sort_order,i.id
```


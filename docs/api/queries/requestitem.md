# SQL catalog: requestitem

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## requestitem.Files

Source: [internal/domain/requestitem/repository.go:65](../../../internal/domain/requestitem/repository.go#L65)

```sql
SELECT id,document_type,slot_code,COALESCE(file_path,''),COALESCE(original_file_name,''),COALESCE(mime_type,''),COALESCE(file_size,0),COALESCE(file_uploaded_at,''),COALESCE(check_result,''),COALESCE(check_note,''),COALESCE(is_updated,0) FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code=?
```

```sql
SELECT f.id,f.document_type,f.slot_code,COALESCE(f.file_path,''),COALESCE(f.original_file_name,''),COALESCE(f.mime_type,''),COALESCE(f.file_size,0),COALESCE(f.file_uploaded_at,''),COALESCE(f.check_result,''),COALESCE(f.check_note,''),COALESCE(f.is_updated,0) FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id INNER JOIN m_document_request r ON r.id=i.request_id WHERE r.id=? AND r.mobile_user_uuid=? AND i.document_master_request_item_code=?
```

## requestitem.ProfileItems

Source: [internal/domain/requestitem/repository.go:45](../../../internal/domain/requestitem/repository.go#L45)

```sql
SELECT m.id,m.document_master_items_code,m.storage_scope,m.sort_order,COALESCE(p.file_uploaded,0),COALESCE(p.file_path,''),COALESCE(p.file_uploaded_at,''),COALESCE(p.check_result,''),COALESCE(p.check_note,'') FROM m_document_master_request_item m LEFT JOIN (SELECT document_master_request_item_code,MIN(id) id FROM m_document_profile_request_item WHERE mobile_user_uuid=? GROUP BY document_master_request_item_code) x ON x.document_master_request_item_code=m.document_master_items_code LEFT JOIN m_document_profile_request_item p ON p.id=x.id WHERE m.is_active='YES' AND m.storage_scope='PROFILE' ORDER BY m.sort_order,m.document_master_items_code
```

## requestitem.Renewal

Source: [internal/domain/requestitem/repository.go:121](../../../internal/domain/requestitem/repository.go#L121)

```sql
SELECT r.id,i.id,s.document_status_code,COALESCE(i.approve_status,'') FROM m_document_request r INNER JOIN m_document_request_items i ON i.request_id=r.id INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.request_no=? AND r.mobile_user_uuid=? AND r.is_active='YES' AND i.document_master_request_item_code=?
```

## requestitem.UpsertProfile

Source: [internal/domain/requestitem/repository.go:92](../../../internal/domain/requestitem/repository.go#L92)

## requestitem.UpsertRenewal

Source: [internal/domain/requestitem/repository.go:129](../../../internal/domain/requestitem/repository.go#L129)

## requestitem.ValidateItem

Source: [internal/domain/requestitem/repository.go:26](../../../internal/domain/requestitem/repository.go#L26)

```sql
SELECT storage_scope FROM m_document_master_request_item WHERE document_master_items_code=? AND is_active='YES'
```

## requestitem.saveLocked

Source: [internal/domain/requestitem/repository.go:150](../../../internal/domain/requestitem/repository.go#L150)

```sql
SELECT storage_scope FROM m_document_master_request_item WHERE document_master_items_code=? AND is_active='YES' FOR UPDATE
```

```sql
SELECT s.document_status_code,COALESCE(i.approve_status,'') FROM m_document_request_items i INNER JOIN m_document_request r ON r.id=i.request_id INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE i.id=? AND i.document_master_request_item_code=? AND r.is_active='YES' FOR UPDATE
```

```sql
UPDATE m_document_profile_request_item SET file_uploaded=0 WHERE mobile_user_uuid=? AND document_master_request_item_code=?
```

```sql
UPDATE m_document_request_item_files SET file_uploaded=0 WHERE request_item_id=? AND document_master_request_item_code=?
```

## requestitem.upsertProfile

Source: [internal/domain/requestitem/repository.go:97](../../../internal/domain/requestitem/repository.go#L97)

```sql
SELECT id FROM m_document_profile_request_item WHERE mobile_user_uuid=? AND document_master_request_item_code=? AND document_type=? AND slot_code=?
```

```sql
INSERT INTO m_document_profile_request_item(id,mobile_user_uuid,document_master_request_item_code,document_type,slot_code,file_uploaded,file_path,original_file_name,mime_type,file_size,file_uploaded_at) SELECT ?,?,?,?, ?,1,?,?,?,?,NOW() FROM m_document_master_request_item WHERE document_master_items_code=? AND is_active='YES'
```

```sql
UPDATE m_document_profile_request_item SET file_uploaded=1,file_path=?,original_file_name=?,mime_type=?,file_size=?,file_uploaded_at=NOW(),check_result=NULL,check_note=NULL,is_updated=1 WHERE id=?
```

## requestitem.upsertRenewal

Source: [internal/domain/requestitem/repository.go:134](../../../internal/domain/requestitem/repository.go#L134)

```sql
SELECT id FROM m_document_request_item_files WHERE request_item_id=? AND document_type=? AND slot_code=?
```

```sql
INSERT INTO m_document_request_item_files(id,request_item_id,document_master_request_item_code,document_type,slot_code,file_uploaded,file_path,original_file_name,mime_type,file_size,file_uploaded_at) VALUES(?,?,?,?,?,1,?,?,?,?,NOW())
```

```sql
UPDATE m_document_request_item_files SET file_uploaded=1,file_path=?,original_file_name=?,mime_type=?,file_size=?,file_uploaded_at=NOW(),check_result=NULL,check_note=NULL,is_updated=1 WHERE id=?
```


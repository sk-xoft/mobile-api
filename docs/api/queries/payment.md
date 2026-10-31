# SQL catalog: payment

Static SQL literals/fragments จาก source ณ 2026-10-31 สำหรับอ้างอิงเอกสารเท่านั้น
ไม่ใช่ migration หรือ script พร้อม execute; `?` ต้อง bind ผ่าน application
บาง method รวม SQL หลาย branch, dynamic column/table, IN placeholders หรือ WHERE suffix
ให้ดู source ที่ลิงก์เพื่อทราบ arguments และเงื่อนไขที่เรียกแต่ละ statement

## payment.Acquire

Source: [internal/platform/mysql/payment/store.go:153](../../../internal/platform/mysql/payment/store.go#L153)

```sql
UPDATE mobile_payment_attempts SET lease_owner=?,lease_until=DATE_ADD(NOW(6),INTERVAL 60 SECOND) WHERE id=? AND (lease_until IS NULL OR lease_until<NOW(6)) AND review_required=0 AND status<>'SUCCESS' AND (last_checked_at IS NULL OR last_checked_at<=DATE_SUB(NOW(6),INTERVAL 5 SECOND)) AND (? OR next_check_at<=NOW(6))
```

## payment.Apply

Source: [internal/platform/mysql/payment/store.go:233](../../../internal/platform/mysql/payment/store.go#L233)

```sql
UPDATE mobile_payment_attempts SET review_required=1 WHERE id=?
```

```sql
UPDATE mobile_payment_request_state SET review_required=1 WHERE request_id=?
```

```sql
SELECT paid_attempt_id,review_required FROM mobile_payment_request_state WHERE request_id=? FOR UPDATE
```

```sql
SELECT id FROM m_document_status WHERE document_status_code='PENDING_DOCUMENT_REVIEW' AND is_active='YES' LIMIT 1
```

```sql
UPDATE m_document_request SET document_status_id=?,updated_at=NOW() WHERE id=?
```

```sql
INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,'PAYMENT_SUCCESS','PAYMENT_PENDING','PENDING_DOCUMENT_REVIEW','Payment completed',?)
```

```sql
UPDATE mobile_payment_request_state SET paid_attempt_id=? WHERE request_id=?
```

```sql
SELECT COUNT(*) FROM m_payment_transaction WHERE id=?
```

```sql
INSERT INTO m_payment_transaction(id,request_id,transaction_no,transaction_type,channel,payment_method,amount,currency,status,provider,provider_charge_id,provider_source_id,provider_transaction_id,provider_status,provider_response,idempotency_key,description,return_uri,authorize_uri,failure_code,failure_message,is_livemode,expires_at) VALUES(?,?,?,'CHARGE',?,?,?,'THB',?,'OMISE',?,?,?,?,?,?,?,?,?,?,?,?,?)
```

```sql
UPDATE m_payment_transaction SET status=?,provider_charge_id=?,provider_source_id=?,provider_transaction_id=?,provider_status=?,provider_response=?,return_uri=?,authorize_uri=?,failure_code=?,failure_message=?,is_livemode=?,expires_at=COALESCE(?,expires_at),paid_at=CASE WHEN ?='SUCCESS' THEN COALESCE(paid_at,NOW()) ELSE paid_at END,failed_at=CASE WHEN ?='FAILED' THEN COALESCE(failed_at,NOW()) ELSE failed_at END,updated_at=NOW() WHERE id=?
```

```sql
UPDATE mobile_payment_attempts SET provider_charge_id=?,charge_json=?,status=?,creation_state='CONFIRMED',review_required=?,failures=0,last_checked_at=NOW(6),next_check_at=TIMESTAMPADD(SECOND,?,NOW(6)),lease_owner=NULL,lease_until=NULL WHERE id=?
```

```sql
UPDATE mobile_payment_request_state SET active_attempt_id=NULL WHERE request_id=? AND active_attempt_id=?
```

## payment.ByKey

Source: [internal/platform/mysql/payment/store.go:73](../../../internal/platform/mysql/payment/store.go#L73)

## payment.Current

Source: [internal/platform/mysql/payment/store.go:76](../../../internal/platform/mysql/payment/store.go#L76)

## payment.Find

Source: [internal/platform/mysql/payment/store.go:79](../../../internal/platform/mysql/payment/store.go#L79)

## payment.Get

Source: [internal/platform/mysql/payment/store.go:70](../../../internal/platform/mysql/payment/store.go#L70)

## payment.MarkSent

Source: [internal/platform/mysql/payment/store.go:161](../../../internal/platform/mysql/payment/store.go#L161)

```sql
UPDATE mobile_payment_attempts SET review_required=1,lease_owner=NULL,lease_until=NULL WHERE id=? AND lease_owner=?
```

```sql
UPDATE mobile_payment_attempts a INNER JOIN mobile_payment_request_state s ON s.request_id=a.request_id SET a.creation_state='SENT',a.sent_at=NOW(6) WHERE a.id=? AND a.lease_owner=? AND a.creation_state='RESERVED' AND s.active_attempt_id=a.id AND s.paid_attempt_id IS NULL AND s.review_required=0
```

## payment.Request

Source: [internal/platform/mysql/payment/store.go:36](../../../internal/platform/mysql/payment/store.go#L36)

## payment.Reserve

Source: [internal/platform/mysql/payment/store.go:82](../../../internal/platform/mysql/payment/store.go#L82)

```sql
INSERT INTO mobile_payment_request_state(request_id) VALUES(?) ON DUPLICATE KEY UPDATE request_id=VALUES(request_id)
```

```sql
SELECT active_attempt_id,paid_attempt_id,review_required FROM mobile_payment_request_state WHERE request_id=? FOR UPDATE
```

```sql
SELECT status FROM mobile_payment_attempts WHERE id=?
```

```sql
SELECT COUNT(*) FROM m_document_request_items i INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code WHERE i.request_id=? AND m.storage_scope='REQUEST' AND NOT EXISTS (SELECT 1 FROM m_document_request_item_files f WHERE f.request_item_id=i.id AND f.file_uploaded=1 AND (f.check_result IS NULL OR f.check_result<>'fix'))
```

```sql
INSERT INTO mobile_payment_attempts(id,transaction_no,request_id,idempotency_key,fingerprint,provider_ref_id,amount_satang,input_json,status,creation_state,created_at,next_check_at) VALUES(?,?,?,?,?,?,?,?,'PROCESSING','RESERVED',?,NOW(6))
```

```sql
UPDATE mobile_payment_request_state SET active_attempt_id=? WHERE request_id=?
```

## payment.Retry

Source: [internal/platform/mysql/payment/store.go:188](../../../internal/platform/mysql/payment/store.go#L188)

```sql
UPDATE mobile_payment_request_state SET review_required=1 WHERE request_id=?
```

```sql
UPDATE mobile_payment_attempts SET failures=failures+1,review_required=(review_required OR ?),last_checked_at=NOW(6),next_check_at=TIMESTAMPADD(SECOND,?,NOW(6)),lease_owner=NULL,lease_until=NULL WHERE id=? AND (?='' OR lease_owner=?)
```

## payment.readAttempt

Source: [internal/platform/mysql/payment/store.go:46](../../../internal/platform/mysql/payment/store.go#L46)

```sql
SELECT a.id,a.transaction_no,a.request_id,r.request_no,r.mobile_user_uuid,a.idempotency_key,a.fingerprint,a.provider_ref_id,a.amount_satang,a.input_json,a.charge_json,a.status,a.creation_state,a.is_legacy,(a.review_required OR COALESCE(s.review_required,0) OR r.is_active<>'YES'),a.created_at,a.sent_at,a.last_checked_at,a.failures,(s.paid_attempt_id IS NOT NULL) FROM mobile_payment_attempts a INNER JOIN m_document_request r ON r.id=a.request_id LEFT JOIN mobile_payment_request_state s ON s.request_id=a.request_id WHERE
```

## payment.readRequest

Source: [internal/platform/mysql/payment/store.go:24](../../../internal/platform/mysql/payment/store.go#L24)

```sql
SELECT r.id,r.request_no,r.mobile_user_uuid,s.document_status_code,ROUND(r.amount*100),r.is_active='YES' FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.id=?
```

```sql
FOR UPDATE
```


-- Stop old payment writers first. Back up the database and audit duplicate
-- (request_id,idempotency_key) / provider_charge_id values before running.
-- Do not discard duplicates to make this migration pass: investigate them.
-- Fail on truncation rather than silently changing imported identifiers/keys.
SET @payment_previous_sql_mode=@@SESSION.sql_mode;
SET SESSION sql_mode=CONCAT_WS(',',NULLIF(@@SESSION.sql_mode,''),'STRICT_ALL_TABLES');
CREATE TABLE mobile_payment_attempts (
 id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 transaction_no VARCHAR(64) NOT NULL,
 request_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 idempotency_key VARCHAR(100) COLLATE utf8mb4_bin NOT NULL,
 fingerprint CHAR(64) NOT NULL DEFAULT '',
 provider_ref_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
 provider_charge_id VARCHAR(255) DEFAULT NULL,
 amount_satang BIGINT NOT NULL,
 input_json LONGTEXT NOT NULL,
 charge_json LONGTEXT DEFAULT NULL,
 status VARCHAR(24) NOT NULL,
 creation_state VARCHAR(24) NOT NULL,
 is_legacy BOOLEAN NOT NULL DEFAULT FALSE,
 review_required BOOLEAN NOT NULL DEFAULT FALSE,
 failures INT NOT NULL DEFAULT 0,
 created_at DATETIME(6) NOT NULL,
 sent_at DATETIME(6) DEFAULT NULL,
 last_checked_at DATETIME(6) DEFAULT NULL,
 next_check_at DATETIME(6) NOT NULL,
 lease_owner CHAR(36) DEFAULT NULL,
 lease_until DATETIME(6) DEFAULT NULL,
 UNIQUE KEY uq_payment_key(request_id,idempotency_key),
 UNIQUE KEY uq_payment_ref(provider_ref_id),
 UNIQUE KEY uq_payment_charge(provider_charge_id),
 KEY idx_payment_due(next_check_at,lease_until),
 KEY idx_payment_request(request_id,created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE mobile_payment_request_state (
 request_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin PRIMARY KEY,
 active_attempt_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
 paid_attempt_id CHAR(36) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL,
 review_required BOOLEAN NOT NULL DEFAULT FALSE
) ENGINE=InnoDB;
-- Provider details are fetched afresh; preserve the existing transaction IDs.
INSERT INTO mobile_payment_attempts
 (id,transaction_no,request_id,idempotency_key,provider_ref_id,provider_charge_id,
 amount_satang,input_json,charge_json,status,creation_state,is_legacy,
 review_required,created_at,next_check_at)
SELECT p.id,p.transaction_no,p.request_id,COALESCE(p.idempotency_key,p.id),p.id,
 NULLIF(p.provider_charge_id,''),ROUND(p.amount*100),
 JSON_OBJECT('paymentMethod',p.payment_method,'idempotencyKey',COALESCE(p.idempotency_key,p.id),'returnUri',COALESCE(p.return_uri,''),'platformType',''),
 JSON_OBJECT('id',COALESCE(p.provider_charge_id,''),'status',COALESCE(p.provider_status,''),
 'amount',ROUND(p.amount*100),'currency',p.currency,'livemode',IF(p.is_livemode,CAST('true' AS JSON),CAST('false' AS JSON)),
 'expires_at',IF(p.expires_at IS NULL,'0001-01-01T00:00:00Z',DATE_FORMAT(p.expires_at,'%Y-%m-%dT%H:%i:%sZ')),
 'QRCodeURI',IF(JSON_VALID(p.provider_response),COALESCE(JSON_UNQUOTE(JSON_EXTRACT(p.provider_response,'$.source.scannable_code.image.download_uri')),''),''),
 'AuthorizeURI',COALESCE(p.authorize_uri,''),'ReturnURI',COALESCE(p.return_uri,'')),
 p.status,'CONFIRMED',TRUE,IF(p.provider='OMISE' AND NULLIF(p.provider_charge_id,'') IS NOT NULL,FALSE,TRUE),
 COALESCE(p.created_at,NOW(6)),NOW(6)
FROM m_payment_transaction p WHERE p.transaction_type='CHARGE';
INSERT INTO mobile_payment_request_state(request_id,active_attempt_id,paid_attempt_id,review_required)
SELECT request_id,
 MAX(CASE WHEN status NOT IN ('SUCCESS','EXPIRED','FAILED') THEN id END),
 MIN(CASE WHEN status='SUCCESS' THEN id END),
 (SUM(status NOT IN ('SUCCESS','EXPIRED','FAILED'))>1 OR SUM(status='SUCCESS')>1 OR MAX(review_required)=1)
FROM mobile_payment_attempts GROUP BY request_id;

SET SESSION sql_mode=@payment_previous_sql_mode;

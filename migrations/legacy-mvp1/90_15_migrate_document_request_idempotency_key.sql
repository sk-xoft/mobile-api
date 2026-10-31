ALTER TABLE m_document_request
    ADD COLUMN idempotency_key VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL AFTER document_code,
    ADD UNIQUE KEY uq_docreq_user_idempotency_key (mobile_user_uuid, idempotency_key);

-- Rollback:
-- ALTER TABLE m_document_request DROP INDEX uq_docreq_user_idempotency_key;
-- ALTER TABLE m_document_request DROP COLUMN idempotency_key;

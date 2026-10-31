-- Verified hot-path indexes from security/performance immediate actions.
-- Rollback:
-- DROP INDEX idx_cert_mobile_doc ON m_certificates;
-- DROP INDEX idx_cert_mobile_end_date ON m_certificates;
-- DROP INDEX idx_documents_status_type_seq ON m_documents;
-- DROP INDEX idx_transaction_logs_trans_id ON t_transaction_logs;

CREATE INDEX idx_cert_mobile_doc
    ON m_certificates (CERT_MOBILE_UUID, CERT_DOCUMENT_CODE);

CREATE INDEX idx_cert_mobile_end_date
    ON m_certificates (CERT_MOBILE_UUID, CERT_END_DATE);

CREATE INDEX idx_documents_status_type_seq
    ON m_documents (DOCUMENT_STATUS, DOCUMENT_TYPE, DOCUMENT_SEQ);

CREATE INDEX idx_transaction_logs_trans_id
    ON t_transaction_logs (TRANS_ID);

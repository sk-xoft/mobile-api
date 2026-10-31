ALTER TABLE m_document_request
    ADD COLUMN is_active VARCHAR(3) NOT NULL DEFAULT 'YES' AFTER is_resubmit,
    ADD KEY idx_docreq_active_user (is_active, mobile_user_uuid),
    ADD CONSTRAINT chk_docreq_active
        CHECK (is_active IN ('YES', 'NO'));

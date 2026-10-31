-- Performance indexes for /v1/documents-renewals/requests/validate-and-create.
-- Rollback:
-- DROP INDEX idx_docreq_validate_active_user_doc_submitted ON m_document_request;
-- DROP INDEX idx_delivery_address_default_lookup ON m_delivery_address;
-- DROP INDEX idx_doc_reqitem_files_batch_lookup ON m_document_request_item_files;
-- DROP INDEX idx_profile_reqitem_batch_lookup ON m_document_profile_request_item;

CREATE INDEX idx_docreq_validate_active_user_doc_submitted
    ON m_document_request (
        mobile_user_uuid,
        document_code,
        is_active,
        submitted_at
    );

CREATE INDEX idx_delivery_address_default_lookup
    ON m_delivery_address (
        mobile_user_uuid,
        is_default,
        is_active,
        updated_at
    );

CREATE INDEX idx_doc_reqitem_files_batch_lookup
    ON m_document_request_item_files (
        request_item_id,
        document_type,
        slot_code
    );

CREATE INDEX idx_profile_reqitem_batch_lookup
    ON m_document_profile_request_item (
        mobile_user_uuid,
        document_master_request_item_code,
        document_type,
        slot_code
    );

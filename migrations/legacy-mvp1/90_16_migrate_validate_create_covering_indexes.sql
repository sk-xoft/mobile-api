-- Covering indexes for optimized /v1/documents-renewals/requests/validate-and-create queries.
-- Rollback:
-- DROP INDEX idx_docreq_validate_user_doc_active_submitted_id ON m_document_request;
-- DROP INDEX idx_profile_reqitem_validate_state ON m_document_profile_request_item;
-- DROP INDEX idx_doc_reqitem_files_validate_state ON m_document_request_item_files;

CREATE INDEX idx_docreq_validate_user_doc_active_submitted_id
    ON m_document_request (
        mobile_user_uuid,
        document_code,
        is_active,
        submitted_at,
        id
    );

CREATE INDEX idx_profile_reqitem_validate_state
    ON m_document_profile_request_item (
        mobile_user_uuid,
        document_master_request_item_code,
        document_type,
        slot_code,
        file_uploaded,
        check_result
    );

CREATE INDEX idx_doc_reqitem_files_validate_state
    ON m_document_request_item_files (
        request_item_id,
        document_type,
        slot_code,
        file_uploaded,
        check_result
    );

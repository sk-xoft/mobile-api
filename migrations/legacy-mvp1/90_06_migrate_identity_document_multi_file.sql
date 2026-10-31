-- Reuse m_document_profile_request_item: one row represents one upload slot.
ALTER TABLE m_document_profile_request_item
    DROP INDEX uq_profile_reqitem_mobile_master,
    ADD COLUMN document_type VARCHAR(20) NULL AFTER document_master_request_item_code,
    ADD COLUMN slot_code VARCHAR(20) NULL AFTER document_type,
    ADD COLUMN original_file_name VARCHAR(255) NULL AFTER file_path,
    ADD COLUMN mime_type VARCHAR(100) NULL AFTER original_file_name,
    ADD COLUMN file_size BIGINT NULL AFTER mime_type,
    ADD UNIQUE KEY uq_profile_reqitem_mobile_master_slot
        (mobile_user_uuid, document_master_request_item_code, document_type, slot_code);

UPDATE m_document_profile_request_item
SET document_type = 'GENERAL', slot_code = 'MAIN',
    original_file_name = COALESCE(NULLIF(SUBSTRING_INDEX(file_path, '/', -1), ''), 'legacy-file'),
    mime_type = CASE WHEN file_uploaded = 1 THEN 'application/octet-stream' ELSE NULL END,
    file_size = CASE WHEN file_uploaded = 1 THEN 1 ELSE NULL END
WHERE document_master_request_item_code <> 'MRI001';

-- Existing MRI001 files cannot be classified safely as FRONT/BACK/PASSPORT.
UPDATE m_document_profile_request_item
SET document_type = NULL, slot_code = NULL, file_uploaded = 0
WHERE document_master_request_item_code = 'MRI001';

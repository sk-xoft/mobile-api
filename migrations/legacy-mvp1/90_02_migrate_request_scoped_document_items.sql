ALTER TABLE m_document_master_request_item
    ADD COLUMN storage_scope VARCHAR(10) NOT NULL DEFAULT 'PROFILE' AFTER sort_order,
    ADD CONSTRAINT chk_doc_master_reqitem_storage_scope
        CHECK (storage_scope IN ('PROFILE', 'REQUEST'));

CREATE TABLE m_document_request_item_files (
    id                                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_item_id                     CHAR(36)        NOT NULL,
    document_master_request_item_code   VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    document_type                       VARCHAR(20)     NULL,
    slot_code                           VARCHAR(20)     NULL,
    sort_order                          TINYINT         NOT NULL DEFAULT 1,
    file_uploaded                       TINYINT(1)      NOT NULL DEFAULT 0,
    file_path                           VARCHAR(500)    NULL,
    original_file_name                  VARCHAR(255)    NULL,
    mime_type                           VARCHAR(100)    NULL,
    file_size                           BIGINT          NULL,
    file_uploaded_at                    DATETIME        NULL,
    check_result                        VARCHAR(10)     NULL,
    check_note                          TEXT            NULL,
    is_updated                          TINYINT(1)      NOT NULL DEFAULT 0,
    checked_at                          DATETIME        NULL,
    checked_by                          CHAR(36)        NULL,
    created_at                          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_doc_reqitem_files_request_item (request_item_id, sort_order),
    KEY idx_doc_reqitem_files_master_code (document_master_request_item_code),
    KEY idx_doc_reqitem_files_check_result (check_result),
    UNIQUE KEY uq_doc_reqitem_files_slot
        (request_item_id, document_type, slot_code),
    CONSTRAINT chk_doc_reqitem_files_file_uploaded
        CHECK (file_uploaded IN (0, 1)),
    CONSTRAINT chk_doc_reqitem_files_is_updated
        CHECK (is_updated IN (0, 1)),
    CONSTRAINT chk_doc_reqitem_files_check_result
        CHECK (check_result IS NULL OR check_result IN ('pass', 'fix')),
    CONSTRAINT chk_doc_reqitem_files_fix_note
        CHECK (check_result <> 'fix' OR check_note IS NOT NULL),
    CONSTRAINT fk_doc_reqitem_files_request_item
        FOREIGN KEY (request_item_id) REFERENCES m_document_request_items (id),
    CONSTRAINT fk_doc_reqitem_files_master_code
        FOREIGN KEY (document_master_request_item_code)
            REFERENCES m_document_master_request_item (document_master_items_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

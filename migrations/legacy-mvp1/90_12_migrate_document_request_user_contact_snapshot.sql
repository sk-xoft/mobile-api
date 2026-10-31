-- Snapshot applicant contact fields on renewal requests.
-- Existing rows stay NULL; new requests are populated from m_mobile_users via application flow.

ALTER TABLE m_document_request
    ADD COLUMN mobile_number VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL
        AFTER mobile_user_uuid,
    ADD COLUMN email VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL
        AFTER mobile_number;

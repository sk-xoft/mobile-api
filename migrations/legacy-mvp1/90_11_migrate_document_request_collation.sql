-- Align renewal request logical key collations with legacy/mobile and document master tables.
-- Fixes Illegal mix of collations errors when joining renewal requests with
-- m_document_profile_request_item, m_document_setting_requires, and price/document master data.

ALTER TABLE m_document_request
    MODIFY mobile_user_uuid VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    MODIFY document_code VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL;

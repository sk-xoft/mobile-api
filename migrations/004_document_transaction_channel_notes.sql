-- Apply once after m_document_transaction exists.
-- Preserve note and existing history; new columns default to NULL.
-- This migration does not change API read/write behavior.
ALTER TABLE m_document_transaction
    ADD COLUMN note_mobile TEXT NULL COMMENT 'ข้อความสำหรับผู้ใช้ Mobile',
    ADD COLUMN note_web TEXT NULL COMMENT 'ข้อความสำหรับเจ้าหน้าที่ Web';

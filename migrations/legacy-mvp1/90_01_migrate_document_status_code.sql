ALTER TABLE m_document_status
    ADD COLUMN document_status_code VARCHAR(50) NULL AFTER id;

UPDATE m_document_status
SET document_status_code = CASE name_en
    WHEN 'Payment Pending' THEN 'PAYMENT_PENDING'
    WHEN 'Pending Document Review' THEN 'PENDING_DOCUMENT_REVIEW'
    WHEN 'Pending Applicant Correction' THEN 'PENDING_APPLICANT_CORRECTION'
    WHEN 'Pending Marine Department Result' THEN 'PENDING_MARINE_DEPARTMENT_RESULT'
    WHEN 'Pending Department Document Pickup' THEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP'
    WHEN 'Delivering' THEN 'DELIVERING'
    WHEN 'Delivered' THEN 'DELIVERED'
    WHEN 'Cancelled' THEN 'CANCELLED'
    ELSE UPPER(REPLACE(name_en, ' ', '_'))
END
WHERE document_status_code IS NULL;

ALTER TABLE m_document_status
    MODIFY document_status_code VARCHAR(50) NOT NULL,
    ADD UNIQUE KEY uq_document_status_code (document_status_code);

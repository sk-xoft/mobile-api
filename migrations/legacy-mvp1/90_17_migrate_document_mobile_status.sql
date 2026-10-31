ALTER TABLE m_document_status
    ADD COLUMN document_mobile_status_code VARCHAR(50) NULL AFTER document_status_code,
    ADD COLUMN document_mobile_status_name_th VARCHAR(255) NULL AFTER document_mobile_status_code,
    ADD COLUMN document_mobile_status_name_en VARCHAR(255) NULL AFTER document_mobile_status_name_th;

UPDATE m_document_status
SET document_mobile_status_code = CASE document_status_code
        WHEN 'PAYMENT_PENDING' THEN 'DOCUMENT_REVIEW'
        WHEN 'PENDING_DOCUMENT_REVIEW' THEN 'DOCUMENT_REVIEW'
        WHEN 'PENDING_APPLICANT_CORRECTION' THEN 'DOCUMENT_REVIEW'
        WHEN 'PENDING_MARINE_DEPARTMENT_RESULT' THEN 'MARINE_DEPARTMENT_RESULT'
        WHEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP' THEN 'DEPARTMENT_DOCUMENT_PICKUP'
        WHEN 'DELIVERING' THEN 'DELIVERING'
        WHEN 'DELIVERED' THEN 'DELIVERED'
        ELSE NULL
    END,
    document_mobile_status_name_th = CASE document_status_code
        WHEN 'PAYMENT_PENDING' THEN 'ตรวจเอกสาร'
        WHEN 'PENDING_DOCUMENT_REVIEW' THEN 'ตรวจเอกสาร'
        WHEN 'PENDING_APPLICANT_CORRECTION' THEN 'ตรวจเอกสาร'
        WHEN 'PENDING_MARINE_DEPARTMENT_RESULT' THEN 'รอผลจากกรมเจ้าท่า'
        WHEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP' THEN 'รอรับเอกสารจากกรมเจ้าท่า'
        WHEN 'DELIVERING' THEN 'กำลังจัดส่ง'
        WHEN 'DELIVERED' THEN 'จัดส่งสำเร็จ'
        ELSE NULL
    END,
    document_mobile_status_name_en = CASE document_status_code
        WHEN 'PAYMENT_PENDING' THEN 'Document Review'
        WHEN 'PENDING_DOCUMENT_REVIEW' THEN 'Document Review'
        WHEN 'PENDING_APPLICANT_CORRECTION' THEN 'Document Review'
        WHEN 'PENDING_MARINE_DEPARTMENT_RESULT' THEN 'Marine Department Result'
        WHEN 'PENDING_DEPARTMENT_DOCUMENT_PICKUP' THEN 'Department Document Pickup'
        WHEN 'DELIVERING' THEN 'Delivering'
        WHEN 'DELIVERED' THEN 'Delivered'
        ELSE NULL
    END
WHERE document_status_code IN (
    'PAYMENT_PENDING',
    'PENDING_DOCUMENT_REVIEW',
    'PENDING_APPLICANT_CORRECTION',
    'PENDING_MARINE_DEPARTMENT_RESULT',
    'PENDING_DEPARTMENT_DOCUMENT_PICKUP',
    'DELIVERING',
    'DELIVERED'
);

ALTER TABLE m_document_request_items
    ADD COLUMN request_no VARCHAR(20) NULL AFTER request_id;

UPDATE m_document_request_items i
INNER JOIN m_document_request r ON r.id = i.request_id
SET i.request_no = r.request_no;

ALTER TABLE m_document_request_items
    MODIFY COLUMN request_no VARCHAR(20) NOT NULL,
    ADD KEY idx_doc_reqitems_request_no (request_no);

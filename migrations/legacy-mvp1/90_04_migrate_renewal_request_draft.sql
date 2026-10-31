ALTER TABLE m_document_status
    ADD COLUMN is_mobile_visible VARCHAR(3) NOT NULL DEFAULT 'YES' AFTER is_active,
    ADD CONSTRAINT chk_document_status_mobile_visible
        CHECK (is_mobile_visible IN ('YES', 'NO'));

INSERT INTO m_document_status (
    document_status_code,
    name_th,
    name_en,
    css_color,
    is_active,
    is_mobile_visible
)
VALUES ('PAYMENT_PENDING', 'รอชำระเงิน', 'Payment Pending', '#999999', 'YES', 'NO')
ON DUPLICATE KEY UPDATE
    document_status_code = VALUES(document_status_code),
    css_color = VALUES(css_color),
    is_active = 'YES',
    is_mobile_visible = 'NO';

CREATE TABLE m_document_request_sequence (
    period CHAR(4) NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (period),
    CONSTRAINT chk_docreq_sequence_number CHECK (last_number >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

ALTER TABLE m_document_request
    ADD COLUMN price_setting_id CHAR(36) NULL AFTER document_status_id,
    ADD COLUMN delivery_address_id CHAR(36) NULL AFTER price_setting_id,
    ADD KEY idx_docreq_price_setting (price_setting_id),
    ADD KEY idx_docreq_delivery_address (delivery_address_id),
    ADD CONSTRAINT fk_docreq_price_setting
        FOREIGN KEY (price_setting_id) REFERENCES m_document_prices_setting (id),
    ADD CONSTRAINT fk_docreq_delivery_address
        FOREIGN KEY (delivery_address_id) REFERENCES m_delivery_address (id);

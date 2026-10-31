-- Snapshot delivery address per renewal request.
-- New requests are populated by application flow.
-- Existing rows are backfilled only when m_document_request.delivery_address_id points to an address.

CREATE TABLE IF NOT EXISTS m_document_request_delivery_address (
    id                          CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_id                  CHAR(36)        NOT NULL,
    source_delivery_address_id  CHAR(36)        NULL,
    mobile_user_uuid            VARCHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    first_name                  VARCHAR(100)    NOT NULL,
    last_name                   VARCHAR(100)    NOT NULL,
    address_line                VARCHAR(500)    NOT NULL,
    province                    VARCHAR(100)    NOT NULL,
    district                    VARCHAR(100)    NOT NULL,
    sub_district                VARCHAR(100)    NOT NULL,
    postal_code                 VARCHAR(10)     NOT NULL,
    mobile_number               VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
    created_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_docreq_delivery_address_request (request_id),
    KEY idx_docreq_delivery_address_source (source_delivery_address_id),
    KEY idx_docreq_delivery_address_user (mobile_user_uuid),

    CONSTRAINT fk_docreq_delivery_address_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id),
    CONSTRAINT fk_docreq_delivery_address_source
        FOREIGN KEY (source_delivery_address_id) REFERENCES m_delivery_address (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

INSERT INTO m_document_request_delivery_address (
    request_id,
    source_delivery_address_id,
    mobile_user_uuid,
    first_name,
    last_name,
    address_line,
    province,
    district,
    sub_district,
    postal_code,
    mobile_number
)
SELECT
    r.id,
    da.id,
    r.mobile_user_uuid,
    da.first_name,
    da.last_name,
    da.address_line,
    da.province,
    da.district,
    da.sub_district,
    da.postal_code,
    COALESCE(r.mobile_number, mu.MOBILE_NUMBER)
FROM m_document_request r
INNER JOIN m_delivery_address da ON da.id = r.delivery_address_id
LEFT JOIN m_mobile_users mu ON mu.MOBILE_UUID = r.mobile_user_uuid
LEFT JOIN m_document_request_delivery_address snap ON snap.request_id = r.id
WHERE r.delivery_address_id IS NOT NULL
  AND snap.id IS NULL;

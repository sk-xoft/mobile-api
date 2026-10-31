CREATE TABLE m_document_status (
    id              CHAR(36)        NOT NULL DEFAULT (UUID()),
    document_status_code VARCHAR(50) NOT NULL,
    document_mobile_status_code VARCHAR(50) NULL,
    document_mobile_status_name_th VARCHAR(255) NULL,
    document_mobile_status_name_en VARCHAR(255) NULL,
    name_th         VARCHAR(255)    NOT NULL,
    name_en         VARCHAR(255)    NOT NULL,
    css_color       VARCHAR(100)    NOT NULL,
    is_active       VARCHAR(3)      NOT NULL DEFAULT 'YES',
    is_mobile_visible VARCHAR(3)    NOT NULL DEFAULT 'YES',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_document_status_code (document_status_code),
    UNIQUE KEY uq_document_status_name_th (name_th),
    UNIQUE KEY uq_document_status_name_en (name_en),
    CONSTRAINT chk_document_status_active
        CHECK (is_active IN ('YES', 'NO')),
    CONSTRAINT chk_document_status_mobile_visible
        CHECK (is_mobile_visible IN ('YES', 'NO'))
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_request (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_no          VARCHAR(20)     NOT NULL,
    mobile_user_uuid    VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    mobile_number       VARCHAR(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
    email               VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
    document_code       VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    idempotency_key     VARCHAR(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
    document_status_id  CHAR(36)        NOT NULL,
    price_setting_id    CHAR(36)        NULL,
    delivery_address_id CHAR(36)        NULL,
    is_resubmit         TINYINT(1)      NOT NULL DEFAULT 0,     -- 1 = ผู้ยื่น resubmit หลังแก้ไข
    is_active           VARCHAR(3)      NOT NULL DEFAULT 'YES', -- YES = active, NO = soft deleted
    amount              DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,  -- ยอดชำระ (บาท)
    submitted_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    submitted_by        CHAR(36)        NULL,                   -- NULL = ลูกเรือยื่นเอง
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_docreq_request_no (request_no),
    UNIQUE KEY uq_docreq_user_idempotency_key (mobile_user_uuid, idempotency_key),
    KEY idx_docreq_mobile_user (mobile_user_uuid),
    KEY idx_docreq_document_code (document_code),
    KEY idx_docreq_price_setting (price_setting_id),
    KEY idx_docreq_delivery_address (delivery_address_id),
    KEY idx_docreq_active_user (is_active, mobile_user_uuid),
    KEY idx_docreq_status_submitted (document_status_id, submitted_at),
    KEY idx_docreq_validate_user_doc_active_submitted_id
        (mobile_user_uuid, document_code, is_active, submitted_at, id),
    CONSTRAINT chk_docreq_amount
        CHECK (amount >= 0),
    CONSTRAINT chk_docreq_resubmit
        CHECK (is_resubmit IN (0, 1)),
    CONSTRAINT chk_docreq_active
        CHECK (is_active IN ('YES', 'NO')),
    CONSTRAINT fk_docreq_status
        FOREIGN KEY (document_status_id) REFERENCES m_document_status (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_prices_setting (
    id                      CHAR(36)        NOT NULL DEFAULT (UUID()),
    document_code           VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    government_fee          DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    document_processing_fee DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    shipping_fee            DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    shipping_discount       DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    service_fee_discount    DECIMAL(10, 2)  NOT NULL DEFAULT 0.00,
    effective_from          DATE             NOT NULL,
    effective_to            DATE             NULL,
    is_active               VARCHAR(3)      NOT NULL DEFAULT 'YES',
    created_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uq_docprice_document_effective (document_code, effective_from),
    KEY idx_docprice_active_effective (document_code, is_active, effective_from, effective_to),
    CONSTRAINT chk_docprice_fee
        CHECK (
            government_fee >= 0
            AND document_processing_fee >= 0
            AND shipping_fee >= 0
            AND shipping_discount >= 0
            AND service_fee_discount >= 0
        ),
    CONSTRAINT chk_docprice_active
        CHECK (is_active IN ('YES', 'NO')),
    CONSTRAINT chk_docprice_effective_period
        CHECK (effective_to IS NULL OR effective_to >= effective_from),
    CONSTRAINT fk_docprice_document
        FOREIGN KEY (document_code) REFERENCES m_documents (DOCUMENT_CODE)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

ALTER TABLE m_document_request
    ADD CONSTRAINT fk_docreq_price_setting
        FOREIGN KEY (price_setting_id) REFERENCES m_document_prices_setting (id);

CREATE TABLE m_document_request_sequence (
    period          CHAR(4)     NOT NULL,
    last_number     INT         NOT NULL DEFAULT 0,
    updated_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (period),
    CONSTRAINT chk_docreq_sequence_number CHECK (last_number >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_payment_transaction (
    id                      CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_id              CHAR(36)        NOT NULL,
    parent_transaction_id   CHAR(36)        NULL,

    transaction_no          VARCHAR(40)     NOT NULL,
    transaction_type        VARCHAR(10)     NOT NULL DEFAULT 'CHARGE',
                                            -- CHARGE | REFUND
    channel                 VARCHAR(30)     NOT NULL,
                                            -- PROMPTPAY | CREDIT_CARD | MOBILE_BANKING
    payment_method          VARCHAR(50)     NOT NULL,
                                            -- promptpay | card | internet_banking_scb | ...

    amount                  DECIMAL(12, 2)  NOT NULL,
    currency                CHAR(3)         NOT NULL DEFAULT 'THB',
    refunded_amount         DECIMAL(12, 2)  NOT NULL DEFAULT 0.00,
                                            -- ยอดคืนสะสม ใช้กับ CHARGE เท่านั้น
    status                  VARCHAR(20)     NOT NULL DEFAULT 'PENDING',
                                            -- PENDING | PROCESSING | SUCCESS | FAILED |
                                            -- EXPIRED | PARTIALLY_REFUNDED | REFUNDED | CANCELLED

    provider                VARCHAR(20)     NOT NULL DEFAULT 'OMISE',
    provider_charge_id      VARCHAR(100)    NULL, -- Omise chrg_... / chrg_test_...
    provider_source_id      VARCHAR(100)    NULL, -- Omise src_... / src_test_...
    provider_refund_id      VARCHAR(100)    NULL, -- Omise rfnd_... / rfnd_test_...
    provider_transaction_id VARCHAR(100)    NULL, -- Omise trxn_... (ถ้ามี)
    provider_status         VARCHAR(30)     NULL, -- status ดิบล่าสุดจาก Omise
    provider_response       JSON            NULL, -- response ล่าสุด (ห้ามมี secret/token)

    idempotency_key         VARCHAR(100)    NOT NULL,
    description             VARCHAR(255)    NULL,
    return_uri              VARCHAR(500)    NULL,
    authorize_uri           VARCHAR(500)    NULL, -- URL สำหรับ 3DS / bank redirect

    bank_code               VARCHAR(30)     NULL,
    card_brand              VARCHAR(30)     NULL,
    card_last_digits        CHAR(4)         NULL,
    failure_code            VARCHAR(100)    NULL,
    failure_message         VARCHAR(500)    NULL,

    is_livemode             TINYINT(1)      NOT NULL DEFAULT 0,
    expires_at              DATETIME        NULL,
    paid_at                 DATETIME        NULL,
    failed_at               DATETIME        NULL,
    refunded_at             DATETIME        NULL,
    created_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP
                                            ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_payment_transaction_no (transaction_no),
    UNIQUE KEY uq_payment_idempotency_key (idempotency_key),
    UNIQUE KEY uq_payment_provider_charge (provider, provider_charge_id),
    UNIQUE KEY uq_payment_provider_refund (provider, provider_refund_id),
    KEY idx_payment_request_created (request_id, created_at),
    KEY idx_payment_status_updated (status, updated_at),
    KEY idx_payment_parent (parent_transaction_id),

    CONSTRAINT chk_payment_amount
        CHECK (amount > 0),
    CONSTRAINT chk_payment_refunded_amount
        CHECK (refunded_amount >= 0 AND refunded_amount <= amount),
    CONSTRAINT chk_payment_type
        CHECK (transaction_type IN ('CHARGE', 'REFUND')),
    CONSTRAINT chk_payment_channel
        CHECK (channel IN ('PROMPTPAY', 'CREDIT_CARD', 'MOBILE_BANKING')),
    CONSTRAINT chk_payment_status
        CHECK (status IN (
            'PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'EXPIRED',
            'PARTIALLY_REFUNDED', 'REFUNDED', 'CANCELLED'
        )),
    CONSTRAINT chk_payment_currency
        CHECK (currency = 'THB'),
    CONSTRAINT chk_payment_livemode
        CHECK (is_livemode IN (0, 1)),
    CONSTRAINT fk_payment_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id),
    CONSTRAINT fk_payment_parent
        FOREIGN KEY (parent_transaction_id) REFERENCES m_payment_transaction (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_master_request_item (
    id                          CHAR(36)        NOT NULL DEFAULT (UUID()),
    document_master_items_code  VARCHAR(10)     CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    document_master_items_name  CHAR(36)        NOT NULL,
    sort_order                  TINYINT         NOT NULL DEFAULT 1,
    storage_scope               VARCHAR(10)     NOT NULL DEFAULT 'PROFILE',
    is_active                   VARCHAR(3)      NOT NULL DEFAULT 'YES',
    created_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_doc_master_reqitem_code (document_master_items_code),
    KEY idx_doc_master_reqitem_active_sort (is_active, sort_order),

    CONSTRAINT chk_doc_master_reqitem_active
        CHECK (is_active IN ('YES', 'NO')),
    CONSTRAINT chk_doc_master_reqitem_storage_scope
        CHECK (storage_scope IN ('PROFILE', 'REQUEST'))
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_request_items (
    id                                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_id                          CHAR(36)        NOT NULL,
    request_no                          VARCHAR(20)     NOT NULL,
    document_master_request_item_code   VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    approve_status                      VARCHAR(10)     NOT NULL DEFAULT 'PENDING',
    note                                TEXT            NULL,
    created_at                          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_doc_reqitems_request_master (request_id, document_master_request_item_code),
    KEY idx_doc_reqitems_request_status (request_id, approve_status),
    KEY idx_doc_reqitems_request_no (request_no),
    KEY idx_doc_reqitems_master_code (document_master_request_item_code),

    CONSTRAINT chk_doc_reqitems_approve_status
        CHECK (approve_status IN ('PENDING', 'PASS', 'FIX')),
    CONSTRAINT chk_doc_reqitems_fix_note
        CHECK (approve_status <> 'FIX' OR note IS NOT NULL),
    CONSTRAINT fk_doc_reqitems_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id),
    CONSTRAINT fk_doc_reqitems_master_code
        FOREIGN KEY (document_master_request_item_code) REFERENCES m_document_master_request_item (document_master_items_code)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_profile_request_item (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    mobile_user_uuid    VARCHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    document_master_request_item_code VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    document_type       VARCHAR(20)     NULL,
    slot_code           VARCHAR(20)     NULL,
    sort_order          TINYINT         NOT NULL DEFAULT 1, -- ลำดับที่ในตาราง (1, 2, 3, 4)
    file_uploaded       TINYINT(1)      NOT NULL DEFAULT 0,
    file_path           VARCHAR(500)    NULL,             -- path หรือ URL ของไฟล์ที่อัปโหลด
    original_file_name  VARCHAR(255)    NULL,
    mime_type           VARCHAR(100)    NULL,
    file_size           BIGINT          NULL,
    file_uploaded_at    DATETIME        NULL,
    check_result        VARCHAR(10)     NULL,             -- 'pass' | 'fix' | NULL (ยังไม่ตรวจ)
    check_note          TEXT            NULL,             -- หมายเหตุเมื่อผล = 'fix' (required)
    is_updated          TINYINT(1)      NOT NULL DEFAULT 0, -- 1 = ผู้ยื่น resubmit ไฟล์ใหม่
    checked_at          DATETIME        NULL,
    checked_by          CHAR(36)        NULL,
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_profile_reqitem_mobile_user_sort (mobile_user_uuid, sort_order),
    KEY idx_profile_reqitem_master_code (document_master_request_item_code),
    KEY idx_profile_reqitem_check_result (check_result),
    KEY idx_profile_reqitem_validate_state
        (mobile_user_uuid, document_master_request_item_code,
         document_type, slot_code, file_uploaded, check_result),
    UNIQUE KEY uq_profile_reqitem_mobile_master_slot
        (mobile_user_uuid, document_master_request_item_code, document_type, slot_code),
    CONSTRAINT chk_profile_reqitem_file_uploaded
        CHECK (file_uploaded IN (0, 1)),
    CONSTRAINT chk_profile_reqitem_is_updated
        CHECK (is_updated IN (0, 1)),
    CONSTRAINT chk_profile_reqitem_check_result
        CHECK (check_result IS NULL OR check_result IN ('pass', 'fix')),
    CONSTRAINT chk_profile_reqitem_fix_note
        CHECK (check_result <> 'fix' OR check_note IS NOT NULL),
    CONSTRAINT fk_profile_reqitem_mobile_user
        FOREIGN KEY (mobile_user_uuid) REFERENCES m_mobile_users (MOBILE_UUID),
    CONSTRAINT fk_profile_reqitem_master_code
        FOREIGN KEY (document_master_request_item_code) REFERENCES m_document_master_request_item (document_master_items_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    KEY idx_doc_reqitem_files_validate_state
        (request_item_id, document_type, slot_code, file_uploaded, check_result),
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
        FOREIGN KEY (document_master_request_item_code) REFERENCES m_document_master_request_item (document_master_items_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_setting_requires (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    document_code       VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    document_master_request_item_code VARCHAR(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    sort_order          TINYINT         NOT NULL DEFAULT 1,
    is_required         TINYINT(1)      NOT NULL DEFAULT 1,
    is_active           VARCHAR(3)      NOT NULL DEFAULT 'YES',
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_doc_setting_requires (document_code, document_master_request_item_code),
    KEY idx_doc_setting_requires_doc_sort (document_code, sort_order),
    KEY idx_doc_setting_requires_item (document_master_request_item_code),

    CONSTRAINT chk_doc_setting_requires_required
        CHECK (is_required IN (0, 1)),
    CONSTRAINT chk_doc_setting_requires_active
        CHECK (is_active IN ('YES', 'NO')),

    CONSTRAINT fk_doc_setting_requires_document
        FOREIGN KEY (document_code) REFERENCES m_documents (DOCUMENT_CODE),
    CONSTRAINT fk_doc_setting_requires_item
        FOREIGN KEY (document_master_request_item_code) REFERENCES m_document_master_request_item (document_master_items_code)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_document_transaction (
    id              CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_id      CHAR(36)        NOT NULL,
    action          VARCHAR(50)     NOT NULL,   -- รหัส action (ดู Action Reference)
    from_status     VARCHAR(50)     NULL,       -- status ก่อนเปลี่ยน (NULL = สร้างใหม่)
    to_status       VARCHAR(50)     NOT NULL,   -- status หลังเปลี่ยน
    note            TEXT            NULL,       -- หมายเหตุเพิ่มเติม (optional)
    actioned_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actioned_by     CHAR(36)        NULL,       -- NULL = ระบบ / ลูกเรือ

    PRIMARY KEY (id),
    KEY idx_doctx_request_actioned (request_id, actioned_at),
    KEY idx_doctx_action (action),
    CONSTRAINT chk_doctx_action
        CHECK (action IN (
            'CREATE', 'PAYMENT_SUCCESS', 'SEND_BACK', 'RESUBMIT', 'CHECK_DOCS',
            'SUBMIT_TO_DEPT', 'RECORD_DEPT_RESULT', 'RECEIVE_FROM_DEPT',
            'RECORD_DELIVERY', 'DELIVERY_COMPLETE', 'CANCEL'
        )),
    CONSTRAINT fk_doctx_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_dept_submission (
    id                      CHAR(36)    NOT NULL DEFAULT (UUID()),
    request_id              CHAR(36)    NOT NULL,
    submitted_to_dept_date  DATE        NOT NULL,   -- วันที่ยื่น Google Form กรมเจ้าท่า
    submitted_by            CHAR(36)    NOT NULL,
    available_from_date     DATE        NULL,       -- วันที่รับเอกสารได้ตั้งแต่ (กรมแจ้งมา)
    received_from_dept_date DATE        NULL,       -- วันที่รับเอกสารจริง
    recorded_at             DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_deptsubmit_request (request_id),  -- บังคับ one-to-one
    KEY idx_deptsubmit_submitted_date (submitted_to_dept_date),
    CONSTRAINT fk_deptsubmit_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_mobile_number_history (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    mobile_user_uuid    VARCHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    old_mobile_number   VARCHAR(10)     NULL,
    new_mobile_number   VARCHAR(10)     NOT NULL,
    changed_by          VARCHAR(255)    NOT NULL,
    changed_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_mobile_number_history_user_changed (mobile_user_uuid, changed_at),
    CONSTRAINT fk_mobile_number_history_user
        FOREIGN KEY (mobile_user_uuid) REFERENCES m_mobile_users (MOBILE_UUID)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

CREATE TABLE m_delivery_address (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    mobile_user_uuid    VARCHAR(36) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
    first_name          VARCHAR(100)    NOT NULL,
    last_name           VARCHAR(100)    NOT NULL,
    address_line        VARCHAR(500)    NOT NULL,   -- บ้านเลขที่, ซอย, หมู่, ถนน
    province            VARCHAR(100)    NOT NULL,
    district            VARCHAR(100)    NOT NULL,   -- เขต/อำเภอ
    sub_district        VARCHAR(100)    NOT NULL,   -- แขวง/ตำบล
    postal_code         VARCHAR(10)     NOT NULL,
    is_default          TINYINT(1)      NOT NULL DEFAULT 0,
    is_active           VARCHAR(3)      NOT NULL DEFAULT 'YES',
    default_owner_uuid  VARCHAR(36) GENERATED ALWAYS AS (
        CASE WHEN is_default = 1 AND is_active = 'YES' THEN mobile_user_uuid ELSE NULL END
    ) STORED,
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    KEY idx_delivery_address_mobile_user (mobile_user_uuid, is_active),
    KEY idx_delivery_address_postal_code (postal_code),
    UNIQUE KEY uq_delivery_address_active_default (default_owner_uuid),

    CONSTRAINT chk_delivery_address_default
        CHECK (is_default IN (0, 1)),
    CONSTRAINT chk_delivery_address_active
        CHECK (is_active IN ('YES', 'NO')),
    CONSTRAINT fk_delivery_address_mobile_user
        FOREIGN KEY (mobile_user_uuid) REFERENCES m_mobile_users (MOBILE_UUID)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

ALTER TABLE m_document_request
    ADD CONSTRAINT fk_docreq_delivery_address
        FOREIGN KEY (delivery_address_id) REFERENCES m_delivery_address (id);

CREATE TABLE m_document_request_delivery_address (
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

CREATE TABLE m_delivery (
    id                  CHAR(36)        NOT NULL DEFAULT (UUID()),
    request_id          CHAR(36)        NOT NULL,
    delivery_address_id CHAR(36)        NULL,
    tracking_no         VARCHAR(50)     NOT NULL,               -- เช่น 'EF123456789TH'
    carrier             VARCHAR(100)    NOT NULL DEFAULT 'Thailand Post',
    shipped_date        DATE            NOT NULL,               -- วันที่นำส่งไปรษณีย์
    delivery_status     VARCHAR(20)     NOT NULL DEFAULT 'in_transit',
    shipped_recorded_at DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    shipped_by          CHAR(36)        NOT NULL,
    delivered_at        DATETIME        NULL,                   -- NULL = ยังส่งไม่ถึง
    created_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (id),
    UNIQUE KEY uq_delivery_request (request_id),    -- บังคับ one-to-one
    KEY idx_delivery_address (delivery_address_id),
    KEY idx_delivery_tracking_no (tracking_no),
    KEY idx_delivery_status_updated (delivery_status, updated_at),
    CONSTRAINT chk_delivery_status
        CHECK (delivery_status IN ('pending', 'in_transit', 'delivered', 'failed', 'returned')),
    CONSTRAINT fk_delivery_request
        FOREIGN KEY (request_id) REFERENCES m_document_request (id),
    CONSTRAINT fk_delivery_address
        FOREIGN KEY (delivery_address_id) REFERENCES m_delivery_address (id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;

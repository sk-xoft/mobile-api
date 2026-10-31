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

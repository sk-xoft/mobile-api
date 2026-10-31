ALTER TABLE m_document_setting_requires
    ADD INDEX idx_doc_setting_requires_validate (
        document_code,
        is_required,
        is_active,
        sort_order,
        document_master_request_item_code
    );


ALTER TABLE m_mobile_users ADD INDEX idx_username (USERNAME);
ALTER TABLE m_mobile_users ADD INDEX idx_email (EMAIL);
ALTER TABLE m_mobile_users ADD INDEX idx_mobile_uuid (MOBILE_UUID);

ALTER TABLE t_session ADD UNIQUE INDEX ux_t_session_client_session_id (CLIENT_SESSION_ID);
ALTER TABLE t_session ADD INDEX ix_t_session_user_id (USER_ID);

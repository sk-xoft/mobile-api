ALTER TABLE m_payment_transaction
    DROP CHECK chk_payment_channel;

ALTER TABLE m_payment_transaction
    ADD CONSTRAINT chk_payment_channel
        CHECK (channel IN ('PROMPTPAY', 'CREDIT_CARD', 'MOBILE_BANKING'));

ALTER TABLE m_document_transaction
    DROP CHECK chk_doctx_action;

ALTER TABLE m_document_transaction
    ADD CONSTRAINT chk_doctx_action
        CHECK (action IN (
            'CREATE', 'PAYMENT_SUCCESS', 'SEND_BACK', 'RESUBMIT', 'CHECK_DOCS',
            'SUBMIT_TO_DEPT', 'RECORD_DEPT_RESULT', 'RECEIVE_FROM_DEPT',
            'RECORD_DELIVERY', 'DELIVERY_COMPLETE', 'CANCEL'
        ));

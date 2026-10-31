-- Existing deployments: add effective periods to renewal prices.
-- Existing rows become effective from their creation date.
ALTER TABLE m_document_prices_setting
    ADD COLUMN effective_from DATE NULL AFTER service_fee_discount,
    ADD COLUMN effective_to DATE NULL AFTER effective_from;

UPDATE m_document_prices_setting
SET effective_from = DATE(created_at)
WHERE effective_from IS NULL;

ALTER TABLE m_document_prices_setting
    MODIFY effective_from DATE NOT NULL,
    DROP INDEX uq_docprice_document_code,
    ADD UNIQUE KEY uq_docprice_document_effective (document_code, effective_from),
    ADD KEY idx_docprice_active_effective (document_code, is_active, effective_from, effective_to),
    ADD CONSTRAINT chk_docprice_effective_period
        CHECK (effective_to IS NULL OR effective_to >= effective_from);

-- Prevent more than one active default delivery address per mobile user.
-- MySQL UNIQUE indexes allow multiple NULL values, so non-default/inactive rows do not conflict.
ALTER TABLE m_delivery_address
    ADD COLUMN default_owner_uuid VARCHAR(36) GENERATED ALWAYS AS (
        CASE WHEN is_default = 1 AND is_active = 'YES' THEN mobile_user_uuid ELSE NULL END
    ) STORED,
    ADD UNIQUE KEY uq_delivery_address_active_default (default_owner_uuid);

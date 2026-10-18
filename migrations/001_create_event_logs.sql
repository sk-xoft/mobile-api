CREATE TABLE IF NOT EXISTS event_logs (
    id CHAR(36) NOT NULL,
    event_name VARCHAR(100) NOT NULL,
    actor_user_id VARCHAR(100) NULL,
    aggregate_type VARCHAR(100) NULL,
    aggregate_id VARCHAR(100) NULL,
    request_id VARCHAR(100) NULL,
    trace_id VARCHAR(100) NULL,
    metadata JSON NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_event_logs_name_created (event_name, created_at),
    KEY idx_event_logs_actor_created (actor_user_id, created_at),
    KEY idx_event_logs_aggregate (aggregate_type, aggregate_id),
    KEY idx_event_logs_request (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

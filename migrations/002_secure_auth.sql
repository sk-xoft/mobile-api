-- Apply before deploying the authentication hardening release.
-- Existing sessions are deliberately invalidated; all users must log in again.
CREATE TABLE IF NOT EXISTS mobile_auth_tokens (
 token_hash CHAR(64) PRIMARY KEY,
 user_id VARCHAR(255) NOT NULL,
 purpose VARCHAR(16) NOT NULL,
 expires_at DATETIME NOT NULL,
 INDEX idx_auth_token_user (user_id, purpose)
) ENGINE=InnoDB;
CREATE TABLE IF NOT EXISTS mobile_auth_sessions (
 session_id VARCHAR(255) PRIMARY KEY,
 refresh_hash CHAR(64) NOT NULL UNIQUE,
 expires_at DATETIME NOT NULL
) ENGINE=InnoDB;
UPDATE t_session SET IS_ONLINE = 'NO';

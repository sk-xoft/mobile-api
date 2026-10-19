package authmysql

import (
	"context"
	"database/sql"
	"errors"
	domain "github.com/smartseaman/mobile-api/internal/domain/auth"
	"time"
)

func (r *Repository) CreateActionToken(ctx context.Context, userID, purpose, hash string, expires time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Lock the account so concurrent requests cannot leave multiple usable links.
	var id string
	if err := tx.QueryRowContext(ctx, "SELECT MOBILE_UUID FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE", userID).Scan(&id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose=?", userID, purpose); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO mobile_auth_tokens(token_hash,user_id,purpose,expires_at) VALUES(?,?,?,?)", hash, userID, purpose, expires); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ConsumeActionToken(ctx context.Context, hash, purpose, password string) (domain.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, err
	}
	defer tx.Rollback()
	// Always lock the user before locking the token, matching issuance and password changes.
	var userID string
	err = tx.QueryRowContext(ctx, "SELECT user_id FROM mobile_auth_tokens WHERE token_hash=? AND purpose=? AND expires_at>NOW()", hash, purpose).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	var user domain.User
	err = tx.QueryRowContext(ctx, "SELECT MOBILE_UUID,USERNAME,EMAIL,USER_STATUS FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE", userID).Scan(&user.MobileUUID, &user.Username, &user.Email, &user.Status)
	if err != nil {
		return domain.User{}, err
	}
	var valid string
	err = tx.QueryRowContext(ctx, "SELECT token_hash FROM mobile_auth_tokens WHERE token_hash=? AND purpose=? AND expires_at>NOW() FOR UPDATE", hash, purpose).Scan(&valid)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	if purpose == "activation" {
		if user.Status != "I" {
			return domain.User{}, domain.ErrNotFound
		}
		_, err = tx.ExecContext(ctx, "UPDATE m_mobile_users SET USER_STATUS='A',UPDATE_DATE=NOW(),UPDATE_BY=? WHERE MOBILE_UUID=?", user.Username, user.MobileUUID)
	} else if purpose == "reset" {
		_, err = tx.ExecContext(ctx, "UPDATE m_mobile_users SET PASSWORD=? WHERE MOBILE_UUID=?", password, user.MobileUUID)
		if err == nil {
			_, err = tx.ExecContext(ctx, "UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=?", user.MobileUUID)
		}
	} else {
		return domain.User{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.User{}, err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose=?", user.MobileUUID, purpose); err != nil {
		return domain.User{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateSecureSession(ctx context.Context, sessionID, userID, token, refreshHash, correlationID string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	if err := tx.QueryRowContext(ctx, "SELECT USER_STATUS FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE", userID).Scan(&status); err != nil {
		return err
	}
	if status != "A" {
		return domain.ErrNotFound
	}
	if _, err := tx.ExecContext(ctx, "UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=? AND IS_ONLINE='YES'", userID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO t_session (CLIENT_SESSION_ID,USER_ID,LOGIN_TIME,DEVICE_MODEL,TOKEN,CREATE_DATE,CREATE_BY,UPDATE_DATE,UPDATE_BY,CORRELATIONID,IS_ONLINE) VALUES(?,?,?,'MOBILE',?,?,?,?,?,?,'YES')`, sessionID, userID, now, token, now, userID, now, userID, correlationID)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO mobile_auth_sessions(session_id,refresh_hash,expires_at) VALUES(?,?,?)", sessionID, refreshHash, now.Add(30*24*time.Hour)); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) FindRefreshSession(ctx context.Context, hash string) (string, string, error) {
	var sessionID, userID string
	err := r.db.QueryRowContext(ctx, `SELECT s.CLIENT_SESSION_ID,s.USER_ID FROM t_session s INNER JOIN mobile_auth_sessions a ON a.session_id=s.CLIENT_SESSION_ID INNER JOIN m_mobile_users u ON u.MOBILE_UUID=s.USER_ID WHERE a.refresh_hash=? AND a.expires_at>NOW() AND s.IS_ONLINE='YES' AND u.USER_STATUS='A' LIMIT 1`, hash).Scan(&sessionID, &userID)
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return sessionID, userID, err
}

func (r *Repository) RotateRefreshSession(ctx context.Context, sessionID, oldHash, newHash, token string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE mobile_auth_sessions a INNER JOIN t_session s ON s.CLIENT_SESSION_ID=a.session_id INNER JOIN m_mobile_users u ON u.MOBILE_UUID=s.USER_ID SET a.refresh_hash=?,s.TOKEN=? WHERE a.session_id=? AND a.refresh_hash=? AND a.expires_at>NOW() AND s.IS_ONLINE='YES' AND u.USER_STATUS='A'`, newHash, token, sessionID, oldHash)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

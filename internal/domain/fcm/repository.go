package fcm

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Exists(ctx context.Context, userID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_fcm_notifications WHERE USER_MOBILE_UUID=?", userID).Scan(&count)
	return count > 0, err
}

func (r *Repository) Insert(ctx context.Context, userID, token string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO m_fcm_notifications (USER_MOBILE_UUID, TOKEN_FCM, CREATE_DATE, UPDATE_DATE) VALUES (?, ?, NOW(), NOW())", userID, token)
	return err
}

func (r *Repository) Update(ctx context.Context, userID, token string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE m_fcm_notifications SET TOKEN_FCM=?, UPDATE_DATE=NOW() WHERE USER_MOBILE_UUID=?", token, userID)
	return err
}

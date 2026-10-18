package notification

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("notification not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) List(ctx context.Context, userID string) ([]Item, error) {
	rows, e := r.db.QueryContext(ctx, `SELECT id,READ_STATUS,NOTI_TYPE,MOBILE_USER_UUID,TITLE_MESSAGE,BODY_MESSAGE,SUCCESS,READ_DATE,RETRY,NOTI_DATE,VALUE_ID FROM m_send_notifications WHERE MOBILE_USER_UUID=? ORDER BY CREATE_DATE DESC`, userID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	result := make([]Item, 0)
	for rows.Next() {
		var i Item
		var readDate, notiDate sql.NullTime
		var retry, value sql.NullString
		if e := rows.Scan(&i.ID, &i.ReadStatus, &i.NotiType, &i.MobileUserUUID, &i.TitleMessage, &i.BodyMessage, &i.Success, &readDate, &retry, &notiDate, &value); e != nil {
			return nil, e
		}
		if readDate.Valid {
			i.ReadDate = readDate.Time.Format("2006-01-02 15:04:05")
		}
		if notiDate.Valid {
			i.NotiDate = notiDate.Time.Format("02/01/2006")
		}
		i.Retry, i.ValueID = retry.String, value.String
		result = append(result, i)
	}
	return result, rows.Err()
}
func (r *Repository) Count(ctx context.Context, userID string) (int, error) {
	var n int
	e := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_send_notifications WHERE READ_STATUS='NO' AND MOBILE_USER_UUID=?", userID).Scan(&n)
	return n, e
}
func (r *Repository) MarkID(ctx context.Context, userID, id string) (bool, error) {
	res, e := r.db.ExecContext(ctx, "UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE id=? AND MOBILE_USER_UUID=?", id, userID)
	if e != nil {
		return false, e
	}
	n, e := res.RowsAffected()
	return n > 0, e
}
func (r *Repository) MarkAll(ctx context.Context, userID string) error {
	_, e := r.db.ExecContext(ctx, "UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE MOBILE_USER_UUID=?", userID)
	return e
}
func (r *Repository) MarkValue(ctx context.Context, userID, valueID, notiType string) error {
	_, e := r.db.ExecContext(ctx, "UPDATE m_send_notifications SET READ_STATUS='YES',READ_DATE=NOW() WHERE MOBILE_USER_UUID=? AND VALUE_ID=? AND NOTI_TYPE=?", userID, valueID, notiType)
	return e
}

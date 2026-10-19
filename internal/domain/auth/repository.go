package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindUserByUsername(ctx context.Context, username string) (User, error) {
	return r.findUser(ctx, "USERNAME", username)
}

func (r *Repository) FindUserByUUID(ctx context.Context, uuid string) (User, error) {
	return r.findUser(ctx, "MOBILE_UUID", uuid)
}

func (r *Repository) findUser(ctx context.Context, column, value string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		"SELECT MOBILE_UUID, USERNAME, PASSWORD, EMAIL, USER_STATUS, FIRST_NAME, LAST_NAME FROM m_mobile_users WHERE "+column+" = ? LIMIT 1",
		value,
	).Scan(&user.MobileUUID, &user.Username, &user.Password, &user.Email, &user.Status, &user.FirstName, &user.LastName)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	return r.findUser(ctx, "EMAIL", email)
}

func (r *Repository) Register(ctx context.Context, request RegisterRequest, mobileUUID, passwordHash, displayName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var existing int
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM m_mobile_users WHERE EMAIL = ?", request.Email).Scan(&existing); err != nil {
		return err
	}
	if existing > 0 {
		return &BusinessError{Code: "MA00006", Detail: "email"}
	}
	var nextID sql.NullInt64
	if err := tx.QueryRowContext(ctx, "SELECT MAX(CAST(SMART_SEAMAN_ID AS DECIMAL) + 1) FROM m_mobile_users").Scan(&nextID); err != nil {
		return err
	}
	memberID := int64(1)
	if nextID.Valid {
		memberID = nextID.Int64
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO m_mobile_users
		(SMART_SEAMAN_ID, MOBILE_UUID, USERNAME, PASSWORD, FIRST_NAME, LAST_NAME, COMPANY_CODE, POSITION_CODE,
		 EMAIL, MOBILE_NUMBER, PROFILE_PICTURE, USER_STATUS, LAST_LOGON, CREATE_BY, CREATE_DATE, DATE_OF_BIRTH, DISPLAY_NAME, DISPLAY_TYPE)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'default', 'I', NULL, ?, NOW(), ?, ?, 'NAME')`,
		fmt.Sprintf("%05d", memberID), mobileUUID, request.Email, passwordHash, request.FirstName, request.LastName,
		request.CompanyCode, request.PositionCode, request.Email, request.MobileNumber, request.Email, nullable(request.DateOfBirth), displayName)
	if err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT DOCUMENT_CODE FROM m_documents WHERE DOCUMENT_DEFAULT_FLAG = 'Y' ORDER BY DOCUMENT_SEQ")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var documentCode string
		if err := rows.Scan(&documentCode); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO m_certificates
			(CERT_MOBILE_UUID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE, CERT_FILE, ORIGINAL_FILE_NAME,
			 CERT_STATUS, CREATE_DATE, CREATE_BY, UPDATE_DATE, UPDATE_BY)
			VALUES (?, ?, NULL, NULL, NULL, NULL, 'N', NOW(), ?, NOW(), ?)`, mobileUUID, documentCode, request.Email, request.Email)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (r *Repository) UpdatePassword(ctx context.Context, mobileUUID, password string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE m_mobile_users SET PASSWORD=? WHERE MOBILE_UUID=?", password, mobileUUID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=?", mobileUUID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM mobile_auth_tokens WHERE user_id=? AND purpose='reset'", mobileUUID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Message(ctx context.Context, code, language string) string {
	var english, thai sql.NullString
	err := r.db.QueryRowContext(ctx,
		"SELECT MESSAGE_DESCRIPTION_EN, MESSAGE_DESCRIPTION_TH FROM m_message_code WHERE MESSAGE_CODE = ? LIMIT 1", code,
	).Scan(&english, &thai)
	if err != nil {
		return "Message code : " + code + " is not matching in master."
	}
	if language == "EN" || language == "en" {
		return english.String
	}
	return thai.String
}

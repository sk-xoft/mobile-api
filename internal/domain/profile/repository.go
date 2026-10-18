package profile

import (
	"context"
	"database/sql"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) FindUser(ctx context.Context, uuid string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `SELECT MOBILE_UUID, SMART_SEAMAN_ID, USERNAME, EMAIL, FIRST_NAME, LAST_NAME,
		MOBILE_NUMBER, COMPANY_CODE, POSITION_CODE, PROFILE_PICTURE, USER_STATUS, DATE_OF_BIRTH
		FROM m_mobile_users WHERE MOBILE_UUID = ? LIMIT 1`, uuid).Scan(
		&user.UUID, &user.SmartID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.Mobile, &user.CompanyCode, &user.PositionCode, &user.ProfilePicture, &user.Status, &user.DateOfBirth)
	return user, err
}

func (r *Repository) Descriptions(ctx context.Context, companyCode, positionCode, language string) (string, string) {
	var company, position sql.NullString
	companyColumn := "COMPANY_NAME_TH"
	positionColumn := "POSITION_NAME_TH"
	if language == "EN" || language == "en" {
		companyColumn = "COMPANY_NAME_EN"
		positionColumn = "POSITION_NAME_EN"
	}
	_ = r.db.QueryRowContext(ctx, "SELECT "+companyColumn+" FROM m_companys WHERE COMPANY_CODE = ? LIMIT 1", companyCode).Scan(&company)
	_ = r.db.QueryRowContext(ctx, "SELECT "+positionColumn+" FROM m_positions WHERE POSITION_CODE = ? LIMIT 1", positionCode).Scan(&position)
	return company.String, position.String
}

func (r *Repository) Update(ctx context.Context, user User, request UpdateRequest, imageName string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentMobile string
	if err := tx.QueryRowContext(ctx, "SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID = ? FOR UPDATE", user.UUID).Scan(&currentMobile); err != nil {
		return err
	}
	if currentMobile != request.MobileNumber {
		if _, err := tx.ExecContext(ctx, `INSERT INTO m_mobile_number_history (id, mobile_user_uuid, old_mobile_number, new_mobile_number, changed_by) VALUES (UUID(), ?, ?, ?, ?)`, user.UUID, currentMobile, request.MobileNumber, user.Username); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE m_mobile_users SET FIRST_NAME=?, LAST_NAME=?, COMPANY_CODE=?, POSITION_CODE=?, MOBILE_NUMBER=?, DATE_OF_BIRTH=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?`, request.FirstName, request.LastName, request.CompanyCode, request.PositionCode, request.MobileNumber, nullable(request.DateOfBirth), user.Username, user.UUID)
	if err != nil {
		return err
	}
	if imageName != "" {
		if _, err := tx.ExecContext(ctx, "UPDATE m_mobile_users SET PROFILE_PICTURE=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?", imageName, user.Username, user.UUID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) UpdateStatus(ctx context.Context, email, status string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, `SELECT MOBILE_UUID, SMART_SEAMAN_ID, USERNAME, EMAIL, FIRST_NAME, LAST_NAME, MOBILE_NUMBER, COMPANY_CODE, POSITION_CODE, PROFILE_PICTURE, USER_STATUS, DATE_OF_BIRTH FROM m_mobile_users WHERE EMAIL=? LIMIT 1`, email).Scan(&user.UUID, &user.SmartID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.Mobile, &user.CompanyCode, &user.PositionCode, &user.ProfilePicture, &user.Status, &user.DateOfBirth)
	if err != nil {
		return User{}, err
	}
	_, err = r.db.ExecContext(ctx, "UPDATE m_mobile_users SET USER_STATUS=?, UPDATE_DATE=NOW(), UPDATE_BY=? WHERE MOBILE_UUID=?", status, user.Username, user.UUID)
	return user, err
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

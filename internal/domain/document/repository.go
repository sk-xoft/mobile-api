package document

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

type User struct{ UUID, Username, Email string }
type Certificate struct{ ID, DocumentCode, StartDate, EndDate, FileName, OriginalFileName, Status string }

func (r *Repository) FindUser(ctx context.Context, userID string) (User, error) {
	var user User
	err := r.db.QueryRowContext(ctx, "SELECT MOBILE_UUID, USERNAME, EMAIL FROM m_mobile_users WHERE MOBILE_UUID = ? LIMIT 1", userID).Scan(&user.UUID, &user.Username, &user.Email)
	return user, err
}

func (r *Repository) Certificate(ctx context.Context, userID, documentCode string) (Certificate, error) {
	var item Certificate
	err := r.db.QueryRowContext(ctx, `SELECT CERT_ID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE,
		CERT_FILE, ORIGINAL_FILE_NAME, CERT_STATUS FROM m_certificates
		WHERE CERT_MOBILE_UUID = ? AND CERT_DOCUMENT_CODE = ? LIMIT 1`, userID, documentCode).Scan(
		&item.ID, &item.DocumentCode, &item.StartDate, &item.EndDate, &item.FileName, &item.OriginalFileName, &item.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Certificate{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateCertificate(ctx context.Context, userID, username string, request CreateRequest, fileName string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO m_certificates
		(CERT_MOBILE_UUID, CERT_DOCUMENT_CODE, CERT_START_DATE, CERT_END_DATE, CERT_FILE, ORIGINAL_FILE_NAME,
		 CERT_STATUS, CREATE_DATE, CREATE_BY, UPDATE_DATE, UPDATE_BY)
		VALUES (?, ?, ?, ?, ?, ?, 'A', NOW(), ?, NOW(), ?)`, userID, request.DocumentCode,
		request.CertStartDate, nullableEnd(request.CertEndDate), fileName, request.FileCertName, username, username)
	return err
}

func (r *Repository) UpdateCertificate(ctx context.Context, userID, username string, request UpdateRequest, fileName string) error {
	if fileName == "" {
		_, err := r.db.ExecContext(ctx, `UPDATE m_certificates SET CERT_START_DATE=?, CERT_END_DATE=?, CERT_STATUS='A', UPDATE_DATE=NOW(), UPDATE_BY=?
			WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?`, request.CertStartDate, nullableEnd(request.CertEndDate), username, userID, request.DocumentCode)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE m_certificates SET CERT_START_DATE=?, CERT_END_DATE=?, CERT_FILE=?, ORIGINAL_FILE_NAME=?, CERT_STATUS='A', UPDATE_DATE=NOW(), UPDATE_BY=?
		WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?`, request.CertStartDate, nullableEnd(request.CertEndDate), fileName, request.FileCertName, username, userID, request.DocumentCode)
	return err
}

func (r *Repository) DeleteCertificate(ctx context.Context, userID, documentCode string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM m_certificates WHERE CERT_MOBILE_UUID=? AND CERT_DOCUMENT_CODE=?", userID, documentCode)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Page(ctx context.Context, userID, documentType string, offset int) ([]Item, int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,
		md.DOCUMENT_TYPE, md.DOCUMENT_DEFAULT_FLAG, md.DOCUMENT_RENEWAL_FLAG, md.DOCUMENT_SEQ, md.DOCUMENT_MOBILE_FLAG, md.DOCUMENT_COURSE_CODE,
		mc.CERT_START_DATE, mc.CERT_END_DATE, mc.ORIGINAL_FILE_NAME
		FROM m_documents md LEFT JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND md.DOCUMENT_TYPE=? AND mc.CERT_MOBILE_UUID IS NOT NULL ORDER BY md.DOCUMENT_SEQ LIMIT 10 OFFSET ?`, userID, documentType, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := scanItem(rows, &item, false); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM m_documents md
		LEFT JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND md.DOCUMENT_TYPE=? AND mc.CERT_MOBILE_UUID IS NOT NULL`, userID, documentType).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) CloseToExpiration(ctx context.Context, userID string, offset int) ([]Item, int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,
		md.DOCUMENT_TYPE, md.DOCUMENT_DEFAULT_FLAG, md.DOCUMENT_RENEWAL_FLAG, md.DOCUMENT_SEQ, md.DOCUMENT_MOBILE_FLAG, md.DOCUMENT_COURSE_CODE,
		mc.CERT_START_DATE, mc.CERT_END_DATE, mc.ORIGINAL_FILE_NAME,
		CASE WHEN EXISTS (SELECT 1 FROM m_document_request dr INNER JOIN m_document_status ds ON ds.id=dr.document_status_id
			WHERE dr.mobile_user_uuid=? AND dr.document_code=md.DOCUMENT_CODE AND dr.is_active='YES'
			AND ds.document_status_code NOT IN ('DELIVERED','CANCELLED')) THEN 'Y' ELSE 'N' END
		FROM m_documents md INNER JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=?
		WHERE md.DOCUMENT_STATUS='A' AND mc.CERT_END_DATE <= NOW() + INTERVAL 18 MONTH
		ORDER BY mc.CERT_END_DATE, md.DOCUMENT_SEQ LIMIT 10 OFFSET ?`, userID, userID, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := scanItem(rows, &item, true); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM m_documents md INNER JOIN m_certificates mc ON mc.CERT_DOCUMENT_CODE=md.DOCUMENT_CODE AND mc.CERT_MOBILE_UUID=? WHERE md.DOCUMENT_STATUS='A' AND mc.CERT_END_DATE <= NOW() + INTERVAL 18 MONTH`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return items, total, rows.Err()
}

func scanItem(rows *sql.Rows, item *Item, includeRenewalRequestFlag bool) error {
	var (
		documentID, documentSeq      sql.NullInt64
		documentCode, nameTh, nameEn sql.NullString
		description, documentType    sql.NullString
		defaultFlag, renewalFlag     sql.NullString
		mobileFlag, courseCode       sql.NullString
		start, end, filename         sql.NullString
		renewalRequestFlag           sql.NullString
	)

	destinations := []any{
		&documentID, &documentCode, &nameTh, &nameEn, &description,
		&documentType, &defaultFlag, &renewalFlag, &documentSeq, &mobileFlag, &courseCode,
		&start, &end, &filename,
	}
	if includeRenewalRequestFlag {
		destinations = append(destinations, &renewalRequestFlag)
	}
	if err := rows.Scan(destinations...); err != nil {
		return err
	}

	item.DocumentID = nullableIntString(documentID)
	item.DocumentCode = documentCode.String
	item.DocumentNameTh = nameTh.String
	item.DocumentNameEn = nameEn.String
	item.DocumentDescription = description.String
	item.DocumentType = documentType.String
	item.DocumentDefaultFlag = defaultFlag.String
	item.DocumentRenewalFlag = renewalFlag.String
	item.DocumentSeq = nullableIntString(documentSeq)
	item.DocumentMobileFlag = mobileFlag.String
	item.DocumentCourseCode = courseCode.String
	item.CertStartDate = start.String
	item.CertEndDate = end.String
	item.CertFileName = filename.String
	item.DocumentRenewalRequestFlag = renewalRequestFlag.String
	return nil
}

func nullableIntString(value sql.NullInt64) string {
	if !value.Valid {
		return ""
	}
	return strconv.FormatInt(value.Int64, 10)
}

var ErrNotFound = errors.New("not found")

func nullableEnd(value string) any {
	if value == "" || value == "9999-99-99" {
		return nil
	}
	return value
}

package master

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Companies(ctx context.Context) ([]Company, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT COMPANY_CODE, COMPANY_NAME_EN, COMPANY_NAME_TH
		FROM m_companys WHERE COMPANY_TYPE = 'Shipping Company' AND COMPANY_MOBILE_FLAG = 'Y'
		AND COMPANY_STATUS = 'A' ORDER BY COMPANY_TYPE, COMPANY_SEQ`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Company, 0)
	for rows.Next() {
		var code, english, thai sql.NullString
		if err := rows.Scan(&code, &english, &thai); err != nil {
			return nil, err
		}
		result = append(result, Company{Code: code.String, Name: thai.String + "\x00" + english.String})
	}
	return result, rows.Err()
}

func (r *Repository) Positions(ctx context.Context) ([]Position, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT POSITION_CODE, POSITION_NAME_EN, POSITION_NAME_TH FROM m_positions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Position, 0)
	for rows.Next() {
		var code, english, thai sql.NullString
		if err := rows.Scan(&code, &english, &thai); err != nil {
			return nil, err
		}
		result = append(result, Position{Code: code.String, Name: thai.String + "\x00" + english.String})
	}
	return result, rows.Err()
}

func (r *Repository) Documents(ctx context.Context, documentType string) ([]Document, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT DOCUMENT_CODE, DOCUMENT_NAME_EN, DOCUMENT_NAME_TH
		FROM m_documents WHERE DOCUMENT_TYPE = ? ORDER BY DOCUMENT_SEQ`, documentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Document, 0)
	for rows.Next() {
		var code, english, thai sql.NullString
		if err := rows.Scan(&code, &english, &thai); err != nil {
			return nil, err
		}
		result = append(result, Document{Code: code.String, Name: thai.String + "\x00" + english.String, NameTH: thai.String})
	}
	return result, rows.Err()
}

func (r *Repository) Courses(ctx context.Context) ([]Course, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT COURSE_CODE, COURSE_NAME_EN, COURSE_NAME_TH
		FROM m_course_name WHERE COURSE_STATUS = 'A' ORDER BY COURSE_SEQ`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Course, 0)
	for rows.Next() {
		var code, english, thai sql.NullString
		if err := rows.Scan(&code, &english, &thai); err != nil {
			return nil, err
		}
		result = append(result, Course{Code: code.String, NameEN: english.String, NameTH: thai.String})
	}
	return result, rows.Err()
}

func (r *Repository) Provinces(ctx context.Context) ([]Address, error) {
	return r.addresses(ctx, `SELECT code, name_in_thai, name_in_english, NULL FROM provinces ORDER BY name_in_thai`)
}

func (r *Repository) Districts(ctx context.Context, provinceCode int) ([]Address, error) {
	return r.addresses(ctx, `SELECT d.code, d.name_in_thai, d.name_in_english, NULL
		FROM districts d INNER JOIN provinces p ON p.id = d.province_id
		WHERE p.code = ? ORDER BY d.name_in_thai`, provinceCode)
}

func (r *Repository) Subdistricts(ctx context.Context, districtCode int) ([]Address, error) {
	return r.addresses(ctx, `SELECT s.code, s.name_in_thai, s.name_in_english, s.zip_code
		FROM subdistricts s INNER JOIN districts d ON d.id = s.district_id
		WHERE d.code = ? ORDER BY s.name_in_thai`, districtCode)
}

func (r *Repository) addresses(ctx context.Context, query string, args ...any) ([]Address, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Address, 0)
	for rows.Next() {
		var code, thai, english sql.NullString
		var postal sql.NullInt64
		if err := rows.Scan(&code, &thai, &english, &postal); err != nil {
			return nil, err
		}
		address := Address{Code: code.String, Name: thai.String, NameTH: thai.String, NameEN: english.String}
		if postal.Valid {
			address.PostalCode = fmt.Sprint(postal.Int64)
		}
		result = append(result, address)
	}
	return result, rows.Err()
}

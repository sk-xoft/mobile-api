package catalog

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("catalog item not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Banners(ctx context.Context) ([]Banner, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT BANNER_ID, BANNER_FILE_NAME FROM m_banners ORDER BY CREATE_DATE DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Banner, 0)
	for rows.Next() {
		var item Banner
		var id, fileName sql.NullString
		if err := rows.Scan(&id, &fileName); err != nil {
			return nil, err
		}
		item.BannerID, item.BannerFileName = id.String, fileName.String
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Banner(ctx context.Context, id string) (Banner, error) {
	var item Banner
	var bannerID, fileName sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT BANNER_ID,BANNER_FILE_NAME FROM m_banners WHERE BANNER_ID=?", id).Scan(&bannerID, &fileName)
	if errors.Is(err, sql.ErrNoRows) {
		return Banner{}, ErrNotFound
	}
	item.BannerID, item.BannerFileName = bannerID.String, fileName.String
	return item, err
}
func (r *Repository) UserSmartID(ctx context.Context, userID string) (string, error) {
	var value sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT SMART_SEAMAN_ID FROM m_mobile_users WHERE MOBILE_UUID=? LIMIT 1", userID).Scan(&value)
	return value.String, err
}
func (r *Repository) Vouchers(ctx context.Context, smartID string) ([]Voucher, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mv.VOUCHER_ID,mv.VOUCHER_TITLE,mv.VOUCHER_DETAILS,mv.VOUCHER_START_DATE,mv.VOUCHER_END_DATE FROM m_vouchers mv INNER JOIN m_voucher_details mvd ON mv.VOUCHER_ID=mvd.VOUCHER_ID WHERE mv.VOUCHER_STATUS='A' AND mv.VOUCHER_TYPE='PERSONAL' AND mvd.SMART_SEAMAN_ID=? UNION ALL SELECT VOUCHER_ID,VOUCHER_TITLE,VOUCHER_DETAILS,VOUCHER_START_DATE,VOUCHER_END_DATE FROM m_vouchers WHERE VOUCHER_STATUS='A' AND VOUCHER_TYPE='GLOBAL' ORDER BY VOUCHER_ID`, smartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Voucher, 0)
	for rows.Next() {
		var item Voucher
		if err := rows.Scan(&item.VoucherID, &item.Title, &item.Details, &item.StartDate, &item.EndDate); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Voucher(ctx context.Context, id string) (Voucher, string, string, error) {
	var item Voucher
	var picture, qr sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT VOUCHER_ID,VOUCHER_TITLE,VOUCHER_DETAILS,VOUCHER_START_DATE,VOUCHER_END_DATE,VOUCHER_PICTURE,VOUCHER_QRCODE FROM m_vouchers WHERE VOUCHER_STATUS='A' AND VOUCHER_ID=?", id).Scan(&item.VoucherID, &item.Title, &item.Details, &item.StartDate, &item.EndDate, &picture, &qr)
	if errors.Is(err, sql.ErrNoRows) {
		return Voucher{}, "", "", ErrNotFound
	}
	return item, picture.String, qr.String, err
}

const schoolSelect = `SELECT COMPANY_CODE, COMPANY_NAME_EN, COMPANY_NAME_TH, COMPANY_COLOUR, COURSE_COLOUR, COURSE_ID, COURSE_CODE, COURSE_NAME_EN, COURSE_NAME_TH, COURSE_TYPE, COURSE_ONLINE_DATE, COURSE_ONSITE_DATE, COURSE_TOTAL_DAYS, COURSE_PRICE, COMPANY_LINE, COMPANY_FACEBOOK, COMPANY_PHONE1, COURSE_START_DATE, COURSE_END_DATE, DATE_FOR_CHECK FROM list_course_school_training `

func (r *Repository) Schools(ctx context.Context, courseCode string, detail bool, companyCode string) ([]School, error) {
	query := schoolSelect + "WHERE COURSE_CODE=? "
	args := []any{courseCode}
	if detail {
		query = schoolSelect + "WHERE COMPANY_CODE=? AND COURSE_CODE=? "
		args = []any{companyCode, courseCode}
	} else {
		query += "AND DATE_FORMAT(NOW(), '%Y%m%d') <= DATE_FOR_CHECK "
	}
	query += "ORDER BY COMPANY_CODE, COURSE_CODE, DATE_FOR_CHECK"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]School, 0)
	for rows.Next() {
		var item School
		if err := rows.Scan(&item.CompanyCode, &item.CompanyNameEn, &item.CompanyNameTh, &item.CompanyColour, &item.CourseColour, &item.CourseID, &item.CourseCode, &item.CourseNameEn, &item.CourseNameTh, &item.CourseType, &item.CourseOnlineDate, &item.CourseOnsiteDate, &item.CourseTotalDays, &item.CoursePrice, &item.CompanyLine, &item.CompanyFacebook, &item.CompanyPhone1, &item.CourseStartDate, &item.CourseEndDate, &item.DateForCheck); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

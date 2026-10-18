package delivery

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("delivery address not found")

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) UserMobile(ctx context.Context, userID string) (string, error) {
	var mobile string
	err := r.db.QueryRowContext(ctx, "SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=? LIMIT 1", userID).Scan(&mobile)
	return mobile, err
}

func (r *Repository) ValidAddress(ctx context.Context, province, district, subDistrict, postalCode string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provinces p
		INNER JOIN districts d ON d.province_code=p.code
		INNER JOIN subdistricts s ON s.district_code=d.code
		WHERE p.code=? AND d.code=? AND s.code=? AND s.zip_code=?`, province, district, subDistrict, postalCode).Scan(&count)
	return count == 1, err
}

func (r *Repository) Default(ctx context.Context, userID string) (Address, error) {
	var address Address
	err := r.db.QueryRowContext(ctx, `SELECT da.id, da.mobile_user_uuid, da.first_name, da.last_name, da.address_line,
		da.province, da.district, da.sub_district, da.postal_code, mu.MOBILE_NUMBER,
		CONCAT_WS(' ', da.address_line, CONCAT('ตำบล', sd.name_in_thai), CONCAT('อำเภอ', d.name_in_thai), CONCAT('จังหวัด', p.name_in_thai), da.postal_code), da.is_default
		FROM m_delivery_address da INNER JOIN m_mobile_users mu ON mu.MOBILE_UUID=da.mobile_user_uuid
		LEFT JOIN provinces p ON p.code=da.province LEFT JOIN districts d ON d.code=da.district LEFT JOIN subdistricts sd ON sd.code=da.sub_district
		WHERE da.mobile_user_uuid=? AND da.is_default=1 AND da.is_active='YES' ORDER BY da.updated_at DESC, da.id LIMIT 1`, userID).Scan(
		&address.ID, &address.UserID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province,
		&address.District, &address.SubDistrict, &address.PostalCode, &address.Mobile, &address.Description, &address.IsDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return Address{}, ErrNotFound
	}
	return address, err
}

func (r *Repository) Find(ctx context.Context, userID, addressID string) (Address, error) {
	var address Address
	err := r.db.QueryRowContext(ctx, `SELECT id, mobile_user_uuid, first_name, last_name, address_line, province, district,
		sub_district, postal_code, is_default FROM m_delivery_address WHERE id=? AND mobile_user_uuid=? AND is_active='YES'`, addressID, userID).Scan(
		&address.ID, &address.UserID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province,
		&address.District, &address.SubDistrict, &address.PostalCode, &address.IsDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return Address{}, ErrNotFound
	}
	return address, err
}

func (r *Repository) Count(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_delivery_address WHERE mobile_user_uuid=? AND is_active='YES'", userID).Scan(&count)
	return count, err
}

func (r *Repository) ClearDefault(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE m_delivery_address SET is_default=0, updated_at=NOW() WHERE mobile_user_uuid=? AND is_active='YES' AND is_default=1", userID)
	return err
}

func (r *Repository) Create(ctx context.Context, address Address) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO m_delivery_address
		(id, mobile_user_uuid, first_name, last_name, address_line, province, district, sub_district, postal_code, is_default, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'YES')`, address.ID, address.UserID, address.FirstName, address.LastName,
		address.AddressLine, address.Province, address.District, address.SubDistrict, address.PostalCode, address.IsDefault)
	return err
}

func (r *Repository) Update(ctx context.Context, address Address) error {
	result, err := r.db.ExecContext(ctx, `UPDATE m_delivery_address SET first_name=?, last_name=?, address_line=?, province=?, district=?,
		sub_district=?, postal_code=?, is_default=?, updated_at=NOW() WHERE id=? AND mobile_user_uuid=? AND is_active='YES'`,
		address.FirstName, address.LastName, address.AddressLine, address.Province, address.District, address.SubDistrict,
		address.PostalCode, address.IsDefault, address.ID, address.UserID)
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

func (r *Repository) RenewalRequest(ctx context.Context, userID, requestNo string) (string, error) {
	var requestID string
	err := r.db.QueryRowContext(ctx, "SELECT id FROM m_document_request WHERE request_no=? AND mobile_user_uuid=? AND is_active='YES' LIMIT 1", requestNo, userID).Scan(&requestID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return requestID, err
}

func (r *Repository) SnapshotExists(ctx context.Context, requestID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_document_request_delivery_address WHERE request_id=? AND mobile_user_uuid=?", requestID, userID).Scan(&count)
	return count > 0, err
}

func (r *Repository) InsertSnapshot(ctx context.Context, requestID, userID, mobile string, address Address) (string, error) {
	id := uuid.NewString()
	_, err := r.db.ExecContext(ctx, `INSERT INTO m_document_request_delivery_address
		(id, request_id, source_delivery_address_id, mobile_user_uuid, first_name, last_name, address_line,
		province, district, sub_district, postal_code, mobile_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, requestID, address.ID, userID, address.FirstName,
		address.LastName, address.AddressLine, address.Province, address.District, address.SubDistrict, address.PostalCode, mobile)
	return id, err
}

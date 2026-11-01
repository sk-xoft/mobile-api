package delivery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("delivery address not found")

type queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}
type Repository struct {
	db   queryer
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db, pool: db} }

func (r *Repository) LockUser(ctx context.Context, userID string) (string, error) {
	var mobile string
	err := r.db.QueryRowContext(ctx, "SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=? FOR UPDATE", userID).Scan(&mobile)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return mobile, err
}

func (r *Repository) ValidAddress(ctx context.Context, province, district, subDistrict, postalCode string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provinces p
		INNER JOIN districts d ON d.province_id=p.id
		INNER JOIN subdistricts s ON s.district_id=d.id
		WHERE p.code=? AND d.code=? AND s.code=? AND CAST(s.zip_code AS CHAR)=?`, province, district, subDistrict, postalCode).Scan(&count)
	return count > 0, err
}

func (r *Repository) Default(ctx context.Context, userID string) (Address, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT da.id, da.mobile_user_uuid, da.first_name, da.last_name, da.address_line,
 da.province, da.district, da.sub_district, da.postal_code, mu.MOBILE_NUMBER,
 CONCAT_WS(' ', da.address_line, CONCAT('ตำบล', sd.name_in_thai), CONCAT('อำเภอ', d.name_in_thai), CONCAT('จังหวัด', p.name_in_thai), da.postal_code), da.is_default
 FROM m_delivery_address da INNER JOIN m_mobile_users mu ON mu.MOBILE_UUID=da.mobile_user_uuid
 LEFT JOIN provinces p ON p.code=da.province LEFT JOIN districts d ON d.code=da.district LEFT JOIN subdistricts sd ON sd.code=da.sub_district
 WHERE da.mobile_user_uuid=? AND da.is_default=1 AND da.is_active='YES' ORDER BY da.updated_at DESC, da.id`, userID)
	if err != nil {
		return Address{}, err
	}
	defer rows.Close()
	var address Address
	count := 0
	for rows.Next() {
		count++
		if err = rows.Scan(&address.ID, &address.UserID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province,
			&address.District, &address.SubDistrict, &address.PostalCode, &address.Mobile, &address.Description, &address.IsDefault); err != nil {
			return Address{}, err
		}
	}
	if err = rows.Err(); err != nil {
		return Address{}, err
	}
	if count == 0 {
		return Address{}, ErrNotFound
	}
	if count != 1 {
		return Address{}, fmt.Errorf("duplicate active default delivery address")
	}
	return address, nil
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
	_, err = result.RowsAffected()
	return err
}

func (r *Repository) RenewalRequest(ctx context.Context, userID, requestNo string) (string, error) {
	var requestID string
	key := "request_no"
	value := strings.ToUpper(strings.TrimSpace(requestNo))
	if id, err := uuid.Parse(requestNo); err == nil {
		key, value = "id", id.String()
	}
	err := r.db.QueryRowContext(ctx, "SELECT id FROM m_document_request WHERE "+key+"=? AND mobile_user_uuid=? AND is_active='YES' FOR UPDATE", value, userID).Scan(&requestID)
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
	var sourceID any
	if address.ID != "" {
		sourceID = address.ID
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO m_document_request_delivery_address
		(id, request_id, source_delivery_address_id, mobile_user_uuid, first_name, last_name, address_line,
		province, district, sub_district, postal_code, mobile_number)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, requestID, sourceID, userID, address.FirstName,
		address.LastName, address.AddressLine, address.Province, address.District, address.SubDistrict, address.PostalCode, mobile)
	return id, err
}

// Transact exposes only a transaction-bound repository to the service callback.
func (r *Repository) Transact(ctx context.Context, action func(repositoryPort) error) error {
	if r.pool == nil {
		return fmt.Errorf("nested delivery transaction")
	}
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = action(&Repository{db: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) LockAddresses(ctx context.Context, userID string) error {
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM m_delivery_address WHERE mobile_user_uuid=? AND is_active='YES' FOR UPDATE", userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *Repository) FindSnapshot(ctx context.Context, userID, addressID string) (Address, error) {
	var address Address
	err := r.db.QueryRowContext(ctx, `SELECT id,mobile_user_uuid,first_name,last_name,address_line,province,district,sub_district,postal_code,mobile_number FROM m_document_request_delivery_address WHERE id=? AND mobile_user_uuid=? FOR UPDATE`, addressID, userID).Scan(
		&address.ID, &address.UserID, &address.FirstName, &address.LastName, &address.AddressLine, &address.Province, &address.District, &address.SubDistrict, &address.PostalCode, &address.Mobile)
	if errors.Is(err, sql.ErrNoRows) {
		return Address{}, ErrNotFound
	}
	return address, err
}

func (r *Repository) UpdateSnapshot(ctx context.Context, address Address) error {
	// The owned row was locked above. MySQL may report zero changes for an identical update.
	_, err := r.db.ExecContext(ctx, `UPDATE m_document_request_delivery_address SET first_name=?,last_name=?,address_line=?,province=?,district=?,sub_district=?,postal_code=? WHERE id=? AND mobile_user_uuid=?`, address.FirstName, address.LastName, address.AddressLine, address.Province, address.District, address.SubDistrict, address.PostalCode, address.ID, address.UserID)
	return err
}

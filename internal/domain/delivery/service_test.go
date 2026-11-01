package delivery

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const owner = "user-1"
const addressID = "11111111-1111-4111-8111-111111111111"

func deliveryTest(t *testing.T) (*Service, sqlmock.Sqlmock, context.Context) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	// Every multi-query regression uses a single connection: reads must use tx,
	// and lock result sets must be consumed/closed before the next statement.
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(func() {
		cancel()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		db.Close()
	})
	return NewService(NewRepository(db)), mock, ctx
}
func request(defaultValue bool) Request {
	return Request{FirstName: " First ", LastName: "Last", AddressLine: "Line", Province: "10", District: "1001", SubDistrict: "100101", PostalCode: "10100", IsDefault: &defaultValue}
}
func validMaster(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM provinces p\s+INNER JOIN districts d ON d.province_id=p.id\s+INNER JOIN subdistricts s ON s.district_id=d.id\s+WHERE p.code=\? AND d.code=\? AND s.code=\? AND CAST\(s.zip_code AS CHAR\)=\?`).WithArgs("10", "1001", "100101", "10100").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
}
func locks(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT MOBILE_NUMBER FROM m_mobile_users WHERE MOBILE_UUID=\? FOR UPDATE`).WithArgs(owner).WillReturnRows(sqlmock.NewRows([]string{"mobile"}).AddRow("0800000000"))
	mock.ExpectQuery(`SELECT id FROM m_delivery_address WHERE mobile_user_uuid=\? AND is_active='YES' FOR UPDATE`).WithArgs(owner).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(addressID)).RowsWillBeClosed()
}
func count(mock sqlmock.Sqlmock, n int) {
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM m_delivery_address WHERE mobile_user_uuid=\? AND is_active='YES'`).WithArgs(owner).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
}
func clear(mock sqlmock.Sqlmock) {
	mock.ExpectExec(`UPDATE m_delivery_address SET is_default=0.*WHERE mobile_user_uuid=\?`).WithArgs(owner).WillReturnResult(sqlmock.NewResult(0, 1))
}
func activeRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user", "first", "last", "line", "province", "district", "sub", "postal", "default"})
}
func snapshotRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "user", "first", "last", "line", "province", "district", "sub", "postal", "mobile"})
}
func findActive(mock sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(`SELECT id, mobile_user_uuid,.*FROM m_delivery_address WHERE id=\? AND mobile_user_uuid=\? AND is_active='YES'`).WithArgs(addressID, owner)
}

func TestCreateDefaultRollbackOnInsertError(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	failure := errors.New("insert failed")
	validMaster(m)
	m.ExpectBegin()
	locks(m)
	count(m, 1)
	clear(m)
	m.ExpectExec(`INSERT INTO m_delivery_address`).WithArgs(sqlmock.AnyArg(), owner, "First", "Last", "Line", "10", "1001", "100101", "10100", true).WillReturnError(failure)
	m.ExpectRollback()
	result, err := s.Create(ctx, owner, request(true))
	if !errors.Is(err, failure) || result.ID != "" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
func TestCreateFirstAddressBecomesDefault(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	validMaster(m)
	m.ExpectBegin()
	locks(m)
	count(m, 0)
	clear(m)
	m.ExpectExec(`INSERT INTO m_delivery_address`).WithArgs(sqlmock.AnyArg(), owner, "First", "Last", "Line", "10", "1001", "100101", "10100", true).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	result, err := s.Create(ctx, owner, request(false))
	if err != nil || !result.IsDefault || result.FirstName != "First" || result.Mobile != "0800000000" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
func TestUpdateDefaultRollbackOnWriteError(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	failure := errors.New("update failed")
	validMaster(m)
	m.ExpectBegin()
	locks(m)
	findActive(m).WillReturnRows(activeRows().AddRow(addressID, owner, "Old", "Last", "Line", "10", "1001", "100101", "10100", true))
	clear(m)
	m.ExpectExec(`UPDATE m_delivery_address SET first_name=.*WHERE id=\? AND mobile_user_uuid=\? AND is_active='YES'`).WithArgs("First", "Last", "Line", "10", "1001", "100101", "10100", true, addressID, owner).WillReturnError(failure)
	m.ExpectRollback()
	_, err := s.Update(ctx, owner, addressID, request(true))
	if !errors.Is(err, failure) {
		t.Fatal(err)
	}
}
func TestUpdateSnapshotPreservesMobileAndDoesNotClearDefault(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	validMaster(m)
	m.ExpectBegin()
	locks(m)
	findActive(m).WillReturnRows(activeRows())
	m.ExpectQuery(`SELECT id,mobile_user_uuid,.*FROM m_document_request_delivery_address WHERE id=\? AND mobile_user_uuid=\? FOR UPDATE`).WithArgs(addressID, owner).WillReturnRows(snapshotRows().AddRow(addressID, owner, "Old", "Last", "Line", "10", "1001", "100101", "10100", "0899999999"))
	m.ExpectExec(`UPDATE m_document_request_delivery_address SET first_name=.*WHERE id=\? AND mobile_user_uuid=\?`).WithArgs("First", "Last", "Line", "10", "1001", "100101", "10100", addressID, owner).WillReturnResult(sqlmock.NewResult(0, 0))
	m.ExpectCommit()
	result, err := s.Update(ctx, owner, addressID, request(true))
	if err != nil || result.Mobile != "0899999999" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
func TestSnapshotOwnershipAndLookupFailures(t *testing.T) {
	for _, mode := range []string{"foreign-snapshot", "database-error"} {
		t.Run(mode, func(t *testing.T) {
			s, m, ctx := deliveryTest(t)
			validMaster(m)
			m.ExpectBegin()
			locks(m)
			var expected error
			if mode == "database-error" {
				expected = errors.New("lookup failed")
				findActive(m).WillReturnError(expected)
			} else {
				expected = ErrNotFound
				findActive(m).WillReturnRows(activeRows())
				m.ExpectQuery(`SELECT id,mobile_user_uuid,.*FROM m_document_request_delivery_address WHERE id=\? AND mobile_user_uuid=\? FOR UPDATE`).WithArgs(addressID, owner).WillReturnRows(snapshotRows())
			}
			m.ExpectRollback()
			_, err := s.Update(ctx, owner, addressID, request(false))
			if !errors.Is(err, expected) {
				t.Fatal(err)
			}
		})
	}
}
func renewalLock(m sqlmock.Sqlmock, key, value string) {
	m.ExpectQuery(`SELECT id FROM m_document_request WHERE `+key+`=\? AND mobile_user_uuid=\? AND is_active='YES' FOR UPDATE`).WithArgs(value, owner).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("request-id"))
}
func snapshotCount(m sqlmock.Sqlmock, n int) {
	m.ExpectQuery(`SELECT COUNT\(\*\) FROM m_document_request_delivery_address WHERE request_id=\? AND mobile_user_uuid=\?`).WithArgs("request-id", owner).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(n))
}
func TestRenewalSnapshotUUIDAndNullableSource(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	validMaster(m)
	m.ExpectBegin()
	renewalLock(m, "id", addressID)
	snapshotCount(m, 0)
	locks(m)
	count(m, 1)
	m.ExpectExec(`INSERT INTO m_document_request_delivery_address`).WithArgs(sqlmock.AnyArg(), "request-id", nil, owner, "First", "Last", "Line", "10", "1001", "100101", "10100", "0800000000").WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectCommit()
	result, err := s.CreateForRenewal(ctx, owner, addressID, request(true))
	if err != nil || result.ID == "" || result.IsDefault {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
func TestRenewalFirstAddressRollsBackWithSnapshot(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	failure := errors.New("snapshot failed")
	validMaster(m)
	m.ExpectBegin()
	renewalLock(m, "request_no", "261000024")
	snapshotCount(m, 0)
	locks(m)
	count(m, 0)
	clear(m)
	m.ExpectExec(`INSERT INTO m_delivery_address`).WithArgs(sqlmock.AnyArg(), owner, "First", "Last", "Line", "10", "1001", "100101", "10100", true).WillReturnResult(sqlmock.NewResult(0, 1))
	m.ExpectExec(`INSERT INTO m_document_request_delivery_address`).WithArgs(sqlmock.AnyArg(), "request-id", sqlmock.AnyArg(), owner, "First", "Last", "Line", "10", "1001", "100101", "10100", "0800000000").WillReturnError(failure)
	m.ExpectRollback()
	_, err := s.CreateForRenewal(ctx, owner, "261000024", request(false))
	if !errors.Is(err, failure) {
		t.Fatal(err)
	}
}
func TestRenewalDuplicateSnapshotHasNoWrites(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	validMaster(m)
	m.ExpectBegin()
	renewalLock(m, "request_no", "261000024")
	snapshotCount(m, 1)
	m.ExpectRollback()
	_, err := s.CreateForRenewal(ctx, owner, "261000024", request(false))
	var business *BusinessError
	if !errors.As(err, &business) || business.Code != "MA00017" {
		t.Fatal(err)
	}
}
func TestValidationRejectsOversizedNamesAndNonNumericPostalCode(t *testing.T) {
	for _, mode := range []string{"name", "postal"} {
		t.Run(mode, func(t *testing.T) {
			s, _, ctx := deliveryTest(t)
			r := request(false)
			if mode == "name" {
				r.FirstName = strings.Repeat("ก", 101)
			} else {
				r.PostalCode = "10x00"
			}
			_, err := s.Create(ctx, owner, r)
			var business *BusinessError
			if !errors.As(err, &business) || business.Code != "MA00005" {
				t.Fatal(err)
			}
		})
	}
}
func TestDefaultRejectsDuplicates(t *testing.T) {
	s, m, ctx := deliveryTest(t)
	rows := sqlmock.NewRows([]string{"id", "user", "first", "last", "line", "province", "district", "sub", "postal", "mobile", "description", "default"})
	for _, id := range []string{"one", "two"} {
		rows.AddRow(id, owner, "First", "Last", "Line", "10", "1001", "100101", "10100", "0800000000", "address", true)
	}
	m.ExpectQuery(`SELECT da.id,.*WHERE da.mobile_user_uuid=\? AND da.is_default=1 AND da.is_active='YES' ORDER BY da.updated_at DESC, da.id$`).WithArgs(owner).WillReturnRows(rows).RowsWillBeClosed()
	if _, err := s.Default(ctx, owner); err == nil {
		t.Fatal("expected duplicate error")
	}
}

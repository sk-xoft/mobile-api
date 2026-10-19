package renewal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockRenewalRepository(t *testing.T) (*Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewRepository(db, nil), mock, func() { db.Close() }
}

func TestRenewalRepositoryStatusesDocumentsAndPrice(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, document_status_code, document_mobile_status_code")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "document_status_code", "document_mobile_status_code", "document_mobile_status_name_th", "document_mobile_status_name_en", "name_th", "name_en", "css_color"}).
			AddRow("s1", "PAYMENT_PENDING", "PAYMENT", "รอชำระ", "Payment pending", "รอชำระ", "Payment pending", "#fff"))
	statuses, err := r.Statuses(context.Background())
	if err != nil || len(statuses) != 1 || statuses[0].Code != "PAYMENT_PENDING" {
		t.Fatalf("statuses = %#v, %v", statuses, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DOCUMENT_CODE, DOCUMENT_NAME_TH, DOCUMENT_NAME_EN")).
		WillReturnRows(sqlmock.NewRows([]string{"code", "th", "en"}).AddRow("DOC001", nil, "Certificate"))
	documents, err := r.Documents(context.Background())
	if err != nil || len(documents) != 1 || documents[0].DocumentName != "" || documents[0].DocumentNameEn != "Certificate" {
		t.Fatalf("documents = %#v, %v", documents, err)
	}

	effective := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.document_code, p.government_fee")).
		WithArgs("DOC001").
		WillReturnRows(sqlmock.NewRows([]string{"document_code", "government_fee", "document_processing_fee", "shipping_fee", "shipping_discount", "service_fee_discount", "effective_from", "effective_to"}).
			AddRow("DOC001", 100, 50, 25, 10, 5, effective, nil))
	price, err := r.Price(context.Background(), "DOC001")
	if err != nil || price.Total != 160 || price.DocumentCode != "DOC001" {
		t.Fatalf("price = %#v, %v", price, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRenewalRepositoryOwnedRequestAndLatestActiveNotFound(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	ownedQuery := regexp.QuoteMeta("SELECT r.id,r.request_no,r.mobile_user_uuid")
	mock.ExpectQuery(ownedQuery).WithArgs("missing", "user-1").WillReturnError(sql.ErrNoRows)
	if _, err := r.ownedRequest(context.Background(), "user-1", "id", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ownedRequest error = %v", err)
	}
	mock.ExpectQuery(ownedQuery).WithArgs("user-1", "DOC001").WillReturnError(sql.ErrNoRows)
	if _, err := r.latestActive(context.Background(), "user-1", "DOC001"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("latestActive error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func expectOwnedRequest(mock sqlmock.Sqlmock, key string, args ...driver.Value) {
	expectation := mock.ExpectQuery(regexp.QuoteMeta("SELECT r.id,r.request_no,r.mobile_user_uuid")).WithArgs(args...)
	expectation.WillReturnRows(sqlmock.NewRows([]string{"id", "request_no", "mobile_user_uuid", "mobile_number", "email", "document_code", "document_status_id", "document_status_code", "name_en", "idempotency_key", "amount", "is_resubmit", "submitted_at"}).
		AddRow("request-1", "260100001", "user-1", "0812345678", "crew@example.com", "DOC001", "status-1", "PAYMENT_PENDING", "Payment pending", "idem-1", 180.0, false, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
}

func TestRenewalRepositoryListDetailTimeline(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.id,r.request_no,r.document_code")).WithArgs("user-1", 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_no", "document_code", "document_status_code", "name_en", "amount", "submitted_at"}).AddRow("r1", "260100001", "DOC001", "PAYMENT_PENDING", "Payment pending", 180, when))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_document_request")).WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))
	items, total, err := r.list(context.Background(), "user-1", 0)
	if err != nil || total != 1 || len(items) != 1 || items[0].SubmittedAt != when.Format(time.RFC3339) {
		t.Fatalf("list = %#v, %d, %v", items, total, err)
	}

	expectOwnedRequest(mock, "request_no", "260100001", "user-1")
	detail, err := r.detail(context.Background(), "user-1", "260100001")
	if err != nil || detail.RequestID != "request-1" || detail.Amount != 180 {
		t.Fatalf("detail = %#v, %v", detail, err)
	}

	expectOwnedRequest(mock, "request_no", "260100001", "user-1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT action,from_status,to_status,note,actioned_at")).WithArgs("request-1").
		WillReturnRows(sqlmock.NewRows([]string{"action", "from_status", "to_status", "note", "actioned_at"}).AddRow("CREATE", "", "PAYMENT_PENDING", "created", when))
	timeline, err := r.timeline(context.Background(), "user-1", "260100001")
	if err != nil || len(timeline.Items) != 1 || timeline.Items[0].Action != "CREATE" {
		t.Fatalf("timeline = %#v, %v", timeline, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRenewalRepositoryCountsAndDocumentInfo(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_TH,''),COALESCE(DOCUMENT_NAME_EN,'')")).
		WithArgs("DOC001").
		WillReturnRows(sqlmock.NewRows([]string{"name", "name_th", "name_en"}).AddRow("ไทย", "ไทย", "English"))
	name, nameTh, nameEn, err := r.documentInfo(context.Background(), "DOC001")
	if err != nil || name != "ไทย" || nameTh != "ไทย" || nameEn != "English" {
		t.Fatalf("documentInfo = %q, %q, %q, %v", name, nameTh, nameEn, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_document_setting_requires")).
		WithArgs("DOC001").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(3))
	count, err := r.requiredItemCount(context.Background(), "DOC001")
	if err != nil || count != 3 {
		t.Fatalf("requiredItemCount = %d, %v", count, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

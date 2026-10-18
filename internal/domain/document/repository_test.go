package document

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRepositoryPageScansNullableDocumentFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := regexp.QuoteMeta("SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,")
	mock.ExpectQuery(query).
		WithArgs("user-1", "COT", 0).
		WillReturnRows(sqlmock.NewRows([]string{"DOCUMENT_ID", "DOCUMENT_CODE", "DOCUMENT_NAME_TH", "DOCUMENT_NAME_EN", "DOCUMENT_DESCRIPTION", "DOCUMENT_TYPE", "DOCUMENT_DEFAULT_FLAG", "DOCUMENT_RENEWAL_FLAG", "DOCUMENT_SEQ", "DOCUMENT_MOBILE_FLAG", "DOCUMENT_COURSE_CODE", "CERT_START_DATE", "CERT_END_DATE", "ORIGINAL_FILE_NAME"}).
			AddRow(1, "DOC001", nil, "English", nil, "COT", "N", "YES", nil, "Y", nil, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_documents md")).
		WithArgs("user-1", "COT").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	items, total, err := NewRepository(db).Page(context.Background(), "user-1", "COT", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].DocumentID != "1" || items[0].DocumentNameTh != "" || items[0].DocumentSeq != "" {
		t.Fatalf("unexpected page: total=%d items=%#v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryCloseToExpirationScansRenewalFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := regexp.QuoteMeta("SELECT md.DOCUMENT_ID, md.DOCUMENT_CODE, md.DOCUMENT_NAME_TH, md.DOCUMENT_NAME_EN, md.DOCUMENT_DESCRIPTION,")
	mock.ExpectQuery(query).
		WithArgs("user-1", "user-1", 0).
		WillReturnRows(sqlmock.NewRows([]string{"DOCUMENT_ID", "DOCUMENT_CODE", "DOCUMENT_NAME_TH", "DOCUMENT_NAME_EN", "DOCUMENT_DESCRIPTION", "DOCUMENT_TYPE", "DOCUMENT_DEFAULT_FLAG", "DOCUMENT_RENEWAL_FLAG", "DOCUMENT_SEQ", "DOCUMENT_MOBILE_FLAG", "DOCUMENT_COURSE_CODE", "CERT_START_DATE", "CERT_END_DATE", "ORIGINAL_FILE_NAME", "DOCUMENT_RENEWAL_PROCESSING_FLAG"}).
			AddRow(2, "DOC002", "ไทย", "English", "desc", "COT", "N", "YES", 2, "Y", "COURSE", nil, time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), "cert.pdf", "N"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_documents md INNER JOIN m_certificates mc")).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	items, total, err := NewRepository(db).CloseToExpiration(context.Background(), "user-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].DocumentRenewalRequestFlag != "N" || items[0].DocumentSeq != "2" {
		t.Fatalf("unexpected result: total=%d items=%#v", total, items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNullableEnd(t *testing.T) {
	if nullableEnd("") != nil || nullableEnd("9999-99-99") != nil || nullableEnd("2026-01-01") != "2026-01-01" {
		t.Fatal("nullableEnd() did not normalize open-ended dates")
	}
}

package renewal

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func validateItemsRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "profile_id", "code", "scope", "type", "name", "sort", "uploaded", "status", "path", "uploaded_at", "check", "note"})
}
func validateFilesRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"code", "id", "type", "slot", "path", "name", "mime", "size", "uploaded", "check", "note", "updated"})
}
func validateAddressRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "first", "last", "address", "province", "district", "subdistrict", "postal", "mobile", "description"})
}

func TestValidateResponseBatchesFilesWithOneConnection(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	row := requestRow{ID: "request", UserID: "owner", No: "260100001", Mobile: "0812345678", Email: "crew@example.com", IdempotencyKey: "idem"}
	mock.ExpectQuery("SELECT COALESCE\\(DOCUMENT_NAME_TH").WithArgs("DOC001").WillReturnRows(sqlmock.NewRows([]string{"name", "th", "en"}).AddRow("ไทย", "ไทย", "English"))
	mock.ExpectQuery("SELECT CERT_END_DATE").WithArgs("owner", "DOC001").WillReturnRows(sqlmock.NewRows([]string{"end"}).AddRow(nil))
	mock.ExpectQuery("SELECT a.id,a.first_name").WithArgs("request", "owner").WillReturnRows(validateAddressRows()).RowsWillBeClosed()
	items := validateItemsRows()
	// Ten items still need only two file queries, including interleaved scopes.
	profileArgs := []driver.Value{"owner"}
	requestArgs := []driver.Value{"request", "owner"}
	for i := 0; i < 10; i++ {
		code := fmt.Sprintf("MRI%03d", i+1)
		scope := "PROFILE"
		if i%2 == 1 {
			scope = "REQUEST"
			requestArgs = append(requestArgs, code)
		} else {
			profileArgs = append(profileArgs, code)
		}
		items.AddRow(fmt.Sprint(i), "profile", code, scope, "GENERAL", code, i, 0, "MISSING", "", "", "", "")
	}
	mock.ExpectQuery("SELECT i.id,COALESCE").WithArgs("owner", "request").WillReturnRows(items).RowsWillBeClosed()
	profileFiles := validateFilesRows().AddRow("MRI001", "front", "ID_CARD", "FRONT", "front.jpg", "front.jpg", "image/jpeg", 10, "01/01/2026 10:00", "", "", false).
		AddRow("MRI003", "general", "GENERAL", "MAIN", "general.pdf", "general.pdf", "application/pdf", 20, "01/01/2026 10:01", "FIX", "replace", true).
		AddRow("MRI001", "back", "ID_CARD", "BACK", "back.jpg", "back.jpg", "image/jpeg", 10, "01/01/2026 10:02", "", "", false)
	mock.ExpectQuery("SELECT document_master_request_item_code,id.*mobile_user_uuid=\\?.* IN ").WithArgs(profileArgs...).WillReturnRows(profileFiles).RowsWillBeClosed()
	requestFiles := validateFilesRows().AddRow("MRI002", "request-file", "GENERAL", "MAIN", "request.pdf", "request.pdf", "application/pdf", 30, "01/01/2026 10:00", "", "", false)
	mock.ExpectQuery("SELECT i.document_master_request_item_code,f.id.*r.mobile_user_uuid=\\?.* IN ").WithArgs(requestArgs...).WillReturnRows(requestFiles).RowsWillBeClosed()
	response, err := r.validateResponse(ctx, row, "DOC001", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Items) != 10 || response.Items[0].ItemCode != "MRI001" || response.Items[0].DocumentStatus != "COMPLETE" {
		t.Fatalf("response=%+v", response)
	}
	if files := response.Items[0].Files; len(files) != 2 || files[0].FileID != "front" || files[1].FileID != "back" {
		t.Fatalf("file ordering/grouping: %+v", files)
	}
	if item := response.Items[1]; item.FilePath != "request.pdf" || item.FileUploaded != 1 || item.DocumentStatus != "COMPLETE" {
		t.Fatalf("request item: %+v", item)
	}
	if response.Items[2].DocumentStatus != "NEED_FIX" || response.Items[3].DocumentStatus != "MISSING" {
		t.Fatal("incorrect status")
	}
	if response.IdempotencyKey == nil || *response.IdempotencyKey != "idem" {
		t.Fatal("lost idempotency key")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateFilesBatchSkipsUnusedScopeAndPropagatesReadErrors(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	readErr := errors.New("file stream failed")
	rows := validateFilesRows().AddRow("MRI002", "f", "GENERAL", "MAIN", "path", "file", "application/pdf", 1, "", "", "", false).RowError(0, readErr)
	mock.ExpectQuery("SELECT i.document_master_request_item_code,f.id").WithArgs("request", "owner", "MRI002").WillReturnRows(rows).RowsWillBeClosed()
	_, err := r.validateFilesBatch(context.Background(), requestRow{ID: "request", UserID: "owner"}, []ValidateItem{{ItemCode: "MRI002", StorageScope: "REQUEST"}})
	if !errors.Is(err, readErr) {
		t.Fatalf("error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCreateAvoidsDuplicateReadsAndUsesOneConnection(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	mock.ExpectQuery("SELECT MOBILE_NUMBER,EMAIL").WithArgs("owner").WillReturnRows(sqlmock.NewRows([]string{"mobile", "email"}).AddRow("0812345678", "crew@example.com"))
	mock.ExpectQuery("SELECT id FROM m_document_status").WithArgs("PAYMENT_PENDING").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("status"))
	mock.ExpectQuery("SELECT COALESCE\\(DOCUMENT_NAME_TH").WithArgs("DOC001").WillReturnRows(sqlmock.NewRows([]string{"name", "th", "en"}).AddRow("ไทย", "ไทย", "English"))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM m_document_setting_requires").WithArgs("DOC001").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// This read must finish before Begin, avoiding a second connection during the transaction.
	mock.ExpectQuery("SELECT id,first_name,last_name,address_line").WithArgs("owner").WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO m_document_request_sequence").WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT last_number.*FOR UPDATE").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"number"}).AddRow(1))
	mock.ExpectExec("INSERT INTO m_document_request\\(").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "owner", "0812345678", "crew@example.com", "DOC001", "status", "DOC001", nil, 180.0, "idem").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO m_document_request_items").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "DOC001").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO m_document_transaction").WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "CREATE", nil, "PAYMENT_PENDING", "Unpaid draft created", "owner").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT CERT_END_DATE").WithArgs("owner", "DOC001").WillReturnRows(sqlmock.NewRows([]string{"end"}).AddRow(nil))
	mock.ExpectQuery("SELECT a.id,a.first_name").WithArgs(sqlmock.AnyArg(), "owner").WillReturnRows(validateAddressRows()).RowsWillBeClosed()
	mock.ExpectQuery("SELECT i.id,COALESCE").WithArgs("owner", sqlmock.AnyArg()).WillReturnRows(validateItemsRows().AddRow("item", "", "MRI002", "REQUEST", "", "Document", 1, 0, "MISSING", "", "", "", "")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT i.document_master_request_item_code,f.id").WithArgs(sqlmock.AnyArg(), "owner", "MRI002").WillReturnRows(validateFilesRows()).RowsWillBeClosed()
	response, err := r.validateCreate(ctx, "owner", "DOC001", "idem", Price{Total: 180})
	if err != nil {
		t.Fatal(err)
	}
	if response.RequestID == "" || response.RequestNo == "" || response.DocumentNameEn != "English" || response.MobileNumber != "0812345678" || response.Items[0].DocumentStatus != "MISSING" {
		t.Fatalf("response=%+v", response)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type validateServiceRepository struct {
	repositoryPort
	lookupErr, activeErr error
	existing, active     requestRow
}

func (r *validateServiceRepository) validateExisting(context.Context, string, string, string) (requestRow, error) {
	return r.existing, r.lookupErr
}
func (r *validateServiceRepository) latestActive(context.Context, string, string) (requestRow, error) {
	return r.active, r.activeErr
}
func (r *validateServiceRepository) validateResponse(_ context.Context, row requestRow, code, key string) (ValidateResponse, error) {
	return ValidateResponse{RequestID: row.ID, DocumentCode: code, IdempotencyKey: stringPointer(key)}, nil
}

func TestValidateStopsOnDatabaseErrorsAndPreservesReplay(t *testing.T) {
	failure := errors.New("database unavailable")
	for _, r := range []*validateServiceRepository{{lookupErr: failure}, {lookupErr: ErrNotFound, activeErr: failure}} {
		_, err := NewService(r).Validate(context.Background(), "owner", "doc001", "idem")
		if !errors.Is(err, failure) {
			t.Fatalf("error=%v", err)
		}
	}
	r := &validateServiceRepository{existing: requestRow{ID: "existing", Code: "DOC001"}}
	response, err := NewService(r).Validate(context.Background(), "owner", " doc001 ", " idem ")
	if err != nil || response.RequestID != "existing" || *response.IdempotencyKey != "idem" {
		t.Fatalf("response=%+v,error=%v", response, err)
	}
	if _, err := NewService(r).Validate(context.Background(), "owner", "OTHER", "idem"); err == nil {
		t.Fatal("idempotency key accepted for another document")
	}
	r = &validateServiceRepository{lookupErr: ErrNotFound, active: requestRow{ID: "active", Code: "DOC001", IdempotencyKey: "old-key"}}
	response, err = NewService(r).Validate(context.Background(), "owner", "DOC001", "new-key")
	if err != nil || response.RequestID != "active" || *response.IdempotencyKey != "old-key" {
		t.Fatalf("active replay=%+v,error=%v", response, err)
	}
}

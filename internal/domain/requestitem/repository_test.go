package requestitem

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newRequestItemDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return db, mock, func() { db.Close() }
}

func TestRepositoryActiveAndProfileItems(t *testing.T) {
	db, mock, closeDB := newRequestItemDB(t)
	defer closeDB()
	r := NewRepository(db)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_document_master_request_item")).WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	if ok, err := r.active(ctx, "A"); err != nil || !ok {
		t.Fatalf("active()=(%v,%v)", ok, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM m_document_master_request_item")).WithArgs("B").WillReturnError(errors.New("count"))
	if _, err := r.active(ctx, "B"); err == nil {
		t.Fatal("active should return query errors")
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id,m.document_master_items_code")).WithArgs("user").WillReturnRows(
		sqlmock.NewRows([]string{"id", "code", "scope", "sort", "uploaded", "path", "uploaded_at", "result", "note"}).
			AddRow("1", "A", "PROFILE", 1, 0, "", "", "", "").AddRow("2", "B", "PROFILE", 2, 1, "p", "now", "OK", "note"))
	items, err := r.ProfileItems(ctx, "user")
	if err != nil || len(items) != 2 || items[0].DocumentStatus != "MISSING" || items[1].DocumentStatus != "COMPLETE" {
		t.Fatalf("ProfileItems()=(%+v,%v)", items, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryFilesAndRenewal(t *testing.T) {
	db, mock, closeDB := newRequestItemDB(t)
	defer closeDB()
	r := NewRepository(db)
	ctx := context.Background()
	columns := []string{"id", "type", "slot", "path", "name", "mime", "size", "uploaded", "result", "note", "updated"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,document_type,slot_code")).WithArgs("u", "A").WillReturnRows(sqlmock.NewRows(columns).AddRow("f", "PDF", "MAIN", "p", "n", "application/pdf", int64(10), "now", "OK", "", int64(1)))
	files, err := r.Files(ctx, "u", "A", "")
	if err != nil || len(files) != 1 || !files[0].IsUpdated {
		t.Fatalf("profile Files()=(%+v,%v)", files, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT f.id,f.document_type,f.slot_code")).WithArgs("request", "u", "A").WillReturnRows(sqlmock.NewRows(columns).AddRow("f2", "ID_CARD", "FRONT", "p2", "n2", "image/png", int64(2), "now", "", "", int64(0)))
	if files, err = r.Files(ctx, "u", "A", "request"); err != nil || len(files) != 1 || files[0].FileID != "f2" {
		t.Fatalf("renewal Files()=(%+v,%v)", files, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.id,i.id,s.document_status_code")).WithArgs("RN", "u", "A").WillReturnRows(sqlmock.NewRows([]string{"request", "item", "status", "approve"}).AddRow("r", "i", "PAYMENT_PENDING", ""))
	rid, iid, status, approve, err := r.Renewal(ctx, "u", "RN", "A")
	if err != nil || rid != "r" || iid != "i" || status != "PAYMENT_PENDING" || approve != "" {
		t.Fatalf("Renewal()=(%q,%q,%q,%q,%v)", rid, iid, status, approve, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT r.id,i.id,s.document_status_code")).WithArgs("missing", "u", "A").WillReturnError(sql.ErrNoRows)
	if _, _, _, _, err = r.Renewal(ctx, "u", "missing", "A"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Renewal missing error=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryUpserts(t *testing.T) {
	db, mock, closeDB := newRequestItemDB(t)
	defer closeDB()
	r := NewRepository(db)
	ctx := context.Background()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM m_document_profile_request_item")).WithArgs("u", "A", "GENERAL", "MAIN").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO m_document_profile_request_item")).WithArgs(sqlmock.AnyArg(), "u", "A", "GENERAL", "MAIN", "key", "name", "mime", int64(3), "A").WillReturnResult(sqlmock.NewResult(1, 1))
	if id, err := r.upsertProfile(ctx, "u", "A", "GENERAL", "MAIN", "key", "name", "mime", 3); err != nil || id == "" {
		t.Fatalf("profile insert=(%q,%v)", id, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM m_document_profile_request_item")).WithArgs("u", "A", "GENERAL", "MAIN").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("existing"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE m_document_profile_request_item SET file_uploaded=1")).WithArgs("key2", "name2", "mime2", int64(4), "existing").WillReturnResult(sqlmock.NewResult(1, 1))
	if id, err := r.upsertProfile(ctx, "u", "A", "GENERAL", "MAIN", "key2", "name2", "mime2", 4); err != nil || id != "existing" {
		t.Fatalf("profile update=(%q,%v)", id, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM m_document_request_item_files")).WithArgs("item", "GENERAL", "MAIN").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO m_document_request_item_files")).WithArgs(sqlmock.AnyArg(), "item", "A", "GENERAL", "MAIN", "key", "name", "mime", int64(3)).WillReturnResult(sqlmock.NewResult(1, 1))
	if id, err := r.upsertRenewal(ctx, "item", "A", "GENERAL", "MAIN", "key", "name", "mime", 3); err != nil || id == "" {
		t.Fatalf("renewal insert=(%q,%v)", id, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM m_document_request_item_files")).WithArgs("item", "GENERAL", "MAIN").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("file"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE m_document_request_item_files SET file_uploaded=1")).WithArgs("key2", "name2", "mime2", int64(4), "file").WillReturnResult(sqlmock.NewResult(1, 1))
	if id, err := r.upsertRenewal(ctx, "item", "A", "GENERAL", "MAIN", "key2", "name2", "mime2", 4); err != nil || id != "file" {
		t.Fatalf("renewal update=(%q,%v)", id, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

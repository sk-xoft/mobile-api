package requestitem

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCompletionUsesMatchingUploadedSlots(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		files      []File
		want       bool
	}{
		{"passport main", "PASSPORT", []File{{DocumentType: "PASSPORT", SlotCode: "MAIN", UploadedAt: "now"}}, true},
		{"wrong type", "ID_CARD", []File{{DocumentType: "PASSPORT", SlotCode: "MAIN", UploadedAt: "now"}}, false},
		{"missing upload", "GENERAL", []File{{DocumentType: "GENERAL", SlotCode: "MAIN"}}, false},
		{"wrong slot", "GENERAL", []File{{DocumentType: "GENERAL", SlotCode: "FRONT", UploadedAt: "now"}}, false},
		{"identity main", "ID_CARD", []File{{DocumentType: "ID_CARD", SlotCode: "MAIN", UploadedAt: "now"}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := complete(tc.kind, tc.files); got != tc.want {
				t.Fatalf("complete=%v want %v", got, tc.want)
			}
		})
	}
}

func TestCurrentFilesKeepsLatestType(t *testing.T) {
	files := currentFiles([]File{{DocumentType: "ID_CARD", UploadedAt: "2026-01-01"}, {DocumentType: "PASSPORT", UploadedAt: "2026-02-01"}, {DocumentType: "ID_CARD", UploadedAt: "2026-01-02"}})
	if len(files) != 1 || files[0].DocumentType != "PASSPORT" {
		t.Fatalf("files=%+v", files)
	}
}

func TestUploadRejectsInvalidMasterBeforeStorage(t *testing.T) {
	for _, body := range [][]byte{nil, make([]byte, 10*1024*1024+1), []byte("%PDF-1.7")} {
		storage := &fakeRequestItemStorage{}
		repo := &fakeRequestItemRepo{validateErr: ErrNotFound}
		_, err := NewService(repo, storage, "uploads").Upload(context.Background(), "u", "A", "GENERAL", "MAIN", "f.pdf", body, "")
		if err == nil || storage.key != "" {
			t.Fatalf("err=%v storage=%q", err, storage.key)
		}
	}
}

func TestValidateItem(t *testing.T) {
	for _, tc := range []struct {
		name, actual, scope string
		queryErr, want      error
	}{
		{name: "profile", actual: "PROFILE", scope: "PROFILE"},
		{name: "request", actual: "REQUEST"},
		{name: "wrong scope", actual: "REQUEST", scope: "PROFILE", want: ErrNotFound},
		{name: "inactive", queryErr: sql.ErrNoRows, want: ErrNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, closeDB := newRequestItemDB(t)
			defer closeDB()
			q := mock.ExpectQuery("SELECT storage_scope").WithArgs("A")
			if tc.queryErr != nil {
				q.WillReturnError(tc.queryErr)
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"scope"}).AddRow(tc.actual))
			}
			if err := NewRepository(db).ValidateItem(context.Background(), "A", tc.scope); !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUpsertTransactions(t *testing.T) {
	for _, tc := range []struct {
		name, status, approve string
		profile, reject       bool
	}{
		{name: "profile", profile: true},
		{name: "pending", status: "PAYMENT_PENDING"},
		{name: "correction", status: "PENDING_APPLICANT_CORRECTION", approve: "FIX"},
		{name: "status changed", status: "APPROVED", reject: true},
		{name: "item not correctable", status: "PENDING_APPLICANT_CORRECTION", approve: "PASS", reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, closeDB := newRequestItemDB(t)
			defer closeDB()
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT storage_scope.*FOR UPDATE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"scope"}).AddRow("PROFILE"))
			if !tc.profile {
				mock.ExpectQuery("SELECT s.document_status_code.*FOR UPDATE").WithArgs("item", "A").WillReturnRows(sqlmock.NewRows([]string{"status", "approve"}).AddRow(tc.status, tc.approve))
			}
			if tc.reject {
				mock.ExpectRollback()
			} else {
				table := "m_document_request_item_files"
				if tc.profile {
					table = "m_document_profile_request_item"
				}
				mock.ExpectQuery("SELECT id FROM " + table).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("file"))
				mock.ExpectExec("UPDATE "+table).WithArgs("key", "name", "mime", int64(3), "file").WillReturnResult(sqlmock.NewResult(0, 1))
				owner := "item"
				if tc.profile {
					owner = "u"
				}
				mock.ExpectExec("UPDATE "+table+" SET file_uploaded=0").WithArgs(owner, "A", "GENERAL", "MAIN", "MAIN").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectCommit()
			}
			repo := NewRepository(db)
			var err error
			if tc.profile {
				_, err = repo.UpsertProfile(context.Background(), "u", "A", "GENERAL", "MAIN", "key", "name", "mime", 3)
			} else {
				_, err = repo.UpsertRenewal(context.Background(), "item", "A", "GENERAL", "MAIN", "key", "name", "mime", 3)
			}
			if (err != nil) != tc.reject {
				t.Fatalf("error=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProfileInsertZeroRowsFails(t *testing.T) {
	db, mock, closeDB := newRequestItemDB(t)
	defer closeDB()
	mock.ExpectQuery("SELECT id FROM").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO").WillReturnResult(sqlmock.NewResult(0, 0))
	_, err := NewRepository(db).upsertProfile(context.Background(), "u", "A", "GENERAL", "MAIN", "key", "name", "mime", 3)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type cleanupStorage struct {
	fakeRequestItemStorage
	deleted   string
	deleteErr error
}

func (s *cleanupStorage) Delete(_ context.Context, key string) error {
	s.deleted = key
	return s.deleteErr
}

func TestFailedSaveCleansUploadedObject(t *testing.T) {
	repo := &fakeRequestItemRepo{err: errors.New("database unavailable")}
	storage := &cleanupStorage{}
	svc := NewService(repo, storage, "uploads")
	_, err := svc.Upload(context.Background(), "u", "A", "GENERAL", "MAIN", "f.pdf", []byte("%PDF-1.7"), "")
	if err == nil || storage.deleted == "" || storage.deleted != storage.key {
		t.Fatalf("err=%v deleted=%q key=%q", err, storage.deleted, storage.key)
	}
	storage.deleted = ""
	if err := svc.cleanup(context.Background(), "key", ErrSaveUncertain); !errors.Is(err, ErrSaveUncertain) || storage.deleted != "" {
		t.Fatal("uncertain commit must retain object")
	}
}

func TestListUsesActiveFilesForStatusAndMetadata(t *testing.T) {
	for _, tc := range []struct {
		name     string
		files    []File
		status   string
		uploaded int
	}{
		{"front only", []File{{FileID: "front", DocumentType: "ID_CARD", SlotCode: "FRONT", UploadedAt: "2026-01-01"}}, "MISSING", 1},
		{"front and back", []File{{DocumentType: "ID_CARD", SlotCode: "FRONT", UploadedAt: "2026-01-01"}, {DocumentType: "ID_CARD", SlotCode: "BACK", UploadedAt: "2026-01-02"}}, "COMPLETE", 1},
		{"missing", nil, "MISSING", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRequestItemRepo{items: []Item{{ItemCode: "MRI001", FilePath: "stale", DocumentStatus: "COMPLETE"}}, files: tc.files}
			items, err := NewService(repo, &fakeRequestItemStorage{}, "uploads").List(context.Background(), "u")
			if err != nil || len(items) != 1 {
				t.Fatalf("items=%+v err=%v", items, err)
			}
			if items[0].DocumentStatus != tc.status || items[0].FileUploaded != tc.uploaded || items[0].FilePath == "stale" {
				t.Fatalf("item=%+v", items[0])
			}
		})
	}
}

func TestSaveRollsBackWhenSupersedingFails(t *testing.T) {
	db, mock, closeDB := newRequestItemDB(t)
	defer closeDB()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT storage_scope.*FOR UPDATE").WithArgs("A").WillReturnRows(sqlmock.NewRows([]string{"scope"}).AddRow("PROFILE"))
	mock.ExpectQuery("SELECT id FROM").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("f"))
	mock.ExpectExec("UPDATE .* SET file_uploaded=1").WillReturnResult(sqlmock.NewResult(0, 1))
	cause := errors.New("supersede failed")
	mock.ExpectExec("UPDATE .* SET file_uploaded=0").WillReturnError(cause)
	mock.ExpectRollback()
	_, err := NewRepository(db).UpsertProfile(context.Background(), "u", "A", "GENERAL", "MAIN", "k", "n", "m", 1)
	if !errors.Is(err, cause) {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

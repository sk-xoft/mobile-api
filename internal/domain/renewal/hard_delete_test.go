package renewal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

const hardDeleteLockQuery = "SELECT id,request_no,document_status_id FROM m_document_request WHERE request_no=? AND is_active='YES' FOR UPDATE"

func expectHardDeleteRequest(mock sqlmock.Sqlmock) {
	mock.ExpectBegin()
	mock.ExpectQuery("^" + regexp.QuoteMeta(hardDeleteLockQuery) + "$").
		WithArgs("261000004").
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_no", "document_status_id"}).
			AddRow("other-owner-request", "261000004", "pending-status"))
	mock.ExpectQuery("^" + regexp.QuoteMeta("SELECT document_status_code FROM m_document_status WHERE id=?") + "$").
		WithArgs("pending-status").
		WillReturnRows(sqlmock.NewRows([]string{"document_status_code"}).AddRow("PAYMENT_PENDING"))
}

func expectHardDeleteChildren(mock sqlmock.Sqlmock, lastRows int64) {
	for _, table := range []string{"m_delivery", "m_dept_submission"} {
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?")).
			WithArgs(table).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectExec("^" + regexp.QuoteMeta("DELETE FROM "+table+" WHERE request_id=?") + "$").
			WithArgs("other-owner-request").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	for _, query := range []string{
		"DELETE FROM mobile_payment_request_state WHERE request_id=?",
		"DELETE FROM mobile_payment_attempts WHERE request_id=?",
		"DELETE f FROM m_document_request_item_files f INNER JOIN m_document_request_items i ON i.id=f.request_item_id WHERE i.request_id=?",
		"DELETE FROM m_document_request_items WHERE request_id=?",
		"DELETE FROM m_document_request_delivery_address WHERE request_id=?",
		"UPDATE m_payment_transaction SET parent_transaction_id=NULL WHERE request_id=?",
		"DELETE FROM m_document_transaction WHERE request_id=?",
		"DELETE FROM m_payment_transaction WHERE request_id=?",
	} {
		mock.ExpectExec("^" + regexp.QuoteMeta(query) + "$").WithArgs("other-owner-request").
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectExec("^" + regexp.QuoteMeta("DELETE FROM m_document_request WHERE id=?") + "$").
		WithArgs("other-owner-request").WillReturnResult(sqlmock.NewResult(0, lastRows))
}

func TestHardDeleteHandlerDoesNotFilterBySessionOwner(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	expectHardDeleteRequest(mock)
	expectHardDeleteChildren(mock, 1)
	mock.ExpectCommit()

	server := fiber.New()
	server.Use(func(c *fiber.Ctx) error {
		c.Locals("user.mobile_uuid", "different-authenticated-user")
		return c.Next()
	})
	server.Delete("/v1/document-renewals/requests/:requestNo/hard", NewHandler(NewService(r)).HardDelete)
	response, err := server.Test(httptest.NewRequest("DELETE", "/v1/document-renewals/requests/261000004/hard", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body struct {
		Code string             `json:"code"`
		Data TransitionResponse `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || body.Code != "MA00000" || body.Data.RequestNo != "261000004" || body.Data.Action != "HARD_DELETE" {
		t.Fatalf("unexpected response: status=%d body=%+v", response.StatusCode, body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHardDeleteMissingActiveRequestRollsBack(t *testing.T) {
	for name, queryErr := range map[string]error{"not-found-or-inactive": sql.ErrNoRows, "database-error": errors.New("database unavailable")} {
		t.Run(name, func(t *testing.T) {
			r, mock, closeDB := newMockRenewalRepository(t)
			defer closeDB()
			mock.ExpectBegin()
			mock.ExpectQuery("^" + regexp.QuoteMeta(hardDeleteLockQuery) + "$").WithArgs("261000004").WillReturnError(queryErr)
			mock.ExpectRollback()
			_, err := r.hardDelete(context.Background(), "261000004")
			want := queryErr
			if errors.Is(queryErr, sql.ErrNoRows) {
				want = ErrNotFound
			}
			if !errors.Is(err, want) {
				t.Fatalf("error=%v, want %v", err, want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHardDeleteRollsBackAllChildrenIfParentWasNotDeleted(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	expectHardDeleteRequest(mock)
	expectHardDeleteChildren(mock, 0)
	mock.ExpectRollback()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := r.hardDelete(ctx, "261000004"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error=%v, want ErrNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestHardDeleteChildFailureRollsBack(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	expectHardDeleteRequest(mock)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?")).
		WithArgs("m_delivery").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	failure := errors.New("delete failed")
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM m_delivery WHERE request_id=?")).
		WithArgs("other-owner-request").WillReturnError(failure)
	mock.ExpectRollback()
	if _, err := r.hardDelete(context.Background(), "261000004"); !errors.Is(err, failure) {
		t.Fatalf("error=%v, want original delete failure", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

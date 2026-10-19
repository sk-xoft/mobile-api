package paymentmysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
)

func fixture() domain.Attempt {
	return domain.Attempt{ID: "attempt", No: "PAY001", RequestID: "request", RequestNo: "REQ001", UserID: "owner", Key: "key", Fingerprint: "fingerprint", Reference: "attempt", Status: "PENDING", Creation: "CONFIRMED", Amount: 18000, Input: domain.Input{PaymentMethod: "promptpay", IdempotencyKey: "key"}, CreatedAt: time.Now().UTC(), Charge: domain.Charge{ID: "chrg_1", Reference: "attempt", Status: "pending", Amount: 18000, Currency: "THB", Raw: `{"id":"chrg_1"}`}}
}
func attemptRows(a domain.Attempt) *sqlmock.Rows {
	input, _ := json.Marshal(a.Input)
	charge, _ := json.Marshal(a.Charge)
	return sqlmock.NewRows([]string{"id", "no", "request_id", "request_no", "user", "key", "fingerprint", "ref", "amount", "input", "charge", "status", "creation", "legacy", "review", "created", "sent", "checked", "failures", "paid"}).AddRow(a.ID, a.No, a.RequestID, a.RequestNo, a.UserID, a.Key, a.Fingerprint, a.Reference, a.Amount, string(input), string(charge), a.Status, a.Creation, a.Legacy, a.Review, a.CreatedAt, nil, nil, a.Failures, a.Paid)
}
func mockRequest(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(requestSQL + " FOR UPDATE")).WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"id", "no", "owner", "status", "amount", "active"}).AddRow("request", "REQ001", "owner", "PAYMENT_PENDING", 18000, true))
}
func TestReserveSerializesWithRequestLockAndOneConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	a := fixture()
	a.Status = "PROCESSING"
	a.Creation = "RESERVED"
	a.Charge = domain.Charge{}
	mock.ExpectBegin()
	mockRequest(mock)
	mock.ExpectQuery(regexp.QuoteMeta(attemptSQL)+"a.request_id").WithArgs("request", "key").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO mobile_payment_request_state").WithArgs("request").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT active_attempt_id,paid_attempt_id,review_required.*FOR UPDATE").WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"active", "paid", "review"}).AddRow(nil, nil, false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM m_document_request_items").WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectExec("INSERT INTO mobile_payment_attempts").WithArgs(a.ID, a.No, a.RequestID, a.Key, a.Fingerprint, a.Reference, a.Amount, sqlmock.AnyArg(), a.CreatedAt).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mobile_payment_request_state SET active_attempt_id").WithArgs("attempt", "request").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	_, created, err := NewStore(db).Reserve(ctx, a)
	if err != nil || !created {
		t.Fatalf("created=%v,error=%v", created, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestReserveRejectsActiveAttemptUnderLock(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	a := fixture()
	mock.ExpectBegin()
	mockRequest(mock)
	mock.ExpectQuery(regexp.QuoteMeta(attemptSQL)+"a.request_id").WithArgs("request", "key").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO mobile_payment_request_state").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT active_attempt_id,paid_attempt_id,review_required.*FOR UPDATE").WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"active", "paid", "review"}).AddRow("old", nil, false))
	mock.ExpectQuery("SELECT status FROM mobile_payment_attempts").WithArgs("old").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("PROCESSING"))
	mock.ExpectRollback()
	_, _, err := NewStore(db).Reserve(context.Background(), a)
	var conflict *domain.Error
	if !errors.As(err, &conflict) || conflict.Code != "PAYMENT_IN_PROGRESS" {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func expectLedgerUpdate(mock sqlmock.Sqlmock, a domain.Attempt, c domain.Charge, status string) {
	mock.ExpectExec("UPDATE m_payment_transaction SET status").WithArgs(status, c.ID, c.SourceID, c.TransactionID, c.Status, c.Raw, c.ReturnURI, c.AuthorizeURI, c.FailureCode, c.FailureMessage, c.Live, nullableTime(c.ExpiresAt), status, status, a.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mobile_payment_attempts SET provider_charge_id").WithArgs(c.ID, sqlmock.AnyArg(), status, false, sqlmock.AnyArg(), a.ID).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE mobile_payment_request_state SET active_attempt_id=NULL").WithArgs(a.RequestID, a.ID).WillReturnResult(sqlmock.NewResult(0, 1))
}
func TestApplySuccessCommitsLedgerRequestAndTimelineTogether(t *testing.T) {
	for _, failTimeline := range []bool{false, true} {
		name := "commit"
		if failTimeline {
			name = "rollback"
		}
		t.Run(name, func(t *testing.T) {
			db, mock, _ := sqlmock.New()
			defer db.Close()
			db.SetMaxOpenConns(1)
			a := fixture()
			c := a.Charge
			c.Status = "successful"
			mock.ExpectBegin()
			mockRequest(mock)
			mock.ExpectQuery(regexp.QuoteMeta(attemptSQL) + "a.id").WithArgs(a.ID).WillReturnRows(attemptRows(a))
			mock.ExpectQuery("SELECT paid_attempt_id,review_required.*FOR UPDATE").WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"paid", "review"}).AddRow(nil, false))
			mock.ExpectQuery("SELECT id FROM m_document_status").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("review-status"))
			mock.ExpectExec("UPDATE m_document_request SET document_status_id").WithArgs("review-status", "request").WillReturnResult(sqlmock.NewResult(0, 1))
			timeline := mock.ExpectExec("INSERT INTO m_document_transaction").WithArgs(sqlmock.AnyArg(), "request", "owner")
			if failTimeline {
				timeline.WillReturnError(errors.New("timeline failed"))
				mock.ExpectRollback()
			} else {
				timeline.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectExec("UPDATE mobile_payment_request_state SET paid_attempt_id").WithArgs(a.ID, "request").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM m_payment_transaction").WithArgs(a.ID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("INSERT INTO m_payment_transaction").WithArgs(a.ID, a.RequestID, a.No, "PROMPTPAY", "promptpay", 180.0, "SUCCESS", c.ID, c.SourceID, c.TransactionID, c.Status, c.Raw, a.Key, "Document renewal REQ001", c.ReturnURI, c.AuthorizeURI, c.FailureCode, c.FailureMessage, c.Live, nil).WillReturnResult(sqlmock.NewResult(0, 1))
				expectLedgerUpdate(mock, a, c, "SUCCESS")
				mock.ExpectCommit()
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err := NewStore(db).Apply(ctx, a, c)
			if (err != nil) != failTimeline {
				t.Fatalf("error=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestOutOfOrderResultCannotDowngradeSuccessOrDuplicateTimeline(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	a := fixture()
	a.Status = "SUCCESS"
	a.Charge.Status = "successful"
	incoming := a.Charge
	incoming.Status = "expired"
	mock.ExpectBegin()
	mockRequest(mock)
	mock.ExpectQuery(regexp.QuoteMeta(attemptSQL) + "a.id").WithArgs(a.ID).WillReturnRows(attemptRows(a))
	mock.ExpectQuery("SELECT paid_attempt_id,review_required.*FOR UPDATE").WithArgs("request").WillReturnRows(sqlmock.NewRows([]string{"paid", "review"}).AddRow(a.ID, false))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM m_payment_transaction").WithArgs(a.ID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	expectLedgerUpdate(mock, a, a.Charge, "SUCCESS")
	mock.ExpectCommit()
	if err := NewStore(db).Apply(context.Background(), a, incoming); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
func TestLeaseThrottlesConcurrentPollsAndScansCloseRows(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectExec("UPDATE mobile_payment_attempts SET lease_owner=.*INTERVAL 5 SECOND").WithArgs("lease", "attempt", true).WillReturnResult(sqlmock.NewResult(0, 0))
	acquired, err := NewStore(db).Acquire(context.Background(), "attempt", "lease", true)
	if err != nil || acquired {
		t.Fatalf("acquired=%v,error=%v", acquired, err)
	}
	mock.ExpectQuery("SELECT id FROM mobile_payment_attempts.*INTERVAL 24 HOUR").WithArgs(50).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("attempt")).RowsWillBeClosed()
	ids, err := NewStore(db).Due(context.Background(), 50)
	if err != nil || len(ids) != 1 {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

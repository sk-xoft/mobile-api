package authmysql

import (
	"context"
	"errors"
	"fmt"
	domain "github.com/smartseaman/mobile-api/internal/domain/auth"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConsumeActionTokenRejectsExpiredOrUnknownToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM mobile_auth_tokens.*expires_at>NOW").WithArgs("hash", "reset").WillReturnRows(sqlmock.NewRows([]string{"user_id"}))
	mock.ExpectRollback()
	if _, err := NewRepository(db).ConsumeActionToken(context.Background(), "hash", "reset", "password-hash"); err != domain.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestConsumeResetAtomicallyChangesPasswordRevokesAndConsumes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM mobile_auth_tokens").WithArgs("hash", "reset").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user"))
	mock.ExpectQuery("SELECT MOBILE_UUID,USERNAME,EMAIL,USER_STATUS.*FOR UPDATE").WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"MOBILE_UUID", "USERNAME", "EMAIL", "USER_STATUS"}).AddRow("user", "crew", "crew@example.com", "A"))
	mock.ExpectQuery("SELECT token_hash.*FOR UPDATE").WithArgs("hash", "reset").WillReturnRows(sqlmock.NewRows([]string{"token_hash"}).AddRow("hash"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE m_mobile_users SET PASSWORD=? WHERE MOBILE_UUID=?")).WithArgs("password-hash", "user").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE t_session SET IS_ONLINE='NO'").WithArgs("user").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("DELETE FROM mobile_auth_tokens").WithArgs("user", "reset").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if _, err := NewRepository(db).ConsumeActionToken(context.Background(), "hash", "reset", "password-hash"); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRefreshRotationRejectsReplay(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectExec("UPDATE mobile_auth_sessions.*a.refresh_hash=.*s.IS_ONLINE='YES'.*u.USER_STATUS='A'").WithArgs("new", "jwt", "session", "old").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := NewRepository(db).RotateRefreshSession(context.Background(), "session", "old", "new", "jwt"); err != domain.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestActivationCannotReactivateDisabledAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT user_id FROM mobile_auth_tokens").WithArgs("hash", "activation").WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("user"))
	mock.ExpectQuery("SELECT MOBILE_UUID,USERNAME,EMAIL,USER_STATUS.*FOR UPDATE").WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"MOBILE_UUID", "USERNAME", "EMAIL", "USER_STATUS"}).AddRow("user", "crew", "crew@example.com", "D"))
	mock.ExpectQuery("SELECT token_hash.*FOR UPDATE").WithArgs("hash", "activation").WillReturnRows(sqlmock.NewRows([]string{"token_hash"}).AddRow("hash"))
	mock.ExpectRollback()
	if _, err := NewRepository(db).ConsumeActionToken(context.Background(), "hash", "activation", ""); err != domain.ErrNotFound {
		t.Fatalf("err=%v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLoginSessionOnlyRevokesOnlineSessionsAndRollsBackOnFailure(t *testing.T) {
	for _, failInsert := range []bool{false, true} {
		t.Run(fmt.Sprint(failInsert), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			r := NewRepository(db)
			now := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT USER_STATUS.*FOR UPDATE").WithArgs("user").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("A"))
			mock.ExpectExec(regexp.QuoteMeta("UPDATE t_session SET IS_ONLINE='NO' WHERE USER_ID=? AND IS_ONLINE='YES'")).WithArgs("user").WillReturnResult(sqlmock.NewResult(0, 1))
			insert := mock.ExpectExec("INSERT INTO t_session").WithArgs("session", "user", now, "jwt", now, "user", now, "user", "correlation")
			if failInsert {
				insert.WillReturnError(errors.New("insert failed"))
				mock.ExpectRollback()
			} else {
				insert.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO mobile_auth_sessions").WithArgs("session", "refresh-hash", now.Add(30*24*time.Hour)).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}
			err = r.CreateSecureSession(context.Background(), "session", "user", "jwt", "refresh-hash", "correlation", now)
			if (err != nil) != failInsert {
				t.Fatalf("error=%v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

package renewal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"
)

const timelineStatusQuery = `SELECT t.actioned_at,COALESCE(t.note,'') FROM m_document_transaction t INNER JOIN m_document_status s ON s.document_status_code=t.to_status WHERE t.request_id=? AND s.document_mobile_status_code=? ORDER BY t.actioned_at DESC,t.id DESC LIMIT 1`

func expectTimelineStatusOwner(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(`SELECT r.id,r.request_no,.*WHERE r.request_no=\? AND r.mobile_user_uuid=\? AND r.is_active='YES'$`).
		WithArgs("261000023", "owner").
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_no", "mobile_user_uuid", "mobile_number", "email", "document_code", "document_status_id", "document_status_code", "name_en", "idempotency_key", "amount", "is_resubmit", "submitted_at"}).
			AddRow("request-id", "261000023", "owner", "0800000000", "crew@example.com", "DOC001", "status-id", "PENDING_DOCUMENT_REVIEW", "Pending Document Review", "", 100, false, time.Now()))
}

func expectTimelineStatusItems(mock sqlmock.Sqlmock) {
	mock.ExpectQuery("SELECT i.id,COALESCE").WithArgs("owner", "request-id").
		WillReturnRows(validateItemsRows().
			AddRow("item-1", "profile-1", "MRI001", "PROFILE", "ID_CARD", "Identity document", 1, 1, "COMPLETE", "identity.png", "2026-10-31T10:00:00Z", "", "").
			AddRow("item-2", "profile-2", "MRI002", "PROFILE", "GENERAL", "Photo", 2, 1, "COMPLETE", "photo.png", "2026-10-31T10:00:00Z", "", "").
			AddRow("item-3", "", "MRI003", "PROFILE", "", "Medical certificate", 3, 0, "MISSING", "", "", "", "").
			AddRow("item-4", "", "MRI004", "REQUEST", "", "Sea Service Record", 4, 0, "INCOMPLETE", "", "", "", "")).RowsWillBeClosed()
	mock.ExpectQuery("SELECT document_master_request_item_code,id.*mobile_user_uuid=\\?.* IN ").
		WithArgs("owner", "MRI001", "MRI002", "MRI003").
		WillReturnRows(validateFilesRows().
			AddRow("MRI001", "profile-1", "ID_CARD", "MAIN", "identity.png", "identity.png", "image/png", 123, "31/10/2026 10:00", "", "", true).
			AddRow("MRI002", "profile-2", "GENERAL", "MAIN", "photo.png", "photo.png", "image/png", 456, "31/10/2026 10:00", "", "", false)).RowsWillBeClosed()
	mock.ExpectQuery("SELECT i.document_master_request_item_code,f.id.*r.mobile_user_uuid=\\?.* IN ").
		WithArgs("request-id", "owner", "MRI004").
		WillReturnRows(validateFilesRows().AddRow("MRI004", "request-file", "GENERAL", "MAIN", "sea-service.pdf", "sea-service.pdf", "application/pdf", 789, "31/10/2026 11:11", "", "", false)).RowsWillBeClosed()
}

func TestTimelineStatusHandlerReturnsLatestEventAndBatchedItems(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	expectTimelineStatusOwner(mock)
	at := time.Date(2026, 10, 31, 10, 30, 0, 0, time.UTC)
	mock.ExpectQuery("^"+regexp.QuoteMeta(timelineStatusQuery)+"$").WithArgs("request-id", "DOCUMENT_REVIEW").
		WillReturnRows(sqlmock.NewRows([]string{"actioned_at", "note"}).AddRow(at, "Documents received"))
	expectTimelineStatusItems(mock)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error { c.Locals("user.mobile_uuid", "owner"); return c.Next() })
	app.Get("/v1/document-renewals/:requestNo/timeline/:documentMobileStatusCode/status", NewHandler(NewService(r)).TimelineStatus)
	response, err := app.Test(httptest.NewRequest("GET", "/v1/document-renewals/261000023/timeline/DOCUMENT_REVIEW/status", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body struct {
		Code string                 `json:"code"`
		Data TimelineStatusResponse `json:"data"`
	}
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(bytes, &body); err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Data struct {
			Timeline map[string]any `json:"timeline"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bytes, &contract); err != nil {
		t.Fatal(err)
	}
	if contract.Data.Timeline["date_time"] != "2026-10-31T10:30:00Z" || len(contract.Data.Timeline) != 2 {
		t.Fatalf("timeline must contain date_time and note: %+v", contract.Data.Timeline)
	}
	if response.StatusCode != 200 || body.Code != "MA00000" {
		t.Fatalf("status=%d, code=%s", response.StatusCode, body.Code)
	}
	v := body.Data
	if v.RequestNo != "261000023" || v.DocumentMobileStatusCode != "DOCUMENT_REVIEW" || v.Timeline.DateTime != "2026-10-31T10:30:00Z" || v.Timeline.Note != "Documents received" {
		t.Fatalf("unexpected timeline: %+v", v)
	}
	if len(v.Items) != 4 || v.Items[0].ProfileID != "profile-1" || v.Items[0].DocumentStatus != "COMPLETE" || !v.Items[0].Files[0].IsUpdated || v.Items[2].DocumentStatus != "MISSING" {
		t.Fatalf("unexpected profile items: %+v", v.Items)
	}
	last := v.Items[3]
	if last.StorageScope != "REQUEST" || last.DocumentCode != "DOC001" || last.DocumentStatus != "COMPLETE" || last.FileUploaded != 1 || last.FilePath != "sea-service.pdf" || len(last.Files) != 1 || last.Files[0].FileSize != 789 {
		t.Fatalf("unexpected request item: %+v", last)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineStatusWithoutHistoryKeepsItemsAndEmptyTimeline(t *testing.T) {
	r, mock, closeDB := newMockRenewalRepository(t)
	defer closeDB()
	r.db.SetMaxOpenConns(1)
	expectTimelineStatusOwner(mock)
	mock.ExpectQuery(regexp.QuoteMeta(timelineStatusQuery)).WithArgs("request-id", "DELIVERED").WillReturnError(sql.ErrNoRows)
	expectTimelineStatusItems(mock)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	v, err := NewService(r).TimelineStatus(ctx, "owner", " 261000023 ", " delivered ")
	if err != nil || v.Timeline != (StatusTimeline{}) || v.Items == nil || len(v.Items) != 4 || v.DocumentMobileStatusCode != "DELIVERED" {
		t.Fatalf("response=%+v, error=%v", v, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineStatusRejectsInvalidInputBeforeQuery(t *testing.T) {
	for _, test := range []struct{ requestNo, code string }{
		{"", "DOCUMENT_REVIEW"}, {"bad/request", "DOCUMENT_REVIEW"}, {"123456789012345678901", "DOCUMENT_REVIEW"}, {"261000023", ""}, {"261000023", "PAYMENT_PENDING"}, {"261000023", "UNSUPPORTED"},
	} {
		_, err := NewService(nil).TimelineStatus(context.Background(), "owner", test.requestNo, test.code)
		var business *BusinessError
		if !errors.As(err, &business) || business.Code != "MA00005" {
			t.Fatalf("input=%+v error=%v", test, err)
		}
	}
}

func TestTimelineStatusOwnershipAndQueryFailures(t *testing.T) {
	t.Run("ownership-or-inactive", func(t *testing.T) {
		r, mock, closeDB := newMockRenewalRepository(t)
		defer closeDB()
		mock.ExpectQuery(`SELECT r.id,r.request_no,.*WHERE r.request_no=\? AND r.mobile_user_uuid=\? AND r.is_active='YES'$`).
			WithArgs("261000023", "other-user").WillReturnError(sql.ErrNoRows)
		if _, err := NewService(r).TimelineStatus(context.Background(), "other-user", "261000023", "DOCUMENT_REVIEW"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("error=%v, want ErrNotFound", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("history-read-failure", func(t *testing.T) {
		r, mock, closeDB := newMockRenewalRepository(t)
		defer closeDB()
		expectTimelineStatusOwner(mock)
		failure := errors.New("history query failed")
		mock.ExpectQuery(regexp.QuoteMeta(timelineStatusQuery)).WithArgs("request-id", "DOCUMENT_REVIEW").WillReturnError(failure)
		if _, err := NewService(r).TimelineStatus(context.Background(), "owner", "261000023", "DOCUMENT_REVIEW"); !errors.Is(err, failure) {
			t.Fatalf("error=%v, want query failure", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

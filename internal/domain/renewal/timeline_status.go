package renewal

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type StatusTimeline struct {
	DateTime string `json:"date_time"`
	Note     string `json:"note"`
}

type TimelineStatusResponse struct {
	RequestNo                string         `json:"requestNo"`
	DocumentMobileStatusCode string         `json:"documentMobileStatusCode"`
	Timeline                 StatusTimeline `json:"timeline"`
	Items                    []ValidateItem `json:"items"`
}

func (r *Repository) timelineStatus(ctx context.Context, userID, no, mobileStatusCode string) (TimelineStatusResponse, error) {
	row, err := r.ownedRequest(ctx, userID, "request_no", no)
	if err != nil {
		return TimelineStatusResponse{}, err
	}
	response := TimelineStatusResponse{
		RequestNo: row.No, DocumentMobileStatusCode: mobileStatusCode,
		Items: []ValidateItem{},
	}
	var actionedAt time.Time
	err = r.db.QueryRowContext(ctx, `SELECT t.actioned_at,COALESCE(t.note,'') FROM m_document_transaction t INNER JOIN m_document_status s ON s.document_status_code=t.to_status WHERE t.request_id=? AND s.document_mobile_status_code=? ORDER BY t.actioned_at DESC,t.id DESC LIMIT 1`, row.ID, mobileStatusCode).
		Scan(&actionedAt, &response.Timeline.Note)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return TimelineStatusResponse{}, err
	}
	if err == nil {
		response.Timeline.DateTime = actionedAt.Format(time.RFC3339)
	}
	response.Items, err = r.requestItems(ctx, row, row.Code)
	if err != nil {
		return TimelineStatusResponse{}, err
	}
	return response, nil
}

package paymentmysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	domain "github.com/smartseaman/mobile-api/internal/domain/payment"
)

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) *Store { return &Store{db: db} }

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

const requestSQL = `SELECT r.id,r.request_no,r.mobile_user_uuid,s.document_status_code,ROUND(r.amount*100),r.is_active='YES' FROM m_document_request r INNER JOIN m_document_status s ON s.id=r.document_status_id WHERE r.id=?`

func readRequest(ctx context.Context, q queryer, id string, lock bool) (domain.Request, error) {
	var r domain.Request
	query := requestSQL
	if lock {
		query += " FOR UPDATE"
	}
	err := q.QueryRowContext(ctx, query, id).Scan(&r.ID, &r.No, &r.UserID, &r.Status, &r.Amount, &r.Active)
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrNotFound
	}
	return r, err
}
func (s *Store) Request(ctx context.Context, id, user string) (domain.Request, error) {
	r, err := readRequest(ctx, s.db, id, false)
	if err == nil && r.UserID != user {
		return domain.Request{}, domain.ErrNotFound
	}
	return r, err
}

const attemptSQL = `SELECT a.id,a.transaction_no,a.request_id,r.request_no,r.mobile_user_uuid,a.idempotency_key,a.fingerprint,a.provider_ref_id,a.amount_satang,a.input_json,a.charge_json,a.status,a.creation_state,a.is_legacy,(a.review_required OR COALESCE(s.review_required,0) OR r.is_active<>'YES'),a.created_at,a.sent_at,a.last_checked_at,a.failures,(s.paid_attempt_id IS NOT NULL) FROM mobile_payment_attempts a INNER JOIN m_document_request r ON r.id=a.request_id LEFT JOIN mobile_payment_request_state s ON s.request_id=a.request_id WHERE `

func readAttempt(ctx context.Context, q queryer, condition string, args ...any) (domain.Attempt, error) {
	var a domain.Attempt
	var input string
	var charge sql.NullString
	var checked, sent sql.NullTime
	err := q.QueryRowContext(ctx, attemptSQL+condition, args...).Scan(&a.ID, &a.No, &a.RequestID, &a.RequestNo, &a.UserID, &a.Key, &a.Fingerprint, &a.Reference, &a.Amount, &input, &charge, &a.Status, &a.Creation, &a.Legacy, &a.Review, &a.CreatedAt, &sent, &checked, &a.Failures, &a.Paid)
	if errors.Is(err, sql.ErrNoRows) {
		return a, domain.ErrNotFound
	}
	if err != nil {
		return a, err
	}
	if err = json.Unmarshal([]byte(input), &a.Input); err != nil {
		return a, err
	}
	if charge.Valid {
		if err = json.Unmarshal([]byte(charge.String), &a.Charge); err != nil {
			return a, err
		}
	}
	a.SentAt = sent.Time
	a.CheckedAt = checked.Time
	return a, nil
}
func (s *Store) Get(ctx context.Context, id string) (domain.Attempt, error) {
	return readAttempt(ctx, s.db, "a.id=?", id)
}
func (s *Store) ByKey(ctx context.Context, id, key string) (domain.Attempt, error) {
	return readAttempt(ctx, s.db, "a.request_id=? AND a.idempotency_key=?", id, key)
}
func (s *Store) Current(ctx context.Context, id string) (domain.Attempt, error) {
	return readAttempt(ctx, s.db, `a.request_id=? ORDER BY (a.id=s.paid_attempt_id) DESC,(a.id=s.active_attempt_id) DESC,a.created_at DESC,a.id DESC LIMIT 1`, id)
}
func (s *Store) Find(ctx context.Context, c domain.Charge) (domain.Attempt, error) {
	return readAttempt(ctx, s.db, "a.provider_charge_id=? OR a.provider_ref_id=? LIMIT 1", c.ID, c.Reference)
}
func (s *Store) Reserve(ctx context.Context, a domain.Attempt) (domain.Attempt, bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return a, false, err
	}
	defer tx.Rollback()
	req, err := readRequest(ctx, tx, a.RequestID, true)
	if err != nil {
		return a, false, err
	}
	if req.UserID != a.UserID {
		return a, false, domain.ErrNotFound
	}
	replay, err := readAttempt(ctx, tx, "a.request_id=? AND a.idempotency_key=?", a.RequestID, a.Key)
	if err == nil {
		if (replay.Fingerprint != "" && replay.Fingerprint != a.Fingerprint) || (replay.Fingerprint == "" && (replay.Input.PaymentMethod != a.Input.PaymentMethod || replay.Input.ReturnURI != a.Input.ReturnURI || replay.Input.PlatformType != a.Input.PlatformType)) {
			return a, false, domain.Conflict("PAYMENT_IDEMPOTENCY_CONFLICT")
		}
		return replay, false, tx.Commit()
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return a, false, err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO mobile_payment_request_state(request_id) VALUES(?) ON DUPLICATE KEY UPDATE request_id=VALUES(request_id)", a.RequestID); err != nil {
		return a, false, err
	}
	var active, paid sql.NullString
	var review bool
	err = tx.QueryRowContext(ctx, "SELECT active_attempt_id,paid_attempt_id,review_required FROM mobile_payment_request_state WHERE request_id=? FOR UPDATE", a.RequestID).Scan(&active, &paid, &review)
	if err != nil {
		return a, false, err
	}
	if review {
		return a, false, domain.Conflict("PAYMENT_RECONCILIATION_REQUIRED")
	}
	if paid.Valid {
		return a, false, domain.Conflict("PAYMENT_ALREADY_PAID")
	}
	if !req.Active || req.Status != "PAYMENT_PENDING" || req.Amount != a.Amount || req.Amount <= 0 {
		return a, false, domain.Conflict("PAYMENT_REQUEST_NOT_PAYABLE")
	}
	if active.Valid {
		var status string
		if err = tx.QueryRowContext(ctx, "SELECT status FROM mobile_payment_attempts WHERE id=?", active.String).Scan(&status); err != nil {
			return a, false, err
		}
		if status != "EXPIRED" && status != "FAILED" {
			return a, false, domain.Conflict("PAYMENT_IN_PROGRESS")
		}
	}
	var incomplete int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM m_document_request_items i INNER JOIN m_document_master_request_item m ON m.document_master_items_code=i.document_master_request_item_code WHERE i.request_id=? AND m.storage_scope='REQUEST' AND NOT EXISTS (SELECT 1 FROM m_document_request_item_files f WHERE f.request_item_id=i.id AND f.file_uploaded=1 AND (f.check_result IS NULL OR f.check_result<>'fix'))`, a.RequestID).Scan(&incomplete)
	if err != nil {
		return a, false, err
	}
	if incomplete > 0 {
		return a, false, domain.Conflict("PAYMENT_DOCUMENTS_INCOMPLETE")
	}
	input, err := json.Marshal(a.Input)
	if err != nil {
		return a, false, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO mobile_payment_attempts(id,transaction_no,request_id,idempotency_key,fingerprint,provider_ref_id,amount_satang,input_json,status,creation_state,created_at,next_check_at) VALUES(?,?,?,?,?,?,?,?,'PROCESSING','RESERVED',?,NOW(6))`, a.ID, a.No, a.RequestID, a.Key, a.Fingerprint, a.Reference, a.Amount, string(input), a.CreatedAt)
	if err != nil {
		return a, false, err
	}
	if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET active_attempt_id=? WHERE request_id=?", a.ID, a.RequestID); err != nil {
		return a, false, err
	}
	return a, true, tx.Commit()
}
func (s *Store) Acquire(ctx context.Context, id, owner string, force bool) (bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE mobile_payment_attempts SET lease_owner=?,lease_until=DATE_ADD(NOW(6),INTERVAL 60 SECOND) WHERE id=? AND (lease_until IS NULL OR lease_until<NOW(6)) AND review_required=0 AND status<>'SUCCESS' AND (last_checked_at IS NULL OR last_checked_at<=DATE_SUB(NOW(6),INTERVAL 5 SECOND)) AND (? OR next_check_at<=NOW(6))`, owner, id, force)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}
func (s *Store) MarkSent(ctx context.Context, a domain.Attempt) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	req, err := readRequest(ctx, tx, a.RequestID, true)
	if err != nil {
		return false, err
	}
	if !req.Active || req.Status != "PAYMENT_PENDING" || req.Amount != a.Amount {
		_, err = tx.ExecContext(ctx, "UPDATE mobile_payment_attempts SET review_required=1,lease_owner=NULL,lease_until=NULL WHERE id=? AND lease_owner=?", a.ID, a.Lease)
		if err != nil {
			return false, err
		}
		return false, tx.Commit()
	}
	result, err := tx.ExecContext(ctx, `UPDATE mobile_payment_attempts a INNER JOIN mobile_payment_request_state s ON s.request_id=a.request_id SET a.creation_state='SENT',a.sent_at=NOW(6) WHERE a.id=? AND a.lease_owner=? AND a.creation_state='RESERVED' AND s.active_attempt_id=a.id AND s.paid_attempt_id IS NULL AND s.review_required=0`, a.ID, a.Lease)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, tx.Commit()
}
func (s *Store) Retry(ctx context.Context, a domain.Attempt, review bool) error {
	seconds := 30
	for n := 0; n < a.Failures && seconds < 300; n++ {
		seconds *= 2
	}
	if seconds > 300 {
		seconds = 300
	}
	seconds += int(time.Now().UnixNano()%5000) / 1000
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = readRequest(ctx, tx, a.RequestID, true); err != nil {
		return err
	}
	if review {
		if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET review_required=1 WHERE request_id=?", a.RequestID); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE mobile_payment_attempts SET failures=failures+1,review_required=(review_required OR ?),last_checked_at=NOW(6),next_check_at=TIMESTAMPADD(SECOND,?,NOW(6)),lease_owner=NULL,lease_until=NULL WHERE id=? AND (?='' OR lease_owner=?)`, review, seconds, a.ID, a.Lease, a.Lease)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func (s *Store) Due(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM mobile_payment_attempts WHERE review_required=0 AND status<>'SUCCESS' AND next_check_at<=NOW(6) AND (lease_until IS NULL OR lease_until<NOW(6)) AND (status NOT IN ('EXPIRED','FAILED') OR created_at>DATE_SUB(NOW(6),INTERVAL 24 HOUR)) ORDER BY next_check_at,id LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) Apply(ctx context.Context, a domain.Attempt, c domain.Charge) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	req, err := readRequest(ctx, tx, a.RequestID, true)
	if err != nil {
		return err
	}
	current, err := readAttempt(ctx, tx, "a.id=?", a.ID)
	if err != nil {
		return err
	}
	if current.Charge.ID != "" && current.Charge.ID != c.ID {
		if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_attempts SET review_required=1 WHERE id=?", a.ID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET review_required=1 WHERE request_id=?", a.RequestID); err != nil {
			return err
		}
		return tx.Commit()
	}
	// Request locks serialize webhook, polling and worker writes. Success is final.
	status := domain.Status(c)
	if current.Status == "SUCCESS" || ((current.Status == "EXPIRED" || current.Status == "FAILED") && status != "SUCCESS") {
		status = current.Status
		c = current.Charge
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return err
	}
	var paid sql.NullString
	var requestReview bool
	err = tx.QueryRowContext(ctx, "SELECT paid_attempt_id,review_required FROM mobile_payment_request_state WHERE request_id=? FOR UPDATE", a.RequestID).Scan(&paid, &requestReview)
	if err != nil {
		return err
	}
	review := current.Review
	if status == "SUCCESS" {
		if paid.Valid && paid.String != a.ID {
			review = true
			requestReview = true
		}
		if !paid.Valid {
			if req.Active && req.Status == "PAYMENT_PENDING" && req.Amount == current.Amount && !review && !requestReview {
				var nextStatus string
				if err = tx.QueryRowContext(ctx, "SELECT id FROM m_document_status WHERE document_status_code='PENDING_DOCUMENT_REVIEW' AND is_active='YES' LIMIT 1").Scan(&nextStatus); err != nil {
					return err
				}
				if _, err = tx.ExecContext(ctx, "UPDATE m_document_request SET document_status_id=?,updated_at=NOW() WHERE id=?", nextStatus, a.RequestID); err != nil {
					return err
				}
				if _, err = tx.ExecContext(ctx, `INSERT INTO m_document_transaction(id,request_id,action,from_status,to_status,note,actioned_by) VALUES(?,?,'PAYMENT_SUCCESS','PAYMENT_PENDING','PENDING_DOCUMENT_REVIEW','Payment completed',?)`, uuid.NewString(), a.RequestID, req.UserID); err != nil {
					return err
				}
			} else {
				review = true
				requestReview = true
			}
			if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET paid_attempt_id=? WHERE request_id=?", a.ID, a.RequestID); err != nil {
				return err
			}
		}
	}
	if requestReview {
		if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET review_required=1 WHERE request_id=?", a.RequestID); err != nil {
			return err
		}
	}
	// Keep the existing ledger as the accounting record; reservations live in the sidecar.
	var exists int
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM m_payment_transaction WHERE id=?", a.ID).Scan(&exists); err != nil {
		return err
	}
	channel := "PROMPTPAY"
	if len(a.Input.PaymentMethod) > 15 && a.Input.PaymentMethod[:15] == "mobile_banking_" {
		channel = "MOBILE_BANKING"
	}
	if exists == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO m_payment_transaction(id,request_id,transaction_no,transaction_type,channel,payment_method,amount,currency,status,provider,provider_charge_id,provider_source_id,provider_transaction_id,provider_status,provider_response,idempotency_key,description,return_uri,authorize_uri,failure_code,failure_message,is_livemode,expires_at) VALUES(?,?,?,'CHARGE',?,?,?,'THB',?,'OMISE',?,?,?,?,?,?,?,?,?,?,?,?,?)`, a.ID, a.RequestID, a.No, channel, a.Input.PaymentMethod, float64(a.Amount)/100, status, c.ID, c.SourceID, c.TransactionID, c.Status, c.Raw, a.Key, "Document renewal "+req.No, c.ReturnURI, c.AuthorizeURI, c.FailureCode, c.FailureMessage, c.Live, nullableTime(c.ExpiresAt))
		if err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE m_payment_transaction SET status=?,provider_charge_id=?,provider_source_id=?,provider_transaction_id=?,provider_status=?,provider_response=?,return_uri=?,authorize_uri=?,failure_code=?,failure_message=?,is_livemode=?,expires_at=COALESCE(?,expires_at),paid_at=CASE WHEN ?='SUCCESS' THEN COALESCE(paid_at,NOW()) ELSE paid_at END,failed_at=CASE WHEN ?='FAILED' THEN COALESCE(failed_at,NOW()) ELSE failed_at END,updated_at=NOW() WHERE id=?`, status, c.ID, c.SourceID, c.TransactionID, c.Status, c.Raw, c.ReturnURI, c.AuthorizeURI, c.FailureCode, c.FailureMessage, c.Live, nullableTime(c.ExpiresAt), status, status, a.ID)
	if err != nil {
		return err
	}
	delay := 5
	if status == "EXPIRED" || status == "FAILED" {
		delay = 300
	}
	_, err = tx.ExecContext(ctx, `UPDATE mobile_payment_attempts SET provider_charge_id=?,charge_json=?,status=?,creation_state='CONFIRMED',review_required=?,failures=0,last_checked_at=NOW(6),next_check_at=TIMESTAMPADD(SECOND,?,NOW(6)),lease_owner=NULL,lease_until=NULL WHERE id=?`, c.ID, string(raw), status, review, delay, a.ID)
	if err != nil {
		return err
	}
	if status == "SUCCESS" || status == "EXPIRED" || status == "FAILED" {
		if _, err = tx.ExecContext(ctx, "UPDATE mobile_payment_request_state SET active_attempt_id=NULL WHERE request_id=? AND active_attempt_id=?", a.RequestID, a.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func nullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}

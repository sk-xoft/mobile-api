package payment

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store    Store
	provider Provider
	live     bool
	now      func() time.Time
}

func NewService(store Store, provider Provider, live bool) *Service {
	return &Service{store: store, provider: provider, live: live, now: time.Now}
}
func fingerprint(in Input) string {
	in.IdempotencyKey = ""
	data, _ := json.Marshal(in)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func Status(c Charge) string {
	if c.Status == "successful" {
		return "SUCCESS"
	}
	if c.Status == "expired" || c.Expired {
		return "EXPIRED"
	}
	if c.Status == "failed" {
		return "FAILED"
	}
	if c.Status == "pending" {
		return "PENDING"
	}
	return "PROCESSING"
}
func (s *Service) response(a Attempt) Response {
	now := s.now().UTC()
	c := a.Charge
	expired := a.Status == "EXPIRED" || c.Expired || (!c.ExpiresAt.IsZero() && !now.Before(c.ExpiresAt))
	p := Response{RequestID: a.RequestID, RequestNo: a.RequestNo, TransactionID: a.ID, TransactionNo: a.No, Status: a.Status, Channel: "PROMPTPAY", PaymentMethod: a.Input.PaymentMethod, Amount: float64(a.Amount) / 100, Currency: "THB", Provider: "OMISE", ProviderChargeID: c.ID, AuthorizeURI: c.AuthorizeURI, ReturnURI: c.ReturnURI, QRCodeURI: c.QRCodeURI, FailureCode: c.FailureCode, FailureMessage: c.FailureMessage, ServerTime: now.Format(time.RFC3339), QRExpired: expired, RetryAfterSeconds: 5, NextAction: "CHECK_PAYMENT"}
	if strings.HasPrefix(a.Input.PaymentMethod, "mobile_banking_") {
		p.Channel = "MOBILE_BANKING"
	}
	if !c.ExpiresAt.IsZero() {
		p.ExpiresAt = c.ExpiresAt.UTC().Format(time.RFC3339)
	}
	switch {
	case a.Review:
		p.NextAction = "CONTACT_SUPPORT"
		p.RetryAfterSeconds = 0
	case a.Status == "SUCCESS" || a.Paid:
		p.NextAction = "COMPLETE"
		p.RetryAfterSeconds = 0
	case a.Status == "EXPIRED" || a.Status == "FAILED":
		p.NextAction = "CREATE_NEW_QR"
		p.RetryAfterSeconds = 0
	case a.Status == "PENDING" && !expired:
		p.NextAction = "WAIT_FOR_PAYMENT"
	default:
		p.Status = "PROCESSING"
	}
	if expired || a.Review || a.Paid || a.Status == "SUCCESS" || a.Status == "FAILED" || a.Status == "EXPIRED" {
		p.QRCodeURI = ""
		p.AuthorizeURI = ""
	}
	return p
}
func (s *Service) Create(ctx context.Context, userID, requestID string, in Input) (Response, error) {
	in.PaymentMethod = strings.ToLower(strings.TrimSpace(in.PaymentMethod))
	in.IdempotencyKey = strings.TrimSpace(in.IdempotencyKey)
	if in.IdempotencyKey == "" || len(in.IdempotencyKey) > 100 || (in.PaymentMethod != "promptpay" && !strings.HasPrefix(in.PaymentMethod, "mobile_banking_")) {
		return Response{}, Conflict("PAYMENT_INVALID_INPUT")
	}
	if !s.provider.Ready() {
		return Response{}, Conflict("PAYMENT_UNAVAILABLE")
	}
	request, err := s.store.Request(ctx, requestID, userID)
	if err != nil {
		return Response{}, err
	}
	requestID = request.ID
	existing, err := s.store.ByKey(ctx, requestID, in.IdempotencyKey)
	if err == nil {
		expected := existing.Fingerprint
		if expected == "" {
			expected = fingerprint(existing.Input)
		}
		if expected != fingerprint(in) {
			return Response{}, Conflict("PAYMENT_IDEMPOTENCY_CONFLICT")
		}
		return s.refreshResponse(ctx, existing, true)
	}
	if !errors.Is(err, ErrNotFound) {
		return Response{}, err
	}
	// Reconcile the previous attempt before deciding whether a replacement is safe.
	current, err := s.store.Current(ctx, requestID)
	if err == nil {
		if _, err = s.reconcile(ctx, current, true); err != nil {
			return Response{}, err
		}
	} else if !errors.Is(err, ErrNotFound) {
		return Response{}, err
	}
	id := uuid.NewString()
	a := Attempt{ID: id, No: "PAY" + strings.ToUpper(strings.ReplaceAll(id, "-", ""))[:20], RequestID: requestID, RequestNo: request.No, UserID: userID, Key: in.IdempotencyKey, Fingerprint: fingerprint(in), Reference: id, Status: "PROCESSING", Creation: "RESERVED", Input: in, Amount: request.Amount, CreatedAt: s.now().UTC()}
	reserved, created, err := s.store.Reserve(ctx, a)
	if err != nil {
		var conflict *Error
		if errors.As(err, &conflict) && conflict.Code == "PAYMENT_IN_PROGRESS" {
			if cur, e := s.store.Current(ctx, requestID); e == nil {
				p := s.response(cur)
				conflict.Payment = &p
			}
		}
		return Response{}, err
	}
	return s.refreshResponse(ctx, reserved, created)
}
func (s *Service) Current(ctx context.Context, userID, requestID string) (Response, error) {
	request, err := s.store.Request(ctx, requestID, userID)
	if err != nil {
		return Response{}, err
	}
	a, err := s.store.Current(ctx, request.ID)
	if err != nil {
		return Response{}, err
	}
	return s.refreshResponse(ctx, a, true)
}
func (s *Service) Get(ctx context.Context, userID, requestID, transactionID string) (Response, error) {
	request, err := s.store.Request(ctx, requestID, userID)
	if err != nil {
		return Response{}, err
	}
	requestID = request.ID
	a, err := s.store.Get(ctx, transactionID)
	if err != nil {
		return Response{}, err
	}
	if a.RequestID != requestID || a.UserID != userID {
		return Response{}, ErrNotFound
	}
	return s.refreshResponse(ctx, a, true)
}
func (s *Service) refreshResponse(ctx context.Context, a Attempt, force bool) (Response, error) {
	updated, err := s.reconcile(ctx, a, force)
	if err != nil {
		return Response{}, err
	}
	return s.response(updated), nil
}
func (s *Service) reconcile(ctx context.Context, a Attempt, force bool) (Attempt, error) {
	if a.Status == "SUCCESS" {
		return a, nil
	}
	token := uuid.NewString()
	ok, err := s.store.Acquire(ctx, a.ID, token, force)
	if err != nil || !ok {
		return a, err
	}
	a, err = s.store.Get(ctx, a.ID)
	if err != nil {
		return a, err
	}
	a.Lease = token
	var charge Charge
	if a.Creation == "RESERVED" {
		sent, e := s.store.MarkSent(ctx, a)
		if e != nil {
			return a, e
		}
		if !sent {
			return a, nil
		}
		charge, err = s.provider.Create(ctx, a)
	} else if a.Charge.ID != "" {
		charge, err = s.provider.Retrieve(ctx, a.Charge.ID)
	} else {
		charge, err = s.provider.Find(ctx, a)
	}
	if err != nil {
		// A timeout never proves that no charge was created. Never POST again here.
		if e := s.store.Retry(ctx, a, false); e != nil {
			return a, e
		}
		if s.now().Sub(a.CreatedAt) > 15*time.Minute {
			zap.L().Error("payment reconciliation overdue", zap.String("transaction_id", a.ID))
		}
		return s.store.Get(ctx, a.ID)
	}
	if err = s.validate(a, charge); err != nil {
		zap.L().Error("payment verification rejected", zap.String("transaction_id", a.ID), zap.Error(err))
		if e := s.store.Retry(ctx, a, true); e != nil {
			return a, e
		}
		return s.store.Get(ctx, a.ID)
	}
	if err = s.store.Apply(ctx, a, charge); err != nil {
		return a, err
	}
	return s.store.Get(ctx, a.ID)
}
func (s *Service) validate(a Attempt, c Charge) error {
	if c.ID == "" || c.Amount != a.Amount || c.Currency != "THB" || c.Live != s.live {
		return fmt.Errorf("charge identity, amount, currency or mode mismatch")
	}
	if a.Charge.ID != "" && a.Charge.ID != c.ID {
		return fmt.Errorf("charge ID mismatch")
	}
	if !a.Legacy && c.Reference != a.Reference {
		return fmt.Errorf("charge reference mismatch")
	}
	if c.RequestNo != "" && c.RequestNo != a.RequestNo {
		return fmt.Errorf("request reference mismatch")
	}
	return nil
}
func (s *Service) Webhook(ctx context.Context, raw, signature, timestamp string) error {
	if !s.provider.VerifySignature(raw, signature, timestamp) {
		return Conflict("OMISE_INVALID_SIGNATURE")
	}
	var event struct {
		Key  string `json:"key"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return err
	}
	if event.Key != "charge.complete" && event.Key != "charge.create" && event.Key != "charge.update" && event.Key != "charge.expire" {
		return nil
	}
	if event.Data.ID == "" {
		return Conflict("PAYMENT_INVALID_INPUT")
	}
	charge, err := s.provider.Retrieve(ctx, event.Data.ID)
	if err != nil {
		return err
	}
	a, err := s.store.Find(ctx, charge)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err = s.validate(a, charge); err != nil {
		return s.store.Retry(ctx, a, true)
	}
	return s.store.Apply(ctx, a, charge)
}

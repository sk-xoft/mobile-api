package renewal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

func (s *Service) CreateRequest(ctx context.Context, userID string, input CreateRequest) (CreateResponse, error) {
	if strings.TrimSpace(input.DocumentCode) == "" || strings.TrimSpace(input.DeliveryAddressID) == "" {
		return CreateResponse{}, &BusinessError{Code: "MA00005", Detail: "documentCode/deliveryAddressId"}
	}
	price, e := s.Price(ctx, input.DocumentCode)
	if e != nil {
		return CreateResponse{}, e
	}
	return s.repository.create(ctx, userID, input, price)
}

func (s *Service) Validate(ctx context.Context, userID, documentCode, idempotencyKey string) (ValidateResponse, error) {
	documentCode = strings.ToUpper(strings.TrimSpace(documentCode))
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if documentCode == "" || len(documentCode) > 10 {
		return ValidateResponse{}, &BusinessError{Code: "MA00005", Detail: "documentCode"}
	}
	if idempotencyKey != "" && len(idempotencyKey) > 100 {
		return ValidateResponse{}, &BusinessError{Code: "MA00005", Detail: "idempotencyKey"}
	}
	if idempotencyKey != "" {
		if existing, lookupErr := s.repository.validateExisting(ctx, userID, "idempotency_key", idempotencyKey); lookupErr == nil {
			if existing.Code != documentCode {
				return ValidateResponse{}, &BusinessError{Code: "MA00005", Detail: "idempotencyKey"}
			}
			return s.repository.validateResponse(ctx, existing, documentCode, idempotencyKey)
		}
	}
	var active requestRow
	if latest, lookupErr := s.repository.latestActive(ctx, userID, documentCode); lookupErr == nil {
		active = latest
		return s.repository.validateResponse(ctx, active, documentCode, active.IdempotencyKey)
	}
	price, err := s.Price(ctx, documentCode)
	if err != nil {
		return ValidateResponse{}, err
	}
	return s.repository.validateCreate(ctx, userID, documentCode, idempotencyKey, price)
}
func (s *Service) ListRequests(ctx context.Context, userID string, offset int) (Page, error) {
	if offset < 0 {
		return Page{}, &BusinessError{Code: "MA00005", Detail: "offSet"}
	}
	items, total, e := s.repository.list(ctx, userID, offset)
	return Page{ItemTotal: total, Last: offset+len(items) >= total, Items: items}, e
}
func (s *Service) Resubmit(ctx context.Context, userID, no string) (TransitionResponse, error) {
	row, err := s.repository.ownedRequest(ctx, userID, "request_no", no)
	if err != nil {
		return TransitionResponse{}, err
	}
	if row.StatusCode != "PENDING_APPLICANT_CORRECTION" {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	fixCount, err := s.repository.countFixItems(ctx, row.ID)
	if err != nil {
		return TransitionResponse{}, err
	}
	if fixCount == 0 {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "documentRequestItems"}
	}
	incompleteCount, err := s.repository.countIncompleteFixItems(ctx, row.ID)
	if err != nil {
		return TransitionResponse{}, err
	}
	if incompleteCount > 0 {
		return TransitionResponse{}, &BusinessError{Code: "MA00004", Detail: "correctedDocumentItems"}
	}
	if err := s.repository.resetFixItems(ctx, row.ID); err != nil {
		return TransitionResponse{}, err
	}
	return s.repository.transition(ctx, userID, no, "PENDING_DOCUMENT_REVIEW", "RESUBMIT")
}
func (s *Service) DeleteRequest(ctx context.Context, userID, no string) (TransitionResponse, error) {
	row, err := s.repository.ownedRequest(ctx, userID, "request_no", no)
	if err != nil {
		return TransitionResponse{}, err
	}
	if row.StatusCode != "PAYMENT_PENDING" {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	return s.repository.transition(ctx, userID, no, "CANCELLED", "CANCEL")
}
func (s *Service) UpdateMobile(ctx context.Context, userID, no, mobile string) (TransitionResponse, error) {
	if len(mobile) != 10 || mobile[0] != '0' {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "mobileNumber"}
	}
	return s.repository.updateMobile(ctx, userID, no, mobile)
}
func (s *Service) Detail(ctx context.Context, userID, no string) (Detail, error) {
	return s.repository.detail(ctx, userID, no)
}
func (s *Service) Timeline(ctx context.Context, userID, no string) (Timeline, error) {
	return s.repository.timeline(ctx, userID, no)
}
func (s *Service) Stage(ctx context.Context, userID, no string) (StageResponse, error) {
	row, e := s.repository.ownedRequest(ctx, userID, "request_no", no)
	if e != nil {
		return StageResponse{}, e
	}
	codes := []string{"PAYMENT_PENDING", "PENDING_DOCUMENT_REVIEW", "PENDING_MARINE_DEPARTMENT_RESULT", "PENDING_DEPARTMENT_DOCUMENT_PICKUP", "DELIVERING", "DELIVERED"}
	stages := make([]Stage, 0, len(codes))
	for i, code := range codes {
		stages = append(stages, Stage{Code: code, Name: code, Step: i + 1, Current: code == row.StatusCode})
	}
	return StageResponse{RequestNo: no, CurrentStatus: row.StatusCode, Stages: stages}, nil
}
func (s *Service) Payment(ctx context.Context, userID, requestID string, input PaymentRequest) (PaymentResponse, error) {
	if strings.TrimSpace(input.PaymentMethod) == "" || strings.TrimSpace(input.IdempotencyKey) == "" {
		return PaymentResponse{}, &BusinessError{Code: "MA00005", Detail: "paymentMethod/idempotencyKey"}
	}
	row, err := s.repository.ownedRequest(ctx, userID, "id", requestID)
	if err != nil {
		return PaymentResponse{}, err
	}
	if row.StatusCode != "PAYMENT_PENDING" {
		return PaymentResponse{}, &BusinessError{Code: "MA00005", Detail: "documentStatus"}
	}
	incompleteCount, err := s.repository.countIncompleteRequestItems(ctx, requestID)
	if err != nil {
		return PaymentResponse{}, err
	}
	if incompleteCount > 0 {
		return PaymentResponse{}, &BusinessError{Code: "MA00004", Detail: "requestDocumentItems"}
	}
	return s.repository.payment(ctx, userID, requestID, input)
}
func (s *Service) PaymentStatus(ctx context.Context, userID, requestID, transactionID string) (PaymentResponse, error) {
	p, e := s.repository.paymentStatus(ctx, userID, requestID, transactionID)
	if errors.Is(e, context.Canceled) {
		return p, e
	}
	return p, e
}
func (s *Service) HardDelete(ctx context.Context, userID, requestNo string) (TransitionResponse, error) {
	requestNo = strings.ToUpper(strings.TrimSpace(requestNo))
	if requestNo == "" || len(requestNo) > 20 {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "requestNo"}
	}
	return s.repository.hardDelete(ctx, userID, requestNo)
}
func (s *Service) Webhook(ctx context.Context, raw, signature, timestamp string) error {
	if s.repository.Omise() == nil || !s.repository.Omise().VerifySignature(raw, signature, timestamp) {
		return &BusinessError{Code: "OMISE_INVALID_SIGNATURE", Detail: "signature"}
	}
	var payload struct {
		Key  string `json:"key"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.Key != "charge.complete" || payload.Data.ID == "" {
		return nil
	}
	charge, err := s.repository.Omise().RetrieveCharge(payload.Data.ID)
	if err != nil {
		return err
	}
	return s.repository.updateWebhook(ctx, charge)
}

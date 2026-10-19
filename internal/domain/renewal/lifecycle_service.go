package renewal

import (
	"context"
	"errors"
	payments "github.com/smartseaman/mobile-api/internal/domain/payment"
	"go.uber.org/zap"
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
		existing, lookupErr := s.repository.validateExisting(ctx, userID, "idempotency_key", idempotencyKey)
		if lookupErr != nil && !errors.Is(lookupErr, ErrNotFound) {
			return ValidateResponse{}, lookupErr
		}
		if lookupErr == nil {
			if existing.Code != documentCode {
				return ValidateResponse{}, &BusinessError{Code: "MA00005", Detail: "idempotencyKey"}
			}
			return s.repository.validateResponse(ctx, existing, documentCode, idempotencyKey)
		}
	}
	var active requestRow
	latest, lookupErr := s.repository.latestActive(ctx, userID, documentCode)
	if lookupErr != nil && !errors.Is(lookupErr, ErrNotFound) {
		return ValidateResponse{}, lookupErr
	}
	if lookupErr == nil {
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
	no = strings.ToUpper(strings.TrimSpace(no))
	if no == "" || len(no) > 20 || !validRequestNo(no) {
		return StageResponse{}, &BusinessError{Code: "MA00005", Detail: "requestNo"}
	}
	row, e := s.repository.ownedRequest(ctx, userID, "request_no", no)
	if e != nil {
		zap.L().Error("renewal stage request lookup failed", zap.String("request_no", no), zap.Error(e))
		return StageResponse{}, e
	}
	current, e := s.repository.stageCurrentStatus(ctx, row.ID)
	if e != nil {
		zap.L().Error("renewal stage current status lookup failed", zap.String("request_no", no), zap.Error(e))
		return StageResponse{}, e
	}
	statuses, e := s.repository.Statuses(ctx)
	if e != nil {
		zap.L().Error("renewal stage statuses lookup failed", zap.String("request_no", no), zap.Error(e))
		return StageResponse{}, e
	}
	ladder := make(map[int]statusRow, 5)
	for _, status := range statuses {
		if step := progress(status.MobileCode); step > 0 && status.MobileCode != "" {
			if _, exists := ladder[step]; !exists {
				ladder[step] = status
			}
		}
	}
	if current.DocumentStatusCode == "" {
		return StageResponse{}, &BusinessError{Code: "MA00013", Detail: "documentRenewalStatus"}
	}
	if len(ladder) != 5 {
		return StageResponse{}, &BusinessError{Code: "MA00013", Detail: "documentRenewalStageLadder"}
	}
	var currentStep *int
	if current.Step != nil {
		value := *current.Step
		currentStep = &value
	}
	stages := make([]Stage, 0, 5)
	for step := 1; step <= 5; step++ {
		status, exists := ladder[step]
		if !exists || status.MobileCode == "" {
			return StageResponse{}, &BusinessError{Code: "MA00013", Detail: "documentRenewalStageLadder"}
		}
		state := "PENDING"
		if currentStep != nil {
			if step < *currentStep {
				state = "DONE"
			} else if step == *currentStep {
				state = "CURRENT"
			}
		}
		stages = append(stages, Stage{Step: step, DocumentMobileStatusCode: status.MobileCode, NameTh: status.MobileNameTh, NameEn: status.MobileNameEn, State: state})
	}
	return StageResponse{RequestNo: row.No, CurrentStatus: current, Stages: stages}, nil
}

func validRequestNo(value string) bool {
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return true
}
func (s *Service) Payment(ctx context.Context, userID, requestID string, input PaymentRequest) (PaymentResponse, error) {
	if s.payments == nil {
		return PaymentResponse{}, payments.Conflict("PAYMENT_UNAVAILABLE")
	}
	return s.payments.Create(ctx, userID, requestID, input)
}
func (s *Service) PaymentStatus(ctx context.Context, userID, requestID, transactionID string) (PaymentResponse, error) {
	if s.payments == nil {
		return PaymentResponse{}, payments.Conflict("PAYMENT_UNAVAILABLE")
	}
	return s.payments.Get(ctx, userID, requestID, transactionID)
}
func (s *Service) CurrentPayment(ctx context.Context, userID, requestID string) (PaymentResponse, error) {
	if s.payments == nil {
		return PaymentResponse{}, payments.Conflict("PAYMENT_UNAVAILABLE")
	}
	return s.payments.Current(ctx, userID, requestID)
}
func (s *Service) HardDelete(ctx context.Context, userID, requestNo string) (TransitionResponse, error) {
	requestNo = strings.ToUpper(strings.TrimSpace(requestNo))
	if requestNo == "" || len(requestNo) > 20 {
		return TransitionResponse{}, &BusinessError{Code: "MA00005", Detail: "requestNo"}
	}
	return s.repository.hardDelete(ctx, userID, requestNo)
}
func (s *Service) Webhook(ctx context.Context, raw, signature, timestamp string) error {
	if s.payments == nil {
		return payments.Conflict("PAYMENT_UNAVAILABLE")
	}
	return s.payments.Webhook(ctx, raw, signature, timestamp)
}

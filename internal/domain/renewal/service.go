package renewal

import (
	"context"
	"database/sql"
	payments "github.com/smartseaman/mobile-api/internal/domain/payment"
	"strings"
)

type Service struct {
	repository repositoryPort
	payments   *payments.Service
}

type repositoryPort interface {
	Statuses(context.Context) ([]statusRow, error)
	stageCurrentStatus(context.Context, string) (StageCurrentStatus, error)
	Documents(context.Context) ([]Document, error)
	Price(context.Context, string) (Price, error)
	create(context.Context, string, CreateRequest, Price) (CreateResponse, error)
	validateExisting(context.Context, string, string, string) (requestRow, error)
	ownedRequest(context.Context, string, string, string) (requestRow, error)
	latestActive(context.Context, string, string) (requestRow, error)
	validateResponse(context.Context, requestRow, string, string) (ValidateResponse, error)
	validateCreate(context.Context, string, string, string, Price) (ValidateResponse, error)
	list(context.Context, string, int) ([]Summary, int, error)
	transition(context.Context, string, string, string, string) (TransitionResponse, error)
	countIncompleteRequestItems(context.Context, string) (int, error)
	countFixItems(context.Context, string) (int, error)
	countIncompleteFixItems(context.Context, string) (int, error)
	resetFixItems(context.Context, string) error
	updateMobile(context.Context, string, string, string) (TransitionResponse, error)
	detail(context.Context, string, string) (Detail, error)
	timeline(context.Context, string, string) (Timeline, error)
	hardDelete(context.Context, string, string) (TransitionResponse, error)
}

func NewService(repository repositoryPort, paymentServices ...*payments.Service) *Service {
	service := &Service{repository: repository}
	if len(paymentServices) > 0 {
		service.payments = paymentServices[0]
	}
	return service
}

func (s *Service) Statuses(ctx context.Context) ([]Status, error) {
	rows, err := s.repository.Statuses(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Status, 0, len(rows))
	for _, row := range rows {
		item := Status{ID: row.ID, DocumentStatusCode: row.Code, NameTh: row.NameTh, NameEn: row.NameEn, CSSColor: row.CSSColor, Correction: row.NameEn == "Pending Applicant Correction", Terminal: row.NameEn == "Cancelled" || row.NameEn == "Delivered"}
		if step := progress(row.MobileCode); step > 0 {
			item.ProgressStep = &step
			item.MobileStatus = &MobileStatus{Code: row.MobileCode, NameTh: row.MobileNameTh, NameEn: row.MobileNameEn, Step: step}
		}
		result = append(result, item)
	}
	return result, nil
}
func (s *Service) Documents(ctx context.Context, language string) ([]Document, error) {
	result, err := s.repository.Documents(ctx)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(language, "en") {
		for i := range result {
			result[i].DocumentName = result[i].DocumentNameEn
		}
	}
	return result, nil
}
func (s *Service) Price(ctx context.Context, code string) (Price, error) {
	if strings.TrimSpace(code) == "" {
		return Price{}, &BusinessError{Code: "MA00005", Detail: "documentCode"}
	}
	result, err := s.repository.Price(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if err == sql.ErrNoRows {
		return Price{}, &BusinessError{Code: "MA00016", Detail: "documentCode"}
	}
	return result, err
}
func progress(code string) int {
	switch code {
	case "DOCUMENT_REVIEW":
		return 1
	case "MARINE_DEPARTMENT_RESULT":
		return 2
	case "DEPARTMENT_DOCUMENT_PICKUP":
		return 3
	case "DELIVERING":
		return 4
	case "DELIVERED":
		return 5
	}
	return 0
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return e.Code + ": " + e.Detail }

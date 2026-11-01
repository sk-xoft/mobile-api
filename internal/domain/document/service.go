package document

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smartseaman/mobile-api/internal/filecontent"
)

type Service struct {
	repository   repositoryPort
	storage      Storage
	pathTemplate string
}

type repositoryPort interface {
	FindUser(context.Context, string) (User, error)
	Certificate(context.Context, string, string) (Certificate, error)
	CreateCertificate(context.Context, string, string, CreateRequest, string) error
	UpdateCertificate(context.Context, string, string, UpdateRequest, string) error
	DeleteCertificate(context.Context, string, string) error
	Page(context.Context, string, string, int) ([]Item, int, error)
	CloseToExpiration(context.Context, string, int) ([]Item, int, error)
}

func NewService(repository repositoryPort, storage Storage, pathTemplate string) *Service {
	return &Service{repository: repository, storage: storage, pathTemplate: pathTemplate}
}

func (s *Service) Page(ctx context.Context, userID, documentType string, offset int) (PageResponse, error) {
	if offset < 0 {
		return PageResponse{}, &BusinessError{Code: "MA00005", Detail: "offSet"}
	}
	items, total, err := s.repository.Page(ctx, userID, documentType, offset)
	if err != nil {
		return PageResponse{}, err
	}
	for i := range items {
		setDisplayDates(&items[i])
		items[i].CertExpiredFlag = expired(items[i].CertEndDate)
	}
	return PageResponse{ItemTotal: total, Last: offset+len(items) >= total, Items: items}, nil
}

func (s *Service) CloseToExpiration(ctx context.Context, userID string, offset int) (PageResponse, error) {
	if offset < 0 {
		return PageResponse{}, &BusinessError{Code: "MA00005", Detail: "offSet"}
	}
	items, total, err := s.repository.CloseToExpiration(ctx, userID, offset)
	if err != nil {
		return PageResponse{}, err
	}
	for i := range items {
		setDisplayDates(&items[i])
		items[i].CertExpiredFlag = expired(items[i].CertEndDate)
	}
	return PageResponse{ItemTotal: total, Last: offset+len(items) >= total, Items: items}, nil
}

func (s *Service) Create(ctx context.Context, userID string, request CreateRequest) (Response, error) {
	if err := validateDates(request.CertStartDate, request.CertEndDate); err != nil {
		return Response{}, err
	}
	if request.DocumentCode == "" || request.FileCert == "" {
		return Response{}, &BusinessError{Code: "MA00004", Detail: "documentCode/fileCert"}
	}
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	if _, err := s.repository.Certificate(ctx, userID, request.DocumentCode); err == nil {
		return Response{}, &BusinessError{Code: "MA00017", Detail: request.DocumentCode}
	} else if !errors.Is(err, ErrNotFound) {
		return Response{}, err
	}
	content, contentType, err := decodeDocument(request.FileCert)
	if err != nil {
		return Response{}, err
	}
	fileName := uuid.NewString()
	if err := s.storage.Put(ctx, s.key(userID, fileName), contentType, content); err != nil {
		return Response{}, err
	}
	if err := s.repository.CreateCertificate(ctx, userID, user.Username, request, fileName); err != nil {
		return Response{}, err
	}
	return responseFromCreate(request), nil
}

func (s *Service) Update(ctx context.Context, userID string, request UpdateRequest) (Response, error) {
	if err := validateDates(request.CertStartDate, request.CertEndDate); err != nil {
		return Response{}, err
	}
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	if _, err := s.repository.Certificate(ctx, userID, request.DocumentCode); err != nil {
		if errors.Is(err, ErrNotFound) {
			return Response{}, &BusinessError{Code: "MA00016", Detail: request.DocumentCode}
		}
		return Response{}, err
	}
	fileName := ""
	if request.IsChangeFile == "Y" {
		content, contentType, decodeErr := decodeDocument(request.FileCert)
		if decodeErr != nil {
			return Response{}, decodeErr
		}
		fileName = uuid.NewString()
		if err := s.storage.Put(ctx, s.key(userID, fileName), contentType, content); err != nil {
			return Response{}, err
		}
	}
	if err := s.repository.UpdateCertificate(ctx, userID, user.Username, request, fileName); err != nil {
		return Response{}, err
	}
	return Response{DocumentCode: request.DocumentCode, CertStartDate: request.CertStartDate, CertEndDateType: request.CertEndDateType, CertEndDate: request.CertEndDate, FileCertName: request.FileCertName}, nil
}

func (s *Service) Delete(ctx context.Context, userID, documentCode string) (string, error) {
	if err := s.repository.DeleteCertificate(ctx, userID, documentCode); err != nil {
		return "", err
	}
	return "success", nil
}

func (s *Service) Edit(ctx context.Context, userID, documentCode string) (Response, error) {
	item, err := s.repository.Certificate(ctx, userID, documentCode)
	if err != nil {
		return Response{}, err
	}
	endType, endDate := "A", item.EndDate
	if endDate == "" {
		endType, endDate = "N", "9999-99-99"
	}
	return Response{DocumentCode: item.DocumentCode, CertStartDate: item.StartDate, CertEndDateType: endType, CertEndDate: endDate, FileCertName: item.OriginalFileName}, nil
}

func (s *Service) View(ctx context.Context, userID, documentCode string) ([]byte, string, error) {
	item, err := s.repository.Certificate(ctx, userID, documentCode)
	if err != nil {
		return nil, "", err
	}
	if item.FileName == "" {
		return nil, "", &BusinessError{Code: "MA00016", Detail: documentCode}
	}
	content, _, err := s.storage.Get(ctx, s.key(userID, item.FileName))
	if err != nil {
		return nil, "", err
	}
	return filecontent.DecodeStored(content)
}

func (s *Service) key(userID, fileName string) string {
	if strings.Contains(s.pathTemplate, "%s") {
		return fmt.Sprintf(s.pathTemplate, userID, fileName)
	}
	return strings.Trim(s.pathTemplate, "/") + "/" + userID + "/" + fileName
}

func responseFromCreate(request CreateRequest) Response {
	return Response{DocumentCode: request.DocumentCode, CertStartDate: request.CertStartDate, CertEndDateType: request.CertEndDateType, CertEndDate: request.CertEndDate, FileCertName: request.FileCertName}
}

func validateDates(start, end string) error {
	if _, err := time.Parse("2006-01-02", start); err != nil {
		return &BusinessError{Code: "MA00005", Detail: start}
	}
	if end != "" && end != "9999-99-99" {
		if _, err := time.Parse("2006-01-02", end); err != nil {
			return &BusinessError{Code: "MA00005", Detail: end}
		}
	}
	return nil
}

func expired(value string) string {
	if value == "" {
		return "N"
	}
	date, err := time.Parse("2006-01-02", value[:10])
	if err != nil || !date.Before(time.Now()) {
		return "N"
	}
	return "Y"
}

func setDisplayDates(item *Item) {
	if item.CertStartDate == "" && item.CertEndDate == "" {
		item.DisYear, item.DisMonth, item.DisDay = "", "", ""
		return
	}
	if item.CertEndDate == "" {
		item.DisYear, item.DisMonth, item.DisDay = "-", "-", "-"
		return
	}
	end, err := time.Parse("2006-01-02", item.CertEndDate[:10])
	if err != nil {
		item.DisYear, item.DisMonth, item.DisDay = "-", "-", "-"
		return
	}
	today := time.Now()
	years := end.Year() - today.Year()
	months := int(end.Month()) - int(today.Month())
	days := end.Day() - today.Day()
	if days < 0 {
		months--
	}
	if months < 0 {
		years--
		months += 12
	}
	if years < 0 {
		years, months, days = 0, 0, 0
	}
	item.DisYear, item.DisMonth, item.DisDay = fmt.Sprint(years), fmt.Sprint(months), fmt.Sprint(maxInt(days, 0))
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return e.Code + ": " + e.Detail }

func decodeDocument(value string) ([]byte, string, error) {
	contentType, encoded := "", strings.TrimSpace(value)
	if strings.HasPrefix(encoded, "data:") {
		comma, marker := strings.Index(encoded, ","), strings.Index(encoded, ";base64")
		if comma < 0 || marker < 0 || marker > comma {
			return nil, "", &BusinessError{Code: "MA00005", Detail: "fileCert"}
		}
		contentType, encoded = encoded[5:marker], encoded[comma+1:]
	}
	content, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(content) == 0 || len(content) > 10*1024*1024 {
		return nil, "", &BusinessError{Code: "MA00005", Detail: "fileCert"}
	}
	detected := ""
	if len(content) >= 8 && string(content[:8]) == "\x89PNG\r\n\x1a\n" {
		detected = "image/png"
	}
	if len(content) >= 3 && content[0] == 0xff && content[1] == 0xd8 && content[2] == 0xff {
		detected = "image/jpeg"
	}
	if len(content) >= 5 && string(content[:5]) == "%PDF-" {
		detected = "application/pdf"
	}
	if detected == "" || (contentType != "" && contentType != detected) {
		return nil, "", &BusinessError{Code: "MA00005", Detail: "fileCert"}
	}
	return content, detected, nil
}

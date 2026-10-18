package delivery

import (
	"context"
	"github.com/google/uuid"
	"strings"
)

type Service struct{ repository repositoryPort }

type repositoryPort interface {
	UserMobile(context.Context, string) (string, error)
	ValidAddress(context.Context, string, string, string, string) (bool, error)
	Default(context.Context, string) (Address, error)
	Find(context.Context, string, string) (Address, error)
	Count(context.Context, string) (int, error)
	ClearDefault(context.Context, string) error
	Create(context.Context, Address) error
	Update(context.Context, Address) error
	RenewalRequest(context.Context, string, string) (string, error)
	SnapshotExists(context.Context, string, string) (bool, error)
	InsertSnapshot(context.Context, string, string, string, Address) (string, error)
}

func NewService(repository repositoryPort) *Service { return &Service{repository: repository} }

func (s *Service) Default(ctx context.Context, userID string) (Response, error) {
	address, err := s.repository.Default(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	return toResponse(address), nil
}

func (s *Service) Create(ctx context.Context, userID string, request Request) (Response, error) {
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	mobile, err := s.repository.UserMobile(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	count, err := s.repository.Count(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	isDefault := count == 0 || (request.IsDefault != nil && *request.IsDefault)
	if isDefault {
		if err := s.repository.ClearDefault(ctx, userID); err != nil {
			return Response{}, err
		}
	}
	address := toAddress(uuid.NewString(), userID, request)
	address.IsDefault = isDefault
	address.Mobile = mobile
	if err := s.repository.Create(ctx, address); err != nil {
		return Response{}, err
	}
	return toResponse(address), nil
}

func (s *Service) Update(ctx context.Context, userID, addressID string, request Request) (Response, error) {
	if _, err := uuid.Parse(addressID); err != nil {
		return Response{}, &BusinessError{Code: "MA00008", Detail: "addressId"}
	}
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	current, err := s.repository.Find(ctx, userID, addressID)
	if err != nil {
		return Response{}, err
	}
	if current.IsDefault && (request.IsDefault == nil || !*request.IsDefault) {
		return Response{}, &BusinessError{Code: "MA00005", Detail: "isDefault"}
	}
	if request.IsDefault != nil && *request.IsDefault {
		if err := s.repository.ClearDefault(ctx, userID); err != nil {
			return Response{}, err
		}
	}
	address := toAddress(addressID, userID, request)
	address.IsDefault = request.IsDefault != nil && *request.IsDefault
	address.Mobile = current.Mobile
	if err := s.repository.Update(ctx, address); err != nil {
		return Response{}, err
	}
	return toResponse(address), nil
}

func (s *Service) CreateForRenewal(ctx context.Context, userID, requestNo string, request Request) (Response, error) {
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	requestID, err := s.repository.RenewalRequest(ctx, userID, requestNo)
	if err != nil {
		return Response{}, err
	}
	exists, err := s.repository.SnapshotExists(ctx, requestID, userID)
	if err != nil {
		return Response{}, err
	}
	if exists {
		return Response{}, &BusinessError{Code: "MA00017", Detail: "deliveryAddress"}
	}
	address := toAddress("", userID, request)
	mobile, err := s.repository.UserMobile(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	snapshotID, err := s.repository.InsertSnapshot(ctx, requestID, userID, mobile, address)
	if err != nil {
		return Response{}, err
	}
	address.ID, address.Mobile = snapshotID, mobile
	return toResponse(address), nil
}

func (s *Service) validate(ctx context.Context, request Request) error {
	if strings.TrimSpace(request.FirstName) == "" || strings.TrimSpace(request.LastName) == "" || strings.TrimSpace(request.AddressLine) == "" ||
		strings.TrimSpace(request.Province) == "" || strings.TrimSpace(request.District) == "" || strings.TrimSpace(request.SubDistrict) == "" ||
		len(request.PostalCode) != 5 || request.IsDefault == nil {
		return &BusinessError{Code: "MA00005", Detail: "address"}
	}
	valid, err := s.repository.ValidAddress(ctx, request.Province, request.District, request.SubDistrict, request.PostalCode)
	if err != nil {
		return err
	}
	if !valid {
		return &BusinessError{Code: "MA00005", Detail: "address"}
	}
	return nil
}

func toAddress(id, userID string, request Request) Address {
	return Address{ID: id, UserID: userID, FirstName: strings.TrimSpace(request.FirstName), LastName: strings.TrimSpace(request.LastName),
		AddressLine: strings.TrimSpace(request.AddressLine), Province: strings.TrimSpace(request.Province), District: strings.TrimSpace(request.District),
		SubDistrict: strings.TrimSpace(request.SubDistrict), PostalCode: request.PostalCode}
}

func toResponse(address Address) Response {
	return Response{ID: address.ID, FirstName: address.FirstName, LastName: address.LastName, AddressLine: address.AddressLine,
		Province: address.Province, District: address.District, SubDistrict: address.SubDistrict, PostalCode: address.PostalCode,
		Mobile: address.Mobile, Description: address.Description, IsDefault: address.IsDefault}
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return e.Code + ": " + e.Detail }

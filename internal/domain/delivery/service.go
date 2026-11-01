package delivery

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type Service struct{ repository repositoryPort }

type repositoryPort interface {
	Transact(context.Context, func(repositoryPort) error) error
	LockUser(context.Context, string) (string, error)
	LockAddresses(context.Context, string) error
	FindSnapshot(context.Context, string, string) (Address, error)
	UpdateSnapshot(context.Context, Address) error
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

// write returns data only after commit succeeds. Every callback read uses the same tx.
func (s *Service) write(ctx context.Context, action func(repositoryPort) (Address, error)) (Response, error) {
	var address Address
	err := s.repository.Transact(ctx, func(repo repositoryPort) error {
		var err error
		address, err = action(repo)
		return err
	})
	if err != nil {
		return Response{}, err
	}
	return toResponse(address), nil
}

func lockAddresses(ctx context.Context, repo repositoryPort, userID string) (string, error) {
	mobile, err := repo.LockUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if err = repo.LockAddresses(ctx, userID); err != nil {
		return "", err
	}
	return mobile, nil
}

func (s *Service) Create(ctx context.Context, userID string, request Request) (Response, error) {
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	return s.write(ctx, func(repo repositoryPort) (Address, error) {
		mobile, err := lockAddresses(ctx, repo, userID)
		if err != nil {
			return Address{}, err
		}
		count, err := repo.Count(ctx, userID)
		if err != nil {
			return Address{}, err
		}
		address := toAddress(uuid.NewString(), userID, request)
		address.IsDefault = count == 0 || *request.IsDefault
		address.Mobile = mobile
		if address.IsDefault {
			if err = repo.ClearDefault(ctx, userID); err != nil {
				return Address{}, err
			}
		}
		return address, repo.Create(ctx, address)
	})
}

func (s *Service) Update(ctx context.Context, userID, addressID string, request Request) (Response, error) {
	if _, err := uuid.Parse(addressID); err != nil {
		return Response{}, &BusinessError{Code: "MA00008", Detail: "addressId"}
	}
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	return s.write(ctx, func(repo repositoryPort) (Address, error) {
		mobile, err := lockAddresses(ctx, repo, userID)
		if err != nil {
			return Address{}, err
		}
		current, err := repo.Find(ctx, userID, addressID)
		address := toAddress(addressID, userID, request)
		if errors.Is(err, ErrNotFound) {
			snapshot, err := repo.FindSnapshot(ctx, userID, addressID)
			if err != nil {
				return Address{}, err
			}
			address.Mobile = snapshot.Mobile
			address.IsDefault = *request.IsDefault
			return address, repo.UpdateSnapshot(ctx, address)
		}
		if err != nil {
			return Address{}, err
		}
		if current.IsDefault && !*request.IsDefault {
			return Address{}, &BusinessError{Code: "MA00005", Detail: "isDefault"}
		}
		address.IsDefault, address.Mobile = *request.IsDefault, mobile
		if address.IsDefault {
			if err = repo.ClearDefault(ctx, userID); err != nil {
				return Address{}, err
			}
		}
		return address, repo.Update(ctx, address)
	})
}

func (s *Service) CreateForRenewal(ctx context.Context, userID, requestNo string, request Request) (Response, error) {
	if err := s.validate(ctx, request); err != nil {
		return Response{}, err
	}
	requestNo = strings.TrimSpace(requestNo)
	if requestNo == "" || len(requestNo) > 36 {
		return Response{}, &BusinessError{Code: "MA00005", Detail: "requestNo"}
	}
	return s.write(ctx, func(repo repositoryPort) (Address, error) {
		// Lock the request before the user, as in the legacy renewal flow.
		requestID, err := repo.RenewalRequest(ctx, userID, requestNo)
		if err != nil {
			return Address{}, err
		}
		exists, err := repo.SnapshotExists(ctx, requestID, userID)
		if err != nil {
			return Address{}, err
		}
		if exists {
			return Address{}, &BusinessError{Code: "MA00017", Detail: "deliveryAddress"}
		}
		mobile, err := lockAddresses(ctx, repo, userID)
		if err != nil {
			return Address{}, err
		}
		count, err := repo.Count(ctx, userID)
		if err != nil {
			return Address{}, err
		}
		address := toAddress("", userID, request)
		address.Mobile, address.IsDefault = mobile, count == 0
		if count == 0 {
			address.ID = uuid.NewString()
			if err = repo.ClearDefault(ctx, userID); err != nil {
				return Address{}, err
			}
			if err = repo.Create(ctx, address); err != nil {
				return Address{}, err
			}
		}
		snapshotID, err := repo.InsertSnapshot(ctx, requestID, userID, mobile, address)
		if err != nil {
			return Address{}, err
		}
		address.ID = snapshotID
		return address, nil
	})
}

func (s *Service) validate(ctx context.Context, request Request) error {
	if strings.TrimSpace(request.FirstName) == "" || strings.TrimSpace(request.LastName) == "" || strings.TrimSpace(request.AddressLine) == "" ||
		strings.TrimSpace(request.Province) == "" || strings.TrimSpace(request.District) == "" || strings.TrimSpace(request.SubDistrict) == "" ||
		!validPostalCode(request.PostalCode) || request.IsDefault == nil ||
		utf8.RuneCountInString(request.FirstName) > 100 || utf8.RuneCountInString(request.LastName) > 100 ||
		utf8.RuneCountInString(request.AddressLine) > 500 || utf8.RuneCountInString(request.Province) > 100 ||
		utf8.RuneCountInString(request.District) > 100 || utf8.RuneCountInString(request.SubDistrict) > 100 {
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

func validPostalCode(value string) bool {
	if len(value) != 5 {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
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

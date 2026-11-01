package profile

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/smartseaman/mobile-api/internal/filecontent"
)

type Storage interface {
	Put(context.Context, string, string, []byte) error
	Get(context.Context, string) ([]byte, string, error)
}

type Service struct {
	repository  repositoryPort
	storage     Storage
	profilePath string
}

type repositoryPort interface {
	FindUser(context.Context, string) (User, error)
	Descriptions(context.Context, string, string, string) (string, string)
	Update(context.Context, User, UpdateRequest, string) error
	UpdateStatus(context.Context, string, string) (User, error)
}

func NewService(repository repositoryPort, storage Storage, profilePath string) *Service {
	return &Service{repository: repository, storage: storage, profilePath: strings.Trim(profilePath, "/")}
}

func (s *Service) Get(ctx context.Context, userID, language string) (Response, error) {
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return Response{}, err
	}
	company, position := s.repository.Descriptions(ctx, user.CompanyCode, user.PositionCode, language)
	age := ""
	if user.DateOfBirth != "" {
		if birth, parseErr := time.Parse("2006-01-02", user.DateOfBirth); parseErr == nil {
			today := time.Now()
			years := today.Year() - birth.Year()
			if today.YearDay() < birth.YearDay() {
				years--
			}
			age = fmt.Sprint(years)
		}
	}
	return Response{
		FirstName: user.FirstName, LastName: user.LastName, DateOfBirth: user.DateOfBirth, Age: age,
		Mobile: user.Mobile, Email: user.Email, CompanyCode: user.CompanyCode, CompanyDescription: company,
		PositionCode: user.PositionCode, PositionDescription: position, SmartSeamanID: user.SmartID,
		ShortName: firstRuneUpper(user.FirstName) + firstRuneUpper(user.LastName),
	}, nil
}

func (s *Service) Update(ctx context.Context, userID string, request UpdateRequest) (RegisterResponse, error) {
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return RegisterResponse{}, err
	}
	if request.FirstName == "" || request.LastName == "" || request.PositionCode == "" || request.Email == "" || request.MobileNumber == "" {
		return RegisterResponse{}, &BusinessError{Code: "MA00004", Detail: "required field"}
	}
	var imageName string
	if request.IsChangeFile == "Y" {
		content, contentType, decodeErr := decodeImage(request.ImageProfile)
		if decodeErr != nil {
			return RegisterResponse{}, decodeErr
		}
		imageName = uuid.NewString()
		if s.storage == nil {
			return RegisterResponse{}, fmt.Errorf("object storage is not configured")
		}
		if err := s.storage.Put(ctx, s.profilePath+"/"+imageName, contentType, content); err != nil {
			return RegisterResponse{}, err
		}
	}
	if err := s.repository.Update(ctx, user, request, imageName); err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{Username: request.Email, Email: request.Email, MobileNumber: request.MobileNumber}, nil
}

func (s *Service) Image(ctx context.Context, userID string) ([]byte, string, error) {
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	if user.ProfilePicture == "" {
		return nil, "", &BusinessError{Code: "MA00016", Detail: "profile image"}
	}
	content, _, err := s.storage.Get(ctx, s.profilePath+"/"+user.ProfilePicture)
	if err != nil {
		return nil, "", err
	}
	return filecontent.DecodeStored(content)
}

func (s *Service) Deactivate(ctx context.Context, userID string) (RegisterResponse, error) {
	user, err := s.repository.FindUser(ctx, userID)
	if err != nil {
		return RegisterResponse{}, err
	}
	if _, err := s.repository.UpdateStatus(ctx, user.Email, "D"); err != nil {
		return RegisterResponse{}, err
	}
	return RegisterResponse{Username: user.Username, Email: user.Email}, nil
}

func (s *Service) Activate(ctx context.Context, email string) (RegisterResponse, error) {
	return RegisterResponse{}, &BusinessError{Code: "MA00007", Detail: "activation token required"}
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return e.Code + ": " + e.Detail }

func decodeImage(value string) ([]byte, string, error) {
	contentType := ""
	encoded := strings.TrimSpace(value)
	if strings.HasPrefix(encoded, "data:") {
		comma := strings.Index(encoded, ",")
		marker := strings.Index(encoded, ";base64")
		if comma < 0 || marker < 0 || marker > comma {
			return nil, "", &BusinessError{Code: "MA00005", Detail: "imageProfile"}
		}
		contentType = encoded[5:marker]
		encoded = encoded[comma+1:]
	}
	content, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(content) == 0 || len(content) > 10*1024*1024 {
		return nil, "", &BusinessError{Code: "MA00005", Detail: "imageProfile"}
	}
	detected := ""
	if len(content) >= 8 && string(content[:8]) == "\x89PNG\r\n\x1a\n" {
		detected = "image/png"
	}
	if len(content) >= 3 && content[0] == 0xff && content[1] == 0xd8 && content[2] == 0xff {
		detected = "image/jpeg"
	}
	if detected == "" || (contentType != "" && contentType != detected) {
		return nil, "", &BusinessError{Code: "MA00005", Detail: "imageProfile"}
	}
	return content, detected, nil
}

func firstRuneUpper(value string) string {
	for _, character := range value {
		return strings.ToUpper(string(character))
	}
	return ""
}

package content

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

type Service struct {
	repository repositoryPort
	storage    Storage
	formPath   string
	newsPath   string
}

type repositoryPort interface {
	Forms(context.Context) ([]Form, error)
	Form(context.Context, string) (Form, error)
	News(context.Context, string) ([]News, error)
	NewsByID(context.Context, string) (News, string, error)
}

func NewService(repository repositoryPort, storage Storage, formPath, newsPath string) *Service {
	return &Service{repository: repository, storage: storage, formPath: strings.Trim(formPath, "/"), newsPath: strings.Trim(newsPath, "/")}
}

func (s *Service) Forms(ctx context.Context) (FormsResponse, error) {
	forms, err := s.repository.Forms(ctx)
	return FormsResponse{Forms: forms}, err
}

func (s *Service) FormFile(ctx context.Context, id string) ([]byte, error) {
	form, err := s.repository.Form(ctx, id)
	if err != nil {
		return nil, err
	}
	content, _, err := s.storage.Get(ctx, s.formPath+"/"+form.FormFileName)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func (s *Service) News(ctx context.Context) (NewsResponse, error) {
	ship, err := s.repository.News(ctx, "SHIP")
	if err != nil {
		return NewsResponse{}, err
	}
	general, err := s.repository.News(ctx, "GENERAL")
	return NewsResponse{NewsShip: ship, NewsGeneral: general}, err
}

func (s *Service) NewsDetail(ctx context.Context, id string) (News, error) {
	item, _, err := s.repository.NewsByID(ctx, id)
	return item, err
}

func (s *Service) NewsImage(ctx context.Context, id string) ([]byte, error) {
	_, file, err := s.repository.NewsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file == "" {
		return nil, &BusinessError{Code: "MA00016", Detail: "newsId"}
	}
	content, _, err := s.storage.Get(ctx, s.newsPath+"/"+file)
	if err != nil {
		return nil, err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

type BusinessError struct{ Code, Detail string }

func (e *BusinessError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }
func isNotFound(err error) bool        { return errors.Is(err, ErrNotFound) }

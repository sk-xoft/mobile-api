package master

import (
	"context"
	"strings"
)

type Service struct{ repository repositoryPort }

type repositoryPort interface {
	Companies(context.Context) ([]Company, error)
	Positions(context.Context) ([]Position, error)
	Documents(context.Context, string) ([]Document, error)
	Courses(context.Context) ([]Course, error)
	Provinces(context.Context) ([]Address, error)
	Districts(context.Context, int) ([]Address, error)
	Subdistricts(context.Context, int) ([]Address, error)
}

func NewService(repository repositoryPort) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, language string) (MasterResponse, error) {
	companies, err := s.repository.Companies(ctx)
	if err != nil {
		return MasterResponse{}, err
	}
	positions, err := s.repository.Positions(ctx)
	if err != nil {
		return MasterResponse{}, err
	}
	english := strings.EqualFold(language, "EN")
	for i := range companies {
		companies[i].Name = localized(companies[i].Name, english)
	}
	for i := range positions {
		positions[i].Name = localized(positions[i].Name, english)
	}
	return MasterResponse{Company: companies, Position: positions}, nil
}

func (s *Service) Documents(ctx context.Context, language string) (DocumentsResponse, error) {
	documents, err := s.repository.Documents(ctx, "Document")
	if err != nil {
		return DocumentsResponse{}, err
	}
	cot, err := s.repository.Documents(ctx, "COT")
	if err != nil {
		return DocumentsResponse{}, err
	}
	english := strings.EqualFold(language, "EN")
	for i := range documents {
		documents[i].Name = localized(documents[i].Name, english)
	}
	for i := range cot {
		cot[i].Name = localized(cot[i].Name, english)
	}
	return DocumentsResponse{Documents: documents, COT: cot}, nil
}

func (s *Service) Courses(ctx context.Context) (CoursesResponse, error) {
	items, err := s.repository.Courses(ctx)
	return CoursesResponse{Courses: items}, err
}

func (s *Service) Provinces(ctx context.Context, language string) ([]Address, error) {
	items, err := s.repository.Provinces(ctx)
	return localizeAddresses(items, language, err)
}

func (s *Service) Districts(ctx context.Context, language string, code int) ([]Address, error) {
	items, err := s.repository.Districts(ctx, code)
	return localizeAddresses(items, language, err)
}

func (s *Service) Subdistricts(ctx context.Context, language string, code int) ([]Address, error) {
	items, err := s.repository.Subdistricts(ctx, code)
	return localizeAddresses(items, language, err)
}

func localizeAddresses(items []Address, language string, err error) ([]Address, error) {
	if err != nil {
		return nil, err
	}
	english := strings.EqualFold(language, "EN")
	for i := range items {
		if english {
			items[i].Name = items[i].NameEN
		} else {
			items[i].Name = items[i].NameTH
		}
	}
	return items, nil
}

func localized(value string, english bool) string {
	parts := strings.SplitN(value, "\x00", 2)
	if english && len(parts) == 2 {
		return parts[1]
	}
	return parts[0]
}

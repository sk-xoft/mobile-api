package notification

import (
	"context"
	"fmt"
)

type Service struct{ repository repositoryPort }

type repositoryPort interface {
	List(context.Context, string) ([]Item, error)
	Count(context.Context, string) (int, error)
	MarkID(context.Context, string, string) (bool, error)
	MarkAll(context.Context, string) error
	MarkValue(context.Context, string, string, string) error
}

func NewService(r repositoryPort) *Service { return &Service{repository: r} }
func (s *Service) List(ctx context.Context, userID string) (Response, error) {
	v, e := s.repository.List(ctx, userID)
	return Response{Notifications: v}, e
}
func (s *Service) MarkID(ctx context.Context, userID, id string) (CountResponse, error) {
	if id != "0" {
		ok, e := s.repository.MarkID(ctx, userID, id)
		if e != nil {
			return CountResponse{}, e
		}
		if !ok {
			return CountResponse{}, ErrNotFound
		}
	}
	return s.count(ctx, userID)
}
func (s *Service) MarkAll(ctx context.Context, userID string) (CountResponse, error) {
	if e := s.repository.MarkAll(ctx, userID); e != nil {
		return CountResponse{}, e
	}
	return s.count(ctx, userID)
}
func (s *Service) MarkValue(ctx context.Context, userID, value, kind string) (CountResponse, error) {
	if value == "" || kind == "" {
		return CountResponse{}, fmt.Errorf("valueId and notiType are required")
	}
	if e := s.repository.MarkValue(ctx, userID, value, kind); e != nil {
		return CountResponse{}, e
	}
	return s.count(ctx, userID)
}
func (s *Service) count(ctx context.Context, userID string) (CountResponse, error) {
	n, e := s.repository.Count(ctx, userID)
	return CountResponse{Count: fmt.Sprint(n)}, e
}

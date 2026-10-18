package fcm

import "context"

type repositoryPort interface {
	Exists(context.Context, string) (bool, error)
	Insert(context.Context, string, string) error
	Update(context.Context, string, string) error
}

type Service struct{ repository repositoryPort }

func NewService(repository repositoryPort) *Service { return &Service{repository: repository} }

func (s *Service) Update(ctx context.Context, userID, token string) error {
	exists, err := s.repository.Exists(ctx, userID)
	if err != nil {
		return err
	}
	if exists {
		return s.repository.Update(ctx, userID, token)
	}
	return s.repository.Insert(ctx, userID, token)
}

package eventlog

import "context"

type Event struct {
	Name          string
	ActorUserID   string
	AggregateType string
	AggregateID   string
	RequestID     string
	TraceID       string
	Metadata      map[string]any
}

type Repository interface {
	Create(context.Context, Event) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	if s == nil || s.repository == nil || event.Name == "" {
		return nil
	}
	return s.repository.Create(ctx, event)
}

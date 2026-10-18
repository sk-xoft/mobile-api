package profilehttp

import domain "github.com/smartseaman/mobile-api/internal/domain/profile"

type Handler = domain.Handler

func NewHandler(service *domain.Service) *Handler { return domain.NewHandler(service) }

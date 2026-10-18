package masterhttp

import domain "github.com/smartseaman/mobile-api/internal/domain/master"

type Handler = domain.Handler

func NewHandler(service *domain.Service) *Handler { return domain.NewHandler(service) }

package renewalhttp

import domain "github.com/smartseaman/mobile-api/internal/domain/renewal"

type Handler = domain.Handler

func NewHandler(service *domain.Service) *Handler { return domain.NewHandler(service) }

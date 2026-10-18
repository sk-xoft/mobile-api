package deliveryhttp

import domain "github.com/smartseaman/mobile-api/internal/domain/delivery"

type Handler = domain.Handler

func NewHandler(service *domain.Service) *Handler { return domain.NewHandler(service) }

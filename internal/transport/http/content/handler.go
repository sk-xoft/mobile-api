package contenthttp

import domain "github.com/smartseaman/mobile-api/internal/domain/content"

type Handler = domain.Handler

func NewHandler(service *domain.Service) *Handler { return domain.NewHandler(service) }

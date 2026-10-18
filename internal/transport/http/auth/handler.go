package authhttp

import (
	"context"
	domain "github.com/smartseaman/mobile-api/internal/domain/auth"
)

type Handler = domain.Handler

func NewHandler(service *domain.Service, messages interface {
	Message(context.Context, string, string) string
}) *Handler {
	return domain.NewHandler(service, messages)
}

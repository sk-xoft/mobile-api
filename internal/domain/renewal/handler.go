package renewal

import (
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
func (h *Handler) Statuses(c *fiber.Ctx) error {
	data, err := h.service.Statuses(c.UserContext())
	return h.result(c, data, err)
}
func (h *Handler) Documents(c *fiber.Ctx) error {
	data, err := h.service.Documents(c.UserContext(), c.Get("Accept-Language"))
	return h.result(c, data, err)
}
func (h *Handler) Price(c *fiber.Ctx) error {
	data, err := h.service.Price(c.UserContext(), c.Query("documentCode"))
	return h.result(c, data, err)
}
func (h *Handler) result(c *fiber.Ctx, data any, err error) error {
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}

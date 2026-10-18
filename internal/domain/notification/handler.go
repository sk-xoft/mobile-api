package notification

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) List(c *fiber.Ctx) error {
	v, e := h.service.List(c.UserContext(), uid(c))
	if e != nil {
		return h.fail(c, e)
	}
	return httpx.Success(c, v)
}
func (h *Handler) MarkID(c *fiber.Ctx) error {
	var q MarkRequest
	if e := c.BodyParser(&q); e != nil {
		return httpx.Error(c, 400, "MA00005", "notiId")
	}
	v, e := h.service.MarkID(c.UserContext(), uid(c), q.NotiID)
	return h.result(c, v, e)
}
func (h *Handler) MarkValue(c *fiber.Ctx) error {
	var q MarkValueRequest
	if e := c.BodyParser(&q); e != nil {
		return httpx.Error(c, 400, "MA00005", "notification")
	}
	v, e := h.service.MarkValue(c.UserContext(), uid(c), q.ValueID, q.NotiType)
	return h.result(c, v, e)
}
func (h *Handler) MarkAll(c *fiber.Ctx) error {
	v, e := h.service.MarkAll(c.UserContext(), uid(c))
	return h.result(c, v, e)
}
func (h *Handler) Manual(c *fiber.Ctx) error { return httpx.Success(c, "success") }
func (h *Handler) result(c *fiber.Ctx, v any, e error) error {
	if e != nil {
		return h.fail(c, e)
	}
	return httpx.Success(c, v)
}
func (h *Handler) fail(c *fiber.Ctx, e error) error {
	if errors.Is(e, ErrNotFound) {
		return httpx.Error(c, 400, "MA00016", "notification")
	}
	return httpx.Error(c, 500, "MA00013", "Technical exception")
}
func uid(c *fiber.Ctx) string { v, _ := c.Locals("user.mobile_uuid").(string); return v }

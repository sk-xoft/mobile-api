package content

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
func (h *Handler) Forms(c *fiber.Ctx) error {
	data, err := h.service.Forms(c.UserContext())
	return h.result(c, data, err)
}
func (h *Handler) FormFile(c *fiber.Ctx) error {
	data, err := h.service.FormFile(c.UserContext(), c.Query("formId"))
	if err != nil {
		return h.fail(c, err)
	}
	c.Set("Content-Type", "application/pdf")
	return c.Send(data)
}
func (h *Handler) News(c *fiber.Ctx) error {
	data, err := h.service.News(c.UserContext())
	return h.result(c, data, err)
}
func (h *Handler) NewsDetail(c *fiber.Ctx) error {
	data, err := h.service.NewsDetail(c.UserContext(), c.Query("newsId"))
	return h.result(c, data, err)
}
func (h *Handler) NewsImage(c *fiber.Ctx) error {
	data, err := h.service.NewsImage(c.UserContext(), c.Query("newsId"))
	if err != nil {
		return h.fail(c, err)
	}
	c.Set("Content-Type", "image/png")
	return c.Send(data)
}
func (h *Handler) result(c *fiber.Ctx, data any, err error) error {
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) fail(c *fiber.Ctx, err error) error {
	if isNotFound(err) {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00016", "data")
	}
	var business *BusinessError
	if errors.As(err, &business) {
		return httpx.Error(c, fiber.StatusBadRequest, business.Code, business.Detail)
	}
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
}

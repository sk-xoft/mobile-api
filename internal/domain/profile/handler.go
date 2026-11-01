package profile

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	response, err := h.service.Get(c.UserContext(), userID, c.Get(fiber.HeaderAcceptLanguage))
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, response)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var request UpdateRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	userID, _ := c.Locals("user.mobile_uuid").(string)
	response, err := h.service.Update(c.UserContext(), userID, request)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, response)
}

func (h *Handler) Image(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	content, contentType, err := h.service.Image(c.UserContext(), userID)
	if err != nil {
		return h.fail(c, err)
	}
	c.Set(fiber.HeaderContentType, contentType)
	return c.Send(content)
}

func (h *Handler) Deactivate(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	response, err := h.service.Deactivate(c.UserContext(), userID)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, response)
}

func (h *Handler) Activate(c *fiber.Ctx) error {
	var request ActivateRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	response, err := h.service.Activate(c.UserContext(), request.Email)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, response)
}

func (h *Handler) fail(c *fiber.Ctx, err error) error {
	var business *BusinessError
	if errors.As(err, &business) {
		return httpx.Error(c, fiber.StatusBadRequest, business.Code, business.Detail)
	}
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
}

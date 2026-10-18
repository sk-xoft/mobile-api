package delivery

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Default(c *fiber.Ctx) error {
	data, err := h.service.Default(c.UserContext(), userID(c))
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var request Request
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "address"})
	}
	data, err := h.service.Create(c.UserContext(), userID(c), request)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var request Request
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "address"})
	}
	data, err := h.service.Update(c.UserContext(), userID(c), c.Params("addressId"), request)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) CreateForRenewal(c *fiber.Ctx) error {
	var request Request
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "address"})
	}
	data, err := h.service.CreateForRenewal(c.UserContext(), userID(c), c.Params("requestNo"), request)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) fail(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00016", "deliveryAddress")
	}
	var business *BusinessError
	if errors.As(err, &business) {
		return httpx.Error(c, fiber.StatusBadRequest, business.Code, business.Detail)
	}
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
}

func userID(c *fiber.Ctx) string { value, _ := c.Locals("user.mobile_uuid").(string); return value }

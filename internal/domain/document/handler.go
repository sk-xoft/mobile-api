package document

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(typeName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		offset, err := strconv.Atoi(c.Query("offSet"))
		if err != nil {
			return httpx.Error(c, fiber.StatusBadRequest, "MA00005", "offSet")
		}
		userID, _ := c.Locals("user.mobile_uuid").(string)
		data, err := h.service.Page(c.UserContext(), userID, typeName, offset)
		if err != nil {
			return h.fail(c, err)
		}
		return httpx.Success(c, data)
	}
}
func (h *Handler) CloseToExpiration(c *fiber.Ctx) error {
	offset, err := strconv.Atoi(c.Query("offSet"))
	if err != nil {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00005", "offSet")
	}
	userID, _ := c.Locals("user.mobile_uuid").(string)
	data, err := h.service.CloseToExpiration(c.UserContext(), userID, offset)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	userID, _ := c.Locals("user.mobile_uuid").(string)
	data, err := h.service.Create(c.UserContext(), userID, req)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) Update(c *fiber.Ctx) error {
	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	userID, _ := c.Locals("user.mobile_uuid").(string)
	data, err := h.service.Update(c.UserContext(), userID, req)
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	data, err := h.service.Delete(c.UserContext(), userID, c.Query("certCode"))
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) Edit(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	data, err := h.service.Edit(c.UserContext(), userID, c.Query("certCode"))
	if err != nil {
		return h.fail(c, err)
	}
	return httpx.Success(c, data)
}
func (h *Handler) View(c *fiber.Ctx) error {
	userID, _ := c.Locals("user.mobile_uuid").(string)
	content, contentType, err := h.service.View(c.UserContext(), userID, c.Query("certCode"))
	if err != nil {
		return h.fail(c, err)
	}
	c.Type(contentType)
	return c.Send(content)
}
func (h *Handler) fail(c *fiber.Ctx, err error) error {
	var business *BusinessError
	if errors.As(err, &business) {
		return httpx.Error(c, fiber.StatusBadRequest, business.Code, business.Detail)
	}
	if errors.Is(err, ErrNotFound) {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00016", "data")
	}
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
}

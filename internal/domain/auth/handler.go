package auth

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/middleware"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct {
	service  *Service
	messages interface {
		Message(context.Context, string, string) string
	}
}

func NewHandler(service *Service, messages interface {
	Message(context.Context, string, string) string
}) *Handler {
	return &Handler{service: service, messages: messages}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var request RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	response, err := h.service.Register(c.UserContext(), request)
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) RequestPasswordReset(c *fiber.Ctx) error {
	var request ResetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	response, err := h.service.RequestPasswordReset(c.UserContext(), request)
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) CompletePasswordReset(c *fiber.Ctx) error {
	var request ForgotPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	response, err := h.service.CompletePasswordReset(c.UserContext(), request)
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var request LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	response, err := h.service.Login(c.UserContext(), request, c.Get("correlationid"))
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	response, err := h.service.Refresh(c.UserContext(), c.Query("refToken"))
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var request ChangePasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request body"})
	}
	username, _ := c.Locals(middleware.JWTSubject).(string)
	response, err := h.service.ChangePassword(c.UserContext(), username, request)
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) Activate(c *fiber.Ctx) error {
	response, err := h.service.Activate(c.UserContext(), c.Query("uid"))
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

func (h *Handler) success(c *fiber.Ctx, data any) error {
	description := h.messages.Message(c.UserContext(), httpx.SuccessCode, c.Get("Accept-Language"))
	return c.JSON(httpx.Response{Code: httpx.SuccessCode, Description: description, Data: data})
}

func (h *Handler) fail(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrNotFound) {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00007", "Invalid or expired credential")
	}
	var business *BusinessError
	if errors.As(err, &business) {
		return httpx.Error(c, fiber.StatusBadRequest, business.Code, h.messages.Message(c.UserContext(), business.Code, c.Get("Accept-Language")))
	}
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", h.messages.Message(c.UserContext(), "MA00013", c.Get("Accept-Language")))
}

func (h *Handler) ActivateBody(c *fiber.Ctx) error {
	var request struct {
		UID string `json:"uid"`
	}
	if err := c.BodyParser(&request); err != nil || request.UID == "" {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "uid"})
	}
	response, err := h.service.Activate(c.UserContext(), request.UID)
	if err != nil {
		return h.fail(c, err)
	}
	return h.success(c, response)
}

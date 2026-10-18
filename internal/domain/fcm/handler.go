package fcm

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Update(c *fiber.Ctx) error {
	var request UpdateRequest
	if err := c.BodyParser(&request); err != nil || strings.TrimSpace(request.Token) == "" {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00005", "tokenFcm")
	}
	userID, _ := c.Locals("user.mobile_uuid").(string)
	err := h.service.Update(c.UserContext(), userID, request.Token)
	if err != nil {
		return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
	}
	username, _ := c.Locals("user.username").(string)
	return httpx.Success(c, username+" update fcm token is success.")
}

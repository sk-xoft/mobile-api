package master

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *fiber.Ctx) error {
	data, err := h.service.List(c.UserContext(), c.Get(fiber.HeaderAcceptLanguage))
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Documents(c *fiber.Ctx) error {
	data, err := h.service.Documents(c.UserContext(), c.Get(fiber.HeaderAcceptLanguage))
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Courses(c *fiber.Ctx) error {
	data, err := h.service.Courses(c.UserContext())
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Provinces(c *fiber.Ctx) error {
	data, err := h.service.Provinces(c.UserContext(), c.Get(fiber.HeaderAcceptLanguage))
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Districts(c *fiber.Ctx) error {
	code, err := strconv.Atoi(c.Query("provinceCode"))
	if err != nil || code <= 0 {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00005", "provinceCode")
	}
	data, err := h.service.Districts(c.UserContext(), c.Get(fiber.HeaderAcceptLanguage), code)
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func (h *Handler) Subdistricts(c *fiber.Ctx) error {
	code, err := strconv.Atoi(c.Query("districtCode"))
	if err != nil || code <= 0 {
		return httpx.Error(c, fiber.StatusBadRequest, "MA00005", "districtCode")
	}
	data, err := h.service.Subdistricts(c.UserContext(), c.Get(fiber.HeaderAcceptLanguage), code)
	if err != nil {
		return failure(c, err)
	}
	return httpx.Success(c, data)
}

func failure(c *fiber.Ctx, err error) error {
	return httpx.Error(c, fiber.StatusInternalServerError, "MA00013", "Technical exception")
}

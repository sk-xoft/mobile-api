package catalog

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
func (h *Handler) Banners(c *fiber.Ctx) error {
	v, e := h.service.Banners(c.UserContext())
	return h.result(c, v, e)
}
func (h *Handler) BannerImage(c *fiber.Ctx) error {
	v, e := h.service.BannerImage(c.UserContext(), c.Query("bannerId"))
	if e != nil {
		return h.fail(c, e)
	}
	c.Set("Content-Type", detect(v))
	return c.Send(v)
}
func (h *Handler) Vouchers(c *fiber.Ctx) error {
	u, _ := c.Locals("user.mobile_uuid").(string)
	v, e := h.service.Vouchers(c.UserContext(), u)
	return h.result(c, v, e)
}
func (h *Handler) VoucherDetail(c *fiber.Ctx) error {
	v, _, _, e := h.service.Voucher(c.UserContext(), c.Query("voucherId"))
	return h.result(c, v, e)
}
func (h *Handler) VoucherImage(c *fiber.Ctx) error {
	v, e := h.service.VoucherImage(c.UserContext(), c.Query("voucherId"))
	if e != nil {
		return h.fail(c, e)
	}
	c.Set("Content-Type", detect(v))
	return c.Send(v)
}
func (h *Handler) VoucherQR(c *fiber.Ctx) error {
	v, e := h.service.VoucherQR(c.UserContext(), c.Query("voucherId"))
	if e != nil {
		return h.fail(c, e)
	}
	c.Set("Content-Type", detect(v))
	return c.Send(v)
}
func (h *Handler) Schools(c *fiber.Ctx) error {
	v, e := h.service.Schools(c.UserContext(), c.Query("courseCode"))
	return h.result(c, v, e)
}
func (h *Handler) SchoolDetail(c *fiber.Ctx) error {
	v, e := h.service.SchoolDetail(c.UserContext(), c.Query("companyCode"), c.Query("courseCode"))
	return h.result(c, v, e)
}
func (h *Handler) result(c *fiber.Ctx, v any, e error) error {
	if e != nil {
		return h.fail(c, e)
	}
	return httpx.Success(c, v)
}
func (h *Handler) fail(c *fiber.Ctx, e error) error {
	if errors.Is(e, ErrNotFound) {
		return httpx.Error(c, 400, "MA00016", "data")
	}
	return httpx.Error(c, 500, "MA00013", "Technical exception")
}
func detect(v []byte) string {
	if len(v) >= 8 && string(v[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(v) >= 3 && v[0] == 0xff && v[1] == 0xd8 && v[2] == 0xff {
		return "image/jpeg"
	}
	if len(v) >= 5 && string(v[:5]) == "%PDF-" {
		return "application/pdf"
	}
	return "application/octet-stream"
}

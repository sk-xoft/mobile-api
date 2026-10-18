package renewal

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

func (h *Handler) CreateRequest(c *fiber.Ctx) error {
	var q CreateRequest
	if e := c.BodyParser(&q); e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "request"})
	}
	v, e := h.service.CreateRequest(c.UserContext(), uid(c), q)
	return h.result(c, v, e)
}
func (h *Handler) ListRequests(c *fiber.Ctx) error {
	o, e := strconv.Atoi(c.Query("offSet"))
	if e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "offSet"})
	}
	v, e := h.service.ListRequests(c.UserContext(), uid(c), o)
	return h.result(c, v, e)
}
func (h *Handler) Resubmit(c *fiber.Ctx) error {
	v, e := h.service.Resubmit(c.UserContext(), uid(c), c.Params("requestNo"))
	return h.result(c, v, e)
}
func (h *Handler) DeleteRequest(c *fiber.Ctx) error {
	v, e := h.service.DeleteRequest(c.UserContext(), uid(c), c.Params("requestNo"))
	return h.result(c, v, e)
}
func (h *Handler) UpdateMobile(c *fiber.Ctx) error {
	var q MobileRequest
	if e := c.BodyParser(&q); e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "mobileNumber"})
	}
	v, e := h.service.UpdateMobile(c.UserContext(), uid(c), c.Params("requestNo"), q.MobileNumber)
	return h.result(c, v, e)
}
func (h *Handler) Detail(c *fiber.Ctx) error {
	v, e := h.service.Detail(c.UserContext(), uid(c), c.Params("requestNo"))
	return h.result(c, v, e)
}
func (h *Handler) Timeline(c *fiber.Ctx) error {
	v, e := h.service.Timeline(c.UserContext(), uid(c), c.Params("requestNo"))
	return h.result(c, v, e)
}
func (h *Handler) Stage(c *fiber.Ctx) error {
	v, e := h.service.Stage(c.UserContext(), uid(c), c.Params("requestNo"))
	return h.result(c, v, e)
}
func (h *Handler) Payment(c *fiber.Ctx) error {
	var q PaymentRequest
	if e := c.BodyParser(&q); e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "payment"})
	}
	v, e := h.service.Payment(c.UserContext(), uid(c), c.Params("requestId"), q)
	return h.result(c, v, e)
}
func (h *Handler) PaymentStatus(c *fiber.Ctx) error {
	v, e := h.service.PaymentStatus(c.UserContext(), uid(c), c.Params("requestId"), c.Params("transactionId"))
	return h.result(c, v, e)
}
func (h *Handler) HardDelete(c *fiber.Ctx) error {
	v, e := h.service.HardDelete(c.UserContext(), uid(c), c.Params("requestNo"))
	if e != nil {
		log.Printf("renewal hard delete failed request_no=%s: %v", c.Params("requestNo"), e)
	}
	return h.result(c, v, e)
}
func (h *Handler) Validate(c *fiber.Ctx) error {
	var q struct {
		DocumentCode   string `json:"documentCode"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if e := c.BodyParser(&q); e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "documentCode"})
	}
	v, e := h.service.Validate(c.UserContext(), uid(c), q.DocumentCode, q.IdempotencyKey)
	return h.result(c, v, e)
}
func (h *Handler) Webhook(c *fiber.Ctx) error {
	if !json.Valid(c.Body()) {
		return c.SendStatus(400)
	}
	if e := h.service.Webhook(c.UserContext(), string(c.Body()), c.Get("Omise-Signature"), c.Get("Omise-Signature-Timestamp")); e != nil {
		var businessError *BusinessError
		if errors.As(e, &businessError) {
			return c.SendStatus(401)
		}
		return c.SendStatus(500)
	}
	return c.SendStatus(200)
}
func (h *Handler) fail(c *fiber.Ctx, e error) error {
	if errors.Is(e, ErrNotFound) {
		return httpx.Error(c, 400, "MA00016", "data")
	}
	var b *BusinessError
	if errors.As(e, &b) {
		return httpx.Error(c, 400, b.Code, b.Detail)
	}
	return httpx.Error(c, 500, "MA00013", "Technical exception")
}
func uid(c *fiber.Ctx) string { v, _ := c.Locals("user.mobile_uuid").(string); return v }

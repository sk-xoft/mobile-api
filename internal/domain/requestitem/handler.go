package requestitem

import (
	"errors"
	"io"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
func (h *Handler) List(c *fiber.Ctx) error {
	v, e := h.service.List(c.UserContext(), uid(c))
	return h.result(c, v, e)
}
func (h *Handler) Preview(c *fiber.Ctx) error {
	v, e := h.service.Preview(c.UserContext(), uid(c), c.Params("itemCode"), "")
	return h.result(c, v, e)
}
func (h *Handler) RenewalPreview(c *fiber.Ctx) error {
	item := c.Params("documentRequestItemCode")
	no := c.Params("requestNo")
	if no == "" {
		no = c.Query("request_no")
	}
	requestID, _, _, _, e := h.service.repo.Renewal(c.UserContext(), uid(c), no, item)
	if e != nil {
		return h.fail(c, e)
	}
	v, e := h.service.Preview(c.UserContext(), uid(c), item, requestID)
	return h.result(c, v, e)
}
func (h *Handler) Upload(c *fiber.Ctx) error        { return h.upload(c, "") }
func (h *Handler) RenewalUpload(c *fiber.Ctx) error { return h.upload(c, c.Params("requestNo")) }
func (h *Handler) upload(c *fiber.Ctx, no string) error {
	file, e := c.FormFile("file")
	if e != nil {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "file"})
	}
	if file.Size <= 0 || file.Size > 10*1024*1024 {
		return h.fail(c, &BusinessError{Code: "MA00005", Detail: "file"})
	}
	content, e := file.Open()
	if e != nil {
		return h.fail(c, e)
	}
	defer content.Close()
	bytes := make([]byte, file.Size)
	if _, e = io.ReadFull(content, bytes); e != nil {
		return h.fail(c, e)
	}
	itemCode := c.Params("itemCode")
	if itemCode == "" {
		itemCode = c.Params("documentRequestItemCode")
	}
	v, e := h.service.Upload(c.UserContext(), uid(c), itemCode, c.FormValue("documentType", "GENERAL"), c.FormValue("slotCode", "MAIN"), file.Filename, bytes, no)
	return h.result(c, v, e)
}
func (h *Handler) result(c *fiber.Ctx, v any, e error) error {
	if e != nil {
		return h.fail(c, e)
	}
	return httpx.Success(c, v)
}
func (h *Handler) fail(c *fiber.Ctx, e error) error {
	if e != nil {
		slog.Error("request item operation failed",
			"request_id", c.Get("X-Request-ID"),
			"trace_id", c.Locals("trace.id"),
			"method", c.Method(),
			"path", c.Path(),
			"error", e,
		)
	}
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
